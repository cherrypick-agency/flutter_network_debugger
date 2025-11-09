use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

/// HTTP Request structure passed from the debugger
#[derive(Deserialize, Serialize)]
struct HttpRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

/// Script result returned to the debugger
#[derive(Serialize)]
struct ScriptResult {
    modified: bool,
    #[serde(rename = "modifiedRequest")]
    modified_request: Option<HttpRequest>,
    logs: Vec<String>,
}

/// Main plugin function called by Extism runtime
/// This is triggered for every HTTP request that matches the script's rules
#[plugin_fn]
pub fn transform(input: String) -> FnResult<String> {
    let mut logs = Vec::new();
    logs.push("Script started: Add Custom Header".to_string());

    // Parse input JSON (request context from debugger)
    let ctx: serde_json::Value = serde_json::from_str(&input)
        .map_err(|e| {
            WithReturnCode::new(
                Error::msg(format!("Failed to parse input: {}", e)),
                1,
            )
        })?;

    // Extract request from context
    let mut request: HttpRequest = serde_json::from_value(ctx["request"].clone())
        .map_err(|e| {
            WithReturnCode::new(
                Error::msg(format!("Failed to parse request: {}", e)),
                1,
            )
        })?;

    logs.push(format!("Original URL: {}", request.url));
    logs.push(format!("Original headers: {:?}", request.headers));

    // Add custom header
    request.headers.insert(
        "X-Custom-Header".to_string(),
        vec!["My-Value".to_string()],
    );

    logs.push("Added X-Custom-Header: My-Value".to_string());

    // Return modified request
    let result = ScriptResult {
        modified: true,
        modified_request: Some(request),
        logs,
    };

    let json = serde_json::to_string(&result)
        .map_err(|e| Error::msg(format!("Failed to serialize result: {}", e)))?;

    Ok(json)
}
