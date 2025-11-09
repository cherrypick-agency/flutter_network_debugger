package compilers

import (
	"context"
	"fmt"
	"strings"

	"network-debugger/internal/features/scripting/domain"
)

// PythonCompiler compiles Python to WASM using RustPython runtime
// Architecture: Python code → Rust wrapper (with RustPython) → WASM
type PythonCompiler struct {
	rustCompiler *RustCompiler // Composition: reuse Rust compiler
}

// NewPythonCompiler creates a new Python compiler
// Note: PythonCompiler is a wrapper around RustCompiler (Python → Rust → WASM)
// Uses system Rust compiler (not cached) since Python is a wrapper language
func NewPythonCompiler() *PythonCompiler {
	// Create RustCompiler with nil cacheManager (will use system Rust)
	return &PythonCompiler{
		rustCompiler: NewRustCompiler(nil),
	}
}

// Language returns the language identifier
func (c *PythonCompiler) Language() string {
	return "python"
}

// IsAvailable checks if Rust compiler is available (needed for Python→Rust→WASM)
func (c *PythonCompiler) IsAvailable() bool {
	// Python compilation requires Rust compiler (system)
	return c.rustCompiler.IsAvailable()
}

// Compile compiles Python source code to WASM
// Strategy: Wrap Python code in Rust with RustPython runtime
func (c *PythonCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	if !c.IsAvailable() {
		return nil, fmt.Errorf("Python compiler requires Rust toolchain. Install Rust from https://rustup.rs/")
	}

	// Generate Rust wrapper with embedded Python code
	rustCode := c.generateRustWrapper(req.SourceCode)

	// Generate Cargo.toml with RustPython dependency
	cargoToml := c.generateCargoToml()

	// Create modified request for Rust compiler
	rustReq := domain.CompileRequest{
		SourceCode: rustCode,
		Dependencies: map[string]string{
			"Cargo.toml": cargoToml,
		},
		Language: "rust",
		Optimize: req.Optimize,
		Debug:    req.Debug,
	}

	// Delegate to Rust compiler (Dependency Inversion Principle)
	result, err := c.rustCompiler.Compile(ctx, rustReq)
	if err != nil {
		return nil, fmt.Errorf("Python→WASM compilation failed: %w", err)
	}

	// Add Python-specific logs
	result.Logs = append([]string{
		"Python compilation: Python → Rust wrapper → WASM",
		"Using RustPython runtime for execution",
	}, result.Logs...)

	return result, nil
}

// generateRustWrapper creates Rust code that embeds Python script
// Architecture: Template Method Pattern
func (c *PythonCompiler) generateRustWrapper(pythonCode string) string {
	// Escape Python code for embedding in Rust string
	escapedCode := escapeForRustString(pythonCode)

	return fmt.Sprintf(`use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use rustpython_vm as vm;
use vm::{Interpreter, PyResult};

#[derive(Deserialize, Serialize, Clone)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
    body: Vec<u8>,
}

// Initialize Python interpreter
fn init_interpreter() -> Interpreter {
    Interpreter::without_stdlib(Default::default())
}

// Execute Python code with HTTP request context
fn execute_python(interp: &Interpreter, code: &str, req: &HTTPRequest) -> PyResult<HTTPRequest> {
    interp.enter(|vm| {
        // Convert Rust types to Python
        let scope = vm.new_scope_with_builtins();

        // Create request dict for Python
        let req_dict = vm.ctx.new_dict();
        req_dict.set_item("method", vm.ctx.new_str(&req.method), vm)?;
        req_dict.set_item("url", vm.ctx.new_str(&req.url), vm)?;

        // Headers as dict
        let headers_dict = vm.ctx.new_dict();
        for (k, v) in &req.headers {
            let values = vm.ctx.new_list(
                v.iter().map(|s| vm.ctx.new_str(s).into()).collect()
            );
            headers_dict.set_item(k, values, vm)?;
        }
        req_dict.set_item("headers", headers_dict, vm)?;

        // Body as bytes
        let body_bytes = vm.ctx.new_bytes(req.body.clone());
        req_dict.set_item("body", body_bytes, vm)?;

        // Add request to scope
        scope.locals.set_item("request", req_dict, vm)?;

        // Execute Python code
        vm.run_code_string(scope.clone(), code, "<script>".to_owned())?;

        // Get modified request back
        let result_req = scope.locals.get_item("request", vm)?;

        // Convert Python dict back to Rust HTTPRequest
        let py_method = result_req.get_attr("method", vm)?;
        let method = py_method.str(vm)?.to_string();

        let py_url = result_req.get_attr("url", vm)?;
        let url = py_url.str(vm)?.to_string();

        let py_headers = result_req.get_attr("headers", vm)?;
        let mut headers = HashMap::new();
        if let Ok(dict) = py_headers.downcast::<vm::builtins::PyDict>() {
            for (k, v) in dict.into_iter() {
                let key = k.str(vm)?.to_string();
                let mut values = Vec::new();
                if let Ok(list) = v.downcast::<vm::builtins::PyList>() {
                    for item in list.borrow_vec() {
                        values.push(item.str(vm)?.to_string());
                    }
                }
                headers.insert(key, values);
            }
        }

        let py_body = result_req.get_attr("body", vm)?;
        let body = if let Ok(bytes) = py_body.downcast::<vm::builtins::PyBytes>() {
            bytes.borrow_value().to_vec()
        } else {
            req.body.clone()
        };

        Ok(HTTPRequest { method, url, headers, body })
    })
}

#[plugin_fn]
pub fn process(Json(req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Python code embedded at compile time
    const PYTHON_CODE: &str = r#"%s"#;

    // Initialize interpreter
    let interp = init_interpreter();

    // Execute Python code
    match execute_python(&interp, PYTHON_CODE, &req) {
        Ok(modified_req) => Ok(Json(modified_req)),
        Err(e) => {
            log!(LogLevel::Error, "Python execution error: {:?}", e);
            // Return original request on error
            Ok(Json(req))
        }
    }
}
`, escapedCode)
}

