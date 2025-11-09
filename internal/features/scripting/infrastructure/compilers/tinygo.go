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

// TinyGoCompiler compiles Go code to WASM using TinyGo
// This is an ADAPTER in Clean Architecture - implements domain.Compiler interface
// Uses CacheManager to find installed TinyGo compiler
type TinyGoCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	tinygoPath   string              // Cached path to tinygo binary
	goPath       string              // Path to go binary (for go mod download)
}

// NewTinyGoCompiler creates a new TinyGo compiler
// cacheManager is used to locate the downloaded TinyGo compiler
func NewTinyGoCompiler(cacheManager domain.CacheManager) *TinyGoCompiler {
	return &TinyGoCompiler{
		cacheManager: cacheManager,
		goPath:       "go", // Use system go for dependencies
	}
}

// Language returns the language identifier (implements domain.Compiler)
func (c *TinyGoCompiler) Language() string {
	return "go"
}

// IsAvailable checks if TinyGo is installed (implements domain.Compiler)
func (c *TinyGoCompiler) IsAvailable() bool {
	// Check if TinyGo is in cache
	tinygoPath, err := c.getTinyGoBinary()
	if err != nil {
		// Try system tinygo as fallback
		cmd := exec.Command("tinygo", "version")
		if err := cmd.Run(); err != nil {
			return false
		}
		c.tinygoPath = "tinygo" // Use system tinygo
		return true
	}

	c.tinygoPath = tinygoPath
	return true
}

// Compile compiles Go source to WASM (implements domain.Compiler)
func (c *TinyGoCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	start := time.Now()

	// Ensure tinygo is available
	if !c.IsAvailable() {
		return nil, domain.NewCompilationError("TinyGo compiler not installed. Please install TinyGo first.")
	}

	// Create isolated workspace
	ws, err := NewWorkspace(fmt.Sprintf("tinygo-%d", time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Write main.go (main source file)
	if err := ws.WriteFile("main.go", []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("write source code: %w", err)
	}

	// Write additional Go source files from Dependencies map
	// This enables multi-file Go projects (types.go, utils.go, etc.)
	for filename, content := range req.Dependencies {
		// Skip go.mod and go.sum (handled separately below)
		if filename == "go.mod" || filename == "go.sum" {
			continue
		}

		// Write any .go file (Go source files in same package)
		if strings.HasSuffix(filename, ".go") {
			if err := ws.WriteFile(filename, []byte(content)); err != nil {
				return nil, fmt.Errorf("write %s: %w", filename, err)
			}
		}
	}

	// Handle dependencies (go.mod)
	if goMod, ok := req.Dependencies["go.mod"]; ok {
		// Write go.mod
		if err := ws.WriteFile("go.mod", []byte(goMod)); err != nil {
			return nil, fmt.Errorf("write go.mod: %w", err)
		}

		// Download dependencies using standard go command
		output, err := ws.ExecuteCommand(ctx, c.goPath, "mod", "download")
		if err != nil {
			return nil, fmt.Errorf("go mod download failed: %s", string(output))
		}
	} else {
		// Generate minimal go.mod
		if err := c.generateMinimalGoMod(ws); err != nil {
			return nil, fmt.Errorf("generate go.mod: %w", err)
		}
	}

	// Build compilation arguments
	args := []string{"build", "-target=wasi", "-o", "output.wasm"}

	if req.Optimize {
		args = append(args, "-opt=2") // TinyGo optimization level 2
	}

	if !req.Debug {
		args = append(args, "-no-debug") // Strip debug info
	}

	args = append(args, "main.go")

	// Compile with TinyGo
	output, err := ws.ExecuteCommand(ctx, c.tinygoPath, args...)
	if err != nil {
		return nil, c.parseCompilationError("compilation failed", string(output))
	}

	// Read compiled WASM
	wasm, err := ws.ReadFile("output.wasm")
	if err != nil {
		return nil, fmt.Errorf("read compiled WASM: %w", err)
	}

	// Parse compilation logs
	logs := strings.Split(string(output), "\n")

	return &domain.CompileResult{
		WASMBinary: wasm,
		Logs:       logs,
		Duration:   time.Since(start),
		WASMSize:   int64(len(wasm)),
	}, nil
}

// ValidateSyntax validates Go syntax (implements domain.Compiler)
// Uses `go vet` for quick syntax checking
func (c *TinyGoCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if !c.IsAvailable() {
		return fmt.Errorf("TinyGo compiler not available")
	}

	ws, err := NewWorkspace(fmt.Sprintf("go-check-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer ws.Cleanup()

	// Write source
	if err := ws.WriteFile("main.go", []byte(req.SourceCode)); err != nil {
		return err
	}

	// Write minimal go.mod
	if err := c.generateMinimalGoMod(ws); err != nil {
		return err
	}

	// Run go vet for syntax check
	output, err := ws.ExecuteCommand(ctx, c.goPath, "vet", "main.go")
	if err != nil {
		return fmt.Errorf("syntax check failed: %s", string(output))
	}

	return nil
}

// ValidateDependencies validates go.mod structure (implements domain.Compiler)
func (c *TinyGoCompiler) ValidateDependencies(deps map[string]string) error {
	goMod, ok := deps["go.mod"]
	if !ok {
		return nil // No dependencies is valid
	}

	// Basic validation - check for module declaration
	if !strings.Contains(goMod, "module ") {
		return fmt.Errorf("go.mod must contain module declaration")
	}

	return nil
}

// generateMinimalGoMod generates a minimal go.mod for TinyGo
func (c *TinyGoCompiler) generateMinimalGoMod(ws *Workspace) error {
	goMod := `module script

go 1.21

require (
	github.com/extism/go-pdk v1.0.0
)
`
	return ws.WriteFile("go.mod", []byte(goMod))
}

// getTinyGoBinary returns the path to tinygo binary from cache
func (c *TinyGoCompiler) getTinyGoBinary() (string, error) {
	if c.tinygoPath != "" {
		return c.tinygoPath, nil
	}

	// Get compiler path from cache
	compilerPath, err := c.cacheManager.GetCompilerPath("tinygo")
	if err != nil {
		return "", fmt.Errorf("tinygo not found in cache: %w", err)
	}

	// Build path to tinygo binary
	tinygoBinary := filepath.Join(compilerPath, "bin", "tinygo")
	if runtime.GOOS == "windows" {
		tinygoBinary += ".exe"
	}

	return tinygoBinary, nil
}

// parseCompilationError parses TinyGo/Go compiler error output
func (c *TinyGoCompiler) parseCompilationError(message string, output string) error {
	// TODO: Parse Go error format more intelligently
	// Example format:
	// main.go:5:2: undefined: foo

	compErr := domain.NewCompilationError(message)
	compErr.Traceback = output

	// Try to extract error message
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "error") || strings.Contains(line, ":") {
			compErr.Message = line
			break
		}
	}

	return compErr
}
