package persistence

import (
	"context"
	"testing"
	"time"

	"network-debugger/internal/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := db.AutoMigrate(&ComposeLibraryModel{}, &ComposeHistoryEntryModel{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestLibraryRepo_LoadLibrary_EmptyDB(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewLibraryRepo(db)

	lib, err := repo.LoadLibrary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return empty library when record not found
	if lib.Collections == nil {
		t.Error("Collections should not be nil")
	}
	if len(lib.Collections) != 0 {
		t.Errorf("Collections should be empty, got %d", len(lib.Collections))
	}
	if lib.RequestsByID == nil {
		t.Error("RequestsByID should not be nil")
	}
	if len(lib.RequestsByID) != 0 {
		t.Errorf("RequestsByID should be empty, got %d", len(lib.RequestsByID))
	}
}

func TestLibraryRepo_SaveLoad(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewLibraryRepo(db)
	ctx := context.Background()

	// Create a library
	lib := domain.ComposeLibrary{
		Collections: []domain.ComposeCollection{
			{
				ID:   "col1",
				Name: "Test Collection",
				Root: domain.ComposeFolder{
					ID:       "root",
					Name:     "Root",
					Requests: []string{"req1"},
				},
			},
		},
		RequestsByID: map[string]domain.ComposeRequestTemplate{
			"req1": {
				ID:     "req1",
				Name:   "Test Request",
				Method: "GET",
				URL:    "http://example.com",
			},
		},
	}

	// Save
	err := repo.SaveLibrary(ctx, lib)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Load
	loaded, err := repo.LoadLibrary(ctx)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	// Verify
	if len(loaded.Collections) != 1 {
		t.Errorf("Collections count: got %d, want 1", len(loaded.Collections))
	}
	if loaded.Collections[0].ID != "col1" {
		t.Errorf("Collection ID: got %s, want col1", loaded.Collections[0].ID)
	}
	if len(loaded.RequestsByID) != 1 {
		t.Errorf("RequestsByID count: got %d, want 1", len(loaded.RequestsByID))
	}
	if req, ok := loaded.RequestsByID["req1"]; !ok {
		t.Error("req1 not found in RequestsByID")
	} else {
		if req.Method != "GET" {
			t.Errorf("Method: got %s, want GET", req.Method)
		}
		if req.URL != "http://example.com" {
			t.Errorf("URL: got %s, want http://example.com", req.URL)
		}
	}
}

func TestLibraryRepo_SaveLibrary_NilRequestsByID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewLibraryRepo(db)
	ctx := context.Background()

	// Save with nil RequestsByID
	lib := domain.ComposeLibrary{
		Collections:  []domain.ComposeCollection{},
		RequestsByID: nil,
	}

	err := repo.SaveLibrary(ctx, lib)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// Load and verify RequestsByID is not nil
	loaded, err := repo.LoadLibrary(ctx)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.RequestsByID == nil {
		t.Error("RequestsByID should be initialized to empty map, not nil")
	}
}

func TestLibraryRepo_LoadLibrary_InvalidJSON(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewLibraryRepo(db)
	ctx := context.Background()

	// Insert invalid JSON manually
	m := ComposeLibraryModel{
		ID:        1,
		JSON:      "invalid json{{{",
		UpdatedAt: time.Now().UTC(),
	}
	if err := db.Create(&m).Error; err != nil {
		t.Fatalf("failed to insert invalid JSON: %v", err)
	}

	// Load should return empty library instead of error
	lib, err := repo.LoadLibrary(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lib.Collections == nil || lib.RequestsByID == nil {
		t.Error("should return empty library on invalid JSON")
	}
}

func TestHistoryRepo_AppendHistory(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewHistoryRepo(db)
	ctx := context.Background()

	entry := domain.ComposeHistoryEntry{
		ID:       "entry1",
		Template: "tpl1",
		Method:   "POST",
		URL:      "http://api.example.com",
		Status:   200,
		When:     time.Now().UTC(),
	}

	err := repo.AppendHistory(ctx, entry)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// Verify it was saved
	var count int64
	db.Model(&ComposeHistoryEntryModel{}).Count(&count)
	if count != 1 {
		t.Errorf("count: got %d, want 1", count)
	}
}

func TestHistoryRepo_AppendHistory_GeneratesID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewHistoryRepo(db)
	ctx := context.Background()

	// Entry without ID
	entry := domain.ComposeHistoryEntry{
		Method: "GET",
		URL:    "http://example.com",
		Status: 404,
	}

	err := repo.AppendHistory(ctx, entry)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// Verify an ID was generated
	var saved ComposeHistoryEntryModel
	db.First(&saved)
	if saved.ID == "" {
		t.Error("ID should be generated")
	}
}

func TestHistoryRepo_AppendHistory_SetsWhen(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewHistoryRepo(db)
	ctx := context.Background()

	entry := domain.ComposeHistoryEntry{
		ID:     "entry1",
		Method: "DELETE",
		URL:    "http://example.com",
		Status: 204,
	}

	before := time.Now().UTC()
	err := repo.AppendHistory(ctx, entry)
	after := time.Now().UTC()

	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	var saved ComposeHistoryEntryModel
	db.First(&saved)
	if saved.When.IsZero() {
		t.Error("When should be set")
	}
	if saved.When.Before(before) || saved.When.After(after) {
		t.Error("When timestamp out of expected range")
	}
}

func TestHistoryRepo_ListHistory(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewHistoryRepo(db)
	ctx := context.Background()

	// Insert multiple entries
	entries := []domain.ComposeHistoryEntry{
		{ID: "e1", Method: "GET", URL: "http://a.com", Status: 200, When: time.Now().UTC().Add(-3 * time.Hour)},
		{ID: "e2", Method: "POST", URL: "http://b.com", Status: 201, When: time.Now().UTC().Add(-2 * time.Hour)},
		{ID: "e3", Method: "PUT", URL: "http://c.com", Status: 204, When: time.Now().UTC().Add(-1 * time.Hour)},
	}
	for _, e := range entries {
		if err := repo.AppendHistory(ctx, e); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}

	// List
	result, next, err := repo.ListHistory(ctx, "", 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("count: got %d, want 3", len(result))
	}

	// Should be ordered by When DESC (newest first)
	if result[0].ID != "e3" {
		t.Errorf("first entry ID: got %s, want e3", result[0].ID)
	}
	if result[2].ID != "e1" {
		t.Errorf("last entry ID: got %s, want e1", result[2].ID)
	}

	if next != "" {
		t.Errorf("next should be empty, got %s", next)
	}
}

func TestHistoryRepo_ListHistory_DefaultLimit(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewHistoryRepo(db)
	ctx := context.Background()

	// List with limit 0 should use default of 50
	result, _, err := repo.ListHistory(ctx, "", 0)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	// Should not error, just return empty list
	if result == nil {
		t.Error("result should not be nil")
	}
}

func TestHistoryRepo_ListHistory_EmptyDB(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewHistoryRepo(db)
	ctx := context.Background()

	result, next, err := repo.ListHistory(ctx, "", 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("count: got %d, want 0", len(result))
	}
	if next != "" {
		t.Errorf("next should be empty, got %s", next)
	}
}
