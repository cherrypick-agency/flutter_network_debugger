package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestTinyGoCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	if compiler.Language() != "go" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "go")
	}
}

// Composer 1.
func TestRustCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	if compiler.Language() != "rust" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "rust")
	}
}

// Composer 1.
func TestAssemblyScriptCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	if compiler.Language() != "assemblyscript" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "assemblyscript")
	}
}

// Composer 1.
func TestCCPPCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	if compiler.Language() != "c" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "c")
	}
}

// Composer 1.
func TestPythonCompiler_Language(t *testing.T) {
	compiler := NewPythonCompiler()

	if compiler == nil {
		t.Fatal("NewPythonCompiler returned nil")
	}

	if compiler.Language() != "python" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "python")
	}
}

// Composer 1.
func TestKotlinCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	if compiler.Language() != "kotlin" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "kotlin")
	}
}

// Composer 1.
func TestSwiftCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	if compiler.Language() != "swift" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "swift")
	}
}

// Composer 1.
func TestZigCompiler_Language(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	if compiler.Language() != "zig" {
		t.Errorf("Language() = %q, want %q", compiler.Language(), "zig")
	}
}

// Composer 1.
func TestNewTinyGoCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	if compiler == nil {
		t.Fatal("NewTinyGoCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}

	if compiler.goPath != "go" {
		t.Errorf("goPath = %q, want %q", compiler.goPath, "go")
	}
}

// Composer 1.
func TestNewRustCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	if compiler == nil {
		t.Fatal("NewRustCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}
}

// Composer 1.
func TestNewAssemblyScriptCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	if compiler == nil {
		t.Fatal("NewAssemblyScriptCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}
}

// Composer 1.
func TestNewPythonCompiler(t *testing.T) {
	compiler := NewPythonCompiler()

	if compiler == nil {
		t.Fatal("NewPythonCompiler returned nil")
	}

	if compiler.rustCompiler == nil {
		t.Error("rustCompiler not set")
	}
}

// Composer 1.
func TestNewCCPPCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	if compiler == nil {
		t.Fatal("NewCCPPCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}
}

// Composer 1.
func TestNewKotlinCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	if compiler == nil {
		t.Fatal("NewKotlinCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}
}

// Composer 1.
func TestNewSwiftCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	if compiler == nil {
		t.Fatal("NewSwiftCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}
}

// Composer 1.
func TestNewZigCompiler(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewZigCompiler(cache)

	if compiler == nil {
		t.Fatal("NewZigCompiler returned nil")
	}

	if compiler.cacheManager != cache {
		t.Error("CacheManager not set correctly")
	}
}

// Mock CacheManager for testing
type mockCacheManager struct{}

func (m *mockCacheManager) GetCacheDir() string {
	return "/tmp/cache"
}

func (m *mockCacheManager) GetCompilerPath(language string) (string, error) {
	return "", nil
}

func (m *mockCacheManager) EnsureCacheDir() error {
	return nil
}

func (m *mockCacheManager) Clear(language string) error {
	return nil
}

func (m *mockCacheManager) ClearAll() error {
	return nil
}

func (m *mockCacheManager) GetCacheSize() (int64, error) {
	return 0, nil
}

func (m *mockCacheManager) GetCompilerSize(language string) (int64, error) {
	return 0, nil
}

func (m *mockCacheManager) IsCompilerCached(language string) bool {
	return false
}

func (m *mockCacheManager) GetCompilerBinaryPath(language string) (string, error) {
	return "", nil
}

// Composer 1.
func TestTinyGoCompiler_ParseCompilationError(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	output := "main.go:5:2: undefined: foo\nmain.go:6:1: syntax error"
	err := compiler.parseCompilationError("compilation failed", output)

	if err == nil {
		t.Fatal("parseCompilationError should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseCompilationError should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}

	if compErr.Traceback != output {
		t.Errorf("Traceback = %q, want %q", compErr.Traceback, output)
	}
}

// Composer 1.
func TestTinyGoCompiler_ParseCompilationError_NoErrorLine(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	output := "some output without error keyword"
	err := compiler.parseCompilationError("compilation failed", output)

	if err == nil {
		t.Fatal("parseCompilationError should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseCompilationError should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestTinyGoCompiler_GenerateMinimalGoMod(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ws, err := NewWorkspace("test-gomod")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	err = compiler.generateMinimalGoMod(ws)
	if err != nil {
		t.Fatalf("generateMinimalGoMod failed: %v", err)
	}

	// Проверяем что go.mod создан
	if !ws.FileExists("go.mod") {
		t.Fatal("go.mod file was not created")
	}

	// Проверяем содержимое
	content, err := ws.ReadFile("go.mod")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(content) == 0 {
		t.Error("go.mod content is empty")
	}

	// Проверяем что содержимое содержит ожидаемые строки
	contentStr := string(content)
	if !contains(contentStr, "module script") {
		t.Error("go.mod should contain 'module script'")
	}

	if !contains(contentStr, "go 1.21") {
		t.Error("go.mod should contain 'go 1.21'")
	}
}

// Composer 1.
func TestRustCompiler_GetRustEnv(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	// Тест без установленных путей
	env := compiler.getRustEnv()
	if len(env) == 0 {
		t.Error("getRustEnv should return at least some environment variables")
	}

	// Устанавливаем пути
	compiler.rustupHome = "/test/rustup"
	compiler.cargoHome = "/test/cargo"

	env = compiler.getRustEnv()
	if len(env) == 0 {
		t.Error("getRustEnv should return environment variables")
	}

	// Проверяем что RUSTUP_HOME установлен
	foundRustupHome := false
	for _, e := range env {
		if contains(e, "RUSTUP_HOME=/test/rustup") {
			foundRustupHome = true
			break
		}
	}
	if !foundRustupHome {
		t.Error("RUSTUP_HOME should be set in environment")
	}

	// Проверяем что CARGO_HOME установлен
	foundCargoHome := false
	for _, e := range env {
		if contains(e, "CARGO_HOME=/test/cargo") {
			foundCargoHome = true
			break
		}
	}
	if !foundCargoHome {
		t.Error("CARGO_HOME should be set in environment")
	}
}

// Composer 1.
func TestRustCompiler_ParseRustError(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	output := "error[E0308]: mismatched types\n --> src/lib.rs:10:5"
	err := compiler.parseRustError(output)

	if err == nil {
		t.Fatal("parseRustError should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseRustError should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}

	if compErr.Traceback != output {
		t.Errorf("Traceback = %q, want %q", compErr.Traceback, output)
	}
}

// Composer 1.
func TestTinyGoCompiler_IsAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	// IsAvailable может вернуть true или false в зависимости от окружения
	// Просто проверяем что метод не падает
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestTinyGoCompiler_GetTinyGoBinary(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	// getTinyGoBinary может вернуть ошибку если tinygo не в кеше
	// Просто проверяем что метод не падает
	_, _ = compiler.getTinyGoBinary()
}

// Composer 1.
func TestRustCompiler_GetRustPaths(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	// getRustPaths может вернуть ошибку если rust не в кеше
	// Просто проверяем что метод не падает
	_, _, _ = compiler.getRustPaths()
}

// Composer 1.
func TestRustCompiler_IsAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	// IsAvailable может вернуть true или false в зависимости от окружения
	// Просто проверяем что метод не падает
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestRustCompiler_ValidateDependencies(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	// Валидный Cargo.toml
	validDeps := map[string]string{
		"Cargo.toml": "[package]\nname = \"test\"\nversion = \"0.1.0\"\n\n[lib]\ncrate-type = [\"cdylib\"]",
	}

	err := compiler.ValidateDependencies(validDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}

	// Невалидный Cargo.toml (нет [package])
	invalidDeps := map[string]string{
		"Cargo.toml": "name = \"test\"",
	}

	err = compiler.ValidateDependencies(invalidDeps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid Cargo.toml")
	}

	// Пустые зависимости
	emptyDeps := map[string]string{}

	err = compiler.ValidateDependencies(emptyDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestRustCompiler_GenerateMinimalCargoToml(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewRustCompiler(cache)

	toml := compiler.generateMinimalCargoToml()

	if toml == "" {
		t.Error("generateMinimalCargoToml() should return non-empty string")
	}

	// Проверяем что содержимое содержит ожидаемые строки
	if !contains(toml, "[package]") {
		t.Error("Cargo.toml should contain '[package]'")
	}

	if !contains(toml, "[lib]") {
		t.Error("Cargo.toml should contain '[lib]'")
	}

	if !contains(toml, "cdylib") {
		t.Error("Cargo.toml should contain 'cdylib'")
	}
}

// Composer 1.
func TestTinyGoCompiler_ValidateDependencies(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	// Валидный go.mod
	validDeps := map[string]string{
		"go.mod": "module test\n\ngo 1.21",
	}

	err := compiler.ValidateDependencies(validDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}

	// Невалидный go.mod (нет module)
	invalidDeps := map[string]string{
		"go.mod": "go 1.21",
	}

	err = compiler.ValidateDependencies(invalidDeps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid go.mod")
	}

	// Пустые зависимости
	emptyDeps := map[string]string{}

	err = compiler.ValidateDependencies(emptyDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ========== AssemblyScript Compiler Tests ==========

func TestAssemblyScriptCompiler_ValidateDependencies(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	// Валидный package.json
	validDeps := map[string]string{
		"package.json": `{"name": "test", "version": "1.0.0"}`,
	}

	err := compiler.ValidateDependencies(validDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}

	// Невалидный JSON
	invalidDeps := map[string]string{
		"package.json": `{invalid json}`,
	}

	err = compiler.ValidateDependencies(invalidDeps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid JSON")
	}

	// Пустые зависимости
	emptyDeps := map[string]string{}

	err = compiler.ValidateDependencies(emptyDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

func TestAssemblyScriptCompiler_ParseCompilationError(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	output := "ERROR TS2304: Cannot find name 'foo'\n  at index.ts:5:10"
	err := compiler.parseCompilationError("compilation failed", output)

	if err == nil {
		t.Fatal("parseCompilationError should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseCompilationError should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}

	if compErr.Traceback != output {
		t.Errorf("Traceback = %q, want %q", compErr.Traceback, output)
	}
}

func TestAssemblyScriptCompiler_GenerateMinimalPackageJSON(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewAssemblyScriptCompiler(cache)

	ws, err := NewWorkspace("test-package-json")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	err = compiler.generateMinimalPackageJSON(ws)
	if err != nil {
		t.Fatalf("generateMinimalPackageJSON failed: %v", err)
	}

	if !ws.FileExists("package.json") {
		t.Fatal("package.json file was not created")
	}

	content, err := ws.ReadFile("package.json")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "assemblyscript") {
		t.Error("package.json should contain 'assemblyscript'")
	}
}

// ========== C/C++ Compiler Tests ==========

func TestCCPPCompiler_ValidateDependencies(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	// Валидные зависимости (header файлы)
	validDeps := map[string]string{
		"header.h":   "#define TEST 1",
		"header.hpp": "#include <iostream>",
		"utils.c":    "int test() { return 0; }",
		"utils.cpp":  "int test() { return 0; }",
	}

	err := compiler.ValidateDependencies(validDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}

	// Невалидное расширение
	invalidDeps := map[string]string{
		"file.txt": "some content",
	}

	err = compiler.ValidateDependencies(invalidDeps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid extension")
	}

	// Пустые зависимости
	emptyDeps := map[string]string{}

	err = compiler.ValidateDependencies(emptyDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// ========== Python Compiler Tests ==========

func TestPythonCompiler_ValidateDependencies(t *testing.T) {
	compiler := NewPythonCompiler()

	// Валидные зависимости
	validDeps := map[string]string{
		"utils.py":         "def test(): pass",
		"requirements.txt": "requests==1.0.0",
	}

	err := compiler.ValidateDependencies(validDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}

	// Невалидная зависимость
	invalidDeps := map[string]string{
		"file.txt": "some content",
	}

	err = compiler.ValidateDependencies(invalidDeps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid dependency")
	}

	// Пустые зависимости
	emptyDeps := map[string]string{}

	err = compiler.ValidateDependencies(emptyDeps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

func TestPythonCompiler_ValidateSyntax_EmptyCode(t *testing.T) {
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

func TestPythonCompiler_ValidateSyntax_BalancedBrackets(t *testing.T) {
	compiler := NewPythonCompiler()
	ctx := context.Background()

	// Несбалансированные скобки
	req := domain.CompileRequest{
		SourceCode: "def test():\n    if True:\n        pass\n)",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error for unbalanced brackets")
	}
}
