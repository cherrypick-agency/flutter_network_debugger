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

// KotlinCompiler compiles Kotlin code to WASM
// This is an ADAPTER in Clean Architecture - implements domain.Compiler interface
// Uses CacheManager to find installed Kotlin compiler and JRE
type KotlinCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	kotlincPath  string              // Cached path to kotlinc binary
	javaHome     string              // Path to JRE home
}

// NewKotlinCompiler creates a new Kotlin compiler
func NewKotlinCompiler(cacheManager domain.CacheManager) *KotlinCompiler {
	return &KotlinCompiler{
		cacheManager: cacheManager,
	}
}

// Language returns the language identifier (implements domain.Compiler)
func (c *KotlinCompiler) Language() string {
	return "kotlin"
}

// IsAvailable checks if Kotlin and JRE are installed (implements domain.Compiler)
func (c *KotlinCompiler) IsAvailable() bool {
	// Check if Kotlin is in cache
	kotlincPath, javaHome, err := c.getKotlinSetup()
	if err != nil {
		// Try system kotlinc as fallback
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "kotlinc", "-version")
		if err := cmd.Run(); err != nil {
			return false
		}
		c.kotlincPath = "kotlinc"

		// Check for JRE
		if jh := os.Getenv("JAVA_HOME"); jh != "" {
			c.javaHome = jh
		}

		return true
	}

	c.kotlincPath = kotlincPath
	c.javaHome = javaHome
	return true
}

// Compile compiles Kotlin source to WASM (implements domain.Compiler)
func (c *KotlinCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	start := time.Now()

	// Ensure kotlin is available
	if !c.IsAvailable() {
		return nil, domain.NewCompilationError("Kotlin compiler not installed. Please install Kotlin first.")
	}

	// Create isolated workspace
	ws, err := NewWorkspace(fmt.Sprintf("kotlin-%d", time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Write main source file
	if err := ws.WriteFile("Main.kt", []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("write source code: %w", err)
	}

	// Write any additional dependencies
	for filename, content := range req.Dependencies {
		if err := ws.WriteFile(filename, []byte(content)); err != nil {
			return nil, fmt.Errorf("write dependency %s: %w", filename, err)
		}
	}

	// Build compilation arguments for Kotlin/Wasm
	args := []string{
		"-Xwasm",                 // Enable Wasm backend (experimental)
		"-Xir-produce-klib-file", // Produce klib for Wasm
		"-output", "output.klib", // Output file
		"Main.kt", // Source file
	}

	// Add optimization flags
	if req.Optimize {
		args = append(args, "-Xopt-in=kotlin.RequiresOptIn")
	}

	// Set JAVA_HOME environment if available
	env := os.Environ()
	if c.javaHome != "" {
		env = append(env, fmt.Sprintf("JAVA_HOME=%s", c.javaHome))
	}

	// Compile with Kotlin
	output, err := ws.ExecuteCommandWithEnv(ctx, env, c.kotlincPath, args...)
	if err != nil {
		return nil, c.parseKotlinError(string(output))
	}

	// Note: Kotlin/Wasm compilation is a two-step process
	// 1. kotlinc produces .klib
	// 2. Need to link .klib to .wasm
	// For now, we'll just return the klib as a placeholder
	// TODO: Implement full Kotlin/Wasm linking

	// Read compiled output
	wasmBytes, err := ws.ReadFile("output.klib")
	if err != nil {
		return nil, fmt.Errorf("read compiled output: %w", err)
	}

	// Parse compilation logs
	logs := strings.Split(string(output), "\n")

	// Extract warnings
	var warnings []string
	for _, line := range logs {
		if strings.Contains(line, "warning:") || strings.Contains(line, "w:") {
			warnings = append(warnings, line)
		}
	}

	// Add notice about Kotlin/Wasm being experimental
	logs = append([]string{
		"NOTE: Kotlin/Wasm is experimental and may not produce fully functional WASM binaries yet.",
		"This output is a Kotlin library (.klib) that needs additional linking steps.",
	}, logs...)

	return &domain.CompileResult{
		WASMBinary: wasmBytes,
		Logs:       logs,
		Warnings:   warnings,
		Duration:   time.Since(start),
		WASMSize:   int64(len(wasmBytes)),
	}, nil
}

// ValidateSyntax validates Kotlin syntax (implements domain.Compiler)
func (c *KotlinCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	// Check if kotlincPath was explicitly set to empty (indicates compiler should not be available)
	wasExplicitlyEmpty := c.kotlincPath == ""

	// Try to get kotlinc path from cache or system
	kotlincPath, javaHome, err := c.getKotlinSetup()
	if err != nil {
		// If getKotlinSetup failed, only try system fallback if kotlincPath wasn't explicitly set to empty
		if wasExplicitlyEmpty {
			// kotlincPath was explicitly set to empty, don't try system fallback
			return fmt.Errorf("Kotlin compiler not available")
		}
		// Try system kotlinc as fallback
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "kotlinc", "-version")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Kotlin compiler not available")
		}
		c.kotlincPath = "kotlinc"
	} else {
		// Only update kotlincPath if it wasn't explicitly set to empty
		if wasExplicitlyEmpty {
			// kotlincPath was explicitly set to empty, don't use the found path
			return fmt.Errorf("Kotlin compiler not available")
		}
		c.kotlincPath = kotlincPath
		c.javaHome = javaHome
	}

	// Final check: if kotlincPath is still empty, compiler is not available
	if c.kotlincPath == "" {
		return fmt.Errorf("Kotlin compiler not available")
	}

	ws, err := NewWorkspace(fmt.Sprintf("kotlin-check-%d", time.Now().UnixNano()))
	if err != nil {
		return err
	}
	defer ws.Cleanup()

	// Write source
	if err := ws.WriteFile("Main.kt", []byte(req.SourceCode)); err != nil {
		return err
	}

	// Run kotlinc with -Xskip-runtime-version-check for syntax check only
	env := os.Environ()
	if c.javaHome != "" {
		env = append(env, fmt.Sprintf("JAVA_HOME=%s", c.javaHome))
	}

	output, err := ws.ExecuteCommandWithEnv(ctx, env, c.kotlincPath, "-Xskip-runtime-version-check", "Main.kt")
	if err != nil {
		return fmt.Errorf("syntax check failed: %s", string(output))
	}

	return nil
}

