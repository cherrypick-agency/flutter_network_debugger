package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Script represents a script entity from the API
type Script struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Runtime     string        `json:"runtime"`
	Code        string        `json:"code"` // base64 encoded
	Language    string        `json:"language"`
	TriggerType string        `json:"triggerType"`
	Priority    int           `json:"priority"`
	Enabled     bool          `json:"enabled"`
	MatchRules  *MatchRules   `json:"matchRules,omitempty"`
	Config      *ScriptConfig `json:"config,omitempty"`
	CreatedAt   string        `json:"createdAt,omitempty"`
	UpdatedAt   string        `json:"updatedAt,omitempty"`
}

// MatchRules defines pattern matching for scripts
type MatchRules struct {
	Methods     []string `json:"methods,omitempty"`
	HostPattern string   `json:"hostPattern,omitempty"`
	PathPattern string   `json:"pathPattern,omitempty"`
	PatternType string   `json:"patternType,omitempty"` // "wildcard" or "regex"
}

// ScriptConfig defines execution limits
type ScriptConfig struct {
	TimeoutMs     int      `json:"timeoutMs,omitempty"`
	MemoryLimitMB int      `json:"memoryLimitMB,omitempty"`
	AllowedHosts  []string `json:"allowedHosts,omitempty"`
}

// loadTestWASM loads a WASM fixture from testdata
func loadTestWASM(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "scripts", "wasm", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load WASM fixture %s: %v", filename, err)
	}
	return data
}

// loadTestDart loads a Dart script from testdata
func loadTestDart(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "scripts", "dart", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load Dart fixture %s: %v", filename, err)
	}
	return data
}

// createScript creates a script via POST /_api/v1/scripts
func createScript(t *testing.T, baseURL string, wasmData []byte, language string) string {
	t.Helper()

	payload := map[string]any{
		"name":        "E2E Test Script - " + language,
		"description": "Automatically created E2E test",
		"runtime":     "extism",
		"code":        base64.StdEncoding.EncodeToString(wasmData),
		"language":    language,
		"triggerType": "request",
		"enabled":     true,
		"priority":    10,
	}

	script := httpPostJSON(t, baseURL+"/_api/v1/scripts", payload)
	var result Script
	if err := json.Unmarshal([]byte(script), &result); err != nil {
		t.Fatalf("failed to parse created script: %v", err)
	}

	if result.ID == "" {
		t.Fatal("created script has empty ID")
	}

	return result.ID
}

// createScriptWithOptions creates a script with custom options
func createScriptWithOptions(t *testing.T, baseURL string, opts map[string]any) string {
	t.Helper()

	// Set defaults
	if _, ok := opts["enabled"]; !ok {
		opts["enabled"] = true
	}
	if _, ok := opts["priority"]; !ok {
		opts["priority"] = 10
	}

	script := httpPostJSON(t, baseURL+"/_api/v1/scripts", opts)
	var result Script
	if err := json.Unmarshal([]byte(script), &result); err != nil {
		t.Fatalf("failed to parse created script: %v", err)
	}

	if result.ID == "" {
		t.Fatal("created script has empty ID")
	}

	return result.ID
}

// getScript retrieves a script via GET /_api/v1/scripts/{id}
func getScript(t *testing.T, baseURL string, id string) Script {
	t.Helper()

	resp, err := http.Get(baseURL + "/_api/v1/scripts/" + id)
	if err != nil {
		t.Fatalf("GET /scripts/%s failed: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /scripts/%s returned %d: %s", id, resp.StatusCode, body)
	}

	var script Script
	if err := json.NewDecoder(resp.Body).Decode(&script); err != nil {
		t.Fatalf("failed to decode script: %v", err)
	}

	return script
}

