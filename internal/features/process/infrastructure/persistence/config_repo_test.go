package persistence

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"network-debugger/internal/features/process/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.AutoMigrate(&ProcessDetectionConfigModel{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// Composer 1.
func TestNewConfigRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	if repo == nil {
		t.Fatal("NewConfigRepo returned nil")
	}

	if repo.db == nil {
		t.Fatal("repo.db is nil")
	}
}

// Composer 1.
func TestConfigRepo_Load_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	ctx := context.Background()
	cfg, err := repo.Load(ctx)

	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Проверяем дефолтные значения
	if cfg.ID != 1 {
		t.Errorf("Load() config.ID = %d, want 1", cfg.ID)
	}

	if !cfg.Enabled {
		t.Error("Load() config.Enabled = false, want true")
	}

	if cfg.UseHelperTool {
		t.Error("Load() config.UseHelperTool = true, want false")
	}

	if cfg.HelperInstalled {
		t.Error("Load() config.HelperInstalled = true, want false")
	}

	if !cfg.CacheEnabled {
		t.Error("Load() config.CacheEnabled = false, want true")
	}

	if cfg.CacheTTLSeconds != 300 {
		t.Errorf("Load() config.CacheTTLSeconds = %d, want 300", cfg.CacheTTLSeconds)
	}

	if !cfg.FallbackEnabled {
		t.Error("Load() config.FallbackEnabled = false, want true")
	}
}

// Composer 1.
func TestConfigRepo_Load_Existing(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	ctx := context.Background()

	// Создаем конфигурацию вручную используя прямой SQL чтобы избежать дефолтных значений GORM
	result := db.Exec(`
		INSERT INTO process_detection_config 
		(id, enabled, use_helper_tool, helper_installed, cache_enabled, cache_ttl_seconds, fallback_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
	`, 1, false, true, true, false, 600, false)
	if result.Error != nil {
		t.Fatalf("Failed to create config: %v", result.Error)
	}
	if result.RowsAffected != 1 {
		t.Fatalf("Expected 1 row affected, got %d", result.RowsAffected)
	}

	cfg, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.ID != 1 {
		t.Errorf("Load() config.ID = %d, want 1", cfg.ID)
	}

	if cfg.Enabled {
		t.Error("Load() config.Enabled = true, want false")
	}

	if !cfg.UseHelperTool {
		t.Error("Load() config.UseHelperTool = false, want true")
	}

	if !cfg.HelperInstalled {
		t.Error("Load() config.HelperInstalled = false, want true")
	}

	if cfg.CacheEnabled {
		t.Error("Load() config.CacheEnabled = true, want false")
	}

	if cfg.CacheTTLSeconds != 600 {
		t.Errorf("Load() config.CacheTTLSeconds = %d, want 600", cfg.CacheTTLSeconds)
	}

	if cfg.FallbackEnabled {
		t.Error("Load() config.FallbackEnabled = true, want false")
	}
}

// Composer 1.
func TestConfigRepo_Save(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	ctx := context.Background()

	cfg := &domain.DetectionConfig{
		ID:              1,
		Enabled:         true,
		UseHelperTool:   true,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTLSeconds: 500,
		FallbackEnabled: true,
		UpdatedAt:       time.Now(),
	}

	err := repo.Save(ctx, cfg)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	// Проверяем что сохранилось
	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load() after Save error = %v, want nil", err)
	}

	if loaded.ID != cfg.ID {
		t.Errorf("Loaded config.ID = %d, want %d", loaded.ID, cfg.ID)
	}

	if loaded.Enabled != cfg.Enabled {
		t.Errorf("Loaded config.Enabled = %v, want %v", loaded.Enabled, cfg.Enabled)
	}

	if loaded.UseHelperTool != cfg.UseHelperTool {
		t.Errorf("Loaded config.UseHelperTool = %v, want %v", loaded.UseHelperTool, cfg.UseHelperTool)
	}

	if loaded.CacheTTLSeconds != cfg.CacheTTLSeconds {
		t.Errorf("Loaded config.CacheTTLSeconds = %d, want %d", loaded.CacheTTLSeconds, cfg.CacheTTLSeconds)
	}
}

// Composer 1.
func TestConfigRepo_Save_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConfigRepo(db)

	ctx := context.Background()

	// Создаем начальную конфигурацию
	initialCfg := &domain.DetectionConfig{
		ID:              1,
		Enabled:         true,
		UseHelperTool:   false,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTLSeconds: 300,
		FallbackEnabled: true,
	}

	if err := repo.Save(ctx, initialCfg); err != nil {
		t.Fatalf("Save() initial error = %v", err)
	}

	// Обновляем конфигурацию
	updatedCfg := &domain.DetectionConfig{
		ID:              1,
		Enabled:         false,
		UseHelperTool:   true,
		HelperInstalled: true,
		CacheEnabled:    false,
		CacheTTLSeconds: 900,
		FallbackEnabled: false,
	}

	if err := repo.Save(ctx, updatedCfg); err != nil {
		t.Fatalf("Save() update error = %v", err)
	}

	// Проверяем что обновилось
	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("Load() after update error = %v", err)
	}

	if loaded.Enabled != updatedCfg.Enabled {
		t.Errorf("Loaded config.Enabled = %v, want %v", loaded.Enabled, updatedCfg.Enabled)
	}

	if loaded.CacheTTLSeconds != updatedCfg.CacheTTLSeconds {
		t.Errorf("Loaded config.CacheTTLSeconds = %d, want %d", loaded.CacheTTLSeconds, updatedCfg.CacheTTLSeconds)
	}
}

// Composer 1.
func TestToDomainConfig(t *testing.T) {
	model := ProcessDetectionConfigModel{
		ID:              1,
		Enabled:         true,
		UseHelperTool:   false,
		HelperInstalled: true,
		CacheEnabled:    true,
		CacheTTLSeconds: 400,
		FallbackEnabled: false,
		UpdatedAt:       time.Now(),
	}

	cfg := toDomainConfig(model)

	if cfg.ID != model.ID {
		t.Errorf("toDomainConfig() ID = %d, want %d", cfg.ID, model.ID)
	}

	if cfg.Enabled != model.Enabled {
		t.Errorf("toDomainConfig() Enabled = %v, want %v", cfg.Enabled, model.Enabled)
	}

	if cfg.CacheTTLSeconds != model.CacheTTLSeconds {
		t.Errorf("toDomainConfig() CacheTTLSeconds = %d, want %d", cfg.CacheTTLSeconds, model.CacheTTLSeconds)
	}
}

// Composer 1.
func TestToModel(t *testing.T) {
	cfg := &domain.DetectionConfig{
		ID:              1,
		Enabled:         false,
		UseHelperTool:   true,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTLSeconds: 700,
		FallbackEnabled: true,
	}

	model := toModel(cfg)

	if model.ID != cfg.ID {
		t.Errorf("toModel() ID = %d, want %d", model.ID, cfg.ID)
	}

	if model.Enabled != cfg.Enabled {
		t.Errorf("toModel() Enabled = %v, want %v", model.Enabled, cfg.Enabled)
	}

	if model.CacheTTLSeconds != cfg.CacheTTLSeconds {
		t.Errorf("toModel() CacheTTLSeconds = %d, want %d", model.CacheTTLSeconds, cfg.CacheTTLSeconds)
	}
}
