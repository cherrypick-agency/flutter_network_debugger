// Пример Rust скрипта для добавления кастомного заголовка
// Компиляция: cargo build --target wasm32-unknown-unknown --release

use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize)]
struct ScriptContext {
    request: Option<HTTPRequest>,
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

    // Логирование через host function
    extism_pdk::log!(
        extism_pdk::LogLevel::Info,
        "Processing request: {} {}",
        req.method,
        req.url
    );

    // Добавляем кастомные заголовки
    req.headers
        .insert("X-Script-Processed".to_string(), vec!["Rust".to_string()]);
    req.headers
        .insert("X-Test-Header".to_string(), vec!["E2E-Test".to_string()]);

    // Возвращаем модифицированный request
    Ok(Json(ScriptResult {
        modified: true,
        modified_request: Some(req),
        logs: Some(vec!["Added headers successfully".to_string()]),
        error: None,
    }))
}
