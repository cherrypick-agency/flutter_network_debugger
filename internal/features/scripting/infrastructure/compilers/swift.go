package compilers

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// SwiftCompiler compiles Swift code to WASM using SwiftWasm SDK
// This is an ADAPTER in Clean Architecture - implements domain.Compiler interface
// Uses CacheManager to find SwiftWasm SDK
type SwiftCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	swiftPath    string              // Path to swift binary (usually system)
	sdkPath      string              // Path to SwiftWasm SDK from cache
}

// NewSwiftCompiler creates a new Swift compiler
func NewSwiftCompiler(cacheManager domain.CacheManager) *SwiftCompiler {
	return &SwiftCompiler{
		cacheManager: cacheManager,
	}
}

// Language returns the language identifier (implements domain.Compiler)
func (c *SwiftCompiler) Language() string {
	return "swift"
}

// IsAvailable checks if Swift and SwiftWasm SDK are installed (implements domain.Compiler)
func (c *SwiftCompiler) IsAvailable() bool {
	// Check system Swift
	cmd := exec.Command("swift", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	c.swiftPath = "swift"

	// Check if SwiftWasm SDK is in cache
	sdkPath, err := c.getSwiftWasmSDK()
	if err != nil {
		return false
	}

	c.sdkPath = sdkPath
	return true
}

// Compile compiles Swift source to WASM (implements domain.Compiler)
func (c *SwiftCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	start := time.Now()

	// Ensure Swift is available
	if !c.IsAvailable() {
		return nil, domain.NewCompilationError("Swift compiler not installed or SwiftWasm SDK not found. Please install both.")
	}

	// Create isolated workspace
	ws, err := NewWorkspace(fmt.Sprintf("swift-%d", time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Write main source file
	if err := ws.WriteFile("main.swift", []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("write source code: %w", err)
	}

	// Write Package.swift if provided, otherwise generate minimal one
	packageSwift := ""
	if pkgSwift, ok := req.Dependencies["Package.swift"]; ok {
		packageSwift = pkgSwift
	} else {
		packageSwift = c.generateMinimalPackageSwift()
	}

	if err := ws.WriteFile("Package.swift", []byte(packageSwift)); err != nil {
		return nil, fmt.Errorf("write Package.swift: %w", err)
	}

	// Write any additional dependencies
	for filename, content := range req.Dependencies {
		if filename != "Package.swift" {
			if err := ws.WriteFile(filename, []byte(content)); err != nil {
				return nil, fmt.Errorf("write dependency %s: %w", filename, err)
			}
		}
	}

	// Build compilation arguments for SwiftWasm
	args := []string{
		"build",
		"--triple", "wasm32-unknown-wasi",
		"--sdk", c.sdkPath,
	}

	// Add optimization flags
	if req.Optimize {
		args = append(args, "-c", "release")
	} else {
		args = append(args, "-c", "debug")
	}

	// Compile with Swift
	cmd := exec.CommandContext(ctx, c.swiftPath, args...)
	cmd.Dir = ws.Path

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, c.parseSwiftError(string(output))
	}

	// Find compiled WASM binary
	// Swift build outputs to .build/wasm32-unknown-wasi/debug/ or /release/
	buildMode := "debug"
	if req.Optimize {
		buildMode = "release"
	}

	wasmPath := filepath.Join(".build", "wasm32-unknown-wasi", buildMode, "main.wasm")

	// Read compiled WASM
	wasmBytes, err := ws.ReadFile(wasmPath)
	if err != nil {
		// Try alternative paths
		wasmPath = filepath.Join(".build", "wasm32-unknown-wasi", buildMode, "Package.wasm")
		wasmBytes, err = ws.ReadFile(wasmPath)
		if err != nil {
			return nil, fmt.Errorf("read compiled WASM: %w", err)
		}
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
		WASMBinary: wasmBytes,
		Logs:       logs,
		Warnings:   warnings,
		Duration:   time.Since(start),
		WASMSize:   int64(len(wasmBytes)),
	}, nil
}

// ValidateSyntax validates Swift syntax (implements domain.Compiler)
func (c *SwiftCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if !c.IsAvailable() {
		return fmt.Errorf("Swift compiler not available")
	}

	ws, err := NewWorkspace(fmt.Sprintf("swift-check-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer ws.Cleanup()

	// Write source
	if err := ws.WriteFile("main.swift", []byte(req.SourceCode)); err != nil {
		return err
	}

	// Run swift with -parse for syntax check only
	cmd := exec.CommandContext(ctx, c.swiftPath, "-frontend", "-parse", "main.swift")
	cmd.Dir = ws.Path

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("syntax check failed: %s", string(output))
	}

	return nil
}

// ValidateDependencies validates dependency structure (implements domain.Compiler)
func (c *SwiftCompiler) ValidateDependencies(deps map[string]string) error {
	// Check that all dependencies are .swift files or Package.swift
	for filename := range deps {
		if !strings.HasSuffix(filename, ".swift") && filename != "Package.swift" {
			return fmt.Errorf("invalid dependency file: %s (must be .swift or Package.swift)", filename)
		}
	}

	// If Package.swift is provided, validate it's not empty
	if pkgSwift, ok := deps["Package.swift"]; ok {
		if strings.TrimSpace(pkgSwift) == "" {
			return fmt.Errorf("Package.swift cannot be empty")
		}
	}

	return nil
}

// getSwiftWasmSDK returns the path to SwiftWasm SDK from cache
func (c *SwiftCompiler) getSwiftWasmSDK() (string, error) {
	if c.sdkPath != "" {
		return c.sdkPath, nil
	}

	// Get SDK path from cache
	sdkPath, err := c.cacheManager.GetCompilerPath("swift")
	if err != nil {
		return "", fmt.Errorf("SwiftWasm SDK not found in cache: %w", err)
	}

	// SwiftWasm SDK should contain swift-wasm-6.1.0-RELEASE.artifactbundle or similar
	// For now, just return the base path
	return sdkPath, nil
}

// generateMinimalPackageSwift generates a minimal Package.swift for standalone files
func (c *SwiftCompiler) generateMinimalPackageSwift() string {
	return `// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "main",
    platforms: [.macOS(.v13)],
    products: [
        .executable(name: "main", targets: ["main"])
    ],
    targets: [
        .executableTarget(
            name: "main",
            path: ".",
            sources: ["main.swift"]
        )
    ]
)
`
}

// parseSwiftError parses Swift compiler error output into CompilationError
func (c *SwiftCompiler) parseSwiftError(output string) error {
	lines := strings.Split(output, "\n")

	// Find first error line
	for _, line := range lines {
		if strings.Contains(line, "error:") {
			// Swift error format: file:line:column: error: message
			parts := strings.SplitN(line, ":", 5)
			if len(parts) >= 4 {
				// Try to parse line and column
				var lineNum, colNum int
				fmt.Sscanf(parts[1], "%d", &lineNum)
				fmt.Sscanf(parts[2], "%d", &colNum)

				message := ""
				if len(parts) >= 5 {
					message = strings.TrimSpace(parts[4])
				} else {
					message = strings.TrimSpace(parts[3])
				}

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
var _ domain.Compiler = (*SwiftCompiler)(nil)
