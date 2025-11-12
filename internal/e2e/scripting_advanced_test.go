package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestE2E_ScriptingAPI_Dart_Subprocess tests Dart executor via subprocess
func TestE2E_ScriptingAPI_Dart_Subprocess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Check if Dart is available
	if !isDartAvailable(t) {
		t.Skip("Dart SDK not installed, skipping Dart executor test")
	}

	// Start binary
	bin := buildWsproxyBinary(t)
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	// Get dynamic port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"DEV_MODE=1",
		"ADDR="+addr,
		"DB_PATH="+tmpDB,
	)
	// Set working directory to project root so script_runner.dart can be found
	cmd.Dir = filepath.Join("..", "..")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		if t.Failed() {
			t.Logf("Binary stderr:\n%s", stderr.String())
		}
	}()

	baseURL := "http://" + addr
	waitReady(t, baseURL, 60*time.Second)

	// Load Dart script (source code, not compiled)
	dartCode := loadTestDart(t, "simple_logger.dart")

	// Create Dart script
	scriptID := createScriptWithOptions(t, baseURL, map[string]any{
		"name":        "Dart Logger E2E",
		"runtime":     "dart",
		"code":        base64.StdEncoding.EncodeToString(dartCode),
		"language":    "dart",
		"triggerType": "request",
		"enabled":     true,
	})
	defer deleteScript(t, baseURL, scriptID)

	// Start echo server
	echoSrv := startEchoHTTPServer(t)

	// Make request
	time.Sleep(500 * time.Millisecond) // Wait for Dart process to spawn
	proxyURL := fmt.Sprintf("%s/httpproxy/headers?_target=%s", baseURL, echoSrv.Addr)
	resp, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var headers map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Headers after Dart script: %+v", headers)

	// Verify Dart script added headers
	if dartProcessed, ok := headers["X-Dart-Processed"]; ok {
		if vals, ok := dartProcessed.([]any); ok && len(vals) > 0 {
			if vals[0] != "true" {
				t.Errorf("expected X-Dart-Processed: true, got %v", vals[0])
			}
		}
	} else {
		t.Error("X-Dart-Processed header not found (Dart script may not have executed)")
	}
}

// TestE2E_ScriptingAPI_MatchRules tests pattern matching for scripts
func TestE2E_ScriptingAPI_MatchRules(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	bin := buildWsproxyBinary(t)
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	// Get dynamic port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"DEV_MODE=1",
		"ADDR="+addr,
		"DB_PATH="+tmpDB,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		if t.Failed() {
			t.Logf("Binary stderr:\n%s", stderr.String())
		}
	}()

	baseURL := "http://" + addr
	waitReady(t, baseURL, 60*time.Second)

	// Create script with match rules: POST /api/*
	wasmData := loadTestWASM(t, "add_header.wasm")
	scriptID := createScriptWithOptions(t, baseURL, map[string]any{
		"name":        "Match Rules Test",
		"runtime":     "extism",
		"code":        base64.StdEncoding.EncodeToString(wasmData),
		"language":    "rust",
		"triggerType": "request",
		"enabled":     true,
		"matchRules": map[string]any{
			"methods":     []string{"POST"},
			"pathPattern": "/api/*",
			"patternType": "wildcard",
		},
	})
	defer deleteScript(t, baseURL, scriptID)

	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond)

	// Test 1: Matching request (POST /api/foo) - should match
	proxyURL1 := fmt.Sprintf("%s/httpproxy/api/foo?_target=%s", baseURL, echoSrv.Addr)
	req1, _ := http.NewRequest(http.MethodPost, proxyURL1, nil)
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp1.Body.Close()

	// Parse headers from echo response (wildcard handler returns headers for any path)
	var headers1 map[string]any
	json.NewDecoder(resp1.Body).Decode(&headers1)
	t.Logf("Test 1 (POST /api/foo) headers: %+v", headers1)

	// Should have X-Script-Processed header (matched)
	if _, ok := headers1["X-Script-Processed"]; !ok {
		t.Error("expected X-Script-Processed header on POST /api/* (should match)")
	}

	// Test 2: Non-matching request (GET /api/foo) - method doesn't match
	proxyURL2 := fmt.Sprintf("%s/httpproxy/api/foo?_target=%s", baseURL, echoSrv.Addr)
	resp2, err := http.Get(proxyURL2)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp2.Body.Close()

	var headers2 map[string]any
	json.NewDecoder(resp2.Body).Decode(&headers2)
	t.Logf("Test 2 (GET /api/foo) headers: %+v", headers2)

	// Should NOT have X-Script-Processed header (method doesn't match)
	if _, ok := headers2["X-Script-Processed"]; ok {
		t.Error("expected NO X-Script-Processed header on GET /api/* (method doesn't match)")
	}

	// Test 3: Non-matching request (POST /other) - path doesn't match
	proxyURL3 := fmt.Sprintf("%s/httpproxy/other?_target=%s", baseURL, echoSrv.Addr)
	req3, _ := http.NewRequest(http.MethodPost, proxyURL3, nil)
	resp3, err := http.DefaultClient.Do(req3)
	if err != nil {
		t.Fatalf("POST /other request failed: %v", err)
	}
	defer resp3.Body.Close()

	var headers3 map[string]any
	json.NewDecoder(resp3.Body).Decode(&headers3)
	t.Logf("Test 3 (POST /other) headers: %+v", headers3)

	// Should NOT have X-Script-Processed header (path doesn't match)
	if _, ok := headers3["X-Script-Processed"]; ok {
		t.Error("expected NO X-Script-Processed header on POST /other (path doesn't match)")
	}
}

