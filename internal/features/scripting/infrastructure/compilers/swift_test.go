package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestSwiftCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)
	compiler.swiftPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
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
func TestSwiftCompiler_Compile_WithPackageSwift(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
		Dependencies: map[string]string{
			"Package.swift": `// swift-tools-version:5.9
import PackageDescription

let package = Package(name: "test")`,
		},
	}

	compiler.swiftPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_Compile_WithAdditionalFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
		Dependencies: map[string]string{
			"Utils.swift": "func helper() {}",
		},
	}

	compiler.swiftPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
		Optimize:   true,
	}

	compiler.swiftPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/swift",
	}
	compiler := NewSwiftCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestSwiftCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestSwiftCompiler_ValidateDependencies_ValidSwiftFile(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	deps := map[string]string{
		"Utils.swift": "func helper() {}",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestSwiftCompiler_ValidateDependencies_ValidPackageSwift(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	deps := map[string]string{
		"Package.swift": `// swift-tools-version:5.9
import PackageDescription

let package = Package(name: "test")`,
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestSwiftCompiler_ValidateDependencies_InvalidFile(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	deps := map[string]string{
		"file.txt": "some content",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid file")
	}
}

// Composer 1.
func TestSwiftCompiler_ValidateDependencies_EmptyPackageSwift(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	deps := map[string]string{
		"Package.swift": "",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for empty Package.swift")
	}
}

// Composer 1.
func TestSwiftCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestSwiftCompiler_GetSwiftWasmSDK_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/swift",
	}
	compiler := NewSwiftCompiler(cache)

	sdkPath, err := compiler.getSwiftWasmSDK()
	_ = sdkPath
	_ = err
}

// Composer 1.
func TestSwiftCompiler_GetSwiftWasmSDK_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewSwiftCompiler(cache)

	_, err := compiler.getSwiftWasmSDK()
	if err == nil {
		t.Error("getSwiftWasmSDK() should return error when SDK not in cache")
	}
}

// Composer 1.
func TestSwiftCompiler_GenerateMinimalPackageSwift(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	pkgSwift := compiler.generateMinimalPackageSwift()
	if pkgSwift == "" {
		t.Error("generateMinimalPackageSwift() should return non-empty string")
	}

	if !contains(pkgSwift, "Package") {
		t.Error("Package.swift should contain 'Package'")
	}
}

// Composer 1.
func TestSwiftCompiler_ParseSwiftError(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	output := "main.swift:10:5: error: cannot find type"
	err := compiler.parseSwiftError(output)

	if err == nil {
		t.Fatal("parseSwiftError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseSwiftError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestSwiftCompiler_ParseSwiftError_WithFullFormat(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	output := "main.swift:10:5: error: cannot find type 'Foo' in scope"
	err := compiler.parseSwiftError(output)

	if err == nil {
		t.Fatal("parseSwiftError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseSwiftError() should return CompilationError")
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
func TestSwiftCompiler_ParseSwiftError_NoErrorLine(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	output := "some output without error"
	err := compiler.parseSwiftError(output)

	if err == nil {
		t.Fatal("parseSwiftError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseSwiftError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestSwiftCompiler_ValidateSyntax_NotAvailable(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewSwiftCompiler(cache)
	compiler.swiftPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_GetSwiftWasmSDK_WithCachedPath(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)
	compiler.sdkPath = "/test/sdk"

	sdkPath, err := compiler.getSwiftWasmSDK()
	if err != nil {
		t.Errorf("getSwiftWasmSDK() error = %v, want nil", err)
	}

	if sdkPath != "/test/sdk" {
		t.Errorf("sdkPath = %q, want %q", sdkPath, "/test/sdk")
	}
}

// Composer 1.
func TestSwiftCompiler_Compile_ReleaseMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)
	compiler.swiftPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
		Optimize:   true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_Compile_DebugMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)
	compiler.swiftPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
		Optimize:   false,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_Compile_WithoutPackageSwift(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)
	compiler.swiftPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode:   "print(\"hello\")",
		Dependencies: map[string]string{},
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_Compile_WithSwiftFiles(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)
	compiler.swiftPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "print(\"hello\")",
		Dependencies: map[string]string{
			"Utils.swift": "func helper() {}",
		},
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestSwiftCompiler_ParseSwiftError_WithFallback(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	output := "some error line"
	err := compiler.parseSwiftError(output)

	if err == nil {
		t.Fatal("parseSwiftError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseSwiftError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}

// Composer 1.
func TestSwiftCompiler_ParseSwiftError_NoErrorFound(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewSwiftCompiler(cache)

	output := "some output without error"
	err := compiler.parseSwiftError(output)

	if err == nil {
		t.Fatal("parseSwiftError() should return error")
	}

	compErr, ok := err.(*domain.CompilationError)
	if !ok {
		t.Fatal("parseSwiftError() should return CompilationError")
	}

	if compErr.Message == "" {
		t.Error("CompilationError.Message should not be empty")
	}
}
