package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewCompilationService(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	if service == nil {
		t.Fatal("NewCompilationService returned nil")
	}

	if service.repo != repo {
		t.Error("Repository not set correctly")
	}

	if service.compilers == nil {
		t.Error("Compilers map not initialized")
	}

	if len(service.compilers) != 0 {
		t.Errorf("Expected empty compilers map, got %d", len(service.compilers))
	}
}

// Composer 1.
func TestCompilationService_RegisterCompiler(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	compiler := &mockCompiler{language: "rust", available: true}
	service.RegisterCompiler(compiler)

	if len(service.compilers) != 1 {
		t.Errorf("Expected 1 compiler, got %d", len(service.compilers))
	}

	if service.compilers["rust"] != compiler {
		t.Error("Compiler not registered correctly")
	}

	compiler2 := &mockCompiler{language: "go", available: false}
	service.RegisterCompiler(compiler2)

	if len(service.compilers) != 2 {
		t.Errorf("Expected 2 compilers, got %d", len(service.compilers))
	}
}

// Composer 1.
func TestCompilationService_CompileScript_Success(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{
		language:  "rust",
		available: true,
		compileResult: &domain.CompileResult{
			WASMBinary: []byte{1, 2, 3},
			WASMSize:   3,
			Duration:   100 * time.Millisecond,
		},
	}

	service.RegisterCompiler(compiler)

	ctx := context.Background()
	result, err := service.CompileScript(ctx, "test-script", false)

	if err != nil {
		t.Errorf("CompileScript() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("CompileScript() returned nil result")
	}

	if len(result.WASMBinary) != 3 {
		t.Errorf("WASMBinary size = %d, want 3", len(result.WASMBinary))
	}

	if script.Code == nil {
		t.Error("Script WASMBinary not updated")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_ScriptNotFound(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "nonexistent", false)

	if err == nil {
		t.Error("CompileScript() should return error when script not found")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_NoSourceCode(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:       "test-script",
		Name:     "Test Script",
		Language: "rust",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{language: "rust", available: true}
	service.RegisterCompiler(compiler)

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "test-script", false)

	if err == nil {
		t.Error("CompileScript() should return error when script has no source code")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_UnsupportedLanguage(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "python",
		SourceCode: "print('hello')",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "test-script", false)

	if err == nil {
		t.Error("CompileScript() should return error for unsupported language")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_CompilerNotAvailable(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{language: "rust", available: false}
	service.RegisterCompiler(compiler)

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "test-script", false)

	if err == nil {
		t.Error("CompileScript() should return error when compiler not available")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_CompilationError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "invalid code",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{
		language:     "rust",
		available:    true,
		compileError: errors.New("compilation failed"),
	}

	service.RegisterCompiler(compiler)

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "test-script", false)

	if err == nil {
		t.Error("CompileScript() should return error when compilation fails")
	}

	if script.CompilationError == "" {
		t.Error("Script should be marked with compilation error")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_WithDependencies(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
		Dependencies: map[string]string{
			"Cargo.toml": "[package]\nname = \"test\"",
		},
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{
		language:  "rust",
		available: true,
		compileResult: &domain.CompileResult{
			WASMBinary: []byte{1, 2, 3},
			WASMSize:   3,
			Duration:   100 * time.Millisecond,
		},
	}

	service.RegisterCompiler(compiler)

	ctx := context.Background()
	result, err := service.CompileScript(ctx, "test-script", false)

	if err != nil {
		t.Errorf("CompileScript() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("CompileScript() returned nil result")
	}
}

// Composer 1.
func TestCompilationService_ValidateSyntax_Success(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	compiler := &mockCompiler{language: "rust", available: true}
	service.RegisterCompiler(compiler)

	err := service.ValidateSyntax(context.Background(), script)

	if err != nil {
		t.Errorf("ValidateSyntax() error = %v, want nil", err)
	}
}

// Composer 1.
func TestCompilationService_ValidateSyntax_NoSourceCode(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		Language: "rust",
	}

	err := service.ValidateSyntax(context.Background(), script)

	if err == nil {
		t.Error("ValidateSyntax() should return error when no source code")
	}
}

// Composer 1.
func TestCompilationService_ValidateSyntax_UnsupportedLanguage(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		Language:   "python",
		SourceCode: "print('hello')",
	}

	err := service.ValidateSyntax(context.Background(), script)

	if err == nil {
		t.Error("ValidateSyntax() should return error for unsupported language")
	}
}

// Composer 1.
func TestCompilationService_ValidateSyntax_CompilerNotAvailable(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	compiler := &mockCompiler{language: "rust", available: false}
	service.RegisterCompiler(compiler)

	err := service.ValidateSyntax(context.Background(), script)

	if err == nil {
		t.Error("ValidateSyntax() should return error when compiler not available")
	}
}

// Composer 1.
func TestCompilationService_ValidateDependencies_Success(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	compiler := &mockCompiler{language: "rust", available: true}
	service.RegisterCompiler(compiler)

	deps := map[string]string{
		"Cargo.toml": "[package]\nname = \"test\"",
	}

	err := service.ValidateDependencies(context.Background(), "rust", deps)

	if err != nil {
		t.Errorf("ValidateDependencies() error = %v, want nil", err)
	}
}

// Composer 1.
func TestCompilationService_ValidateDependencies_UnsupportedLanguage(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	deps := map[string]string{
		"package.json": "{}",
	}

	err := service.ValidateDependencies(context.Background(), "javascript", deps)

	if err == nil {
		t.Error("ValidateDependencies() should return error for unsupported language")
	}
}

// Composer 1.
func TestCompilationService_GetAvailableCompilers(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	compiler1 := &mockCompiler{language: "rust", available: true}
	compiler2 := &mockCompiler{language: "go", available: false}
	compiler3 := &mockCompiler{language: "zig", available: true}

	service.RegisterCompiler(compiler1)
	service.RegisterCompiler(compiler2)
	service.RegisterCompiler(compiler3)

	available := service.GetAvailableCompilers()

	if len(available) != 2 {
		t.Errorf("GetAvailableCompilers() returned %d compilers, want 2", len(available))
	}

	hasRust := false
	hasZig := false
	for _, lang := range available {
		if lang == "rust" {
			hasRust = true
		}
		if lang == "zig" {
			hasZig = true
		}
	}

	if !hasRust {
		t.Error("GetAvailableCompilers() should include rust")
	}

	if !hasZig {
		t.Error("GetAvailableCompilers() should include zig")
	}
}

// Composer 1.
func TestCompilationService_GetAllCompilers(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	compiler1 := &mockCompiler{language: "rust", available: true}
	compiler2 := &mockCompiler{language: "go", available: false}

	service.RegisterCompiler(compiler1)
	service.RegisterCompiler(compiler2)

	all := service.GetAllCompilers()

	if len(all) != 2 {
		t.Errorf("GetAllCompilers() returned %d compilers, want 2", len(all))
	}

	if all["rust"] != true {
		t.Error("GetAllCompilers() should mark rust as available")
	}

	if all["go"] != false {
		t.Error("GetAllCompilers() should mark go as not available")
	}
}

// Composer 1.
func TestCompilationService_IsCompilerAvailable(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	compiler1 := &mockCompiler{language: "rust", available: true}
	compiler2 := &mockCompiler{language: "go", available: false}

	service.RegisterCompiler(compiler1)
	service.RegisterCompiler(compiler2)

	if !service.IsCompilerAvailable("rust") {
		t.Error("IsCompilerAvailable('rust') should return true")
	}

	if service.IsCompilerAvailable("go") {
		t.Error("IsCompilerAvailable('go') should return false")
	}

	if service.IsCompilerAvailable("nonexistent") {
		t.Error("IsCompilerAvailable('nonexistent') should return false")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_SaveErrorOnMarkCompiling(t *testing.T) {
	repo := &mockScriptRepository{
		saveError: errors.New("save failed"),
	}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{
		language:  "rust",
		available: true,
		compileResult: &domain.CompileResult{
			WASMBinary: []byte{1, 2, 3},
			WASMSize:   3,
			Duration:   100 * time.Millisecond,
		},
	}

	service.RegisterCompiler(compiler)

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "test-script", false)

	if err == nil {
		t.Error("CompileScript() should return error when save fails on MarkCompiling")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_SaveErrorOnMarkCompilationSuccess(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{
		language:  "rust",
		available: true,
		compileResult: &domain.CompileResult{
			WASMBinary: []byte{1, 2, 3},
			WASMSize:   3,
			Duration:   100 * time.Millisecond,
		},
	}

	service.RegisterCompiler(compiler)

	repo.saveError = errors.New("save failed")

	ctx := context.Background()
	_, err := service.CompileScript(ctx, "test-script", false)

	if err == nil {
		t.Error("CompileScript() should return error when save fails on MarkCompilationSuccess")
	}
}

// Composer 1.
func TestCompilationService_CompileScript_WithOptimize(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewCompilationService(repo, nil)

	script := &domain.Script{
		ID:         "test-script",
		Name:       "Test Script",
		Language:   "rust",
		SourceCode: "fn main() {}",
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	compiler := &mockCompiler{
		language:  "rust",
		available: true,
		compileResult: &domain.CompileResult{
			WASMBinary: []byte{1, 2, 3},
			WASMSize:   3,
			Duration:   100 * time.Millisecond,
		},
	}

	service.RegisterCompiler(compiler)

	ctx := context.Background()
	result, err := service.CompileScript(ctx, "test-script", true)

	if err != nil {
		t.Errorf("CompileScript() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("CompileScript() returned nil result")
	}
}

// Mock implementations
type mockCompiler struct {
	language      string
	available     bool
	compileResult *domain.CompileResult
	compileError  error
	validateError error
}

func (m *mockCompiler) Language() string {
	return m.language
}

func (m *mockCompiler) IsAvailable() bool {
	return m.available
}

func (m *mockCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	if m.compileError != nil {
		return nil, m.compileError
	}
	return m.compileResult, nil
}

func (m *mockCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	return m.validateError
}

func (m *mockCompiler) ValidateDependencies(deps map[string]string) error {
	return m.validateError
}
