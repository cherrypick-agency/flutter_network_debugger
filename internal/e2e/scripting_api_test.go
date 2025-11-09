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
	"time"
)

// TestE2E_ScriptingAPI_CRUD tests CRUD operations for scripts
func TestE2E_ScriptingAPI_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// 1. Build and start binary
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
	waitReady(t, baseURL, 10*time.Second)

	// 2. Load WASM fixture
	wasmData := loadTestWASM(t, "add_header.wasm")
	if len(wasmData) == 0 {
		t.Fatal("WASM fixture is empty")
	}

	// 3. Create script
	scriptID := createScript(t, baseURL, wasmData, "rust")
	t.Logf("Created script: %s", scriptID)

	// 4. Verify GET /_api/v1/scripts/{id}
	script := getScript(t, baseURL, scriptID)
	if script.Language != "rust" {
		t.Errorf("expected language=rust, got %s", script.Language)
	}
	if script.TriggerType != "request" {
		t.Errorf("expected triggerType=request, got %s", script.TriggerType)
	}
	if !script.Enabled {
		t.Error("expected script to be enabled")
	}

	// 5. List scripts
	scripts := listScripts(t, baseURL)
	if len(scripts) != 1 {
		t.Errorf("expected 1 script, got %d", len(scripts))
	}

	// 6. Update script
	updateScript(t, baseURL, scriptID, map[string]any{
		"description": "Updated via E2E test",
		"priority":    20,
	})

	updatedScript := getScript(t, baseURL, scriptID)
	if updatedScript.Description != "Updated via E2E test" {
		t.Errorf("description not updated: %s", updatedScript.Description)
	}
	if updatedScript.Priority != 20 {
		t.Errorf("priority not updated: %d", updatedScript.Priority)
	}

	// 7. Toggle enabled
	toggleScript(t, baseURL, scriptID, false)
	disabledScript := getScript(t, baseURL, scriptID)
	if disabledScript.Enabled {
		t.Error("expected script to be disabled")
	}

	toggleScript(t, baseURL, scriptID, true)
	enabledScript := getScript(t, baseURL, scriptID)
	if !enabledScript.Enabled {
		t.Error("expected script to be enabled")
	}

	// 8. Delete script
	deleteScript(t, baseURL, scriptID)

	// 9. Verify deletion (404)
	assertScriptNotFound(t, baseURL, scriptID)

	// Verify list is empty
	scriptsAfterDelete := listScripts(t, baseURL)
	if len(scriptsAfterDelete) != 0 {
		t.Errorf("expected 0 scripts after deletion, got %d", len(scriptsAfterDelete))
	}
}

// TestE2E_ScriptingAPI_RequestModification tests script execution on requests
func TestE2E_ScriptingAPI_RequestModification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
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

	// Capture output for debugging
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
			t.Logf("Binary stdout:\n%s", stdout.String())
			t.Logf("Binary stderr:\n%s", stderr.String())
		}
	}()

	baseURL := "http://" + addr
	waitReady(t, baseURL, 10*time.Second)

	// Start upstream echo server
	echoSrv := startEchoHTTPServer(t)
	upstreamURL := echoSrv.Addr

	// Create Rust "add_header" script
	wasmData := loadTestWASM(t, "add_header.wasm")
	scriptID := createScript(t, baseURL, wasmData, "rust")
	defer deleteScript(t, baseURL, scriptID)

	// Verify script was created and can be retrieved
	script := getScript(t, baseURL, scriptID)
	t.Logf("Script created: ID=%s, Name=%s, Enabled=%v, TriggerType=%s",
		script.ID, script.Name, script.Enabled, script.TriggerType)

	// List all scripts to verify it's in the database
	allScripts := listScripts(t, baseURL)
	t.Logf("Total scripts in database: %d", len(allScripts))

	// Wait a bit for script to be loaded
	time.Sleep(500 * time.Millisecond)

	// Make HTTP request through proxy
	proxyURL := fmt.Sprintf("%s/httpproxy/headers?_target=%s", baseURL, upstreamURL)
	resp, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse response (echo server returns headers as JSON)
	var headers map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify script added headers
	t.Logf("Response headers: %+v", headers)

	// Check for X-Script-Processed header
	if scriptProcessed, ok := headers["X-Script-Processed"]; ok {
		if vals, ok := scriptProcessed.([]any); ok && len(vals) > 0 {
			if vals[0] != "Rust" {
				t.Errorf("expected X-Script-Processed: Rust, got %v", vals[0])
			}
		}
	} else {
		t.Error("X-Script-Processed header not found (script may not have executed)")
	}

	// Check for X-Test-Header
	if testHeader, ok := headers["X-Test-Header"]; ok {
		if vals, ok := testHeader.([]any); ok && len(vals) > 0 {
			if vals[0] != "E2E-Test" {
				t.Errorf("expected X-Test-Header: E2E-Test, got %v", vals[0])
			}
		}
	} else {
		t.Error("X-Test-Header not found")
	}
}

