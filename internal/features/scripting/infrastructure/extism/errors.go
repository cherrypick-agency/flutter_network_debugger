package extism

import (
	"fmt"
	"strings"
)

// WASMValidationError represents a validation error with helpful fix suggestions
type WASMValidationError struct {
	Message    string
	Language   string
	FixExample string
	Cause      error
}

func (e *WASMValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Message)
	if e.FixExample != "" {
		sb.WriteString("\n\nExample fix for ")
		sb.WriteString(e.Language)
		sb.WriteString(":\n")
		sb.WriteString(e.FixExample)
	}
	return sb.String()
}

func (e *WASMValidationError) Unwrap() error {
	return e.Cause
}

// NewWASMValidationError creates a validation error with language-specific help
func NewWASMValidationError(language string, cause error) *WASMValidationError {
	return &WASMValidationError{
		Message:    fmt.Sprintf("WASM module validation failed: %v", cause),
		Language:   language,
		FixExample: GetExampleCode(language),
		Cause:      cause,
	}
}

// GetExampleCode returns example code for the given language showing how to export process function
func GetExampleCode(language string) string {
	switch strings.ToLower(language) {
	case "rust":
		return rustExample
	case "go", "tinygo":
		return goExample
	case "assemblyscript", "typescript", "ts":
		return assemblyScriptExample
	case "c":
		return cExample
	case "zig":
		return zigExample
	default:
		return genericExample
	}
}

const rustExample = `use extism_pdk::*;
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
struct ScriptContext {
    request: Option<HttpRequest>,
}

#[derive(Deserialize, Serialize)]
struct HttpRequest {
    method: String,
    url: String,
    headers: std::collections::HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

#[derive(Serialize)]
struct ScriptResult {
    modified: bool,
    #[serde(rename = "modifiedRequest")]
    modified_request: Option<HttpRequest>,
    logs: Vec<String>,
}

#[plugin_fn]
pub fn process(input: String) -> FnResult<String> {
    let ctx: ScriptContext = serde_json::from_str(&input)?;

    // Your logic here
    let result = ScriptResult {
        modified: false,
        modified_request: None,
        logs: vec!["Script executed".to_string()],
    };

    Ok(serde_json::to_string(&result)?)
}`

const goExample = `package main

import (
    "encoding/json"
    "github.com/extism/go-pdk"
)

type ScriptContext struct {
    Request *HTTPRequest ` + "`json:\"request,omitempty\"`" + `
}

type HTTPRequest struct {
    Method  string              ` + "`json:\"method\"`" + `
    URL     string              ` + "`json:\"url\"`" + `
    Headers map[string][]string ` + "`json:\"headers\"`" + `
    Body    []byte              ` + "`json:\"body\"`" + `
}

type ScriptResult struct {
    Modified        bool         ` + "`json:\"modified\"`" + `
    ModifiedRequest *HTTPRequest ` + "`json:\"modifiedRequest,omitempty\"`" + `
    Logs            []string     ` + "`json:\"logs\"`" + `
}

//export process
func process() int32 {
    input := pdk.Input()
    var ctx ScriptContext
    json.Unmarshal(input, &ctx)

    // Your logic here
    result := ScriptResult{
        Modified: false,
        Logs:     []string{"Script executed"},
    }

    output, _ := json.Marshal(result)
    pdk.OutputMemory(output)
    return 0
}

func main() {}`

const assemblyScriptExample = `// AssemblyScript example
declare function Host_inputString(): string;
declare function Host_outputString(s: string): void;

interface ScriptContext {
    request?: HttpRequest;
}

interface HttpRequest {
    method: string;
    url: string;
    headers: Map<string, string[]>;
    body: ArrayBuffer;
}

interface ScriptResult {
    modified: boolean;
    modifiedRequest?: HttpRequest;
    logs: string[];
}

export function process(): i32 {
    const input = Host_inputString();
    const ctx: ScriptContext = JSON.parse(input);

    // Your logic here
    const result: ScriptResult = {
        modified: false,
        logs: ["Script executed"]
    };

    Host_outputString(JSON.stringify(result));
    return 0;
}`

const cExample = `// C example
#include <extism-pdk.h>
#include <string.h>

int32_t EXTISM_EXPORTED_SYMBOL(process)() {
    // Read input
    uint64_t input_len = extism_input_length();
    uint8_t *input = malloc(input_len + 1);
    extism_input_load_u8(input, input_len);
    input[input_len] = '\0';

    // Your logic here
    const char *result = "{\"modified\":false,\"logs\":[\"Script executed\"]}";

    // Write output
    extism_output_set_from_u8((const uint8_t *)result, strlen(result));

    free(input);
    return 0;
}`

const zigExample = `// Zig example
const std = @import("std");
const extism = @import("extism-pdk");

export fn process() i32 {
    const allocator = std.heap.page_allocator;

    // Read input
    const input = extism.getInput(allocator) catch return 1;
    defer allocator.free(input);

    // Your logic here
    const result = "{\"modified\":false,\"logs\":[\"Script executed\"]}";

    // Write output
    extism.output(result);
    return 0;
}`

const genericExample = `Your WASM module must export a function named "process".

For Extism-based scripts, this function should:
1. Read input using the PDK's input function
2. Parse the JSON ScriptContext
3. Perform your modifications
4. Return a JSON ScriptResult

Check the documentation for your specific language's Extism PDK.`

// GetAvailableLanguages returns list of languages with examples
func GetAvailableLanguages() []string {
	return []string{"rust", "go", "assemblyscript", "c", "zig"}
}
