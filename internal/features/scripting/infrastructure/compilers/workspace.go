package compilers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Workspace manages temporary compilation workspace
// This provides isolation for each compilation process (Single Responsibility Principle)
type Workspace struct {
	Path string // Absolute path to workspace directory
}

// NewWorkspace creates a new temporary workspace for compilation
// Each workspace is isolated to prevent conflicts between concurrent compilations
func NewWorkspace(scriptID string) (*Workspace, error) {
	// Create workspace in system temp directory
	tmpDir := filepath.Join(os.TempDir(), "go-proxy-compile", scriptID)

	// Ensure parent directory exists
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("create workspace directory: %w", err)
	}

	return &Workspace{Path: tmpDir}, nil
}

// WriteFile writes a file to the workspace
// Automatically creates parent directories if needed
func (w *Workspace) WriteFile(filename string, content []byte) error {
	path := filepath.Join(w.Path, filename)

	// Create parent directories
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create parent directories for %s: %w", filename, err)
	}

	// Write file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("write file %s: %w", filename, err)
	}

	return nil
}

// ReadFile reads a file from the workspace
func (w *Workspace) ReadFile(filename string) ([]byte, error) {
	path := filepath.Join(w.Path, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}
	return data, nil
}

// FileExists checks if a file exists in the workspace
func (w *Workspace) FileExists(filename string) bool {
	path := filepath.Join(w.Path, filename)
	_, err := os.Stat(path)
	return err == nil
}

// ExecuteCommand executes a command in the workspace directory
// Returns combined stdout and stderr
func (w *Workspace) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = w.Path

	// Capture combined output
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Include output in error for better diagnostics
		return output, fmt.Errorf("command '%s %s' failed: %w\nOutput: %s",
			name, strings.Join(args, " "), err, string(output))
	}

	return output, nil
}

// ExecuteCommandSeparate executes a command and returns stdout and stderr separately
func (w *Workspace) ExecuteCommandSeparate(ctx context.Context, name string, args ...string) (stdout, stderr []byte, err error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = w.Path

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start command: %w", err)
	}

	// Read outputs from pipes
	stdout, _ = io.ReadAll(stdoutPipe)
	stderr, _ = io.ReadAll(stderrPipe)

	if err := cmd.Wait(); err != nil {
		return stdout, stderr, fmt.Errorf("command failed: %w", err)
	}

	return stdout, stderr, nil
}

// Cleanup removes the workspace directory and all its contents
// Should be called with defer after workspace creation
func (w *Workspace) Cleanup() error {
	if w.Path == "" {
		return nil
	}

	if err := os.RemoveAll(w.Path); err != nil {
		return fmt.Errorf("cleanup workspace %s: %w", w.Path, err)
	}

	return nil
}

// ListFiles returns all files in the workspace (recursive)
// Useful for debugging and diagnostics
func (w *Workspace) ListFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(w.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Make path relative to workspace
			relPath, err := filepath.Rel(w.Path, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	return files, err
}
