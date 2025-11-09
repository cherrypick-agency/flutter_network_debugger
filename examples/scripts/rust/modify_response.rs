// Пример Rust скрипта для модификации response
// Use case: убираем чувствительные данные из API responses
// Компиляция: cargo build --target wasm32-unknown-unknown --release

use extism_pdk::*;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;

#[derive(Deserialize)]
struct HTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
    #[serde(with = "serde_bytes")]
    body: Vec<u8>,
}

#[derive(Serialize)]
struct ModifiedHTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
    #[serde(with = "serde_bytes")]
    body: Vec<u8>,
}

#[plugin_fn]
pub fn process(Json(mut resp): Json<HTTPResponse>) -> FnResult<Json<ModifiedHTTPResponse>> {
    extism_pdk::log!(
        extism_pdk::LogLevel::Info,
        "Processing response with status: {}",
        resp.status
    );

    // Проверяем Content-Type
    let is_json = resp
        .headers
        .get("Content-Type")
        .and_then(|v| v.first())
        .map(|ct| ct.contains("application/json"))
        .unwrap_or(false);

    if is_json && !resp.body.is_empty() {
        // Парсим JSON body
        if let Ok(body_str) = String::from_utf8(resp.body.clone()) {
            if let Ok(mut json) = serde_json::from_str::<Value>(&body_str) {
                // Удаляем чувствительные поля
                remove_sensitive_fields(&mut json);

                // Добавляем metadata о обработке
                if let Value::Object(ref mut map) = json {
                    map.insert(
                        "_sanitized".to_string(),
                        Value::Bool(true),
                    );
                    map.insert(
                        "_processedBy".to_string(),
                        Value::String("Network Debugger Script".to_string()),
                    );
                }

                // Сериализуем обратно
                if let Ok(modified_body) = serde_json::to_vec(&json) {
                    resp.body = modified_body;

                    // Обновляем Content-Length
                    resp.headers.insert(
                        "Content-Length".to_string(),
                        vec![resp.body.len().to_string()],
                    );
                }
            }
        }
    }

    // Добавляем заголовок о обработке
    resp.headers.insert(
        "X-Script-Sanitized".to_string(),
        vec!["true".to_string()],
    );

    Ok(Json(ModifiedHTTPResponse {
        status: resp.status,
        headers: resp.headers,
        body: resp.body,
    }))
}

// Рекурсивно удаляем чувствительные поля из JSON
fn remove_sensitive_fields(value: &mut Value) {
    match value {
        Value::Object(map) => {
            // Список чувствительных полей
            let sensitive_keys = [
                "password",
                "token",
                "secret",
                "apiKey",
                "api_key",
                "creditCard",
                "credit_card",
                "ssn",
                "socialSecurityNumber",
                "privateKey",
                "private_key",
            ];

            // Удаляем чувствительные поля
            for key in &sensitive_keys {
                if map.contains_key(*key) {
                    map.insert(key.to_string(), Value::String("***REDACTED***".to_string()));
                }
            }

            // Рекурсивно обрабатываем вложенные объекты
            for (_, v) in map.iter_mut() {
                remove_sensitive_fields(v);
            }
        }
        Value::Array(arr) => {
            // Рекурсивно обрабатываем элементы массива
            for item in arr.iter_mut() {
                remove_sensitive_fields(item);
            }
        }
        _ => {}
    }
}
