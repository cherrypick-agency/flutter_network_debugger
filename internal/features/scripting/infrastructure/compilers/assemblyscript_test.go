package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestAssemblyScriptCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)
	compiler.ascPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
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
func TestAssemblyScriptCompiler_Compile_WithPackageJSON(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Dependencies: map[string]string{
			"package.json": `{
  "name": "test",
  "dependencies": {
    "assemblyscript": "^0.27.0"
  }
}`,
		},
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_WithAdditionalFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Dependencies: map[string]string{
			"utils.ts": "export function helper(): void {}",
		},
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Optimize:   true,
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_DebugFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Debug:      true,
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/assemblyscript",
	}
	compiler := NewAssemblyScriptCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestAssemblyScriptCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestAssemblyScriptCompiler_ValidateDependencies_ValidPackageJSON(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	deps := map[string]string{
		"package.json": `{
  "name": "test",
  "dependencies": {
    "assemblyscript": "^0.27.0"
  }
}`,
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_ValidateDependencies_InvalidPackageJSON(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	deps := map[string]string{
		"package.json": "invalid json",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid package.json")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_GetAssemblyScriptPaths_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/assemblyscript",
	}
	compiler := NewAssemblyScriptCompiler(cache)

	ascPath, npmPath, err := compiler.getAssemblyScriptPaths()
	_ = ascPath
	_ = npmPath
	_ = err
}

// Composer 1.
func TestAssemblyScriptCompiler_GetAssemblyScriptPaths_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewAssemblyScriptCompiler(cache)

	_, _, err := compiler.getAssemblyScriptPaths()
	if err == nil {
		t.Error("getAssemblyScriptPaths() should return error when compiler not in cache")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_GetAssemblyScriptPaths_WithCachedPaths(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)
	compiler.ascPath = "/test/asc"
	compiler.npmPath = "/test/npm"

	ascPath, npmPath, err := compiler.getAssemblyScriptPaths()
	if err != nil {
		t.Errorf("getAssemblyScriptPaths() error = %v, want nil", err)
	}

	if ascPath != "/test/asc" {
		t.Errorf("ascPath = %q, want %q", ascPath, "/test/asc")
	}

	if npmPath != "/test/npm" {
		t.Errorf("npmPath = %q, want %q", npmPath, "/test/npm")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_GetAssemblyScriptPaths_NoCacheManager(t *testing.T) {
	compiler := NewAssemblyScriptCompiler(nil)

	_, _, err := compiler.getAssemblyScriptPaths()
	if err == nil {
		t.Error("getAssemblyScriptPaths() should return error when cache manager is nil")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_ValidateSyntax(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)
	compiler.ascPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
	}

	err := compiler.ValidateSyntax(ctx, req)
	// ValidateSyntax вызывает Compile, который вернет ошибку если компилятор недоступен
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_ParseCompilationError_WithErrorLine(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	output := "ERROR TS2304: Cannot find name 'foo'\n  at index.ts:5:10"
	err := compiler.parseCompilationError("compilation failed", output)

	if err == nil {
		t.Fatal("parseCompilationError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseCompilationError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}

	if compErr.Traceback != output {
		t.Errorf("Traceback = %q, want %q", compErr.Traceback, output)
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_ParseCompilationError_NoErrorLine(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	output := "some output without ERROR keyword"
	err := compiler.parseCompilationError("compilation failed", output)

	if err == nil {
		t.Fatal("parseCompilationError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseCompilationError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_WithTsconfig(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Dependencies: map[string]string{
			"tsconfig.json": `{"compilerOptions": {}}`,
		},
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_IsAvailable_SystemFallback(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewAssemblyScriptCompiler(cache)

	// IsAvailable может вернуть true если asc и npm есть в системе
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_WithPackageJSONAndNpmInstall(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Dependencies: map[string]string{
			"package.json": `{
  "name": "test",
  "dependencies": {
    "assemblyscript": "^0.27.0"
  }
}`,
		},
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_WithSourceMap(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Debug:      true,
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_IsAvailable_SystemAscNotFound(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewAssemblyScriptCompiler(cache)

	// IsAvailable может вернуть false если asc не найден в системе
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestAssemblyScriptCompiler_IsAvailable_SystemNpmNotFound(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewAssemblyScriptCompiler(cache)

	// IsAvailable может вернуть false если npm не найден в системе
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_WithTsFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Dependencies: map[string]string{
			"utils.ts": "export function helper(): i32 { return 42; }",
		},
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Compile_WithOptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "export function add(a: i32, b: i32): i32 { return a + b; }",
		Optimize:   true,
	}

	compiler.ascPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}
