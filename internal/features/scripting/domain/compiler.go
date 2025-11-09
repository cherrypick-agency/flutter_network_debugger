package domain

import (
	"context"
	"fmt"
	"time"
)

// Compiler is the PORT interface for compiling scripts to WASM (Clean Architecture)
// Implementations live in infrastructure layer (ADAPTERS)
type Compiler interface {
	// Language returns the supported language identifier
	// Examples: "rust", "go", "javascript", "typescript", "assemblyscript"
	Language() string

	// Compile compiles source code to WASM binary
	Compile(ctx context.Context, req CompileRequest) (*CompileResult, error)

	// ValidateSyntax validates source code syntax without full compilation
	// Useful for quick feedback during editing
	ValidateSyntax(ctx context.Context, req CompileRequest) error

	// ValidateDependencies validates dependency manifest structure
	// (Cargo.toml, go.mod, package.json)
	ValidateDependencies(deps map[string]string) error

	// IsAvailable checks if compiler toolchain is installed and ready
	IsAvailable() bool
}

// CompileRequest contains all data needed for compilation (Value Object)
type CompileRequest struct {
	SourceCode   string            // Main source code
	Dependencies map[string]string // filename → content (Cargo.toml, go.mod, package.json)
	Language     string            // Language identifier
	Optimize     bool              // Enable compiler optimizations
	Debug        bool              // Include debug symbols
}

// CompileResult contains compilation output (Value Object)
type CompileResult struct {
	WASMBinary []byte        // Compiled WASM binary
	Logs       []string      // Compilation logs (stdout/stderr)
	Warnings   []string      // Compilation warnings
	Duration   time.Duration // Compilation time
	WASMSize   int64         // Binary size in bytes
}

// CompilationError represents a compilation failure (Domain Error)
type CompilationError struct {
	Message   string // Error message
	Line      int    // Line number (if available)
	Column    int    // Column number (if available)
	Code      string // Error code (e.g., "E0308" for Rust)
	Severity  string // "error", "warning", "note"
	Traceback string // Full traceback/stack trace
}

func (e *CompilationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s at line %d:%d: %s", e.Severity, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Severity, e.Message)
}

// NewCompilationError creates a new compilation error
func NewCompilationError(message string) *CompilationError {
	return &CompilationError{
		Message:  message,
		Severity: "error",
	}
}
