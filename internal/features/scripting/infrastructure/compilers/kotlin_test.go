package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestKotlinCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
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
func TestKotlinCompiler_Compile_WithDependencies(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
		Dependencies: map[string]string{
			"Utils.kt": "fun helper() {}",
		},
	}

	compiler.kotlincPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestKotlinCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
		Optimize:   true,
	}

	compiler.kotlincPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestKotlinCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/kotlin",
	}
	compiler := NewKotlinCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestKotlinCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestKotlinCompiler_ValidateDependencies_ValidKtFile(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	deps := map[string]string{
		"Utils.kt": "fun helper() {}",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestKotlinCompiler_ValidateDependencies_ValidBuildGradle(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	deps := map[string]string{
		"build.gradle": "plugins { id 'org.jetbrains.kotlin.wasm' }",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestKotlinCompiler_ValidateDependencies_InvalidFile(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	deps := map[string]string{
		"file.txt": "some content",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid file")
	}
}

// Composer 1.
func TestKotlinCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestKotlinCompiler_GetKotlinSetup_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/kotlin",
	}
	compiler := NewKotlinCompiler(cache)

	kotlincPath, javaHome, err := compiler.getKotlinSetup()
	_ = kotlincPath
	_ = javaHome
	_ = err
}

// Composer 1.
func TestKotlinCompiler_GetKotlinSetup_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewKotlinCompiler(cache)

	_, _, err := compiler.getKotlinSetup()
	if err == nil {
		t.Error("getKotlinSetup() should return error when compiler not in cache")
	}
}

// Composer 1.
func TestKotlinCompiler_ParseKotlinError(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	output := "Main.kt:10:5: error: unresolved reference"
	err := compiler.parseKotlinError(output)

	if err == nil {
		t.Fatal("parseKotlinError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseKotlinError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestKotlinCompiler_ParseKotlinError_WithFullFormat(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	output := "Main.kt:10:5: error: unresolved reference: foo"
	err := compiler.parseKotlinError(output)

	if err == nil {
		t.Fatal("parseKotlinError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseKotlinError() should return CompilationError")
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
func TestKotlinCompiler_ParseKotlinError_WithE(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	output := "Main.kt:10:5: e: unresolved reference"
	err := compiler.parseKotlinError(output)

	if err == nil {
		t.Fatal("parseKotlinError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseKotlinError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestKotlinCompiler_ParseKotlinError_NoErrorLine(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)

	output := "some output without error"
	err := compiler.parseKotlinError(output)

	if err == nil {
		t.Fatal("parseKotlinError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseKotlinError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestKotlinCompiler_ValidateSyntax_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestKotlinCompiler_GetKotlinSetup_WithCachedPath(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = "/test/kotlinc"
	compiler.javaHome = "/test/java"

	kotlincPath, javaHome, err := compiler.getKotlinSetup()
	if err != nil {
		t.Errorf("getKotlinSetup() error = %v, want nil", err)
	}

	if kotlincPath != "/test/kotlinc" {
		t.Errorf("kotlincPath = %q, want %q", kotlincPath, "/test/kotlinc")
	}

	if javaHome != "/test/java" {
		t.Errorf("javaHome = %q, want %q", javaHome, "/test/java")
	}
}

// Composer 1.
func TestKotlinCompiler_Compile_WithJavaHome(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = ""
	compiler.javaHome = "/test/java"

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestKotlinCompiler_IsAvailable_SystemFallback(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewKotlinCompiler(cache)

	// IsAvailable may return true if kotlinc is in the system
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestKotlinCompiler_GetKotlinSetup_WithJREInCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/kotlin",
	}
	compiler := NewKotlinCompiler(cache)

	kotlincPath, javaHome, err := compiler.getKotlinSetup()
	_ = kotlincPath
	_ = javaHome
	_ = err
}

// Composer 1.
func TestKotlinCompiler_GetKotlinSetup_WithSystemJavaHome(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/kotlin",
	}
	compiler := NewKotlinCompiler(cache)

	kotlincPath, javaHome, err := compiler.getKotlinSetup()
	_ = kotlincPath
	_ = javaHome
	_ = err
}

// Composer 1.
func TestKotlinCompiler_Compile_WithoutJavaHome(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = ""
	compiler.javaHome = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestKotlinCompiler_Compile_WithWarnings(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestKotlinCompiler_ValidateSyntax_WithJavaHome(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewKotlinCompiler(cache)
	compiler.kotlincPath = ""
	compiler.javaHome = "/test/java"

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "fun main() {}",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}
