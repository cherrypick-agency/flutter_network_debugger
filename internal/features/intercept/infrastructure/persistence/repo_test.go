package persistence

import (
	"context"
	"testing"

	"network-debugger/internal/features/intercept/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&InterceptRuleModel{}, &InterceptConfigModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestRuleRepository_SaveAllAndList(t *testing.T) {
	db := newTestDB(t)
	repo := NewRuleRepository(db)
	ctx := context.Background()

	rules := []domain.InterceptRule{
		{
			ID:       "r1",
			Enabled:  true,
			Priority: 10,
			Action:   "request",
			When: domain.InterceptWhen{
				Host: &domain.RuleStringMatch{Equals: "example.com"},
			},
		},
		{
			ID:       "r2",
			Enabled:  true,
			Priority: 5,
			Action:   "both",
			When: domain.InterceptWhen{
				Method: []string{"POST"},
			},
		},
	}

	if err := repo.SaveAll(ctx, rules); err != nil {
		t.Fatalf("SaveAll: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 rules, got %d", len(list))
	}
	// Should be ordered by priority ASC
	if list[0].ID != "r2" {
		t.Errorf("expected r2 first (priority 5), got %s", list[0].ID)
	}
	if list[1].ID != "r1" {
		t.Errorf("expected r1 second (priority 10), got %s", list[1].ID)
	}

	// Verify When was serialized/deserialized correctly
	if list[0].When.Method[0] != "POST" {
		t.Errorf("expected Method[0]=POST, got %s", list[0].When.Method[0])
	}
	if list[1].When.Host == nil || list[1].When.Host.Equals != "example.com" {
		t.Errorf("expected Host.Equals=example.com")
	}

	// SaveAll again should replace
	newRules := []domain.InterceptRule{
		{
			ID:       "r3",
			Enabled:  true,
			Priority: 1,
			Action:   "response",
		},
	}
	if err := repo.SaveAll(ctx, newRules); err != nil {
		t.Fatalf("SaveAll replace: %v", err)
	}
	list2, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List after replace: %v", err)
	}
	if len(list2) != 1 || list2[0].ID != "r3" {
		t.Fatalf("expected 1 rule r3 after replace, got %d", len(list2))
	}
}

func TestConfigRepository_LoadAndSave(t *testing.T) {
	db := newTestDB(t)
	repo := NewConfigRepository(db)
	ctx := context.Background()

	// Load should return defaults when no config exists
	cfg, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TimeoutMs != 60000 {
		t.Errorf("expected default TimeoutMs=60000, got %d", cfg.TimeoutMs)
	}
	if cfg.QueueMax != 200 {
		t.Errorf("expected default QueueMax=200, got %d", cfg.QueueMax)
	}

	// Save
	cfg = domain.InterceptConfig{
		Enabled:      true,
		Requests:     true,
		Responses:    true,
		TimeoutMs:    30000,
		QueueMax:     100,
		BodyMaxBytes: 2 << 20,
		Reencode:     true,
		Overflow:     "drop-new",
	}
	if err := repo.Save(ctx, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load again
	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}
	if loaded.TimeoutMs != 30000 {
		t.Errorf("expected TimeoutMs=30000, got %d", loaded.TimeoutMs)
	}
	if loaded.Overflow != "drop-new" {
		t.Errorf("expected Overflow=drop-new, got %s", loaded.Overflow)
	}
	if !loaded.Enabled {
		t.Errorf("expected Enabled=true")
	}

	// Update (upsert)
	cfg.TimeoutMs = 45000
	if err := repo.Save(ctx, cfg); err != nil {
		t.Fatalf("Save update: %v", err)
	}
	loaded2, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load after update: %v", err)
	}
	if loaded2.TimeoutMs != 45000 {
		t.Errorf("expected TimeoutMs=45000, got %d", loaded2.TimeoutMs)
	}
}
