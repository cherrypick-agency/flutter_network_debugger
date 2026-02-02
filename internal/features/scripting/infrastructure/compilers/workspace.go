package compilers

import (
	"bytes"
	"context"
	"fmt"
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

const maxCommandOutputBytes = 2 * 1024 * 1024

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	remaining := b.limit - b.buf.Len()
	if remaining > 0 {
		if len(p) <= remaining {
			_, _ = b.buf.Write(p)
		} else {
			_, _ = b.buf.Write(p[:remaining])
			b.truncated = true
		}
	} else {
		b.truncated = true
	}

	// Важно: всегда "принимаем" все байты, чтобы не блокировать reader процессов.
	return len(p), nil
}

func (b *limitedBuffer) BytesWithNote() []byte {
	out := b.buf.Bytes()
	if !b.truncated {
		return out
	}
	return append(out, []byte("\n[output truncated]\n")...)
}

func sanitizeForFilename(s string) string {
	if s == "" {
		return "unknown"
	}
	const maxLen = 80
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		if b.Len() >= maxLen {
			break
		}
	}
	out := strings.Trim(b.String(), "._-")
	if out == "" {
		return "unknown"
	}
	return out
}

func (w *Workspace) safePath(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename is empty")
	}
	if filepath.IsAbs(filename) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", filename)
	}

	clean := filepath.Clean(filename)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path traversal is not allowed: %s", filename)
	}

	full := filepath.Join(w.Path, clean)
	rel, err := filepath.Rel(w.Path, full)
	if err != nil {
		return "", fmt.Errorf("invalid path %s: %w", filename, err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes workspace: %s", filename)
	}

	return full, nil
}

func mustBeWithinDir(baseDir, path string) error {
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("get abs base dir: %w", err)
	}
	if resolvedBase, err := filepath.EvalSymlinks(baseAbs); err == nil {
		baseAbs = resolvedBase
	}
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("get abs path: %w", err)
	}
	if resolvedPath, err := filepath.EvalSymlinks(pathAbs); err == nil {
		pathAbs = resolvedPath
	}

	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if rel == "." {
		return nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes workspace")
	}
	return nil
}

func mkdirAllNoSymlink(baseDir, dirPath string, perm os.FileMode) error {
	rel, err := filepath.Rel(baseDir, dirPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if rel == "." {
		return nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes workspace")
	}

	cur := baseDir
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		cur = filepath.Join(cur, part)

		fi, err := os.Lstat(cur)
		if err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to follow symlink in path: %s", cur)
			}
			if !fi.IsDir() {
				return fmt.Errorf("path component is not a directory: %s", cur)
			}
			continue
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", cur, err)
		}
		if err := os.Mkdir(cur, perm); err != nil {
			// Если гонка — перепроверим.
			if fi2, err2 := os.Lstat(cur); err2 == nil {
				if fi2.Mode()&os.ModeSymlink != 0 {
					return fmt.Errorf("refusing to follow symlink in path: %s", cur)
				}
				if !fi2.IsDir() {
					return fmt.Errorf("path component is not a directory: %s", cur)
				}
				continue
			}
			return fmt.Errorf("mkdir %s: %w", cur, err)
		}
	}

	return nil
}

func ensureNotSymlink(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to operate on symlink: %s", path)
	}
	return nil
}

// NewWorkspace creates a new temporary workspace for compilation
// Each workspace is isolated to prevent conflicts between concurrent compilations
func NewWorkspace(scriptID string) (*Workspace, error) {
	prefix := "go-proxy-compile-" + sanitizeForFilename(scriptID) + "-"
	tmpDir, err := os.MkdirTemp(os.TempDir(), prefix)
	if err != nil {
		return nil, fmt.Errorf("create workspace directory: %w", err)
	}
	return &Workspace{Path: tmpDir}, nil
}

// WriteFile writes a file to the workspace
// Automatically creates parent directories if needed
func (w *Workspace) WriteFile(filename string, content []byte) error {
	path, err := w.safePath(filename)
	if err != nil {
		return err
	}

	// Create parent directories
	dir := filepath.Dir(path)
	if err := mkdirAllNoSymlink(w.Path, dir, 0755); err != nil {
		return fmt.Errorf("create parent directories for %s: %w", filename, err)
	}
	if err := ensureNotSymlink(path); err != nil {
		return err
	}

	// Write file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("write file %s: %w", filename, err)
	}

	return nil
}

// ReadFile reads a file from the workspace
func (w *Workspace) ReadFile(filename string) ([]byte, error) {
	path, err := w.safePath(filename)
	if err != nil {
		return nil, err
	}
	if err := mustBeWithinDir(w.Path, path); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}
	return data, nil
}

// FileExists checks if a file exists in the workspace
func (w *Workspace) FileExists(filename string) bool {
	path, err := w.safePath(filename)
	if err != nil {
		return false
	}
	_, statErr := os.Stat(path)
	if statErr != nil {
		return false
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	return mustBeWithinDir(w.Path, resolved) == nil
}

// ExecuteCommand executes a command in the workspace directory
// Returns combined stdout and stderr
func (w *Workspace) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	return w.ExecuteCommandWithEnv(ctx, nil, name, args...)
}

// ExecuteCommandWithEnv executes a command with custom environment and returns combined stdout/stderr.
func (w *Workspace) ExecuteCommandWithEnv(ctx context.Context, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = w.Path
	if env != nil {
		cmd.Env = env
	}

	var out limitedBuffer
	out.limit = maxCommandOutputBytes
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.BytesWithNote()
	if err != nil {
		return output, fmt.Errorf("command '%s %s' failed: %w\nOutput: %s",
			name, strings.Join(args, " "), err, string(output))
	}
	return output, nil
}

// ExecuteCommandSeparate executes a command and returns stdout and stderr separately
func (w *Workspace) ExecuteCommandSeparate(ctx context.Context, name string, args ...string) (stdout, stderr []byte, err error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = w.Path

	var outBuf limitedBuffer
	outBuf.limit = maxCommandOutputBytes
	var errBuf limitedBuffer
	errBuf.limit = maxCommandOutputBytes
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return outBuf.BytesWithNote(), errBuf.BytesWithNote(), fmt.Errorf("command failed: %w", err)
	}
	return outBuf.BytesWithNote(), errBuf.BytesWithNote(), nil
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
