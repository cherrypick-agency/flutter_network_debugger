package httpapi

import (
	"bytes"
	"encoding/json"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpoolBody_Success(t *testing.T) {
	deps := &Deps{Cfg: cfgpkg.Config{}}
	data := []byte("test body content")
	reader := bytes.NewReader(data)

	path, err := deps.spoolBody(reader, 1024, "test")

	if err != nil {
		t.Fatalf("spoolBody failed: %v", err)
	}
	if path == "" {
		t.Error("path should not be empty")
	}

	// Verify file exists and contains data
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read spooled file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("content = %q, want %q", content, data)
	}

	// Cleanup
	_ = os.Remove(path)
}

func TestSpoolBody_WithSpoolDir(t *testing.T) {
	// Create a temporary directory for this test
	tempDir := filepath.Join(os.TempDir(), "test-spool-dir")
	err := os.MkdirAll(tempDir, 0o755)
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	deps := &Deps{Cfg: cfgpkg.Config{BodySpoolDir: tempDir}}
	data := []byte("spool test")
	reader := bytes.NewReader(data)

	path, err := deps.spoolBody(reader, 1024, "custom")

	if err != nil {
		t.Fatalf("spoolBody failed: %v", err)
	}
	if !strings.Contains(path, tempDir) {
		t.Errorf("path should contain custom spool dir: %s", path)
	}

	// Verify file is in the custom directory
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read spooled file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("content = %q, want %q", content, data)
	}

	// Cleanup
	_ = os.Remove(path)
}

func TestSpoolBody_MaxSize(t *testing.T) {
	deps := &Deps{Cfg: cfgpkg.Config{}}
	data := []byte("1234567890") // 10 bytes
	reader := bytes.NewReader(data)

	// Only read first 5 bytes
	path, err := deps.spoolBody(reader, 5, "limited")

	if err != nil {
		t.Fatalf("spoolBody failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read spooled file: %v", err)
	}
	if len(content) != 5 {
		t.Errorf("content length = %d, want 5", len(content))
	}
	if string(content) != "12345" {
		t.Errorf("content = %q, want %q", content, "12345")
	}

	// Cleanup
	_ = os.Remove(path)
}

func TestSpoolBody_EmptyReader(t *testing.T) {
	deps := &Deps{Cfg: cfgpkg.Config{}}
	reader := bytes.NewReader([]byte{})

	path, err := deps.spoolBody(reader, 1024, "empty")

	if err != nil {
		t.Fatalf("spoolBody failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read spooled file: %v", err)
	}
	if len(content) != 0 {
		t.Errorf("content should be empty, got %d bytes", len(content))
	}

	// Cleanup
	_ = os.Remove(path)
}

func TestTryCompactJSON_ValidJSON(t *testing.T) {
	input := []byte(`{"name": "test",  "value":  123}`)

	result := tryCompactJSON(input)

	if result == "" {
		t.Error("result should not be empty for valid JSON")
	}

	// Verify it's valid JSON
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(result), &v); err != nil {
		t.Errorf("result is not valid JSON: %v", err)
	}

	// Verify values are preserved
	if v["name"] != "test" {
		t.Errorf("name = %v, want test", v["name"])
	}
	if v["value"].(float64) != 123 {
		t.Errorf("value = %v, want 123", v["value"])
	}
}

func TestTryCompactJSON_InvalidJSON(t *testing.T) {
	input := []byte(`{not valid json}`)

	result := tryCompactJSON(input)

	if result != "" {
		t.Error("result should be empty for invalid JSON")
	}
}

func TestTryCompactJSON_EmptyInput(t *testing.T) {
	input := []byte(``)

	result := tryCompactJSON(input)

	if result != "" {
		t.Error("result should be empty for empty input")
	}
}

func TestTryCompactJSON_Array(t *testing.T) {
	input := []byte(`[1, 2,  3,   4]`)

	result := tryCompactJSON(input)

	if result == "" {
		t.Error("result should not be empty for valid JSON array")
	}

	// Verify it's valid JSON
	var v []interface{}
	if err := json.Unmarshal([]byte(result), &v); err != nil {
		t.Errorf("result is not valid JSON: %v", err)
	}
	if len(v) != 4 {
		t.Errorf("array length = %d, want 4", len(v))
	}
}

func TestAugmentPreviewWithTimings_ValidJSON(t *testing.T) {
	preview := `{"status":200,"method":"GET"}`

	result := augmentPreviewWithTimings(preview, 100, 500)

	if result == preview {
		t.Error("result should be different from input")
	}

	// Parse and verify
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// Check that original fields are preserved
	if m["status"].(float64) != 200 {
		t.Error("status should be preserved")
	}
	if m["method"] != "GET" {
		t.Error("method should be preserved")
	}

	// Check that timings were added
	timings, ok := m["timings"].(map[string]interface{})
	if !ok {
		t.Fatal("timings should be added")
	}
	if timings["ttfbMs"].(float64) != 100 {
		t.Errorf("ttfbMs = %v, want 100", timings["ttfbMs"])
	}
	if timings["totalMs"].(float64) != 500 {
		t.Errorf("totalMs = %v, want 500", timings["totalMs"])
	}
}

func TestAugmentPreviewWithTimings_InvalidJSON(t *testing.T) {
	preview := `{invalid json}`

	result := augmentPreviewWithTimings(preview, 100, 500)

	if result != preview {
		t.Error("result should be unchanged for invalid JSON")
	}
}

func TestAugmentPreviewWithTimings_EmptyPreview(t *testing.T) {
	preview := ``

	result := augmentPreviewWithTimings(preview, 100, 500)

	if result != preview {
		t.Error("result should be unchanged for empty preview")
	}
}

func TestAugmentPreviewWithTimings_ZeroTimings(t *testing.T) {
	preview := `{"status":200}`

	result := augmentPreviewWithTimings(preview, 0, 0)

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	timings, ok := m["timings"].(map[string]interface{})
	if !ok {
		t.Fatal("timings should be added even with zero values")
	}
	if timings["ttfbMs"].(float64) != 0 {
		t.Errorf("ttfbMs = %v, want 0", timings["ttfbMs"])
	}
	if timings["totalMs"].(float64) != 0 {
		t.Errorf("totalMs = %v, want 0", timings["totalMs"])
	}
}

func TestAugmentPreviewWithTimings_NegativeTimings(t *testing.T) {
	preview := `{"status":200}`

	result := augmentPreviewWithTimings(preview, -10, -20)

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	timings, ok := m["timings"].(map[string]interface{})
	if !ok {
		t.Fatal("timings should be added")
	}
	if timings["ttfbMs"].(float64) != -10 {
		t.Errorf("ttfbMs = %v, want -10", timings["ttfbMs"])
	}
	if timings["totalMs"].(float64) != -20 {
		t.Errorf("totalMs = %v, want -20", timings["totalMs"])
	}
}

func TestAugmentPreviewWithTimings_OverwritesExistingTimings(t *testing.T) {
	preview := `{"status":200,"timings":{"old":999}}`

	result := augmentPreviewWithTimings(preview, 100, 500)

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	timings, ok := m["timings"].(map[string]interface{})
	if !ok {
		t.Fatal("timings should be present")
	}

	// Old timing should be overwritten
	if _, exists := timings["old"]; exists {
		t.Error("old timing should be overwritten")
	}

	// New timings should be present
	if timings["ttfbMs"].(float64) != 100 {
		t.Errorf("ttfbMs = %v, want 100", timings["ttfbMs"])
	}
	if timings["totalMs"].(float64) != 500 {
		t.Errorf("totalMs = %v, want 500", timings["totalMs"])
	}
}
