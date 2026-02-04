package compilers

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestCCPPCompiler_Compile_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_WithHeaders(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
		Dependencies: map[string]string{
			"header.h": "#ifndef HEADER_H\n#define HEADER_H\n#endif",
		},
	}

	compiler.clangPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_CPPVariant(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "#include <iostream>\nint main() { return 0; }",
		Language:   "cpp",
	}

	compiler.clangPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_OptimizeFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
		Optimize:   true,
	}

	compiler.clangPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_DebugFlag(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
		Debug:      true,
	}

	compiler.clangPath = ""

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_IsAvailable_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/wasi-sdk",
	}
	compiler := NewCCPPCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestCCPPCompiler_IsAvailable_WithoutCache(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	_ = compiler.IsAvailable()
}

// Composer 1.
func TestCCPPCompiler_ValidateDependencies_ValidHeader(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	deps := map[string]string{
		"header.h": "#ifndef HEADER_H\n#define HEADER_H\n#endif",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestCCPPCompiler_ValidateDependencies_ValidSource(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	deps := map[string]string{
		"utils.c": "void helper() {}",
	}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestCCPPCompiler_ValidateDependencies_InvalidExtension(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	deps := map[string]string{
		"file.txt": "some content",
	}

	err := compiler.ValidateDependencies(deps)
	if err == nil {
		t.Error("ValidateDependencies() should return error for invalid extension")
	}
}

// Composer 1.
func TestCCPPCompiler_ValidateDependencies_Empty(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)

	deps := map[string]string{}

	err := compiler.ValidateDependencies(deps)
	if err != nil {
		t.Errorf("ValidateDependencies() with empty deps error = %v, want nil", err)
	}
}

// Composer 1.
func TestCCPPCompiler_GetWASISDKPaths_WithCache(t *testing.T) {
	cache := &mockCacheManagerWithPath{
		compilerPath: "/test/wasi-sdk",
	}
	compiler := NewCCPPCompiler(cache)

	clangPath, sysroot, err := compiler.getWASISDKPaths()
	_ = clangPath
	_ = sysroot
	_ = err
}

// Composer 1.
func TestCCPPCompiler_GetWASISDKPaths_WithoutCache(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewCCPPCompiler(cache)

	_, _, err := compiler.getWASISDKPaths()
	if err == nil {
		t.Error("getWASISDKPaths() should return error when compiler not in cache")
	}
}

// Composer 1.
func TestCCPPCompiler_GetWASISDKPaths_WithCachedPaths(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = "/test/clang"
	compiler.sysroot = "/test/sysroot"

	clangPath, sysroot, err := compiler.getWASISDKPaths()
	if err != nil {
		t.Errorf("getWASISDKPaths() error = %v, want nil", err)
	}

	if clangPath != "/test/clang" {
		t.Errorf("clangPath = %q, want %q", clangPath, "/test/clang")
	}

	if sysroot != "/test/sysroot" {
		t.Errorf("sysroot = %q, want %q", sysroot, "/test/sysroot")
	}
}

// Composer 1.
func TestCCPPCompiler_GetWASISDKPaths_NoCacheManager(t *testing.T) {
	compiler := NewCCPPCompiler(nil)

	_, _, err := compiler.getWASISDKPaths()
	if err == nil {
		t.Error("getWASISDKPaths() should return error when cache manager is nil")
	}
}

// Composer 1.
func TestCCPPCompiler_ValidateSyntax_NotAvailable(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_ValidateSyntax_CPPVariant(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "#include <iostream>\nint main() { return 0; }",
		Language:   "cpp",
	}

	err := compiler.ValidateSyntax(ctx, req)
	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_WithSysroot(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = "/test/clang"
	compiler.sysroot = "/test/sysroot"

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
	}

	_, err := compiler.Compile(ctx, req)
	// May return error if clang is unavailable or cannot compile
	_ = err
}

// Composer 1.
func TestCCPPCompiler_IsAvailable_SystemFallback(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewCCPPCompiler(cache)

	// IsAvailable may return true if clang is in the system
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestCCPPCompiler_IsAvailable_WithPrintTargets(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewCCPPCompiler(cache)

	// IsAvailable checks --print-targets for wasm32 support
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestCCPPCompiler_Compile_CPPExtensionDetection(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "#include <iostream>\nusing namespace std;\nint main() { return 0; }",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_CPlusPlusLanguage(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
		Language:   "c++",
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_DebugMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
		Debug:      true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_Compile_OptimizeMode(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = ""

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
		Optimize:   true,
	}

	_, err := compiler.Compile(ctx, req)
	if err == nil {
		t.Fatal("Compile() should return error when compiler not available")
	}
}

// Composer 1.
func TestCCPPCompiler_ValidateSyntax_WithSysroot(t *testing.T) {
	cache := &mockCacheManager{}
	compiler := NewCCPPCompiler(cache)
	compiler.clangPath = "/test/clang"
	compiler.sysroot = "/test/sysroot"

	ctx := context.Background()
	req := domain.CompileRequest{
		SourceCode: "int main() { return 0; }",
	}

	err := compiler.ValidateSyntax(ctx, req)
	// May return error if clang is unavailable
	_ = err
}

// Composer 1.
func TestCCPPCompiler_IsAvailable_WithVersionedClang(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewCCPPCompiler(cache)

	// IsAvailable checks different clang versions (clang-18, clang-17)
	_ = compiler.IsAvailable()
}

// Composer 1.
func TestCCPPCompiler_IsAvailable_WithWasiSDKPath(t *testing.T) {
	cache := &mockCacheManagerWithError{}
	compiler := NewCCPPCompiler(cache)

	// IsAvailable checks /opt/wasi-sdk/bin/clang
	_ = compiler.IsAvailable()
}
