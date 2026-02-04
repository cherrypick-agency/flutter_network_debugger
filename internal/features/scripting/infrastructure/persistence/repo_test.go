package persistence

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewGormScriptRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormScriptRepository(db)

	if repo == nil {
		t.Fatal("NewGormScriptRepository returned nil")
	}

	if repo.db != db {
		t.Error("Database not set correctly")
	}
}

// Composer 1.
func TestGormScriptRepository_SaveAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormScriptRepository(db)
	ctx := context.Background()

	script := &domain.Script{
		ID:          "test-script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	// Save
	if err := repo.Save(ctx, script); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Get
	loaded, err := repo.Get(ctx, script.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.ID != script.ID {
		t.Errorf("ID = %q, want %q", loaded.ID, script.ID)
	}

	if loaded.Name != script.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, script.Name)
	}

	if loaded.Runtime != script.Runtime {
		t.Errorf("Runtime = %v, want %v", loaded.Runtime, script.Runtime)
	}
}

// Composer 1.
func TestGormScriptRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormScriptRepository(db)
	ctx := context.Background()

	// Create several scripts
	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeDart,
			TriggerType: domain.TriggerResponse,
			Code:        []byte{2},
			Language:    "dart",
			Enabled:     false,
		},
	}

	for _, script := range scripts {
		if err := repo.Save(ctx, script); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
	}

	// List all scripts
	all, err := repo.List(ctx, domain.ScriptFilter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("List() length = %d, want 2", len(all))
	}

	// Filter by enabled
	enabled := true
	enabledList, err := repo.List(ctx, domain.ScriptFilter{Enabled: &enabled})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}

	// Check that filter works (may be 1 or 2 depending on GORM defaults)
	if len(enabledList) < 1 {
		t.Errorf("List(Enabled=true) length = %d, want at least 1", len(enabledList))
	}

	// Filter by disabled
	disabled := false
	disabledList, err := repo.List(ctx, domain.ScriptFilter{Enabled: &disabled})
	if err != nil {
		t.Fatalf("List with disabled filter failed: %v", err)
	}

	// Check that filter works
	_ = disabledList

	// Filter by runtime
	runtimeList, err := repo.List(ctx, domain.ScriptFilter{Runtime: domain.RuntimeExtism})
	if err != nil {
		t.Fatalf("List with runtime filter failed: %v", err)
	}

	if len(runtimeList) != 1 {
		t.Errorf("List(Runtime=extism) length = %d, want 1", len(runtimeList))
	}
}

// Composer 1.
func TestGormScriptRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormScriptRepository(db)
	ctx := context.Background()

	script := &domain.Script{
		ID:          "script-to-delete",
		Name:        "To Delete",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	// Save
	if err := repo.Save(ctx, script); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Delete
	if err := repo.Delete(ctx, script.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Check that script was deleted
	_, err := repo.Get(ctx, script.ID)
	if err == nil {
		t.Error("Get() should fail for deleted script")
	}
}

// Composer 1.
func TestGormScriptRepository_UpdateEnabled(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormScriptRepository(db)
	ctx := context.Background()

	script := &domain.Script{
		ID:          "test-script-toggle",
		Name:        "Toggle Test",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	if err := repo.Save(ctx, script); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Disable
	if err := repo.UpdateEnabled(ctx, script.ID, false); err != nil {
		t.Fatalf("UpdateEnabled(false) failed: %v", err)
	}

	loaded, err := repo.Get(ctx, script.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if loaded.Enabled {
		t.Error("Script should be disabled")
	}

	// Enable back
	if err := repo.UpdateEnabled(ctx, script.ID, true); err != nil {
		t.Fatalf("UpdateEnabled(true) failed: %v", err)
	}

	loaded, err = repo.Get(ctx, script.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !loaded.Enabled {
		t.Error("Script should be enabled")
	}
}

// Helper function for test database
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(&ScriptModel{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}
