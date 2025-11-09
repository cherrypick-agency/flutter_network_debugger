// Minimal WASM module for testing - returns input unchanged
use extism_pdk::*;

#[plugin_fn]
pub fn process(input: Vec<u8>) -> FnResult<Vec<u8>> {
    // Simply return input unchanged
    Ok(input)
}
