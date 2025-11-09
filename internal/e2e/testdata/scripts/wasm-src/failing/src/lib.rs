// WASM module that always fails (for circuit breaker testing)
// This compiles successfully but panics on execution
use extism_pdk::*;

#[plugin_fn]
pub fn process(_input: Vec<u8>) -> FnResult<Vec<u8>> {
    // Always panic to trigger circuit breaker
    panic!("Intentional failure for circuit breaker testing");
}
