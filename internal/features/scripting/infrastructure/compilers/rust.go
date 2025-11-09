package compilers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// RustCompiler compiles Rust code to WASM using Cargo
// This is an ADAPTER in Clean Architecture - implements domain.Compiler interface
// Uses CacheManager to find installed Rust compiler
type RustCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	cargoPath    string              // Cached path to cargo binary
	rustupPath   string              // Cached path to rustup binary
	rustupHome   string              // RUSTUP_HOME directory
	cargoHome    string              // CARGO_HOME directory
}

// NewRustCompiler creates a new Rust compiler
// cacheManager is used to locate the downloaded Rust compiler
func NewRustCompiler(cacheManager domain.CacheManager) *RustCompiler {
	return &RustCompiler{
		cacheManager: cacheManager,
	}
}

// Language returns the language identifier (implements domain.Compiler)
func (c *RustCompiler) Language() string {
	return "rust"
}

// IsAvailable checks if Rust is installed (implements domain.Compiler)
func (c *RustCompiler) IsAvailable() bool {
	// Check if Rust is in cache
	cargoPath, rustupPath, err := c.getRustPaths()
	if err != nil {
		// Try system cargo as fallback
		cmd := exec.Command("cargo", "--version")
		if err := cmd.Run(); err != nil {
			return false
		}

		// Check wasm32 target with system rustup
		cmd = exec.Command("rustup", "target", "list", "--installed")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		if !strings.Contains(string(output), "wasm32-unknown-unknown") {
			return false
		}

		c.cargoPath = "cargo"
		c.rustupPath = "rustup"
		return true
	}

	// Use cached Rust
	c.cargoPath = cargoPath
	c.rustupPath = rustupPath

	// Verify wasm32 target is installed
	cmd := exec.Command(rustupPath, "target", "list", "--installed")
	cmd.Env = c.getRustEnv()
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "wasm32-unknown-unknown")
}

