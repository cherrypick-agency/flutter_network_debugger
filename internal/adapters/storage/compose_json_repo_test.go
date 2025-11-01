package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"network-debugger/internal/domain"
)

func TestJSONComposeLibraryRepo_LoadSave(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "lib.json")
	r := NewJSONComposeLibraryRepo(p)

	// load empty -> default
	lib, err := r.LoadLibrary(context.Background())
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(lib.Collections) != 0 || len(lib.RequestsByID) != 0 {
		t.Fatalf("default not empty")
	}

	// save sample
	lib.Collections = []domain.ComposeCollection{{ID: "c1", Name: "Default", Root: domain.ComposeFolder{ID: "root", Name: "root"}}}
	lib.RequestsByID = map[string]domain.ComposeRequestTemplate{"r1": {ID: "r1", Name: "Ping", Method: "GET", URL: "http://example.com"}}
	if err := r.SaveLibrary(context.Background(), lib); err != nil {
		t.Fatalf("save: %v", err)
	}

	// reload
	lib2, err := r.LoadLibrary(context.Background())
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(lib2.Collections) != 1 || len(lib2.RequestsByID) != 1 {
		t.Fatalf("unexpected sizes")
	}

	// corrupt file -> recover to empty
	if err := os.WriteFile(p, []byte("{"), 0o644); err != nil {
		t.Fatalf("corrupt write: %v", err)
	}
	lib3, err := r.LoadLibrary(context.Background())
	if err != nil {
		t.Fatalf("load after corrupt: %v", err)
	}
	if lib3.RequestsByID == nil {
		t.Fatalf("requests map nil")
	}
}

func TestJSONComposeHistoryRepo_AppendList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.json")
	r := NewJSONComposeHistoryRepo(p, 3)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := r.AppendHistory(ctx, domain.ComposeHistoryEntry{ID: "e"}); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	items, next, err := r.ListHistory(ctx, "", 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected trim to 3, got %d", len(items))
	}
	_ = next
}

// Additional comprehensive tests

func TestJSONComposeLibraryRepo_EnsureDir(t *testing.T) {
	t.Parallel()

	t.Run("creates nested directories", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "a", "b", "c", "lib.json")
		r := NewJSONComposeLibraryRepo(p)

		lib := domain.ComposeLibrary{
			Collections:  []domain.ComposeCollection{},
			RequestsByID: map[string]domain.ComposeRequestTemplate{},
		}

		err := r.SaveLibrary(context.Background(), lib)
		if err != nil {
			t.Fatalf("save with nested dirs failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Error("file was not created")
		}
	})

	t.Run("handles current directory path", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// Use relative path within tempdir
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(dir)

		r := NewJSONComposeLibraryRepo("lib.json") // relative path

		lib := domain.ComposeLibrary{
			Collections:  []domain.ComposeCollection{},
			RequestsByID: map[string]domain.ComposeRequestTemplate{},
		}

		err := r.SaveLibrary(context.Background(), lib)
		if err != nil {
			t.Fatalf("save with current dir failed: %v", err)
		}
	})
}

func TestJSONComposeLibraryRepo_LoadLibrary_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil RequestsByID is initialized", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "lib.json")

		// Manually write JSON with null RequestsByID
		data := `{"collections":[],"requestsByID":null}`
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		r := NewJSONComposeLibraryRepo(p)
		lib, err := r.LoadLibrary(context.Background())

		if err != nil {
			t.Fatalf("load failed: %v", err)
		}
		if lib.RequestsByID == nil {
			t.Error("RequestsByID should be initialized to empty map")
		}
	})

	t.Run("permission error on file", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("skipping permission test when running as root")
		}
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "lib.json")

		// Create file with no read permissions
		if err := os.WriteFile(p, []byte("{}"), 0o000); err != nil {
			t.Fatalf("write: %v", err)
		}
		defer os.Chmod(p, 0o644) // cleanup

		r := NewJSONComposeLibraryRepo(p)
		_, err := r.LoadLibrary(context.Background())

		if err == nil {
			t.Error("expected permission error")
		}
	})
}

func TestJSONComposeLibraryRepo_SaveLibrary_Atomicity(t *testing.T) {
	t.Parallel()

	t.Run("uses temp file and rename", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		p := filepath.Join(dir, "lib.json")
		r := NewJSONComposeLibraryRepo(p)

		lib := domain.ComposeLibrary{
			Collections:  []domain.ComposeCollection{{ID: "c1"}},
			RequestsByID: map[string]domain.ComposeRequestTemplate{},
		}

		err := r.SaveLibrary(context.Background(), lib)
		if err != nil {
			t.Fatalf("save failed: %v", err)
		}

		// Check temp file is cleaned up
		tmpPath := p + ".tmp"
		if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
			t.Error("temp file should be removed after successful save")
		}

		// Final file should exist
		if _, err := os.Stat(p); err != nil {
			t.Errorf("final file missing: %v", err)
		}
	})
}

func TestNewJSONComposeHistoryRepo_MaxItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"positive value", 100, 100},
		{"zero defaults to 200", 0, 200},
		{"negative defaults to 200", -10, 200},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := NewJSONComposeHistoryRepo("/tmp/test.json", tt.input)

			if r.maxItems != tt.expected {
				t.Errorf("maxItems = %d, want %d", r.maxItems, tt.expected)
			}
		})
	}
}

