package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestE2E_ScriptCompilation_Rust tests the full compilation workflow
func TestE2E_ScriptCompilation_Rust(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Check if Rust is available
	if !isRustAvailable() {
		t.Skip("Rust not installed, skipping Rust compilation test")
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

	// 1. Check available compilers
	t.Log("Step 1: Checking available compilers...")
	resp, err := http.Get(baseURL + "/_api/v1/scripts/compilers")
	if err != nil {
		t.Fatalf("failed to check compilers: %v", err)
	}
	defer resp.Body.Close()

	var compilersResp struct {
		Compilers []string        `json:"compilers"`
		All       map[string]bool `json:"all"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&compilersResp); err != nil {
		t.Fatalf("failed to decode compilers response: %v", err)
	}

	if !compilersResp.All["rust"] {
		t.Skip("Rust compiler not available on this system")
	}
	t.Logf("Available compilers: %v", compilersResp.Compilers)

	// 2. Create script with source code
	t.Log("Step 2: Creating script with source code...")
	sourceCode := `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize)]
struct ScriptContext {
    request: Option<HTTPRequest>,
    #[allow(dead_code)]
    session: Option<SessionInfo>,
}

#[derive(Deserialize, Serialize, Clone)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Option<Vec<u8>>,
}

#[derive(Deserialize)]
#[allow(dead_code)]
struct SessionInfo {
    id: String,
    client_addr: String,
}

#[derive(Serialize)]
struct ScriptResult {
    modified: bool,
    #[serde(rename = "modifiedRequest", skip_serializing_if = "Option::is_none")]
    modified_request: Option<HTTPRequest>,
    #[serde(skip_serializing_if = "Option::is_none")]
    logs: Option<Vec<String>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
}

#[plugin_fn]
pub fn process(Json(ctx): Json<ScriptContext>) -> FnResult<Json<ScriptResult>> {
    let mut req = match ctx.request {
        Some(r) => r,
        None => {
            return Ok(Json(ScriptResult {
                modified: false,
                modified_request: None,
                logs: None,
                error: Some("No request in context".to_string()),
            }))
        }
    };

    log!(LogLevel::Info, "E2E Test: Processing {} {}", req.method, req.url);
    req.headers.insert("X-E2E-Test".to_string(), vec!["Compiled".to_string()]);

    Ok(Json(ScriptResult {
        modified: true,
        modified_request: Some(req),
        logs: Some(vec!["E2E test: Added X-E2E-Test header".to_string()]),
        error: None,
    }))
}`

	createReq := map[string]any{
		"name":        "E2E Compilation Test",
		"description": "Test script for compilation E2E",
		"runtime":     "extism",
		"language":    "rust",
		"sourceCode":  sourceCode,
		"triggerType": "request",
		"priority":    10,
		"enabled":     false,
	}

	body, _ := json.Marshal(createReq)
	resp, err = http.Post(baseURL+"/_api/v1/scripts", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var createResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	scriptID := createResp.ID
	t.Logf("Created script with ID: %s", scriptID)
	defer deleteScript(t, baseURL, scriptID)

	// 3. Compile the script
	t.Log("Step 3: Compiling script to WASM...")
	compileReq := map[string]any{
		"optimize": true,
	}
	body, _ = json.Marshal(compileReq)
	resp, err = http.Post(baseURL+"/_api/v1/scripts/"+scriptID+"/compile", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to compile script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("compilation failed with status %d: %v", resp.StatusCode, errResp)
	}

	var compileResp struct {
		Status          string   `json:"status"`
		WASMSize        int      `json:"wasmSize"`
		CompilationTime string   `json:"compilationTime"`
		Logs            []string `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&compileResp); err != nil {
		t.Fatalf("failed to decode compile response: %v", err)
	}

	if compileResp.Status != "success" {
		t.Fatalf("compilation failed: %+v", compileResp)
	}
	t.Logf("Compilation successful! WASM size: %d bytes, time: %s", compileResp.WASMSize, compileResp.CompilationTime)

	// 4. Enable the script
	t.Log("Step 4: Enabling compiled script...")
	toggleReq := map[string]any{
		"enabled": true,
	}
	body, _ = json.Marshal(toggleReq)
	req, _ := http.NewRequest("PATCH", baseURL+"/_api/v1/scripts/"+scriptID+"/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to enable script: %v", err)
	}
	defer resp.Body.Close()

	// 5. Test the compiled script works
	t.Log("Step 5: Testing compiled script execution...")
	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond) // Give script time to load

	proxyURL := fmt.Sprintf("%s/httpproxy/headers?_target=%s", baseURL, echoSrv.Addr)
	resp, err = http.Get(proxyURL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	var headers map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify the compiled script added the header
	if _, ok := headers["X-E2e-Test"]; !ok {
		if _, ok := headers["x-e2e-test"]; !ok {
			t.Errorf("Expected header X-E2E-Test not found in response. Headers: %+v", headers)
		}
	}

	t.Log("✅ E2E compilation test passed!")
}