// Compile compiles Rust source to WASM (implements domain.Compiler)
func (c *RustCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	start := time.Now()

	// Ensure rust is available
	if !c.IsAvailable() {
		return nil, domain.NewCompilationError("Rust compiler not installed. Please install Rust first.")
	}

	// Create isolated workspace
	ws, err := NewWorkspace(fmt.Sprintf("rust-%d", time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Handle Cargo.toml
	var cargoToml string
	if toml, ok := req.Dependencies["Cargo.toml"]; ok {
		cargoToml = toml
	} else {
		// Generate minimal Cargo.toml
		cargoToml = c.generateMinimalCargoToml()
	}

	if err := ws.WriteFile("Cargo.toml", []byte(cargoToml)); err != nil {
		return nil, fmt.Errorf("write Cargo.toml: %w", err)
	}

	// Write src/lib.rs (main source file)
	if err := ws.WriteFile("src/lib.rs", []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("write source code: %w", err)
	}

	// Write additional source files from Dependencies map
	// This enables multi-file Rust projects (src/types.rs, src/utils.rs, etc.)
	for filename, content := range req.Dependencies {
		// Skip Cargo.toml (already written above)
		if filename == "Cargo.toml" {
			continue
		}

		// Write any file that starts with "src/" (Rust modules)
		if strings.HasPrefix(filename, "src/") {
			if err := ws.WriteFile(filename, []byte(content)); err != nil {
				return nil, fmt.Errorf("write %s: %w", filename, err)
			}
		}
	}

	// Build compilation arguments
	args := []string{"build", "--target", "wasm32-unknown-unknown"}

	if req.Optimize {
		args = append(args, "--release")
	}

	// Compile with Cargo (with Rust env vars if using cached Rust)
	cmd := exec.CommandContext(ctx, c.cargoPath, args...)
	cmd.Dir = ws.Path
	if c.rustupHome != "" && c.cargoHome != "" {
		cmd.Env = c.getRustEnv()
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, c.parseRustError(string(output))
	}

	// Determine WASM path based on build mode
	wasmPath := "target/wasm32-unknown-unknown/debug/script.wasm"
	if req.Optimize {
		wasmPath = "target/wasm32-unknown-unknown/release/script.wasm"
	}

	// Read compiled WASM
	wasm, err := ws.ReadFile(wasmPath)
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

// ValidateSyntax validates Rust syntax using cargo check (implements domain.Compiler)
func (c *RustCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	if !c.IsAvailable() {
		return fmt.Errorf("Rust compiler not available")
	}

	ws, err := NewWorkspace(fmt.Sprintf("rust-check-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer ws.Cleanup()

	// Write Cargo.toml
	cargoToml := c.generateMinimalCargoToml()
	if err := ws.WriteFile("Cargo.toml", []byte(cargoToml)); err != nil {
		return err
	}

	// Write source
	if err := ws.WriteFile("src/lib.rs", []byte(req.SourceCode)); err != nil {
		return err
	}

	// Run cargo check for fast syntax validation
	cmd := exec.CommandContext(ctx, c.cargoPath, "check", "--target", "wasm32-unknown-unknown")
	cmd.Dir = ws.Path
	if c.rustupHome != "" && c.cargoHome != "" {
		cmd.Env = c.getRustEnv()
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("syntax check failed: %s", string(output))
	}

	return nil
}

// ValidateDependencies validates Cargo.toml structure (implements domain.Compiler)
func (c *RustCompiler) ValidateDependencies(deps map[string]string) error {
	cargoToml, ok := deps["Cargo.toml"]
	if !ok {
		return nil // No dependencies is valid
	}

	// Basic validation - check for [package] section
	if !strings.Contains(cargoToml, "[package]") {
		return fmt.Errorf("Cargo.toml must contain [package] section")
	}

	// Check for cdylib crate-type (required for WASM)
	if !strings.Contains(cargoToml, "crate-type") || !strings.Contains(cargoToml, "cdylib") {
		return fmt.Errorf("Cargo.toml must specify crate-type = [\"cdylib\"] in [lib] section")
	}

	return nil
}

// generateMinimalCargoToml generates a minimal Cargo.toml for Extism PDK
func (c *RustCompiler) generateMinimalCargoToml() string {
	return `[package]
name = "script"
version = "0.1.0"
edition = "2021"

[lib]
crate-type = ["cdylib"]

[dependencies]
extism-pdk = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

[profile.release]
opt-level = "z"
lto = true
strip = true
`
}

// getRustPaths returns the paths to cargo and rustup binaries from cache
func (c *RustCompiler) getRustPaths() (cargoPath, rustupPath string, err error) {
	if c.cargoPath != "" && c.rustupPath != "" {
		return c.cargoPath, c.rustupPath, nil
	}

	// If no cache manager, return error (will fallback to system)
	if c.cacheManager == nil {
		return "", "", fmt.Errorf("no cache manager configured")
	}

	// Get compiler path from cache
	compilerPath, err := c.cacheManager.GetCompilerPath("rust")
	if err != nil {
		return "", "", fmt.Errorf("rust not found in cache: %w", err)
	}

	// Set home directories
	c.rustupHome = filepath.Join(compilerPath, "rustup")
	c.cargoHome = filepath.Join(compilerPath, "cargo")

	// Build paths to binaries
	cargoBinary := filepath.Join(c.cargoHome, "bin", "cargo")
	rustupBinary := filepath.Join(c.cargoHome, "bin", "rustup")

	if runtime.GOOS == "windows" {
		cargoBinary += ".exe"
		rustupBinary += ".exe"
	}

	return cargoBinary, rustupBinary, nil
}

// getRustEnv returns environment variables for cached Rust installation
func (c *RustCompiler) getRustEnv() []string {
	env := os.Environ()

	if c.rustupHome != "" {
		env = append(env, fmt.Sprintf("RUSTUP_HOME=%s", c.rustupHome))
	}

	if c.cargoHome != "" {
		env = append(env, fmt.Sprintf("CARGO_HOME=%s", c.cargoHome))
		// Also add cargo/bin to PATH
		cargoPath := filepath.Join(c.cargoHome, "bin")
		pathEnv := fmt.Sprintf("PATH=%s%c%s", cargoPath, filepath.ListSeparator, os.Getenv("PATH"))
		env = append(env, pathEnv)
	}

	return env
}

// parseRustError parses Rust compiler error output
func (c *RustCompiler) parseRustError(output string) error {
	// Rust error format example:
	// error[E0308]: mismatched types
	//  --> src/lib.rs:10:5

	compErr := domain.NewCompilationError("rust compilation failed")
	compErr.Traceback = output

	// Extract first error message
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		// Look for error[E####] pattern
		if strings.Contains(line, "error[E") || strings.Contains(line, "error:") {
			compErr.Message = line

			// Try to extract line/column info from next line
			if i+1 < len(lines) && strings.Contains(lines[i+1], "-->") {
				locationLine := lines[i+1]
				// Parse: --> src/lib.rs:10:5
				parts := strings.Split(locationLine, ":")
				if len(parts) >= 3 {
					// Extract line number
					fmt.Sscanf(parts[1], "%d", &compErr.Line)
					// Extract column number
					fmt.Sscanf(parts[2], "%d", &compErr.Column)
				}
			}

			// Extract error code (E0308)
			if start := strings.Index(line, "[E"); start != -1 {
				if end := strings.Index(line[start:], "]"); end != -1 {
					compErr.Code = line[start+1 : start+end]
				}
			}

			break
		}
	}

	return compErr
}

// OptimizeWASM runs wasm-opt for additional optimization (optional)
// This is called after compilation if wasm-opt is available
func (c *RustCompiler) OptimizeWASM(ctx context.Context, wasmPath string) error {
	// Check if wasm-opt is available
	if _, err := exec.LookPath("wasm-opt"); err != nil {
		return nil // Skip optimization if wasm-opt not found
	}

	// Create optimized output path
	optimizedPath := filepath.Join(filepath.Dir(wasmPath), "optimized.wasm")

	// Run wasm-opt
	cmd := exec.CommandContext(ctx, "wasm-opt", "-Oz", "--strip-debug", wasmPath, "-o", optimizedPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wasm-opt failed: %w", err)
	}

	return nil
}