func TestJSONComposeHistoryRepo_AppendHistory_Timestamping(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.json")
	r := NewJSONComposeHistoryRepo(p, 10)
	ctx := context.Background()

	// Entry without timestamp
	entry := domain.ComposeHistoryEntry{
		ID:     "e1",
		Method: "GET",
		URL:    "http://test.com",
		Status: 200,
	}

	err := r.AppendHistory(ctx, entry)
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// Load and verify timestamp was set
	items, _, err := r.ListHistory(ctx, "", 10)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].When.IsZero() {
		t.Error("timestamp should be set on append")
	}
}

func TestJSONComposeHistoryRepo_ListHistory_Pagination(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.json")
	r := NewJSONComposeHistoryRepo(p, 100)
	ctx := context.Background()

	// Add 10 entries
	for i := 0; i < 10; i++ {
		entry := domain.ComposeHistoryEntry{
			ID:     string(rune('a' + i)),
			Method: "GET",
			Status: 200,
		}
		if err := r.AppendHistory(ctx, entry); err != nil {
			t.Fatalf("append %d failed: %v", i, err)
		}
	}

	t.Run("limit smaller than total", func(t *testing.T) {
		items, next, err := r.ListHistory(ctx, "", 5)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}

		// Should return last 5 items
		if len(items) != 5 {
			t.Errorf("items count = %d, want 5", len(items))
		}

		// Should have next cursor
		if next == "" {
			t.Error("expected next cursor when results are limited")
		}
	})

	t.Run("limit larger than total", func(t *testing.T) {
		items, next, err := r.ListHistory(ctx, "", 20)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}

		if len(items) != 10 {
			t.Errorf("items count = %d, want 10", len(items))
		}

		if next != "" {
			t.Errorf("next should be empty, got %s", next)
		}
	})

	t.Run("zero limit uses default 50", func(t *testing.T) {
		items, _, err := r.ListHistory(ctx, "", 0)
		if err != nil {
			t.Fatalf("list failed: %v", err)
		}

		// Should return all 10 items
		if len(items) != 10 {
			t.Errorf("items count = %d, want 10", len(items))
		}
	})
}

func TestJSONComposeHistoryRepo_ListHistory_EmptyFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.json")
	r := NewJSONComposeHistoryRepo(p, 10)

	items, next, err := r.ListHistory(context.Background(), "", 10)

	if err != nil {
		t.Fatalf("list on empty file failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
	if next != "" {
		t.Errorf("next should be empty, got %s", next)
	}
}

func TestJSONComposeHistoryRepo_ListHistory_CorruptFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.json")

	// Write corrupt JSON
	if err := os.WriteFile(p, []byte("{corrupt"), 0o644); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}

	r := NewJSONComposeHistoryRepo(p, 10)
	items, next, err := r.ListHistory(context.Background(), "", 10)

	// Should recover to empty list on corrupt data
	if err != nil {
		t.Fatalf("list should recover from corrupt, got error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list on corrupt, got %d items", len(items))
	}
	if next != "" {
		t.Errorf("next should be empty, got %s", next)
	}
}

func TestJSONComposeLibraryRepo_SaveLibrary_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("directory creation error", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("skipping permission test when running as root")
		}
		t.Parallel()

		dir := t.TempDir()
		readOnlyDir := filepath.Join(dir, "readonly")
		if err := os.Mkdir(readOnlyDir, 0o444); err != nil {
			t.Fatalf("create readonly dir: %v", err)
		}
		defer os.Chmod(readOnlyDir, 0o755)

		p := filepath.Join(readOnlyDir, "subdir", "lib.json")
		r := NewJSONComposeLibraryRepo(p)

		lib := domain.ComposeLibrary{
			Collections:  []domain.ComposeCollection{},
			RequestsByID: map[string]domain.ComposeRequestTemplate{},
		}

		err := r.SaveLibrary(context.Background(), lib)
		if err == nil {
			t.Error("expected error when directory can't be created")
		}
	})
}

func TestJSONComposeHistoryRepo_SaveAll_MaxItems(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := filepath.Join(dir, "hist.json")
	r := NewJSONComposeHistoryRepo(p, 5)

	ctx := context.Background()

	// Add more items than maxItems
	for i := 0; i < 10; i++ {
		entry := domain.ComposeHistoryEntry{
			ID:     string(rune('a' + i)),
			Method: "GET",
			Status: 200,
		}
		if err := r.AppendHistory(ctx, entry); err != nil {
			t.Fatalf("append %d failed: %v", i, err)
		}
	}

	// Should only keep the last 5 items
	items, _, err := r.ListHistory(ctx, "", 100)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(items) != 5 {
		t.Errorf("expected 5 items after trimming, got %d", len(items))
	}

	// Verify the last 5 items are kept (f, g, h, i, j)
	expectedFirst := string(rune('a' + 5)) // 'f'
	if len(items) > 0 && items[0].ID != expectedFirst {
		t.Errorf("expected first item to be %q, got %q", expectedFirst, items[0].ID)
	}
}