// TestE2E_ScriptingAPI_ValidationError tests invalid WASM rejection
func TestE2E_ScriptingAPI_ValidationError(t *testing.T) {
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
	waitReady(t, baseURL, 10*time.Second)

	// Try to create script with invalid WASM
	invalidWASM := loadTestWASM(t, "invalid.wasm")

	// Try to create script - should fail with 400
	payload, _ := json.Marshal(map[string]any{
		"name":        "Invalid Script",
		"runtime":     "extism",
		"code":        base64.StdEncoding.EncodeToString(invalidWASM),
		"language":    "rust",
		"triggerType": "request",
	})

	req, _ := http.NewRequest(http.MethodPost, baseURL+"/_api/v1/scripts", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Should return 400 Bad Request for invalid WASM
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid WASM, got %d", resp.StatusCode)
	}

	t.Logf("Correctly rejected invalid WASM with status %d", resp.StatusCode)
}

// TestE2E_ScriptingAPI_Noop tests minimal WASM module (passthrough)
func TestE2E_ScriptingAPI_Noop(t *testing.T) {
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
	waitReady(t, baseURL, 10*time.Second)

	// Create noop script (returns input unchanged)
	wasmData := loadTestWASM(t, "noop.wasm")
	scriptID := createScript(t, baseURL, wasmData, "rust")
	defer deleteScript(t, baseURL, scriptID)

	// Start echo server
	echoSrv := startEchoHTTPServer(t)

	// Make request - should pass through unchanged
	proxyURL := fmt.Sprintf("%s/httpproxy/api/test?_target=%s", baseURL, echoSrv.Addr)
	resp, err := http.Get(proxyURL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read raw body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	t.Logf("Response body: %s", string(bodyBytes))

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("expected success: true")
	}
}