// TestE2E_ScriptCompilation_MultiFile_Rust tests multi-file project compilation
func TestE2E_ScriptCompilation_MultiFile_Rust(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// Check if Rust is available
	if !isRustAvailable() {
		t.Skip("Rust not installed, skipping Rust multi-file compilation test")
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

	// 1. Check available compilers
	t.Log("Step 1: Checking available compilers...")
	resp, err := http.Get(baseURL + "/_api/v1/scripts/compilers")
	if err != nil {
		t.Fatalf("failed to check compilers: %v", err)
	}
	defer resp.Body.Close()

	var compilersResp struct {
		Compilers []string        `json:"compilers"`
		All       map[string]bool `json:"all"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&compilersResp); err != nil {
		t.Fatalf("failed to decode compilers response: %v", err)
	}

	if !compilersResp.All["rust"] {
		t.Skip("Rust compiler not available on this system")
	}
	t.Logf("Available compilers: %v", compilersResp.Compilers)

	// 2. Create multi-file project with dependencies
	t.Log("Step 2: Creating multi-file Rust project...")

	// Cargo.toml
	cargoToml := `[package]
name = "script"
version = "0.1.0"
edition = "2021"

[dependencies]
extism-pdk = "1.0"
serde = { version = "1.0", features = ["derive"] }

[lib]
crate-type = ["cdylib"]

[profile.release]
opt-level = "z"
lto = true
strip = true`

	// src/lib.rs
	libRs := `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize)]
struct ScriptContext {
    request: Option<HTTPRequest>,
    #[allow(dead_code)]
    session: Option<SessionInfo>,
}

#[derive(Deserialize, Serialize, Clone)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Option<Vec<u8>>,
}

#[derive(Deserialize)]
#[allow(dead_code)]
struct SessionInfo {
    id: String,
    client_addr: String,
}

