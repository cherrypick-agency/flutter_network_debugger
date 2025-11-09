package domain

import (
	"context"
	"time"
)

// ScriptRuntime identifies the execution environment
type ScriptRuntime string

const (
	RuntimeExtism ScriptRuntime = "extism" // WASM runtime (Rust, Go, JS, etc.)
	RuntimeDart   ScriptRuntime = "dart"   // Dart subprocess runtime
)

// ScriptExecutor is the port/interface for all script runtimes
// This follows Clean Architecture - this is a PORT that adapters implement
type ScriptExecutor interface {
	// Execute runs a script with given input and returns output
	Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error)

	// Runtime returns the runtime type this executor handles
	Runtime() ScriptRuntime

	// Validate checks if the script is valid for this runtime
	Validate(ctx context.Context, script Script) error

	// Close cleans up resources
	Close() error
}

// ExecutionRequest contains input for script execution
type ExecutionRequest struct {
	Script  Script
	Input   []byte         // JSON-encoded ScriptContext
	Context map[string]any // Additional context data
}

// ExecutionResult contains script output
type ExecutionResult struct {
	Output   []byte
	Logs     []string
	Duration time.Duration
	Error    string
}