// TestE2E_ScriptingAPI_ResponseModification tests response modification scripts
func TestE2E_ScriptingAPI_ResponseModification(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Check if Rust is available
	if !isRustAvailable() {
		t.Skip("Rust not installed, skipping response modification test")
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
	waitReady(t, baseURL, 10*time.Second)

	// Rust script that modifies response
	sourceCode := `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize)]
struct ScriptContext {
    response: Option<HTTPResponse>,
}

#[derive(Deserialize, Serialize, Clone)]
struct HTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
}

#[derive(Serialize)]
struct ScriptResult {
    modified: bool,
    #[serde(rename = "modifiedResponse", skip_serializing_if = "Option::is_none")]
    modified_response: Option<HTTPResponse>,
}

#[plugin_fn]
pub fn process(Json(ctx): Json<ScriptContext>) -> FnResult<Json<ScriptResult>> {
    let mut resp = match ctx.response {
        Some(r) => r,
        None => {
            return Ok(Json(ScriptResult {
                modified: false,
                modified_response: None,
            }))
        }
    };

    log!(LogLevel::Info, "Response modification: status {}", resp.status);
    resp.headers.insert("X-Response-Modified".to_string(), vec!["true".to_string()]);
    resp.headers.insert("X-Test-Response".to_string(), vec!["E2E".to_string()]);

    Ok(Json(ScriptResult {
        modified: true,
        modified_response: Some(resp),
    }))
}`

	// Create script with sourceCode (will auto-compile)
	createReq := map[string]any{
		"name":        "E2E Response Test",
		"description": "Test response modification",
		"runtime":     "extism",
		"language":    "rust",
		"sourceCode":  sourceCode,
		"triggerType": "response", // ← Response trigger
		"priority":    10,
		"enabled":     true,
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(baseURL+"/_api/v1/scripts", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status creating script: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var createResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	scriptID := createResp.ID
	t.Logf("Created response script: ID=%s", scriptID)
	defer deleteScript(t, baseURL, scriptID)

	// Wait for compilation to complete (response scripts are auto-compiled)
	time.Sleep(12 * time.Second)

	// Test: proxy a request and check response was modified
	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond)

	proxyURL := fmt.Sprintf("%s/httpproxy/echo?_target=%s", baseURL, echoSrv.Addr)
	resp, err = http.Post(proxyURL, "application/json", bytes.NewReader([]byte(`{"test":true}`)))
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response headers were modified
	if resp.Header.Get("X-Response-Modified") != "true" {
		t.Errorf("Expected X-Response-Modified: true, got: %s", resp.Header.Get("X-Response-Modified"))
		t.Logf("All response headers: %+v", resp.Header)
	}

	if resp.Header.Get("X-Test-Response") != "E2E" {
		t.Errorf("Expected X-Test-Response: E2E, got: %s", resp.Header.Get("X-Test-Response"))
	}

	t.Log("✅ Response modification test passed!")
}

// TestE2E_ScriptingAPI_BothTriggers tests scripts with triggerType: "both"
// that can see and modify both request and response
func TestE2E_ScriptingAPI_BothTriggers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Check if Rust is available
	if !isRustAvailable() {
		t.Skip("Rust not installed, skipping both triggers test")
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
	waitReady(t, baseURL, 10*time.Second)

	// Rust script that sees both request and response
	sourceCode := `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize)]
struct ScriptContext {
    request: Option<HTTPRequest>,
    response: Option<HTTPResponse>,
}

#[derive(Deserialize, Serialize, Clone)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
}

#[derive(Deserialize, Serialize, Clone)]
struct HTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
}

#[derive(Serialize)]
struct ScriptResult {
    modified: bool,
    #[serde(rename = "modifiedRequest", skip_serializing_if = "Option::is_none")]
    modified_request: Option<HTTPRequest>,
    #[serde(rename = "modifiedResponse", skip_serializing_if = "Option::is_none")]
    modified_response: Option<HTTPResponse>,
}

#[plugin_fn]
pub fn process(Json(ctx): Json<ScriptContext>) -> FnResult<Json<ScriptResult>> {
    let mut modified_req = None;
    let mut modified_resp = None;
    let mut was_modified = false;

    // Modify request if present
    if let Some(mut req) = ctx.request {
        log!(LogLevel::Info, "Both triggers: processing request {} {}", req.method, req.url);
        req.headers.insert("X-Request-Modified".to_string(), vec!["true".to_string()]);
        req.headers.insert("X-Both-Triggers".to_string(), vec!["request-phase".to_string()]);
        modified_req = Some(req);
        was_modified = true;
    }

    // Modify response if present
    if let Some(mut resp) = ctx.response {
        log!(LogLevel::Info, "Both triggers: processing response status {}", resp.status);
        resp.headers.insert("X-Response-Modified".to_string(), vec!["true".to_string()]);
        resp.headers.insert("X-Both-Triggers".to_string(), vec!["response-phase".to_string()]);
        modified_resp = Some(resp);
        was_modified = true;
    }

    Ok(Json(ScriptResult {
        modified: was_modified,
        modified_request: modified_req,
        modified_response: modified_resp,
    }))
}`

	// Create script with triggerType: "both"
	createReq := map[string]any{
		"name":        "E2E Both Triggers Test",
		"description": "Test script with both triggers",
		"runtime":     "extism",
		"language":    "rust",
		"sourceCode":  sourceCode,
		"triggerType": "both", // ← Both request and response
		"priority":    10,
		"enabled":     true,
	}

	body, _ := json.Marshal(createReq)
	resp, err := http.Post(baseURL+"/_api/v1/scripts", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status creating script: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var createResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	scriptID := createResp.ID
	t.Logf("Created 'both' triggers script: ID=%s", scriptID)
	defer deleteScript(t, baseURL, scriptID)

	// Wait for compilation to complete
	time.Sleep(12 * time.Second)

	// Test: proxy a request and check both request and response were modified
	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond)

	proxyURL := fmt.Sprintf("%s/httpproxy/echo?_target=%s", baseURL, echoSrv.Addr)
	resp, err = http.Post(proxyURL, "application/json", bytes.NewReader([]byte(`{"test":"both"}`)))
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response to check headers
	responseBody, _ := io.ReadAll(resp.Body)
	t.Logf("Response body: %s", string(responseBody))

	// The request should have been modified (check via echo response which echoes request headers)
	var echoResp map[string]any
	if err := json.Unmarshal(responseBody, &echoResp); err != nil {
		t.Fatalf("failed to parse echo response: %v", err)
	}

	// Check request headers (echoed by server)
	if headers, ok := echoResp["headers"].(map[string]any); ok {
		t.Logf("Echoed request headers: %+v", headers)
		if reqModified, ok := headers["X-Request-Modified"]; ok {
			if vals, ok := reqModified.([]any); ok && len(vals) > 0 {
				if vals[0] != "true" {
					t.Errorf("Expected X-Request-Modified: true in request, got: %v", vals[0])
				}
			}
		} else {
			t.Errorf("X-Request-Modified header not found in echoed request headers")
		}
	}

	// Check response headers
	if resp.Header.Get("X-Response-Modified") != "true" {
		t.Errorf("Expected X-Response-Modified: true in response, got: %s", resp.Header.Get("X-Response-Modified"))
		t.Logf("All response headers: %+v", resp.Header)
	}

	if resp.Header.Get("X-Both-Triggers") != "response-phase" {
		t.Errorf("Expected X-Both-Triggers: response-phase, got: %s", resp.Header.Get("X-Both-Triggers"))
	}

	t.Log("✅ Both triggers test passed!")
}