#[derive(Serialize)]
struct ScriptResult {
    modified: bool,
    #[serde(rename = "modifiedRequest", skip_serializing_if = "Option::is_none")]
    modified_request: Option<HTTPRequest>,
    #[serde(skip_serializing_if = "Option::is_none")]
    logs: Option<Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(ctx): Json<ScriptContext>) -> FnResult<Json<ScriptResult>> {
    let mut req = match ctx.request {
        Some(r) => r,
        None => {
            return Ok(Json(ScriptResult {
                modified: false,
                modified_request: None,
                logs: None,
            }))
        }
    };

    log!(LogLevel::Info, "Multi-file test: Processing {} {}", req.method, req.url);
    req.headers.insert("X-Multi-File-Test".to_string(), vec!["Success".to_string()]);

    Ok(Json(ScriptResult {
        modified: true,
        modified_request: Some(req),
        logs: Some(vec!["Multi-file test: Added X-Multi-File-Test header".to_string()]),
    }))
}`

	dependencies := map[string]string{
		"Cargo.toml": cargoToml,
		"src/lib.rs": libRs,
	}

	createReq := map[string]any{
		"name":         "E2E Multi-File Test",
		"description":  "Test multi-file project compilation",
		"runtime":      "extism",
		"language":     "rust",
		"dependencies": dependencies,
		"triggerType":  "request",
		"priority":     10,
		"enabled":      false,
	}

	body, _ := json.Marshal(createReq)
	resp, err = http.Post(baseURL+"/_api/v1/scripts", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var createResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	scriptID := createResp.ID
	t.Logf("Created multi-file script with ID: %s", scriptID)
	defer deleteScript(t, baseURL, scriptID)

	// 3. Compile the script
	t.Log("Step 3: Compiling multi-file project to WASM...")
	compileReq := map[string]any{
		"optimize": true,
	}
	body, _ = json.Marshal(compileReq)
	resp, err = http.Post(baseURL+"/_api/v1/scripts/"+scriptID+"/compile", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to compile script: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Fatalf("compilation failed with status %d: %v", resp.StatusCode, errResp)
	}

	var compileResp struct {
		Status          string   `json:"status"`
		WASMSize        int      `json:"wasmSize"`
		CompilationTime string   `json:"compilationTime"`
		Logs            []string `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&compileResp); err != nil {
		t.Fatalf("failed to decode compile response: %v", err)
	}

	if compileResp.Status != "success" {
		t.Fatalf("compilation failed: %+v", compileResp)
	}
	t.Logf("Multi-file compilation successful! WASM size: %d bytes, time: %s", compileResp.WASMSize, compileResp.CompilationTime)

	// 4. Enable the script
	t.Log("Step 4: Enabling compiled script...")
	toggleReq := map[string]any{
		"enabled": true,
	}
	body, _ = json.Marshal(toggleReq)
	req, _ := http.NewRequest("PATCH", baseURL+"/_api/v1/scripts/"+scriptID+"/toggle", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to enable script: %v", err)
	}
	defer resp.Body.Close()

	// 5. Test the compiled script works
	t.Log("Step 5: Testing compiled multi-file script execution...")
	echoSrv := startEchoHTTPServer(t)
	time.Sleep(500 * time.Millisecond) // Give script time to load

	proxyURL := fmt.Sprintf("%s/httpproxy/headers?_target=%s", baseURL, echoSrv.Addr)
	resp, err = http.Get(proxyURL)
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	var headers map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&headers); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify the compiled script added the header
	if _, ok := headers["X-Multi-File-Test"]; !ok {
		if _, ok := headers["x-multi-file-test"]; !ok {
			t.Errorf("Expected header X-Multi-File-Test not found in response. Headers: %+v", headers)
		}
	}

	t.Log("✅ E2E multi-file compilation test passed!")
}

// TestE2E_ScriptValidation tests syntax validation without compilation
func TestE2E_ScriptValidation(t *testing.T) {
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

	tests := []struct {
		name      string
		language  string
		code      string
		wantError bool
	}{
		{
			name:      "Valid Rust",
			language:  "rust",
			code:      `pub fn test_func() -> i32 { 42 }`,
			wantError: false,
		},
		{
			name:      "Invalid Rust",
			language:  "rust",
			code:      `fn main( { missing bracket`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip Rust tests if toolchain is not available
			if tt.language == "rust" && !isRustAvailable() {
				t.Skip("Rust toolchain not available (rustc, cargo, or wasm32 target missing)")
			}

			validateReq := map[string]any{
				"language":   tt.language,
				"sourceCode": tt.code,
			}
			body, _ := json.Marshal(validateReq)
			resp, err := http.Post(baseURL+"/_api/v1/scripts/validate", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("failed to validate: %v", err)
			}
			defer resp.Body.Close()

			hasError := resp.StatusCode != http.StatusOK
			if hasError != tt.wantError {
				var result map[string]any
				json.NewDecoder(resp.Body).Decode(&result)
				t.Errorf("wantError=%v, got status=%d, response=%+v", tt.wantError, resp.StatusCode, result)
			}
		})
	}
}

// Helper function to check if Rust is available with full compilation capability
// Checks for rustc, cargo, and wasm32-unknown-unknown target
func isRustAvailable() bool {
	// Check if rustc exists
	if cmd := exec.Command("rustc", "--version"); cmd.Run() != nil {
		return false
	}

	// Check if cargo exists
	if cmd := exec.Command("cargo", "--version"); cmd.Run() != nil {
		return false
	}

	// Check if rustup exists (needed to query installed targets)
	if cmd := exec.Command("rustup", "--version"); cmd.Run() != nil {
		return false
	}

	// Check if wasm32-unknown-unknown target is INSTALLED (not just supported)
	// This matches the server's detection logic in rust.go:52
	cmd := exec.Command("rustup", "target", "list", "--installed")
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(out), "wasm32-unknown-unknown")
}
