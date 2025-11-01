package httpapi

import (
	"testing"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
)

func TestToDTO(t *testing.T) {
	now := time.Now()
	filePath := "/path/to/file"
	blobPath := "blob123"

	rule := mdomain.MapRule{
		ID:                  "test-1",
		Enabled:             true,
		Priority:            1,
		Kind:                mdomain.KindLocal,
		StopProcessing:      true,
		Methods:             []string{"GET", "POST"},
		HostPattern:         "example.com",
		PathPattern:         "/api/*",
		PatternType:         mdomain.PatternGlob,
		FilePath:            &filePath,
		BlobPath:            &blobPath,
		StatusOverride:      200,
		ContentTypeOverride: "application/json",
		TargetURLTemplate:   "http://target.com",
		PreserveHost:        true,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	dto := toDTO(rule)

	if dto.ID != "test-1" {
		t.Errorf("ID = %q, want test-1", dto.ID)
	}
	if !dto.Enabled {
		t.Error("Enabled should be true")
	}
	if dto.Priority != 1 {
		t.Errorf("Priority = %d, want 1", dto.Priority)
	}
	if dto.Kind != "local" {
		t.Errorf("Kind = %q, want local", dto.Kind)
	}
	if !dto.StopProcessing {
		t.Error("StopProcessing should be true")
	}
	if len(dto.Methods) != 2 {
		t.Errorf("len(Methods) = %d, want 2", len(dto.Methods))
	}
	if dto.HostPattern != "example.com" {
		t.Errorf("HostPattern = %q, want example.com", dto.HostPattern)
	}
	if dto.PathPattern != "/api/*" {
		t.Errorf("PathPattern = %q, want /api/*", dto.PathPattern)
	}
	if dto.PatternType != "glob" {
		t.Errorf("PatternType = %q, want glob", dto.PatternType)
	}
	if dto.FilePath == nil || *dto.FilePath != filePath {
		t.Errorf("FilePath = %v, want %s", dto.FilePath, filePath)
	}
	if dto.BlobPath == nil || *dto.BlobPath != blobPath {
		t.Errorf("BlobPath = %v, want %s", dto.BlobPath, blobPath)
	}
	if dto.StatusOverride != 200 {
		t.Errorf("StatusOverride = %d, want 200", dto.StatusOverride)
	}
	if dto.ContentTypeOverride != "application/json" {
		t.Errorf("ContentTypeOverride = %q, want application/json", dto.ContentTypeOverride)
	}
	if dto.TargetURLTemplate != "http://target.com" {
		t.Errorf("TargetURLTemplate = %q, want http://target.com", dto.TargetURLTemplate)
	}
	if !dto.PreserveHost {
		t.Error("PreserveHost should be true")
	}
}

func TestFromDTO(t *testing.T) {
	now := time.Now()
	filePath := "/path/to/file"
	blobPath := "blob123"

	dto := mapRuleDTO{
		ID:                  "test-1",
		Enabled:             true,
		Priority:            1,
		Kind:                "remote",
		StopProcessing:      false,
		Methods:             []string{"  get  ", "POST", "", "   "},
		HostPattern:         "api.example.com",
		PathPattern:         "/v1/*",
		PatternType:         "regex",
		FilePath:            &filePath,
		BlobPath:            &blobPath,
		StatusOverride:      404,
		ContentTypeOverride: "text/plain",
		TargetURLTemplate:   "https://target.com",
		PreserveHost:        false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	rule := fromDTO(dto)

	if rule.ID != "test-1" {
		t.Errorf("ID = %q, want test-1", rule.ID)
	}
	if !rule.Enabled {
		t.Error("Enabled should be true")
	}
	if rule.Priority != 1 {
		t.Errorf("Priority = %d, want 1", rule.Priority)
	}
	if rule.Kind != mdomain.KindRemote {
		t.Errorf("Kind = %q, want remote", rule.Kind)
	}
	if rule.StopProcessing {
		t.Error("StopProcessing should be false")
	}
	// Should normalize methods: trim, uppercase, remove empty
	if len(rule.Methods) != 2 {
		t.Errorf("len(Methods) = %d, want 2 (normalized)", len(rule.Methods))
	}
	if rule.Methods[0] != "GET" {
		t.Errorf("Methods[0] = %q, want GET", rule.Methods[0])
	}
	if rule.Methods[1] != "POST" {
		t.Errorf("Methods[1] = %q, want POST", rule.Methods[1])
	}
	if rule.HostPattern != "api.example.com" {
		t.Errorf("HostPattern = %q, want api.example.com", rule.HostPattern)
	}
	if rule.PathPattern != "/v1/*" {
		t.Errorf("PathPattern = %q, want /v1/*", rule.PathPattern)
	}
	if rule.PatternType != mdomain.PatternRegex {
		t.Errorf("PatternType = %q, want regex", rule.PatternType)
	}
	if rule.FilePath == nil || *rule.FilePath != filePath {
		t.Errorf("FilePath = %v, want %s", rule.FilePath, filePath)
	}
	if rule.BlobPath == nil || *rule.BlobPath != blobPath {
		t.Errorf("BlobPath = %v, want %s", rule.BlobPath, blobPath)
	}
	if rule.StatusOverride != 404 {
		t.Errorf("StatusOverride = %d, want 404", rule.StatusOverride)
	}
	if rule.ContentTypeOverride != "text/plain" {
		t.Errorf("ContentTypeOverride = %q, want text/plain", rule.ContentTypeOverride)
	}
	if rule.TargetURLTemplate != "https://target.com" {
		t.Errorf("TargetURLTemplate = %q, want https://target.com", rule.TargetURLTemplate)
	}
	if rule.PreserveHost {
		t.Error("PreserveHost should be false")
	}
}

func TestFromDTO_EmptyMethods(t *testing.T) {
	dto := mapRuleDTO{
		ID:      "test-1",
		Methods: []string{},
	}

	rule := fromDTO(dto)

	if len(rule.Methods) != 0 {
		t.Errorf("len(Methods) = %d, want 0", len(rule.Methods))
	}
}

func TestFromDTO_NilPointers(t *testing.T) {
	dto := mapRuleDTO{
		ID:       "test-1",
		FilePath: nil,
		BlobPath: nil,
	}

	rule := fromDTO(dto)

	if rule.FilePath != nil {
		t.Error("FilePath should be nil")
	}
	if rule.BlobPath != nil {
		t.Error("BlobPath should be nil")
	}
}
