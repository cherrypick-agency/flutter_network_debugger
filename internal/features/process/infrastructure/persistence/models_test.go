package persistence

import (
	"testing"
	"time"
)

// Composer 1.
func TestProcessDetectionConfigModel_TableName(t *testing.T) {
	model := ProcessDetectionConfigModel{}
	tableName := model.TableName()

	if tableName != "process_detection_config" {
		t.Errorf("TableName() = %q, want %q", tableName, "process_detection_config")
	}
}

// Composer 1.
func TestIconCacheModel_TableName(t *testing.T) {
	model := IconCacheModel{}
	tableName := model.TableName()

	if tableName != "icon_cache" {
		t.Errorf("TableName() = %q, want %q", tableName, "icon_cache")
	}
}

// Composer 1.
func TestProcessDetectionConfigModel_Fields(t *testing.T) {
	now := time.Now()
	model := ProcessDetectionConfigModel{
		ID:              1,
		Enabled:         true,
		UseHelperTool:   false,
		HelperInstalled: true,
		CacheEnabled:    true,
		CacheTTLSeconds: 300,
		FallbackEnabled: false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if model.ID != 1 {
		t.Errorf("ID = %d, want 1", model.ID)
	}

	if !model.Enabled {
		t.Error("Enabled = false, want true")
	}

	if model.UseHelperTool {
		t.Error("UseHelperTool = true, want false")
	}

	if !model.HelperInstalled {
		t.Error("HelperInstalled = false, want true")
	}

	if !model.CacheEnabled {
		t.Error("CacheEnabled = false, want true")
	}

	if model.CacheTTLSeconds != 300 {
		t.Errorf("CacheTTLSeconds = %d, want 300", model.CacheTTLSeconds)
	}

	if model.FallbackEnabled {
		t.Error("FallbackEnabled = true, want false")
	}
}

// Composer 1.
func TestIconCacheModel_Fields(t *testing.T) {
	now := time.Now()
	path := "/path/to/icon.png"
	model := IconCacheModel{
		CacheKey:   "test-key",
		IconFormat: "png",
		IconData:   []byte("test data"),
		IconPath:   &path,
		ExpiresAt:  now.Add(5 * time.Minute),
		CreatedAt:  now,
	}

	if model.CacheKey != "test-key" {
		t.Errorf("CacheKey = %q, want %q", model.CacheKey, "test-key")
	}

	if model.IconFormat != "png" {
		t.Errorf("IconFormat = %q, want %q", model.IconFormat, "png")
	}

	if string(model.IconData) != "test data" {
		t.Errorf("IconData = %q, want %q", string(model.IconData), "test data")
	}

	if model.IconPath == nil || *model.IconPath != path {
		t.Errorf("IconPath = %v, want %q", model.IconPath, path)
	}
}

// Composer 1.
func TestIconCacheModel_NilPath(t *testing.T) {
	model := IconCacheModel{
		CacheKey:   "test-key",
		IconFormat: "png",
		IconData:   []byte("test data"),
		IconPath:   nil,
		ExpiresAt:  time.Now(),
	}

	if model.IconPath != nil {
		t.Error("IconPath should be nil")
	}
}
