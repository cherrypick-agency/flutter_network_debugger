package compilers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// AssemblyScriptCompiler compiles TypeScript/JavaScript to WASM using AssemblyScript
// This is an ADAPTER in Clean Architecture - implements domain.Compiler interface
// Uses CacheManager to find installed AssemblyScript compiler (Node.js + asc)
type AssemblyScriptCompiler struct {
	cacheManager domain.CacheManager // Dependency injection for cache
	ascPath      string              // Cached path to asc binary
	npmPath      string              // Cached path to npm binary
	nodePath     string              // Cached path to node binary
}

// NewAssemblyScriptCompiler creates a new AssemblyScript compiler
// cacheManager is used to locate the downloaded AssemblyScript compiler (Node.js + asc)
func NewAssemblyScriptCompiler(cacheManager domain.CacheManager) *AssemblyScriptCompiler {
	return &AssemblyScriptCompiler{
		cacheManager: cacheManager,
	}
}

// Language returns the language identifier (implements domain.Compiler)
func (c *AssemblyScriptCompiler) Language() string {
	return "assemblyscript"
}

// IsAvailable checks if AssemblyScript is installed (implements domain.Compiler)
func (c *AssemblyScriptCompiler) IsAvailable() bool {
	// Check if AssemblyScript is in cache
	ascPath, npmPath, err := c.getAssemblyScriptPaths()
	if err != nil {
		// Try system asc as fallback
		cmd := exec.Command("asc", "--version")
		if err := cmd.Run(); err != nil {
			return false
		}

		// Check system npm
		cmd = exec.Command("npm", "--version")
		if err := cmd.Run(); err != nil {
			return false
		}

		c.ascPath = "asc"
		c.npmPath = "npm"
		return true
	}

	c.ascPath = ascPath
	c.npmPath = npmPath
	return true
}

// Compile compiles AssemblyScript source to WASM (implements domain.Compiler)
func (c *AssemblyScriptCompiler) Compile(ctx context.Context, req domain.CompileRequest) (*domain.CompileResult, error) {
	start := time.Now()

	// Ensure AssemblyScript is available
	if !c.IsAvailable() {
		return nil, domain.NewCompilationError("AssemblyScript compiler not installed. Please install AssemblyScript first.")
	}

	// Create isolated workspace
	ws, err := NewWorkspace(fmt.Sprintf("as-%d", time.Now().UnixNano()))
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	defer ws.Cleanup()

	// Write source code (main entry point)
	if err := ws.WriteFile("index.ts", []byte(req.SourceCode)); err != nil {
		return nil, fmt.Errorf("write source code: %w", err)
	}

	// Write additional TypeScript source files from Dependencies map
	// This enables multi-file AssemblyScript projects (types.ts, utils.ts, etc.)
	for filename, content := range req.Dependencies {
		// Skip package.json and tsconfig.json (handled separately)
		if filename == "package.json" || filename == "tsconfig.json" {
			continue
		}

		// Write any .ts file (TypeScript/AssemblyScript modules)
		if strings.HasSuffix(filename, ".ts") {
			if err := ws.WriteFile(filename, []byte(content)); err != nil {
				return nil, fmt.Errorf("write %s: %w", filename, err)
			}
		}
	}

	// Handle dependencies if provided
	if packageJSON, ok := req.Dependencies["package.json"]; ok {
		// Write package.json
		if err := ws.WriteFile("package.json", []byte(packageJSON)); err != nil {
			return nil, fmt.Errorf("write package.json: %w", err)
		}

		// Install dependencies
		output, err := ws.ExecuteCommand(ctx, c.npmPath, "install")
		if err != nil {
			return nil, c.parseCompilationError("npm install failed", string(output))
		}
	} else {
		// Generate minimal package.json
		if err := c.generateMinimalPackageJSON(ws); err != nil {
			return nil, fmt.Errorf("generate package.json: %w", err)
		}
	}

	// Build compilation arguments
	args := []string{"index.ts", "-o", "output.wasm", "--target", "release"}

	if req.Optimize {
		args = append(args, "--optimize")
	}

	if req.Debug {
		args = append(args, "--sourceMap")
	}

	// Compile
	output, err := ws.ExecuteCommand(ctx, c.ascPath, args...)
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

// ValidateSyntax validates AssemblyScript syntax (implements domain.Compiler)
// AssemblyScript doesn't have a separate syntax checker, so we do a full compile
func (c *AssemblyScriptCompiler) ValidateSyntax(ctx context.Context, req domain.CompileRequest) error {
	// For AssemblyScript, we need to do a full compilation to check syntax
	// This is not ideal but AssemblyScript doesn't provide a separate syntax checker
	_, err := c.Compile(ctx, req)
	return err
}

// ValidateDependencies validates package.json structure (implements domain.Compiler)
func (c *AssemblyScriptCompiler) ValidateDependencies(deps map[string]string) error {
	packageJSON, ok := deps["package.json"]
	if !ok {
		return nil // No dependencies is valid
	}

	// Validate JSON structure
	var pkg map[string]interface{}
	if err := json.Unmarshal([]byte(packageJSON), &pkg); err != nil {
		return fmt.Errorf("invalid package.json: %w", err)
	}

	return nil
}

// generateMinimalPackageJSON generates a minimal package.json for AssemblyScript
func (c *AssemblyScriptCompiler) generateMinimalPackageJSON(ws *Workspace) error {
	packageJSON := `{
  "name": "script",
  "version": "1.0.0",
  "dependencies": {
    "assemblyscript": "^0.27.0"
  },
  "scripts": {
    "build": "asc index.ts -o output.wasm --optimize"
  }
}
`
	return ws.WriteFile("package.json", []byte(packageJSON))
}

// parseCompilationError parses AssemblyScript compiler error output
func (c *AssemblyScriptCompiler) parseCompilationError(message string, output string) error {
	// TODO: Parse AssemblyScript error format more intelligently
	// Example format:
	// ERROR TS2304: Cannot find name 'foo'
	//   at index.ts:5:10

	compErr := domain.NewCompilationError(message)
	compErr.Traceback = output

	// Try to extract line/column info
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "ERROR") || strings.Contains(line, "error") {
			compErr.Message = line
			break
		}
	}

	return compErr
}

// getAssemblyScriptPaths returns the paths to asc and npm binaries from cache
func (c *AssemblyScriptCompiler) getAssemblyScriptPaths() (ascPath, npmPath string, err error) {
	if c.ascPath != "" && c.npmPath != "" {
		return c.ascPath, c.npmPath, nil
	}

	// If no cache manager, return error (will fallback to system)
	if c.cacheManager == nil {
		return "", "", fmt.Errorf("no cache manager configured")
	}

	// Get compiler path from cache
	compilerPath, err := c.cacheManager.GetCompilerPath("assemblyscript")
	if err != nil {
		return "", "", fmt.Errorf("assemblyscript not found in cache: %w", err)
	}

	// Build paths to binaries
	ascBinary := filepath.Join(compilerPath, "bin", "asc")
	npmBinary := filepath.Join(compilerPath, "bin", "npm")
	nodeBinary := filepath.Join(compilerPath, "bin", "node")

	if runtime.GOOS == "windows" {
		ascBinary = filepath.Join(compilerPath, "asc.cmd")
		npmBinary = filepath.Join(compilerPath, "npm.cmd")
		nodeBinary += ".exe"
	}

	c.nodePath = nodeBinary
	return ascBinary, npmBinary, nil
}
