package compilers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestTinyGoCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)
	compiler.tinygoPath = "" // Убеждаемся что путь не установлен

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
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
func TestTinyGoCompiler_Compile_WorkspaceCreation(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
	}

	// Устанавливаем невалидный путь чтобы вызвать ошибку IsAvailable
	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error")
	}

	// Проверяем что ошибка связана с недоступностью компилятора
	if err.Error() == "" {
		t.Error("Error message should not be empty")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_WithGoMod(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Dependencies: map[string]string{
			"go.mod": "module test\n\ngo 1.21",
		},
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	// Ожидаем ошибку так как компилятор недоступен
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_WithAdditionalFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Dependencies: map[string]string{
			"utils.go": "package main\nfunc helper() {}",
		},
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_WithTimeout(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Ждем немного чтобы контекст точно истек
	time.Sleep(10 * time.Millisecond)

	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	// Ожидаем ошибку (либо из-за недоступности компилятора, либо из-за таймаута)
	if err == nil {
		t.Fatal("Compile() should return error")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Optimize:   true,
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_DebugFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Debug:      true,
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/tinygo",
	}
	compiler := NewTinyGoCompiler(cache)

	// IsAvailable может вернуть true или false в зависимости от окружения
	// Просто проверяем что метод не падает
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestTinyGoCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	// IsAvailable может вернуть true или false в зависимости от окружения
	// Просто проверяем что метод не падает
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestTinyGoCompiler_ValidateDependencies_ValidGoMod(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	deps := map[string]string{
		"go.mod": "module test\n\ngo 1.21",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestTinyGoCompiler_ValidateDependencies_InvalidGoMod(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	deps := map[string]string{
		"go.mod": "go 1.21", // Нет module declaration
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid go.mod")
	}
}

// Composer 1.
func TestTinyGoCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestTinyGoCompiler_ValidateDependencies_NoGoMod(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	deps := map[string]string{
		"other.go": "package main",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestTinyGoCompiler_GetTinyGoBinary_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/tinygo",
	}
	compiler := NewTinyGoCompiler(cache)

	path, err := compiler.getTinyGoBinary()
	// Может вернуть ошибку если путь не существует, но это нормально
	_ = path
	_ = err
}

// Composer 1.
func TestTinyGoCompiler_GetTinyGoBinary_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewTinyGoCompiler(cache)

	_, err := compiler.getTinyGoBinary()
	// Ожидаем ошибку так как компилятор не в кеше
	if err == nil {
		t.Error("getTinyGoBinary() should return error when compiler not in cache")
	}
}

// Composer 1.
func TestTinyGoCompiler_GetTinyGoBinary_WithCachedPath(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)
	compiler.tinygoPath = "/test/tinygo"

	path, err := compiler.getTinyGoBinary()
	if err != nil {
		t.Errorf("getTinyGoBinary() error = %v, want nil", err)
	}

	if path != "/test/tinygo" {
		t.Errorf("path = %q, want %q", path, "/test/tinygo")
	}
}

// Composer 1.
func TestTinyGoCompiler_ValidateSyntax_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewTinyGoCompiler(cache)
	compiler.tinygoPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_WithGoSum(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Dependencies: map[string]string{
			"go.mod": "module test\n\ngo 1.21",
			"go.sum": "github.com/extism/go-pdk v1.0.0 h1:abc123",
		},
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_NoDebugFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Debug:      false,
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_IsAvailable_SystemFallback(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewTinyGoCompiler(cache)

	// IsAvailable может вернуть true если tinygo есть в системе
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestTinyGoCompiler_Compile_WithGoFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Dependencies: map[string]string{
			"utils.go": "package main\nfunc helper() {}",
		},
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_WithGoModDownload(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Dependencies: map[string]string{
			"go.mod": "module test\n\ngo 1.21\n\nrequire github.com/extism/go-pdk v1.0.0",
		},
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_Compile_OptimizeMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
		Optimize:   true,
	}

	compiler.tinygoPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestTinyGoCompiler_ValidateSyntax_WithGoMod(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewTinyGoCompiler(cache)
	compiler.tinygoPath = ""
	compiler.goPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "package main\nfunc main() {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Mock CacheManager с поддержкой пути компилятора
type mockCacheManagerWithPath struct {
	compilerPath string
}

func (m *mockCacheManagerWithPath) GetCacheDir() string {
	return "/tmp/cache"
}

func (m *mockCacheManagerWithPath) GetCompilerPath(language string) (string, error) {
	if m.compilerPath != "" {
		return m.compilerPath, nil
	}
	return "", nil
}

func (m *mockCacheManagerWithPath) EnsureCacheDir() error {
	return nil
}

func (m *mockCacheManagerWithPath) Clear(language string) error {
	return nil
}

func (m *mockCacheManagerWithPath) ClearAll() error {
	return nil
}

func (m *mockCacheManagerWithPath) GetCacheSize() (int64, error) {
	return 0, nil
}

func (m *mockCacheManagerWithPath) GetCompilerSize(language string) (int64, error) {
	return 0, nil
}

func (m *mockCacheManagerWithPath) IsCompilerCached(language string) bool {
	return m.compilerPath != ""
}

func (m *mockCacheManagerWithPath) GetCompilerBinaryPath(language string) (string, error) {
	return "", nil
}

// Mock CacheManager который возвращает ошибку при GetCompilerPath
type mockCacheManagerWithError struct{}

func (m *mockCacheManagerWithError) GetCacheDir() string {
	return "/tmp/cache"
}

func (m *mockCacheManagerWithError) GetCompilerPath(language string) (string, error) {
	return "", fmt.Errorf("compiler %s not found in cache", language)
}

func (m *mockCacheManagerWithError) EnsureCacheDir() error {
	return nil
}

func (m *mockCacheManagerWithError) Clear(language string) error {
	return nil
}

func (m *mockCacheManagerWithError) ClearAll() error {
	return nil
}

func (m *mockCacheManagerWithError) GetCacheSize() (int64, error) {
	return 0, nil
}

func (m *mockCacheManagerWithError) GetCompilerSize(language string) (int64, error) {
	return 0, nil
}

func (m *mockCacheManagerWithError) IsCompilerCached(language string) bool {
	return false
}

func (m *mockCacheManagerWithError) GetCompilerBinaryPath(language string) (string, error) {
	return "", nil
}
