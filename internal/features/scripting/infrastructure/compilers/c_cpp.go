package compilers

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// CCPPCompiler compiles C/C++ to WASM using clang with wasm32-wasi target
// Uses CacheManager to find installed WASI-SDK
type CCPPCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	clangPath    string              // Cached path to clang binary
	sysroot      string              // WASI-SDK sysroot path
}

// NewCCPPCompiler creates a new C/C++ compiler
// cacheManager is used to locate the downloaded WASI-SDK
func NewCCPPCompiler(cacheManager domain.CacheManager) *CCPPCompiler {
	return &CCPPCompiler{
		cacheManager: cacheManager,
	}
}

// Language returns the language identifier
func (c *CCPPCompiler) Language() string {
	return "c" // Also handles C++ based on file extension
}

// IsAvailable checks if clang with WASM target is available
func (c *CCPPCompiler) IsAvailable() bool {
	// Check if WASI-SDK is in cache
	clangPath, sysroot, err := c.getWASISDKPaths()
	if err != nil {
		// Try system clang as fallback
		paths := []string{
			"clang",    // System clang
			"clang-18", // Versioned clang
			"clang-17",
			"/opt/wasi-sdk/bin/clang", // wasi-sdk
		}

		for _, path := range paths {
			cmd := exec.Command(path, "--version")
			if err := cmd.Run(); err == nil {
				// Check if it supports wasm32-wasi target
				cmd = exec.Command(path, "--print-targets")
				output, err := cmd.Output()
				if err == nil && strings.Contains(string(output), "wasm32") {
					c.clangPath = path
					return true
				}
				// Even if --print-targets fails, try to use it
				c.clangPath = path
				return true
			}
		}

		return false
	}

	c.clangPath = clangPath
	c.sysroot = sysroot
	return true
}

// Compile compiles C/C++ source code to WASM
func (c *CCPPCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	if c.clangPath == "" && !c.IsAvailable() {
		return nil, fmt.Errorf("clang with WASM support not found. Install wasi-sdk or clang with wasm32 target")
	}

	// Create workspace
	ws, err := NewWorkspace("c-cpp-compile")
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Determine file extension based on language variant
	ext := ".c"
	if strings.Contains(strings.ToLower(req.SourceCode), "iostream") ||
		strings.Contains(strings.ToLower(req.SourceCode), "std::") {
		ext = ".cpp"
	}
	// User can also specify "cpp" or "c++" as language
	if req.Language == "cpp" || req.Language == "c++" {
		ext = ".cpp"
	}

	sourceFile := "main" + ext
	if err := ws.WriteFile(sourceFile, []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("failed to write source file: %w", err)
	}

	// Write any dependencies (headers, etc.)
	for filename, content := range req.Dependencies {
		if err := ws.WriteFile(filename, []byte(content)); err != nil {
			return nil, fmt.Errorf("failed to write dependency %s: %w", filename, err)
		}
	}

	// Prepare compilation flags
	outputFile := "output.wasm"
	args := []string{
		sourceFile,
		"-o", outputFile,
		"--target=wasm32-wasi", // WASI target for maximum compatibility
		"-nostartfiles",        // Don't link startup files
		"-Wl,--no-entry",       // No main entry point (it's a library)
		"-Wl,--export-all",     // Export all functions
		"-I.",                  // Include current directory for headers
	}

	// Add sysroot if using cached WASI-SDK
	if c.sysroot != "" {
		args = append(args, "--sysroot="+c.sysroot)
	}

	if req.Optimize {
		args = append(args, "-O3", "-flto") // Level 3 optimization + LTO
	} else if req.Debug {
		args = append(args, "-g", "-O0") // Debug symbols, no optimization
	} else {
		args = append(args, "-O2") // Default optimization
	}

	// Execute compilation
	compileStart := time.Now()
	output, err := ws.ExecuteCommand(ctx, c.clangPath, args...)
	compileDuration := time.Since(compileStart)

	logs := []string{string(output)}

	if err != nil {
		// Parse clang errors
		errorMsg := string(output)
		if errorMsg == "" {
			errorMsg = err.Error()
		}
		return nil, &domain.CompilationError{
			Message:   "Compilation failed",
			Traceback: errorMsg,
		}
	}

	// Read compiled WASM
	wasmData, err := ws.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read compiled WASM: %w", err)
	}

	return &domain.CompileResult{
		WASMBinary: wasmData,
		Logs:       logs,
		Duration:   compileDuration,
		WASMSize:   int64(len(wasmData)),
	}, nil
}

// ValidateSyntax validates C/C++ syntax without full compilation
func (c *CCPPCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if c.clangPath == "" && !c.IsAvailable() {
		return fmt.Errorf("clang not available")
	}

	ws, err := NewWorkspace("c-cpp-validate")
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	defer ws.Cleanup()

	ext := ".c"
	if req.Language == "cpp" || req.Language == "c++" {
		ext = ".cpp"
	}

	sourceFile := "main" + ext
	if err := ws.WriteFile(sourceFile, []byte(req.SourceCode)); err != nil {
		return fmt.Errorf("failed to write source: %w", err)
	}

	// Syntax check only
	args := []string{
		"-fsyntax-only", // Only check syntax
		"-Wall",         // All warnings
		sourceFile,
	}

	output, err := ws.ExecuteCommand(ctx, c.clangPath, args...)
	if err != nil {
		return &domain.CompilationError{
			Message:   "Syntax validation failed",
			Traceback: string(output),
		}
	}

	return nil
}

// ValidateDependencies validates dependencies structure
func (c *CCPPCompiler) ValidateDependencies(deps map[string]string) error {
	// For C/C++, dependencies are typically header files
	// Just check that files have reasonable extensions
	for filename := range deps {
		ext := filepath.Ext(filename)
		validExts := []string{".h", ".hpp", ".hxx", ".hh", ".c", ".cpp", ".cc", ".cxx"}
		valid := false
		for _, validExt := range validExts {
			if ext == validExt {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid file extension for C/C++: %s (expected .h, .hpp, .c, .cpp, etc.)", filename)
		}
	}
	return nil
}

// getWASISDKPaths returns the paths to clang binary and sysroot from cache
func (c *CCPPCompiler) getWASISDKPaths() (clangPath, sysroot string, err error) {
	if c.clangPath != "" && c.sysroot != "" {
		return c.clangPath, c.sysroot, nil
	}

	// If no cache manager, return error (will fallback to system)
	if c.cacheManager == nil {
		return "", "", fmt.Errorf("no cache manager configured")
	}

	// Get compiler path from cache
	compilerPath, err := c.cacheManager.GetCompilerPath("wasi-sdk")
	if err != nil {
		return "", "", fmt.Errorf("wasi-sdk not found in cache: %w", err)
	}

	// Build path to clang binary
	clangBinary := filepath.Join(compilerPath, "bin", "clang")
	if runtime.GOOS == "windows" {
		clangBinary += ".exe"
	}

	// WASI-SDK sysroot is in share/wasi-sysroot
	sysrootPath := filepath.Join(compilerPath, "share", "wasi-sysroot")

	return clangBinary, sysrootPath, nil
}
