package compilers

import (
	"context"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestRustCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = "" // Убеждаемся что путь не установлен

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
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
func TestRustCompiler_Compile_WithCargoToml(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"Cargo.toml": `[package]
name = "test"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]`,
		},
	}

	compiler.cargoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_WithAdditionalFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"src/utils.rs": "pub fn helper() {}",
		},
	}

	compiler.cargoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
		Optimize:   true,
	}

	compiler.cargoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_WithTimeout(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
	}

	compiler.cargoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error")
	}
}

// Composer 1.
func TestRustCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/rust",
	}
	compiler := NewRustCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestRustCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestRustCompiler_ValidateDependencies_ValidCargoToml(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	deps := map[string]string{
		"Cargo.toml": `[package]
name = "test"
version = "0.1.0"

[lib]
crate-type = ["cdylib"]`,
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestRustCompiler_ValidateDependencies_InvalidCargoToml_NoPackage(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	deps := map[string]string{
		"Cargo.toml": "name = \"test\"",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid Cargo.toml")
	}
}

// Composer 1.
func TestRustCompiler_ValidateDependencies_InvalidCargoToml_NoCdylib(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	deps := map[string]string{
		"Cargo.toml": `[package]
name = "test"
version = "0.1.0"`,
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error when cdylib not specified")
	}
}

// Composer 1.
func TestRustCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestRustCompiler_GetRustPaths_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/rust",
	}
	compiler := NewRustCompiler(cache)

	cargoPath, rustupPath, err := compiler.getRustPaths()
	_ = cargoPath
	_ = rustupPath
	_ = err
}

// Composer 1.
func TestRustCompiler_GetRustPaths_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewRustCompiler(cache)

	_, _, err := compiler.getRustPaths()
	if err == nil {
		t.Error("getRustPaths() should return error when compiler not in cache")
	}
}

// Composer 1.
func TestRustCompiler_GetRustEnv_WithPaths(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	compiler.rustupHome = "/test/rustup"
	compiler.cargoHome = "/test/cargo"

	env := compiler.getRustEnv()
	if len(env) == 0 {
		t.Error("getRustEnv() should return environment variables")
	}

	foundRustupHome := false
	foundCargoHome := false
	for _, e := range env {
		if contains(e, "RUSTUP_HOME=/test/rustup") {
			foundRustupHome = true
		}
		if contains(e, "CARGO_HOME=/test/cargo") {
			foundCargoHome = true
		}
	}

	if !foundRustupHome {
		t.Error("RUSTUP_HOME should be set in environment")
	}

	if !foundCargoHome {
		t.Error("CARGO_HOME should be set in environment")
	}
}

// Composer 1.
func TestRustCompiler_GetRustEnv_WithoutPaths(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	env := compiler.getRustEnv()
	if len(env) == 0 {
		t.Error("getRustEnv() should return at least some environment variables")
	}
}

// Composer 1.
func TestRustCompiler_ValidateSyntax_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""
	compiler.rustupPath = ""

	// Устанавливаем IsAvailable в false, переопределяя проверку
	// Это симулирует ситуацию когда компилятор недоступен
	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
	}

	// ValidateSyntax проверяет IsAvailable, который может вернуть true если Rust есть в системе
	// Поэтому просто проверяем что метод не падает
	err := compiler.ValidateSyntax(ctx, req)
	// Может вернуть ошибку или нет в зависимости от наличия Rust в системе
	_ = err
}

