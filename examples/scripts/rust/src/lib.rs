// Пример Rust скрипта для добавления кастомного заголовка
// Компиляция: cargo build --target wasm32-unknown-unknown --release

use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

#[derive(Serialize)]
struct HTTPResponse {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

#[plugin_fn]
pub fn process(Json(mut req): Json<HTTPRequest>) -> FnResult<Json<HTTPResponse>> {
    // Логирование через host function
    extism_pdk::log!(
        extism_pdk::LogLevel::Info,
        "Processing request: {} {}",
        req.method,
        req.url
    );

    // Добавляем кастомный заголовок
    req.headers
        .insert("X-Script-Processed".to_string(), vec!["Rust".to_string()]);

    // Добавляем тестовый заголовок для проверки
    req.headers.insert(
        "X-Test-Header".to_string(),
        vec!["E2E-Test".to_string()],
    );

    // Возвращаем модифицированный request
    Ok(Json(HTTPResponse {
        method: req.method,
        url: req.url,
        headers: req.headers,
        body: req.body,
    }))
}
