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
pub fn process(Json(req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Simple passthrough - no modifications
    Ok(Json(req))
}
