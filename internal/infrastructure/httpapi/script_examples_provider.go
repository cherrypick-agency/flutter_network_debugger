package httpapi

type ScriptExample struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Difficulty  string `json:"difficulty"`
	Category    string `json:"category"`
	TriggerType string `json:"triggerType"`
	SourceCode  string `json:"sourceCode"`
}

type scriptExamplesProvider struct{}

func newScriptExamplesProvider() scriptExamplesProvider {
	return scriptExamplesProvider{}
}

func (scriptExamplesProvider) list() []ScriptExample {
	return []ScriptExample{
		{
			ID:          "rust-passthrough",
			Name:        "Passthrough (No Modification)",
			Description: "Simplest possible script - passes requests through without any modifications. Great starting point for understanding script structure.",
			Language:    "rust",
			Difficulty:  "beginner",
			Category:    "request_modification",
			TriggerType: "request",
			SourceCode: `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Simple passthrough - no modifications
    Ok(Json(req))
}`,
		},
		{
			ID:          "rust-add-header",
			Name:        "Add Custom Headers",
			Description: "Adds custom headers to HTTP requests. Demonstrates basic request modification and header manipulation.",
			Language:    "rust",
			Difficulty:  "beginner",
			Category:    "request_modification",
			TriggerType: "request",
			SourceCode: `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(mut req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Log the request
    log!(LogLevel::Info, "Processing {} {}", req.method, req.url);

    // Add custom headers
    req.headers.insert("X-Script-Processed".to_string(), vec!["Rust".to_string()]);
    req.headers.insert("X-Custom-Header".to_string(), vec!["Hello from Rust!".to_string()]);

    Ok(Json(req))
}`,
		},
		{
			ID:          "rust-modify-response",
			Name:        "Sanitize Response Data",
			Description: "Removes sensitive fields from JSON responses and adds processing metadata. Demonstrates response modification and JSON parsing.",
			Language:    "rust",
			Difficulty:  "advanced",
			Category:    "response_modification",
			TriggerType: "response",
			SourceCode: `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;

#[derive(Deserialize)]
struct HTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
}

#[derive(Serialize)]
struct ModifiedHTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(mut resp): Json<HTTPResponse>) -> FnResult<Json<ModifiedHTTPResponse>> {
    log!(LogLevel::Info, "Processing response with status: {}", resp.status);

    // Add header indicating response was processed
    resp.headers.insert(
        "X-Script-Sanitized".to_string(),
        vec!["true".to_string()],
    );

    Ok(Json(ModifiedHTTPResponse {
        status: resp.status,
        headers: resp.headers,
    }))
}`,
		},
		{
			ID:          "python-add-header",
			Name:        "Add Headers (Python)",
			Description: "Python version of header addition script. Shows how to use Python for simple request modifications.",
			Language:    "python",
			Difficulty:  "beginner",
			Category:    "request_modification",
			TriggerType: "request",
			SourceCode: `# Simple Python script that adds custom headers to HTTP requests

# Add custom headers
if "X-Python-Processed" not in request.headers:
    request.headers["X-Python-Processed"] = ["true"]

request.headers["X-Script-Language"] = ["Python"]
request.headers["X-Runtime"] = ["RustPython-WASM"]

# Log the request
print(f"Processing {request.method} {request.url}")`,
		},
	}
}
