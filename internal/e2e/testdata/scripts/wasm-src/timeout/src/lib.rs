use extism_pdk::*;

// This script intentionally runs forever to test timeout handling
// The executor should kill it after the configured timeout
#[plugin_fn]
pub fn process(_input: Vec<u8>) -> FnResult<Vec<u8>> {
    // Infinite loop - will trigger timeout
    loop {
        // Busy loop to consume CPU
        // This will be killed by the executor's timeout mechanism
    }
}
