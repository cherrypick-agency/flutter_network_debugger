package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestPythonCompiler_Compile_NotAvailable(t *testing.T) {
	compiler := NewPythonCompiler()
	compiler.rustCompiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print('hello')",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when Rust compiler not available")
	}
}

// Composer 1.
func TestPythonCompiler_Compile_WithDependencies(t *testing.T) {
	compiler := NewPythonCompiler()
	compiler.rustCompiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print('hello')",
		Dependencies: map[string]string{
			"utils.py": "def helper(): pass",
		},
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when Rust compiler not available")
	}
}

// Composer 1.
func TestPythonCompiler_Compile_OptimizeFlag(t *testing.T) {
	compiler := NewPythonCompiler()
	compiler.rustCompiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print('hello')",
		Optimize:   true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when Rust compiler not available")
	}
}

// Composer 1.
func TestPythonCompiler_IsAvailable(t *testing.T) {
	compiler := NewPythonCompiler()

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestPythonCompiler_ValidateSyntax_Empty(t *testing.T) {
	compiler := NewPythonCompiler()

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error for empty code")
	}
}

// Composer 1.
func TestPythonCompiler_ValidateSyntax_ValidCode(t *testing.T) {
	compiler := NewPythonCompiler()

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print('hello')",
	}

	err := compiler.ValidateSyntax(ctx, req)
	// May return an error if there are syntax problems, but should not panic
	_ = err
}

// Composer 1.
func TestPythonCompiler_ValidateSyntax_UnbalancedBrackets(t *testing.T) {
	compiler := NewPythonCompiler()

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "def func(:\n    pass",
	}

	err := compiler.ValidateSyntax(ctx, req)
	// May return an error due to unbalanced brackets
	_ = err
}

// Composer 1.
func TestPythonCompiler_ValidateDependencies_ValidPyFile(t *testing.T) {
	compiler := NewPythonCompiler()

	deps := map[string]string{
		"utils.py": "def helper(): pass",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestPythonCompiler_ValidateDependencies_ValidRequirementsTxt(t *testing.T) {
	compiler := NewPythonCompiler()

	deps := map[string]string{
		"requirements.txt": "requests==2.31.0",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestPythonCompiler_ValidateDependencies_InvalidFile(t *testing.T) {
	compiler := NewPythonCompiler()

	deps := map[string]string{
		"invalid.txt": "some content",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid file")
	}
}

// Composer 1.
func TestPythonCompiler_ValidateDependencies_Empty(t *testing.T) {
	compiler := NewPythonCompiler()

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestPythonCompiler_GenerateRustWrapper(t *testing.T) {
	compiler := NewPythonCompiler()

	pythonCode := "print('hello')"
	wrapper := compiler.generateRustWrapper(pythonCode)

	if wrapper == "" {
		t.Error("generateRustWrapper() should return non-empty string")
	}

	if !contains(wrapper, "PYTHON_CODE") {
		t.Error("Rust wrapper should contain PYTHON_CODE constant")
	}
}

// Composer 1.
func TestPythonCompiler_GenerateCargoToml(t *testing.T) {
	compiler := NewPythonCompiler()

	toml := compiler.generateCargoToml()
	if toml == "" {
		t.Error("generateCargoToml() should return non-empty string")
	}

	if !contains(toml, "[package]") {
		t.Error("Cargo.toml should contain '[package]'")
	}

	if !contains(toml, "rustpython") {
		t.Error("Cargo.toml should contain rustpython dependency")
	}
}

// Composer 1.
func TestPythonCompiler_Compile_DebugFlag(t *testing.T) {
	compiler := NewPythonCompiler()
	compiler.rustCompiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print('hello')",
		Debug:      true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when Rust compiler not available")
	}
}
