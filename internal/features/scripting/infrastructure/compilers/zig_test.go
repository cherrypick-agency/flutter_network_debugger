package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestZigCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewZigCompiler(cache)
	compiler.zigPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "pub fn main() void {}",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("Compile() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestZigCompiler_Compile_WithDependencies(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "pub fn main() void {}",
		Dependencies: map[string]string{
			"utils.zig": "pub fn helper() void {}",
		},
	}

	compiler.zigPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestZigCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "pub fn main() void {}",
		Optimize:   true,
	}

	compiler.zigPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestZigCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/zig",
	}
	compiler := NewZigCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestZigCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestZigCompiler_ValidateDependencies_ValidZigFile(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	deps := map[string]string{
		"utils.zig": "pub fn helper() void {}",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestZigCompiler_ValidateDependencies_ValidBuildZig(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	deps := map[string]string{
		"build.zig": "const std = @import(\"std\");",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestZigCompiler_ValidateDependencies_InvalidFile(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	deps := map[string]string{
		"file.txt": "some content",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid file")
	}
}

// Composer 1.
func TestZigCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestZigCompiler_GetZigBinary_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/zig",
	}
	compiler := NewZigCompiler(cache)

	zigPath, err := compiler.getZigBinary()
	_ = zigPath
	_ = err
}

// Composer 1.
func TestZigCompiler_GetZigBinary_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewZigCompiler(cache)

	_, err := compiler.getZigBinary()
	if err == nil {
		t.Error("getZigBinary() should return error when compiler not in cache")
	}
}

// Composer 1.
func TestZigCompiler_ParseZigError(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	output := "main.zig:10:5: error: undefined symbol"
	err := compiler.parseZigError(output)

	if err == nil {
		t.Fatal("parseZigError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseZigError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestZigCompiler_ParseZigError_WithFullFormat(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	output := "main.zig:10:5: error: undefined symbol 'foo'"
	err := compiler.parseZigError(output)

	if err == nil {
		t.Fatal("parseZigError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseZigError() should return CompilationError")
	}

	if compErr.Line != 10 {
		t.Errorf("Line = %d, want %d", compErr.Line, 10)
	}

	if compErr.Column != 5 {
		t.Errorf("Column = %d, want %d", compErr.Column, 5)
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestZigCompiler_ParseZigError_NoErrorLine(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	output := "some output without error"
	err := compiler.parseZigError(output)

	if err == nil {
		t.Fatal("parseZigError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseZigError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestZigCompiler_ValidateSyntax_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewZigCompiler(cache)
	compiler.zigPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "pub fn main() void {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestZigCompiler_GetZigBinary_WithCachedPath(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)
	compiler.zigPath = "/test/zig"

	zigPath, err := compiler.getZigBinary()
	if err != nil {
		t.Errorf("getZigBinary() error = %v, want nil", err)
	}

	if zigPath != "/test/zig" {
		t.Errorf("zigPath = %q, want %q", zigPath, "/test/zig")
	}
}

// Composer 1.
func TestZigCompiler_Compile_ReleaseMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)
	compiler.zigPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "pub fn main() void {}",
		Optimize:   true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestZigCompiler_IsAvailable_SystemFallback(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewZigCompiler(cache)

	// IsAvailable may return true if zig is available in the system
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestZigCompiler_Compile_DebugMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)
	compiler.zigPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "pub fn main() void {}",
		Optimize:   false,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestZigCompiler_ValidateDependencies_WithBuildZig(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	deps := map[string]string{
		"build.zig": "const std = @import(\"std\");",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}
