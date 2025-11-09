use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

#[plugin_fn]
pub fn process(Json(mut req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Log the request
    log!(LogLevel::Info, "Processing {} {}", req.method, req.url);
    
    // Add custom headers
    req.headers.insert("X-Script-Processed".to_string(), vec!["Rust".to_string()]);
    req.headers.insert("X-Compiled-With".to_string(), vec!["in-app-compilation".to_string()]);
    req.headers.insert("X-Custom-Header".to_string(), vec!["Hello from Rust WASM!".to_string()]);
    
    Ok(Json(req))
}
