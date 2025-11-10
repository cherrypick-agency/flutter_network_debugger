package domain

import (
	"testing"
	"time"
)

// Composer 1.
func TestMapRule(t *testing.T) {
	now := time.Now()
	filePath := "/path/to/file"
	blobPath := "/blob/path"

	rule := MapRule{
		ID:                  "rule-1",
		Enabled:             true,
		Priority:            10,
		Kind:                KindLocal,
		StopProcessing:      true,
		Methods:             []string{"GET", "POST"},
		HostPattern:         "api.example.com",
		PathPattern:         "/users/*",
		PatternType:         PatternGlob,
		FilePath:            &filePath,
		BlobPath:            &blobPath,
		StatusOverride:      200,
		ContentTypeOverride: "application/json",
		TargetURLTemplate:   "",
		PreserveHost:        false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if rule.ID != "rule-1" {
		t.Errorf("Expected ID 'rule-1', got %q", rule.ID)
	}

	if !rule.Enabled {
		t.Error("Expected Enabled true")
	}

	if rule.Priority != 10 {
		t.Errorf("Expected Priority 10, got %d", rule.Priority)
	}

	if rule.Kind != KindLocal {
		t.Errorf("Expected Kind %v, got %v", KindLocal, rule.Kind)
	}

	if !rule.StopProcessing {
		t.Error("Expected StopProcessing true")
	}

	if len(rule.Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(rule.Methods))
	}

	if rule.HostPattern != "api.example.com" {
		t.Errorf("Expected HostPattern 'api.example.com', got %q", rule.HostPattern)
	}

	if rule.PatternType != PatternGlob {
		t.Errorf("Expected PatternType %v, got %v", PatternGlob, rule.PatternType)
	}
}

// Composer 1.
func TestMapRule_Remote(t *testing.T) {
	now := time.Now()
	rule := MapRule{
		ID:                "rule-2",
		Enabled:           true,
		Priority:          5,
		Kind:              KindRemote,
		StopProcessing:    false,
		Methods:           []string{"PUT"},
		HostPattern:       "*.example.com",
		PathPattern:       "/api/v1/*",
		PatternType:       PatternRegex,
		TargetURLTemplate: "https://backend.example.com{{.Path}}",
		PreserveHost:      true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if rule.Kind != KindRemote {
		t.Errorf("Expected Kind %v, got %v", KindRemote, rule.Kind)
	}

	if rule.TargetURLTemplate == "" {
		t.Error("Expected TargetURLTemplate to be set")
	}

	if !rule.PreserveHost {
		t.Error("Expected PreserveHost true")
	}

	if rule.PatternType != PatternRegex {
		t.Errorf("Expected PatternType %v, got %v", PatternRegex, rule.PatternType)
	}
}

// Composer 1.
func TestPatternType_Constants(t *testing.T) {
	if PatternGlob != "glob" {
		t.Errorf("Expected PatternGlob 'glob', got %q", PatternGlob)
	}

	if PatternRegex != "regex" {
		t.Errorf("Expected PatternRegex 'regex', got %q", PatternRegex)
	}
}

// Composer 1.
func TestKind_Constants(t *testing.T) {
	if KindLocal != "local" {
		t.Errorf("Expected KindLocal 'local', got %q", KindLocal)
	}

	if KindRemote != "remote" {
		t.Errorf("Expected KindRemote 'remote', got %q", KindRemote)
	}
}

// Composer 1.
func TestMapRule_NilPointers(t *testing.T) {
	rule := MapRule{
		ID:      "rule-3",
		Enabled: true,
		Kind:    KindLocal,
	}

	if rule.FilePath != nil {
		t.Error("Expected FilePath to be nil")
	}

	if rule.BlobPath != nil {
		t.Error("Expected BlobPath to be nil")
	}
}
