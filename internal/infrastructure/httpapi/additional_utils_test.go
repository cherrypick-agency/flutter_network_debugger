package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"net/http"
	"testing"
)

// Additional tests for tryDecompress to reach 100% coverage
func TestTryDecompress_GzipSuccess(t *testing.T) {
	// Create valid gzip data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("hello world"))
	gw.Close()

	result, valid := tryDecompress(buf.Bytes(), "gzip")
	if !valid {
		t.Error("expected valid=true for gzip")
	}
	if string(result) != "hello world" {
		t.Errorf("expected 'hello world', got %q", string(result))
	}
}

func TestTryDecompress_DeflateSuccess(t *testing.T) {
	// Create valid deflate data
	var buf bytes.Buffer
	fw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	fw.Write([]byte("test data"))
	fw.Close()

	result, valid := tryDecompress(buf.Bytes(), "deflate")
	if !valid {
		t.Error("expected valid=true for deflate")
	}
	if string(result) != "test data" {
		t.Errorf("expected 'test data', got %q", string(result))
	}
}

func TestTryDecompress_EmptyInput(t *testing.T) {
	result, valid := tryDecompress([]byte{}, "gzip")
	if valid {
		t.Error("expected valid=false for empty input")
	}
	if result != nil {
		t.Error("expected nil result for invalid gzip")
	}
}

// Additional tests for buildHTTPRequestPreview edge cases
func TestBuildHTTPRequestPreview_LargeBody(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com/api", bytes.NewReader(make([]byte, 10000)))
	req.Header.Set("Content-Type", "application/json")

	preview := buildHTTPRequestPreview(req, []byte("large body truncated"))
	if preview == "" {
		t.Error("expected non-empty preview")
	}
}

func TestBuildHTTPRequestPreview_BinaryContentType(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com/upload", bytes.NewReader([]byte{0x00, 0x01, 0x02}))
	req.Header.Set("Content-Type", "application/octet-stream")

	preview := buildHTTPRequestPreview(req, []byte{0x00, 0x01, 0x02})
	if preview == "" {
		t.Error("expected non-empty preview for binary content")
	}
}

func TestBuildHTTPRequestPreview_FormURLEncoded(t *testing.T) {
	req, _ := http.NewRequest("POST", "http://example.com/form", bytes.NewReader([]byte("key=value&foo=bar")))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	preview := buildHTTPRequestPreview(req, []byte("key=value&foo=bar"))
	if preview == "" {
		t.Error("expected non-empty preview")
	}
}

// Additional tests for buildHTTPResponsePreview edge cases
func TestBuildHTTPResponsePreview_JSONResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       nil,
	}
	preview := buildHTTPResponsePreview(resp)
	if preview == "" {
		t.Error("expected non-empty preview")
	}
}

func TestBuildHTTPResponsePreview_HTMLResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       nil,
	}
	preview := buildHTTPResponsePreview(resp)
	if preview == "" {
		t.Error("expected non-empty preview")
	}
}

func TestBuildHTTPResponsePreview_ImageResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"image/png"}},
		Body:       nil,
	}
	preview := buildHTTPResponsePreview(resp)
	if preview == "" {
		t.Error("expected non-empty preview for image")
	}
}

func TestBuildHTTPResponsePreview_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: 204,
		Header:     http.Header{},
		Body:       nil,
	}
	preview := buildHTTPResponsePreview(resp)
	if preview == "" {
		t.Error("expected non-empty preview even for empty body")
	}
}

func TestBuildHTTPResponsePreview_ErrorStatus(t *testing.T) {
	resp := &http.Response{
		StatusCode: 500,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       nil,
	}
	preview := buildHTTPResponsePreview(resp)
	if preview == "" {
		t.Error("expected non-empty preview for error status")
	}
}

// Additional edge case tests
func TestAugmentPreviewWithTimings_LargeTimings(t *testing.T) {
	preview := `{"method":"GET"}`
	result := augmentPreviewWithTimings(preview, 999999, 1000000)
	if result == preview {
		t.Error("expected preview to be augmented with timings")
	}
}

func TestTryCompactJSON_NestedObjects(t *testing.T) {
	input := []byte(`  {  "a": { "b": { "c": "d" } }  }  `)
	result := tryCompactJSON(input)
	expected := `{"a":{"b":{"c":"d"}}}`
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTryCompactJSON_Array(t *testing.T) {
	input := []byte(`  [ 1 , 2 , 3 ]  `)
	result := tryCompactJSON(input)
	expected := `[1,2,3]`
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestTryCompactJSON_ComplexJSON(t *testing.T) {
	input := []byte(`{
		"users": [
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"}
		],
		"total": 2
	}`)
	result := tryCompactJSON(input)
	if result == "" {
		t.Error("expected non-empty compact JSON")
	}
	// Should not contain newlines or extra spaces
	if bytes.Contains([]byte(result), []byte("\n")) || bytes.Contains([]byte(result), []byte("  ")) {
		t.Error("expected compacted JSON without newlines or multiple spaces")
	}
}