// ValidateDependencies validates dependency structure (implements domain.Compiler)
func (c *KotlinCompiler) ValidateDependencies(deps map[string]string) error {
	// Check that all dependencies are .kt files or build files
	for filename := range deps {
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".kt" && filename != "build.gradle" && filename != "build.gradle.kts" {
			return fmt.Errorf("invalid dependency file: %s (must be .kt or build.gradle)", filename)
		}
	}

	return nil
}

// getKotlinSetup returns the paths to kotlinc and JRE from cache
func (c *KotlinCompiler) getKotlinSetup() (string, string, error) {
	if c.kotlincPath != "" {
		return c.kotlincPath, c.javaHome, nil
	}

	// Get compiler path from cache
	compilerPath, err := c.cacheManager.GetCompilerPath("kotlin")
	if err != nil {
		return "", "", fmt.Errorf("kotlin not found in cache: %w", err)
	}

	// Build path to kotlinc binary
	kotlincBinary := filepath.Join(compilerPath, "bin", "kotlinc")
	if runtime.GOOS == "windows" {
		kotlincBinary = filepath.Join(compilerPath, "bin", "kotlinc.bat")
	}

	// Check for JRE in cache
	jrePath := ""
	jreCachePath, err := c.cacheManager.GetCompilerPath("kotlin")
	if err == nil {
		// JRE might be at kotlin/jre
		jreDir := filepath.Join(jreCachePath, "jre")
		if _, err := os.Stat(jreDir); err == nil {
			jrePath = jreDir
		}
	}

	// Fallback to system JAVA_HOME
	if jrePath == "" {
		jrePath = os.Getenv("JAVA_HOME")
	}

	return kotlincBinary, jrePath, nil
}

// parseKotlinError parses Kotlin compiler error output into CompilationError
func (c *KotlinCompiler) parseKotlinError(output string) error {
	lines := strings.Split(output, "\n")

	// Find first error line
	for _, line := range lines {
		if strings.Contains(line, "error:") || strings.Contains(line, "e:") {
			// Kotlin error format: file:line:column: error: message
			parts := strings.SplitN(line, ":", 5)
			if len(parts) >= 4 {
				// Try to parse line and column
				var lineNum, colNum int
				fmt.Sscanf(parts[1], "%d", &lineNum)
				fmt.Sscanf(parts[2], "%d", &colNum)

				message := ""
				if len(parts) >= 5 {
					message = strings.TrimSpace(parts[4])
				} else if len(parts) >= 4 {
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
var _ domain.Compiler = (*KotlinCompiler)(nil)