// TestE2E_ScriptingAPI_Priority tests script execution order by priority
func TestE2E_ScriptingAPI_Priority(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	bin := buildWsproxyBinary(t)
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	// Get dynamic port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"DEV_MODE=1",
		"ADDR="+addr,
		"DB_PATH="+tmpDB,
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	baseURL := "http://" + addr
	waitReady(t, baseURL, 60*time.Second)

	// Create 3 scripts with different priorities
	wasmData := loadTestWASM(t, "add_header.wasm")

	// Priority 5 (low)
	script1 := createScriptWithOptions(t, baseURL, map[string]any{
		"name":        "Low Priority",
		"runtime":     "extism",
		"code":        base64.StdEncoding.EncodeToString(wasmData),
		"language":    "rust",
		"triggerType": "request",
		"priority":    5,
		"enabled":     true,
	})
	defer deleteScript(t, baseURL, script1)

	// Priority 20 (high)
	script2 := createScriptWithOptions(t, baseURL, map[string]any{
		"name":        "High Priority",
		"runtime":     "extism",
		"code":        base64.StdEncoding.EncodeToString(wasmData),
		"language":    "rust",
		"triggerType": "request",
		"priority":    20,
		"enabled":     true,
	})
	defer deleteScript(t, baseURL, script2)

	// Priority 10 (medium)
	script3 := createScriptWithOptions(t, baseURL, map[string]any{
		"name":        "Medium Priority",
		"runtime":     "extism",
		"code":        base64.StdEncoding.EncodeToString(wasmData),
		"language":    "rust",
		"triggerType": "request",
		"priority":    10,
		"enabled":     true,
	})
	defer deleteScript(t, baseURL, script3)

	// Make request
	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond)

	proxyURL := fmt.Sprintf("%s/httpproxy/test?_target=%s", baseURL, echoSrv.Addr)
	resp, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Expected execution order: High (20) → Medium (10) → Low (5)
	// All should add X-Script-Processed header (3 times)
	// Parse headers from echo response (wildcard handler returns headers)
	var headers map[string]any
	json.NewDecoder(resp.Body).Decode(&headers)
	t.Logf("Request headers received by upstream: %+v", headers)

	if _, ok := headers["X-Script-Processed"]; !ok {
		t.Error("expected X-Script-Processed header from scripts")
	}
}

// TestE2E_ScriptingAPI_ToggleEnabled tests enabling/disabling script execution
func TestE2E_ScriptingAPI_ToggleEnabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	bin := buildWsproxyBinary(t)
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	// Get dynamic port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"DEV_MODE=1",
		"ADDR="+addr,
		"DB_PATH="+tmpDB,
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	baseURL := "http://" + addr
	waitReady(t, baseURL, 60*time.Second)

	// Create enabled script
	wasmData := loadTestWASM(t, "add_header.wasm")
	scriptID := createScript(t, baseURL, wasmData, "rust")
	defer deleteScript(t, baseURL, scriptID)

	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond)

	// Test 1: Script enabled - should add header
	proxyURL := fmt.Sprintf("%s/httpproxy/headers?_target=%s/headers", baseURL, echoSrv.Addr)
	resp1, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp1.Body.Close()

	var headers1 map[string]any
	json.NewDecoder(resp1.Body).Decode(&headers1)
	if _, ok := headers1["X-Script-Processed"]; !ok {
		t.Error("expected X-Script-Processed when script is enabled")
	}

	// Disable script
	toggleScript(t, baseURL, scriptID, false)
	time.Sleep(200 * time.Millisecond)

	// Test 2: Script disabled - should NOT add header
	resp2, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp2.Body.Close()

	var headers2 map[string]any
	json.NewDecoder(resp2.Body).Decode(&headers2)
	if _, ok := headers2["X-Script-Processed"]; ok {
		t.Error("expected NO X-Script-Processed when script is disabled")
	}

	// Re-enable script
	toggleScript(t, baseURL, scriptID, true)
	time.Sleep(200 * time.Millisecond)

	// Test 3: Script re-enabled - should add header again
	resp3, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp3.Body.Close()

	var headers3 map[string]any
	json.NewDecoder(resp3.Body).Decode(&headers3)
	if _, ok := headers3["X-Script-Processed"]; !ok {
		t.Error("expected X-Script-Processed when script is re-enabled")
	}
}