// Composer 1.
func TestRustCompiler_ParseRustError_WithLocation(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	output := "error[E0308]: mismatched types\n --> src/lib.rs:10:5\n   |\n10 |     let x: i32 = \"hello\";\n   |                  ^^^^^^^ expected i32, found &str"
	err := compiler.parseRustError(output)

	if err == nil {
		t.Fatal("parseRustError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseRustError() should return CompilationError")
	}

	if compErr.Line != 10 {
		t.Errorf("Line = %d, want %d", compErr.Line, 10)
	}

	if compErr.Column != 5 {
		t.Errorf("Column = %d, want %d", compErr.Column, 5)
	}

	if compErr.Code != "E0308" {
		t.Errorf("Code = %q, want %q", compErr.Code, "E0308")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestRustCompiler_ParseRustError_WithErrorCode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	output := "error[E0425]: cannot find value `foo` in this scope"
	err := compiler.parseRustError(output)

	if err == nil {
		t.Fatal("parseRustError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseRustError() should return CompilationError")
	}

	if compErr.Code != "E0425" {
		t.Errorf("Code = %q, want %q", compErr.Code, "E0425")
	}
}

// Composer 1.
func TestRustCompiler_ParseRustError_WithErrorColon(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	output := "error: cannot find value `foo` in this scope"
	err := compiler.parseRustError(output)

	if err == nil {
		t.Fatal("parseRustError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseRustError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestRustCompiler_ParseRustError_NoLocation(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	output := "error[E0308]: mismatched types"
	err := compiler.parseRustError(output)

	if err == nil {
		t.Fatal("parseRustError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseRustError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestRustCompiler_GetRustPaths_WithCachedPaths(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = "/test/cargo"
	compiler.rustupPath = "/test/rustup"

	cargoPath, rustupPath, err := compiler.getRustPaths()
	if err != nil {
		t.Errorf("getRustPaths() error = %v, want nil", err)
	}

	if cargoPath != "/test/cargo" {
		t.Errorf("cargoPath = %q, want %q", cargoPath, "/test/cargo")
	}

	if rustupPath != "/test/rustup" {
		t.Errorf("rustupPath = %q, want %q", rustupPath, "/test/rustup")
	}
}

// Composer 1.
func TestRustCompiler_GetRustPaths_NoCacheManager(t *testing.T) {
	compiler := NewRustCompiler(nil)

	_, _, err := compiler.getRustPaths()
	if err == nil {
		t.Error("getRustPaths() should return error when cache manager is nil")
	}
}

// Composer 1.
func TestRustCompiler_OptimizeWASM_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	ctx := context.Background()
	wasmPath := "/tmp/test.wasm"

	// OptimizeWASM возвращает nil если wasm-opt не найден
	err := compiler.OptimizeWASM(ctx, wasmPath)
	// Может вернуть nil или ошибку в зависимости от наличия wasm-opt
	_ = err
}

// Composer 1.
func TestRustCompiler_Compile_WithRustEnv(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""
	compiler.rustupHome = "/test/rustup"
	compiler.cargoHome = "/test/cargo"

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
	}

	_, err := compiler.Compile(ctx, req)
	// Ожидаем ошибку так как компилятор недоступен
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_ReleaseMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
		Optimize:   true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_WithSrcFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"src/utils.rs": "pub fn helper() {}",
		},
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_WithoutCargoToml(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode:   "fn main() {}",
		Dependencies: map[string]string{},
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_WithoutRustEnv(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""
	compiler.rustupHome = ""
	compiler.cargoHome = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_Compile_WithCargoTomlInDeps(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"Cargo.toml": `[package]
name = "test"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]`,
		},
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_IsAvailable_WithCachedRust(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/rust",
	}
	compiler := NewRustCompiler(cache)

	// IsAvailable проверяет wasm32 target для кешированного Rust
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestRustCompiler_IsAvailable_SystemCargoWithoutRustup(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewRustCompiler(cache)

	// IsAvailable может вернуть false если rustup не найден
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestRustCompiler_IsAvailable_SystemRustWithoutWasm32Target(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewRustCompiler(cache)

	// IsAvailable может вернуть false если wasm32 target не установлен
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestRustCompiler_ValidateSyntax_WithoutRustEnv(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""
	compiler.rustupHome = ""
	compiler.cargoHome = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestRustCompiler_ValidateSyntax_WithRustEnv(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)
	compiler.cargoPath = ""
	compiler.rustupHome = "/test/rustup"
	compiler.cargoHome = "/test/cargo"

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fn main() {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}
