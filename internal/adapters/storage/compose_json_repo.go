package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
)

// JSONComposeLibraryRepo persists compose library to a single JSON file.
type JSONComposeLibraryRepo struct {
	path string
	mu   sync.RWMutex
}

func NewJSONComposeLibraryRepo(path string) *JSONComposeLibraryRepo {
	return &JSONComposeLibraryRepo{path: path}
}

func (r *JSONComposeLibraryRepo) ensureDir() error {
	dir := filepath.Dir(r.path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func (r *JSONComposeLibraryRepo) LoadLibrary(ctx context.Context) (domain.ComposeLibrary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var lib domain.ComposeLibrary
	f, err := os.Open(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// default empty library
			return domain.ComposeLibrary{Collections: []domain.ComposeCollection{}, RequestsByID: map[string]domain.ComposeRequestTemplate{}}, nil
		}
		return domain.ComposeLibrary{}, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err := dec.Decode(&lib); err != nil {
		// On decode error, return empty library to recover softly
		return domain.ComposeLibrary{Collections: []domain.ComposeCollection{}, RequestsByID: map[string]domain.ComposeRequestTemplate{}}, nil
	}
	if lib.RequestsByID == nil {
		lib.RequestsByID = map[string]domain.ComposeRequestTemplate{}
	}
	return lib, nil
}

func (r *JSONComposeLibraryRepo) SaveLibrary(ctx context.Context, lib domain.ComposeLibrary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.ensureDir(); err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(lib); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, r.path)
}

// JSONComposeHistoryRepo keeps a rolling history JSON file with trimming.
type JSONComposeHistoryRepo struct {
	path     string
	maxItems int
	mu       sync.Mutex
}

func NewJSONComposeHistoryRepo(path string, maxItems int) *JSONComposeHistoryRepo {
	if maxItems <= 0 {
		maxItems = 200
	}
	return &JSONComposeHistoryRepo{path: path, maxItems: maxItems}
}

func (r *JSONComposeHistoryRepo) loadAll() ([]domain.ComposeHistoryEntry, error) {
	f, err := os.Open(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []domain.ComposeHistoryEntry{}, nil
		}
		return nil, err
	}
	defer f.Close()
	var items []domain.ComposeHistoryEntry
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		// recover to empty on decode error
		return []domain.ComposeHistoryEntry{}, nil
	}
	return items, nil
}

func (r *JSONComposeHistoryRepo) saveAll(items []domain.ComposeHistoryEntry) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	tmp := r.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, r.path)
}

func (r *JSONComposeHistoryRepo) AppendHistory(ctx context.Context, e domain.ComposeHistoryEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	items, err := r.loadAll()
	if err != nil {
		return err
	}
	// Add timestamp if missing
	if e.When.IsZero() {
		e.When = time.Now().UTC()
	}
	items = append(items, e)
	if len(items) > r.maxItems {
		items = items[len(items)-r.maxItems:]
	}
	return r.saveAll(items)
}

func (r *JSONComposeHistoryRepo) ListHistory(ctx context.Context, from string, limit int) ([]domain.ComposeHistoryEntry, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items, err := r.loadAll()
	if err != nil {
		return nil, "", err
	}
	if limit <= 0 {
		limit = 50
	}
	// For simplicity, ignore cursor and return last N items; future: real cursor
	n := len(items)
	if n == 0 {
		return []domain.ComposeHistoryEntry{}, "", nil
	}
	start := n - limit
	if start < 0 {
		start = 0
	}
	next := ""
	if start > 0 {
		next = fmt.Sprintf("%d", start)
	}
	out := make([]domain.ComposeHistoryEntry, n-start)
	copy(out, items[start:])
	return out, next, nil
}

// Ensure interfaces are implemented
var _ usecase.ComposeLibraryRepository = (*JSONComposeLibraryRepo)(nil)
var _ usecase.ComposeHistoryRepository = (*JSONComposeHistoryRepo)(nil)