// generateCargoToml creates Cargo.toml with RustPython dependency
func (c *PythonCompiler) generateCargoToml() string {
	return `[package]
name = "python-wasm-plugin"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
extism-pdk = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

# RustPython for Python runtime
rustpython-vm = { version = "0.3", default-features = false, features = ["freeze-stdlib"] }
rustpython-compiler = "0.3"

[profile.release]
opt-level = "z"     # Optimize for size
lto = true          # Enable Link Time Optimization
codegen-units = 1   # Better optimization
strip = true        # Strip symbols
`
}

// ValidateSyntax validates Python syntax
func (c *PythonCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	// Basic Python syntax validation
	// Could use python -m py_compile or ast.parse in future

	// Check for common syntax errors
	code := req.SourceCode

	// Basic checks
	if strings.TrimSpace(code) == "" {
		return &domain.CompilationError{
			Message: "Empty Python code",
		}
	}

	// Check for balanced brackets/parentheses
	if err := validatePythonBrackets(code); err != nil {
		return &domain.CompilationError{
			Message:   "Syntax error",
			Traceback: err.Error(),
		}
	}

	// Check indentation (basic)
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}
		// Check mixed tabs and spaces
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		tabs := len(line) - len(strings.TrimLeft(line, "\t"))
		if spaces > 0 && tabs > 0 {
			return &domain.CompilationError{
				Message:   "Mixed spaces and tabs in indentation",
				Traceback: fmt.Sprintf("Line %d: inconsistent use of tabs and spaces", i+1),
				Line:      i + 1,
			}
		}
	}

	return nil
}

// ValidateDependencies validates Python dependencies
func (c *PythonCompiler) ValidateDependencies(deps map[string]string) error {
	// Python dependencies could be:
	// - requirements.txt
	// - additional .py modules

	for filename := range deps {
		if filename == "requirements.txt" {
			// Valid
			continue
		}
		if strings.HasSuffix(filename, ".py") {
			// Valid Python module
			continue
		}
		return fmt.Errorf("invalid Python dependency: %s (expected .py or requirements.txt)", filename)
	}

	return nil
}

// Helper functions

func escapeForRustString(s string) string {
	// Escape special characters for Rust raw string
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func validatePythonBrackets(code string) error {
	// Stack-based bracket matching
	stack := make([]rune, 0)
	pairs := map[rune]rune{
		')': '(',
		']': '[',
		'}': '{',
	}

	inString := false
	escapeNext := false
	stringChar := rune(0)

	for _, ch := range code {
		// Handle strings
		if escapeNext {
			escapeNext = false
			continue
		}
		if ch == '\\' {
			escapeNext = true
			continue
		}
		if ch == '"' || ch == '\'' {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
			continue
		}

		if inString {
			continue
		}

		// Check brackets
		if ch == '(' || ch == '[' || ch == '{' {
			stack = append(stack, ch)
		} else if ch == ')' || ch == ']' || ch == '}' {
			if len(stack) == 0 {
				return fmt.Errorf("unmatched closing bracket: %c", ch)
			}
			expected := pairs[ch]
			if stack[len(stack)-1] != expected {
				return fmt.Errorf("mismatched brackets: expected %c but got %c", stack[len(stack)-1], ch)
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 0 {
		return fmt.Errorf("unclosed bracket: %c", stack[len(stack)-1])
	}

	return nil
}
