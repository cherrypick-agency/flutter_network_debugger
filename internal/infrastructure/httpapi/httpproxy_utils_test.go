package httpapi

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"network-debugger/internal/infrastructure/config"
)

func TestSpoolBody(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxBytes  int64
		kind      string
		spoolDir  string
		expectErr bool
	}{
		{
			name:      "small body",
			input:     "hello world",
			maxBytes:  100,
			kind:      "req",
			expectErr: false,
		},
		{
			name:      "large body truncated",
			input:     strings.Repeat("a", 1000),
			maxBytes:  50,
			kind:      "resp",
			expectErr: false,
		},
		{
			name:      "empty body",
			input:     "",
			maxBytes:  100,
			kind:      "ws",
			expectErr: false,
		},
		{
			name:      "custom spool dir",
			input:     "test data",
			maxBytes:  100,
			kind:      "req",
			spoolDir:  "",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.spoolDir
			if dir == "" {
				dir = t.TempDir()
			}

			d := &Deps{
				Cfg: config.Config{
					BodySpoolDir: dir,
				},
			}

			reader := bytes.NewReader([]byte(tt.input))
			path, err := d.spoolBody(reader, tt.maxBytes, tt.kind)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if path == "" {
				t.Error("expected non-empty path")
				return
			}

			// Verify file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("spooled file does not exist: %s", path)
			}

			// Verify file content
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("failed to read spooled file: %v", err)
			}

			expectedLen := int64(len(tt.input))
			if expectedLen > tt.maxBytes {
				expectedLen = tt.maxBytes
			}

			if int64(len(content)) != expectedLen {
				t.Errorf("spooled file size = %d, want %d", len(content), expectedLen)
			}

			// Cleanup
			os.Remove(path)
		})
	}
}

func TestSpoolBody_DirectoryCreation(t *testing.T) {
	tempBase := t.TempDir()
	spoolDir := filepath.Join(tempBase, "nested", "spool", "dir")

	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: spoolDir,
		},
	}

	reader := bytes.NewReader([]byte("test"))
	path, err := d.spoolBody(reader, 100, "req")

	if err != nil {
		t.Errorf("spoolBody failed: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path")
	}

	// Verify directory was created
	if _, err := os.Stat(spoolDir); os.IsNotExist(err) {
		t.Error("spool directory was not created")
	}

	// Cleanup
	if path != "" {
		os.Remove(path)
	}
}

func TestTryCompactJSON(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		expect string
	}{
		{
			name:   "valid JSON with whitespace",
			input:  []byte(`  {"key": "value",  "num": 123  }  `),
			expect: `{"key":"value","num":123}`,
		},
		{
			name:   "already compact JSON",
			input:  []byte(`{"a":"b"}`),
			expect: `{"a":"b"}`,
		},
		{
			name:   "invalid JSON",
			input:  []byte(`{not valid json`),
			expect: "",
		},
		{
			name:   "empty input",
			input:  []byte(``),
			expect: "",
		},
		{
			name:   "JSON array",
			input:  []byte(`[  1,  2,  3  ]`),
			expect: `[1,2,3]`,
		},
		{
			name:   "nested JSON",
			input:  []byte(`{"outer": {"inner": "value"}}`),
			expect: `{"outer":{"inner":"value"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tryCompactJSON(tt.input)

			if result != tt.expect {
				t.Errorf("tryCompactJSON() = %q, want %q", result, tt.expect)
			}
		})
	}
}

func TestAugmentPreviewWithTimings(t *testing.T) {
	tests := []struct {
		name    string
		preview string
		ttfb    int64
		total   int64
		wantKey bool
	}{
		{
			name:    "valid JSON preview",
			preview: `{"status":200,"method":"GET"}`,
			ttfb:    100,
			total:   500,
			wantKey: true,
		},
		{
			name:    "empty JSON object",
			preview: `{}`,
			ttfb:    50,
			total:   200,
			wantKey: true,
		},
		{
			name:    "invalid JSON",
			preview: `not json`,
			ttfb:    100,
			total:   500,
			wantKey: false,
		},
		{
			name:    "zero timings",
			preview: `{"data":"test"}`,
			ttfb:    0,
			total:   0,
			wantKey: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := augmentPreviewWithTimings(tt.preview, tt.ttfb, tt.total)

			if !tt.wantKey {
				// Should return original for invalid JSON
				if result != tt.preview {
					t.Errorf("expected original preview for invalid JSON")
				}
				return
			}

			// Parse result and verify timings key exists
			var m map[string]any
			if err := json.Unmarshal([]byte(result), &m); err != nil {
				t.Fatalf("result is not valid JSON: %v", err)
			}

			timings, ok := m["timings"].(map[string]any)
			if !ok {
				t.Fatal("timings key not found or not a map")
			}

			ttfbMs, ok := timings["ttfbMs"].(float64)
			if !ok {
				t.Fatal("ttfbMs not found or not a number")
			}

			totalMs, ok := timings["totalMs"].(float64)
			if !ok {
				t.Fatal("totalMs not found or not a number")
			}

			if int64(ttfbMs) != tt.ttfb {
				t.Errorf("ttfbMs = %v, want %v", ttfbMs, tt.ttfb)
			}

			if int64(totalMs) != tt.total {
				t.Errorf("totalMs = %v, want %v", totalMs, tt.total)
			}
		})
	}
}

func TestTryDecompress(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		contentEnc  string
		expectValid bool
		expectNil   bool
	}{
		{
			name:        "no encoding",
			body:        []byte("plain text"),
			contentEnc:  "",
			expectValid: false,
			expectNil:   true,
		},
		{
			name:        "unknown encoding",
			body:        []byte("data"),
			contentEnc:  "br",
			expectValid: false,
			expectNil:   true,
		},
		{
			name:        "invalid gzip data",
			body:        []byte("not gzip"),
			contentEnc:  "gzip",
			expectValid: false,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := tryDecompress(tt.body, tt.contentEnc)

			if valid != tt.expectValid {
				t.Errorf("tryDecompress() valid = %v, want %v", valid, tt.expectValid)
			}

			if tt.expectNil && result != nil {
				t.Error("expected nil result for invalid decompression")
			}
		})
	}
}

func TestInt64ToInt(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  int
	}{
		{"zero", 0, 0},
		{"positive small", 100, 100},
		{"negative", -1, 0},
		{"negative large", -1000, 0},
		{"max int32", 2147483647, 2147483647},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := int64ToInt(tt.input)
			if got != tt.want {
				t.Errorf("int64ToInt(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}

	// Test overflow case on 64-bit systems
	t.Run("very large positive", func(t *testing.T) {
		veryLarge := int64(1) << 62
		result := int64ToInt(veryLarge)
		// Should return max int value
		if result < 0 {
			t.Errorf("int64ToInt with very large value returned negative: %d", result)
		}
	})
}
