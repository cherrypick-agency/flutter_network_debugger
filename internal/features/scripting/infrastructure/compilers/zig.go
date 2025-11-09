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

// ZigCompiler compiles Zig code to WASM
// This is an ADAPTER in Clean Architecture - implements domain.Compiler interface
// Uses CacheManager to find installed Zig compiler
type ZigCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	zigPath      string              // Cached path to zig binary
}

// NewZigCompiler creates a new Zig compiler
// cacheManager is used to locate the downloaded Zig compiler
func NewZigCompiler(cacheManager domain.CacheManager) *ZigCompiler {
	return &ZigCompiler{
		cacheManager: cacheManager,
	}
}

// Language returns the language identifier (implements domain.Compiler)
func (c *ZigCompiler) Language() string {
	return "zig"
}

// IsAvailable checks if Zig is installed (implements domain.Compiler)
func (c *ZigCompiler) IsAvailable() bool {
	// Check if Zig is in cache
	zigPath, err := c.getZigBinary()
	if err != nil {
		// Try system zig as fallback
		cmd := exec.Command("zig", "version")
		if err := cmd.Run(); err != nil {
			return false
		}
		c.zigPath = "zig" // Use system zig
		return true
	}

	c.zigPath = zigPath
	return true
}

// Compile compiles Zig source to WASM (implements domain.Compiler)
func (c *ZigCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	start := time.Now()

	// Ensure zig is available
	if !c.IsAvailable() {
		return nil, domain.NewCompilationError("Zig compiler not installed. Please install Zig first.")
	}

	// Create isolated workspace
	ws, err := NewWorkspace(fmt.Sprintf("zig-%d", time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Write main source file
	if err := ws.WriteFile("main.zig", []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("write source code: %w", err)
	}

	// Write any additional dependencies
	for filename, content := range req.Dependencies {
		if err := ws.WriteFile(filename, []byte(content)); err != nil {
			return nil, fmt.Errorf("write dependency %s: %w", filename, err)
		}
	}

	// Build compilation arguments
	args := []string{
		"build-lib",
		"main.zig",
		"-target", "wasm32-freestanding",
		"-dynamic",
		"-rdynamic",
		"-O", "Debug", // Default to Debug
	}

	if req.Optimize {
		// Use ReleaseSmall for smallest output
		args[len(args)-1] = "ReleaseSmall"
	}

	// Set output file
	wasmFile := "main.wasm"
	args = append(args, "-femit-bin="+wasmFile)

	// Compile with Zig
	output, err := ws.ExecuteCommand(ctx, c.zigPath, args...)
	if err != nil {
		return nil, c.parseZigError(string(output))
	}

	// Read compiled WASM
	wasm, err := ws.ReadFile(wasmFile)
	if err != nil {
		return nil, fmt.Errorf("read compiled WASM: %w", err)
	}

	// Parse compilation logs
	logs := strings.Split(string(output), "\n")

	// Extract warnings
	var warnings []string
	for _, line := range logs {
		if strings.Contains(line, "warning:") {
			warnings = append(warnings, line)
		}
	}

	return &domain.CompileResult{
		WASMBinary: wasm,
		Logs:       logs,
		Warnings:   warnings,
		Duration:   time.Since(start),
		WASMSize:   int64(len(wasm)),
	}, nil
}

// ValidateSyntax validates Zig syntax (implements domain.Compiler)
func (c *ZigCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if !c.IsAvailable() {
		return fmt.Errorf("Zig compiler not available")
	}

	ws, err := NewWorkspace(fmt.Sprintf("zig-check-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer ws.Cleanup()

	// Write source
	if err := ws.WriteFile("main.zig", []byte(req.SourceCode)); err != nil {
		return err
	}

	// Run zig ast-check for fast syntax validation
	output, err := ws.ExecuteCommand(ctx, c.zigPath, "ast-check", "main.zig")
	if err != nil {
		return fmt.Errorf("syntax check failed: %s", string(output))
	}

	return nil
}

// ValidateDependencies validates dependency structure (implements domain.Compiler)
// Zig doesn't have a standard package manager yet, so we just check for .zig files
func (c *ZigCompiler) ValidateDependencies(deps map[string]string) error {
	// Check that all dependencies are .zig files
	for filename := range deps {
		if !strings.HasSuffix(filename, ".zig") && filename != "build.zig" {
			return fmt.Errorf("invalid dependency file: %s (must be .zig or build.zig)", filename)
		}
	}

	return nil
}

// getZigBinary returns the path to zig binary from cache
func (c *ZigCompiler) getZigBinary() (string, error) {
	if c.zigPath != "" {
		return c.zigPath, nil
	}

	// Get compiler path from cache
	compilerPath, err := c.cacheManager.GetCompilerPath("zig")
	if err != nil {
		return "", fmt.Errorf("zig not found in cache: %w", err)
	}

	// Build path to zig binary
	zigBinary := filepath.Join(compilerPath, "zig")
	if runtime.GOOS == "windows" {
		zigBinary += ".exe"
	}

	return zigBinary, nil
}

// parseZigError parses Zig compiler error output into CompilationError
func (c *ZigCompiler) parseZigError(output string) error {
	lines := strings.Split(output, "\n")

	// Find first error line
	for _, line := range lines {
		if strings.Contains(line, "error:") {
			// Zig error format: filename:line:column: error: message
			parts := strings.SplitN(line, ":", 4)
			if len(parts) >= 4 {
				// Try to parse line and column
				var lineNum, colNum int
				fmt.Sscanf(parts[1], "%d", &lineNum)
				fmt.Sscanf(parts[2], "%d", &colNum)

				message := strings.TrimSpace(parts[3])
				if strings.HasPrefix(message, "error:") {
					message = strings.TrimSpace(strings.TrimPrefix(message, "error:"))
				}

				return &domain.CompilationError{
					Message:   message,
					Line:      lineNum,
					Column:    colNum,
					Severity:  "error",
					Traceback: output,
				}
			}

			// Fallback: just return the error line
			return domain.NewCompilationError(strings.TrimSpace(line))
		}
	}

	// No specific error found, return full output
	return domain.NewCompilationError(output)
}

// Verify interface implementation at compile time
var _ domain.Compiler = (*ZigCompiler)(nil)
