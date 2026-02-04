//go:build darwin
// +build darwin

package icon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"network-debugger/internal/features/process/domain"
)

type darwinExtractor struct{}

// ExtractByPID - extract icon by process PID
func (e *darwinExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	// Get path to application by PID
	cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprint(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process path: %w", err)
	}

	path := strings.TrimSpace(string(output))
	return e.ExtractByPath(ctx, path)
}

// ExtractByPath - extract icon by path to application
func (e *darwinExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	// 1. Find .app bundle
	appPath := findAppBundle(path)
	if appPath == "" {
		return nil, fmt.Errorf("not an application bundle: %s", path)
	}

	// 2. Create temporary directory
	tmpDir, err := os.MkdirTemp("", "icons")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	icnsPath := filepath.Join(tmpDir, "icon.icns")
	pngPath := filepath.Join(tmpDir, "icon.png")

	// 3. Extract .icns using fileicon (if installed)
	cmd := exec.CommandContext(ctx, "fileicon", "get", appPath, "--output", icnsPath)
	if err := cmd.Run(); err != nil {
		// Fallback: try to find Info.plist and icon directly
		return e.extractFromInfoPlist(appPath)
	}

	// 4. Convert to PNG using sips (built-in macOS utility)
	cmd = exec.CommandContext(ctx, "sips", "-s", "format", "png", icnsPath, "--out", pngPath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sips conversion failed: %w", err)
	}

	// 5. Read PNG
	data, err := os.ReadFile(pngPath)
	if err != nil {
		return nil, err
	}

	return &domain.AppIcon{
		Format: "png",
		Data:   data,
	}, nil
}

// findAppBundle - find .app bundle in path
func findAppBundle(path string) string {
	// Look for .app in path, going up the directories
	current := path
	for {
		if strings.HasSuffix(current, ".app") {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current || parent == "/" {
			break
		}
		current = parent
	}
	return ""
}

// extractFromInfoPlist - fallback: extract icon directly from Info.plist
func (e *darwinExtractor) extractFromInfoPlist(appPath string) (*domain.AppIcon, error) {
	// Path to Resources
	resourcesPath := filepath.Join(appPath, "Contents", "Resources")

	// Try to find .icns files
	entries, err := os.ReadDir(resourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Resources: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".icns") {
			icnsPath := filepath.Join(resourcesPath, entry.Name())

			// Convert to PNG via sips
			tmpDir, _ := os.MkdirTemp("", "icons")
			defer os.RemoveAll(tmpDir)

			pngPath := filepath.Join(tmpDir, "icon.png")
			cmd := exec.Command("sips", "-s", "format", "png", icnsPath, "--out", pngPath)
			if err := cmd.Run(); err != nil {
				continue
			}

			// Read PNG
			data, err := os.ReadFile(pngPath)
			if err != nil {
				continue
			}

			return &domain.AppIcon{
				Format: "png",
				Data:   data,
			}, nil
		}
	}

	return nil, fmt.Errorf("no icon found in app bundle")
}