// listScripts retrieves all scripts via GET /_api/v1/scripts
func listScripts(t *testing.T, baseURL string) []Script {
	t.Helper()

	resp, err := http.Get(baseURL + "/_api/v1/scripts")
	if err != nil {
		t.Fatalf("GET /scripts failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("GET /scripts returned %d: %s", resp.StatusCode, body)
	}

	var scripts []Script
	if err := json.NewDecoder(resp.Body).Decode(&scripts); err != nil {
		t.Fatalf("failed to decode scripts list: %v", err)
	}

	return scripts
}

// updateScript updates a script via PUT /_api/v1/scripts/{id}
func updateScript(t *testing.T, baseURL string, id string, updates map[string]any) {
	t.Helper()

	payload, _ := json.Marshal(updates)
	req, _ := http.NewRequest(http.MethodPut, baseURL+"/_api/v1/scripts/"+id, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT /scripts/%s failed: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("PUT /scripts/%s returned %d: %s", id, resp.StatusCode, body)
	}
}

// toggleScript enables/disables a script via PATCH /_api/v1/scripts/{id}/toggle
func toggleScript(t *testing.T, baseURL string, id string, enabled bool) {
	t.Helper()

	payload, _ := json.Marshal(map[string]bool{"enabled": enabled})
	req, _ := http.NewRequest(http.MethodPatch, baseURL+"/_api/v1/scripts/"+id+"/toggle", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH /scripts/%s/toggle failed: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("PATCH /scripts/%s/toggle returned %d: %s", id, resp.StatusCode, body)
	}
}

// deleteScript deletes a script via DELETE /_api/v1/scripts/{id}
func deleteScript(t *testing.T, baseURL string, id string) {
	t.Helper()

	req, _ := http.NewRequest(http.MethodDelete, baseURL+"/_api/v1/scripts/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /scripts/%s failed: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("DELETE /scripts/%s returned %d: %s", id, resp.StatusCode, body)
	}
}

// assertScriptNotFound verifies that getting a script returns 404
func assertScriptNotFound(t *testing.T, baseURL string, id string) {
	t.Helper()

	resp, err := http.Get(baseURL + "/_api/v1/scripts/" + id)
	if err != nil {
		t.Fatalf("GET /scripts/%s failed: %v", id, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for deleted script, got %d", resp.StatusCode)
	}
}

// httpPostJSON performs POST request with JSON payload and returns response body
func httpPostJSON(t *testing.T, url string, payload any) string {
	t.Helper()

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s returned %d: %s", url, resp.StatusCode, responseBody)
	}

	return string(responseBody)
}

// startEchoHTTPServer starts a simple echo HTTP server for testing
func startEchoHTTPServer(t *testing.T) *http.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Echo headers as JSON
	mux.HandleFunc("/headers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(r.Header)
	})

	// Echo request details
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   r.URL.Query(),
			"headers": r.Header,
			"body":    string(body),
		})
	})

	// Simple success response
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
			"message": "upstream response",
		})
	})

	// Wildcard handler - returns headers for any path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return headers for any unmatched path
		json.NewEncoder(w).Encode(r.Header)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	srv := &http.Server{Handler: mux}
	go func() {
		_ = srv.Serve(listener)
	}()

	t.Cleanup(func() {
		_ = srv.Close()
	})

	// Store server URL for easy access
	addr := listener.Addr().String()
	srv.Addr = "http://" + addr

	return srv
}

// isDartAvailable checks if Dart SDK is installed
func isDartAvailable(t *testing.T) bool {
	t.Helper()
	cmd := exec.Command("dart", "--version")
	return cmd.Run() == nil
}

// makeProxyRequest makes HTTP request through the proxy's /httpproxy endpoint
func makeProxyRequest(t *testing.T, proxyBaseURL string, targetURL string, method string, body []byte) *http.Response {
	t.Helper()

	// Use /httpproxy endpoint with _target query param
	proxyURL := fmt.Sprintf("%s/httpproxy%s?_target=%s", proxyBaseURL, "/test", targetURL)

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, _ := http.NewRequest(method, proxyURL, reqBody)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}

	return resp
}
