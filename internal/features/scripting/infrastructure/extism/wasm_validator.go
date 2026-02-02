package extism

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"network-debugger/internal/features/scripting/domain"
)

// ValidateWASMExports validates that a WASM module exports the required "process" function.
// This is called after compilation to catch errors early, before the script is used.
func ValidateWASMExports(ctx context.Context, wasmBytes []byte) error {
	if len(wasmBytes) == 0 {
		return fmt.Errorf("WASM binary is empty")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// Create a minimal wazero runtime just for validation (not execution)
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(context.Background())

	// Compile the module to inspect its exports
	// This does NOT instantiate or execute the module
	compiled, err := runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("invalid WASM binary: %w", err)
	}
	defer compiled.Close(context.Background())

	// Check for "process" function export
	exports := compiled.ExportedFunctions()
	processFunc, hasProcess := exports["process"]

	if !hasProcess {
		// List available exports for debugging
		availableExports := make([]string, 0, len(exports))
		for name := range exports {
			availableExports = append(availableExports, name)
		}
		return &MissingProcessFunctionError{
			AvailableExports: availableExports,
		}
	}

	// Validate function signature
	// Extism PDK expects: process() -> i32 (for Go/AssemblyScript)
	// or process(i64) -> i64 (for Rust with extism-pdk)
	// We accept both signatures
	if err := validateProcessSignature(processFunc); err != nil {
		return err
	}

	return nil
}

// validateProcessSignature checks that the process function has a valid signature
func validateProcessSignature(fn api.FunctionDefinition) error {
	params := fn.ParamTypes()
	results := fn.ResultTypes()

	// Accept multiple valid signatures:
	// 1. () -> () - no params, no return (valid for some PDKs)
	// 2. () -> i32 - no params, returns i32 (Go/AssemblyScript)
	// 3. (i64) -> i64 - takes i64, returns i64 (Rust extism-pdk)

	if len(params) == 0 && len(results) == 0 {
		return nil
	}
	if len(params) == 0 && len(results) == 1 && results[0] == api.ValueTypeI32 {
		return nil
	}
	if len(params) == 1 && params[0] == api.ValueTypeI64 && len(results) == 1 && results[0] == api.ValueTypeI64 {
		return nil
	}

	return fmt.Errorf("invalid process signature (params=%v results=%v)", params, results)
}

// MissingProcessFunctionError indicates the WASM module doesn't export "process" function
type MissingProcessFunctionError struct {
	AvailableExports []string
}

func (e *MissingProcessFunctionError) Error() string {
	if len(e.AvailableExports) == 0 {
		return "WASM module does not export 'process' function (no exports found)"
	}
	return fmt.Sprintf("WASM module does not export 'process' function (available exports: %v)", e.AvailableExports)
}

// ValidateWASMWithLanguage validates WASM and returns language-specific error messages
func ValidateWASMWithLanguage(ctx context.Context, wasmBytes []byte, language string) error {
	err := ValidateWASMExports(ctx, wasmBytes)
	if err == nil {
		return nil
	}

	// Wrap error with language-specific help
	if _, ok := err.(*MissingProcessFunctionError); ok {
		return NewWASMValidationError(language, err)
	}
	// Также полезно для "неправильной сигнатуры" — ошибка похожа на missing export по UX.
	if strings.Contains(err.Error(), "invalid process signature") {
		return NewWASMValidationError(language, err)
	}

	return err
}

// ExtismWASMValidator implements domain.WASMValidator for Extism runtime
type ExtismWASMValidator struct{}

// NewExtismWASMValidator creates a new WASM validator for Extism
func NewExtismWASMValidator() *ExtismWASMValidator {
	return &ExtismWASMValidator{}
}

// ValidateWASM implements domain.WASMValidator
func (v *ExtismWASMValidator) ValidateWASM(wasmBytes []byte, language string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ValidateWASMWithLanguage(ctx, wasmBytes, language)
}

// Verify ExtismWASMValidator implements domain.WASMValidator
var _ domain.WASMValidator = (*ExtismWASMValidator)(nil)
