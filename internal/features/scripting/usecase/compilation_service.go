package usecase

import (
	"context"
	"fmt"
	"log"

	"network-debugger/internal/features/scripting/domain"
)

// CompilationService orchestrates script compilation (USE CASE in Clean Architecture)
// This service coordinates between domain and infrastructure layers
type CompilationService struct {
	compilers     map[string]domain.Compiler // Registered compiler implementations (ADAPTERS)
	repo          domain.ScriptRepository    // Script persistence (PORT)
	wasmValidator domain.WASMValidator       // WASM validation (PORT)
}

// NewCompilationService creates a new compilation service
// Dependency Injection: receives repository through constructor (Dependency Inversion Principle)
func NewCompilationService(repo domain.ScriptRepository, wasmValidator domain.WASMValidator) *CompilationService {
	return &CompilationService{
		compilers:     make(map[string]domain.Compiler),
		repo:          repo,
		wasmValidator: wasmValidator,
	}
}

// RegisterCompiler registers a compiler implementation (ADAPTER)
// This allows for runtime registration of different compilers
// Follows Open/Closed Principle: open for extension, closed for modification
func (s *CompilationService) RegisterCompiler(compiler domain.Compiler) {
	language := compiler.Language()
	s.compilers[language] = compiler
	log.Printf("[CompilationService] Registered compiler for language: %s (available: %v)", language, compiler.IsAvailable())
}

// CompileScript compiles a script and updates it in repository
// This is the main use case for script compilation
func (s *CompilationService) CompileScript(ctx context.Context, scriptID string, optimize bool) (*domain.CompileResult, error) {
	// Load script (read from repository)
	script, err := s.repo.Get(ctx, scriptID)
	if err != nil {
		return nil, fmt.Errorf("load script: %w", err)
	}

	// Validate preconditions - either single file (SourceCode) or multi-file (Dependencies)
	if script.SourceCode == "" && len(script.Dependencies) == 0 {
		return nil, fmt.Errorf("script has no source code or dependencies to compile")
	}

	// Get compiler for language
	compiler, ok := s.compilers[script.Language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s (available: %v)", script.Language, s.GetAvailableCompilers())
	}

	// Check if compiler toolchain is available
	if !compiler.IsAvailable() {
		return nil, fmt.Errorf("compiler for %s is not available (toolchain not installed)", script.Language)
	}

	// Mark script as compiling (domain behavior)
	script.MarkCompiling()
	if err := s.repo.Save(ctx, script); err != nil {
		return nil, fmt.Errorf("update compilation status: %w", err)
	}

	// Prepare compilation request (Value Object)
	req := domain.CompileRequest{
		SourceCode:   script.SourceCode,
		Dependencies: script.Dependencies,
		Language:     script.Language,
		Optimize:     optimize,
	}

	// Execute compilation (delegate to compiler ADAPTER)
	result, err := compiler.Compile(ctx, req)
	if err != nil {
		// Mark compilation as failed (domain behavior)
		script.MarkCompilationError(err)
		s.repo.Save(ctx, script)
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	// Validate WASM exports (check for 'process' function)
	if s.wasmValidator != nil {
		if err := s.wasmValidator.ValidateWASM(result.WASMBinary, script.Language); err != nil {
			script.MarkCompilationError(err)
			script.MarkValidationError(err)
			s.repo.Save(ctx, script)
			return nil, fmt.Errorf("WASM validation failed: %w", err)
		}
	}

	// Mark script as successfully compiled and validated (domain behavior)
	script.MarkCompilationSuccess(result.WASMBinary)
	script.MarkValidationSuccess()
	if err := s.repo.Save(ctx, script); err != nil {
		return nil, fmt.Errorf("save compiled script: %w", err)
	}

	log.Printf("[CompilationService] Successfully compiled script %s (%s) in %v, size: %d bytes",
		script.ID, script.Name, result.Duration, result.WASMSize)

	return result, nil
}

// ValidateSyntax validates script syntax without full compilation
// Useful for providing quick feedback in the editor
func (s *CompilationService) ValidateSyntax(ctx context.Context, script *domain.Script) error {
	if script.SourceCode == "" {
		return fmt.Errorf("no source code to validate")
	}

	compiler, ok := s.compilers[script.Language]
	if !ok {
		return fmt.Errorf("unsupported language: %s", script.Language)
	}

	if !compiler.IsAvailable() {
		return fmt.Errorf("compiler for %s is not available", script.Language)
	}

	req := domain.CompileRequest{
		SourceCode:   script.SourceCode,
		Dependencies: script.Dependencies,
		Language:     script.Language,
	}

	return compiler.ValidateSyntax(ctx, req)
}

// ValidateDependencies validates dependency manifest structure
func (s *CompilationService) ValidateDependencies(ctx context.Context, language string, deps map[string]string) error {
	compiler, ok := s.compilers[language]
	if !ok {
		return fmt.Errorf("unsupported language: %s", language)
	}

	return compiler.ValidateDependencies(deps)
}

// GetAvailableCompilers returns list of available compiler languages
// Only includes compilers with installed toolchains
func (s *CompilationService) GetAvailableCompilers() []string {
	var available []string
	for lang, compiler := range s.compilers {
		if compiler.IsAvailable() {
			available = append(available, lang)
		}
	}
	return available
}

// GetAllCompilers returns all registered compilers (available or not)
func (s *CompilationService) GetAllCompilers() map[string]bool {
	result := make(map[string]bool)
	for lang, compiler := range s.compilers {
		result[lang] = compiler.IsAvailable()
	}
	return result
}

// IsCompilerAvailable checks if a specific compiler is available
func (s *CompilationService) IsCompilerAvailable(language string) bool {
	compiler, ok := s.compilers[language]
	if !ok {
		return false
	}
	return compiler.IsAvailable()
}
