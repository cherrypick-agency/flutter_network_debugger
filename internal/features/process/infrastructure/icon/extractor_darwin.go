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

// ExtractByPID - извлечь иконку по PID процесса
func (e *darwinExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	// Получить путь к приложению по PID
	cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprint(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process path: %w", err)
	}

	path := strings.TrimSpace(string(output))
	return e.ExtractByPath(ctx, path)
}

// ExtractByPath - извлечь иконку по пути к приложению
func (e *darwinExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	// 1. Найти .app bundle
	appPath := findAppBundle(path)
	if appPath == "" {
		return nil, fmt.Errorf("not an application bundle: %s", path)
	}

	// 2. Создать временную директорию
	tmpDir, err := os.MkdirTemp("", "icons")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	icnsPath := filepath.Join(tmpDir, "icon.icns")
	pngPath := filepath.Join(tmpDir, "icon.png")

	// 3. Извлечь .icns с помощью fileicon (если установлен)
	cmd := exec.CommandContext(ctx, "fileicon", "get", appPath, "--output", icnsPath)
	if err := cmd.Run(); err != nil {
		// Fallback: попробовать найти Info.plist и иконку напрямую
		return e.extractFromInfoPlist(appPath)
	}

	// 4. Конвертировать в PNG с помощью sips (встроенная утилита macOS)
	cmd = exec.CommandContext(ctx, "sips", "-s", "format", "png", icnsPath, "--out", pngPath)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sips conversion failed: %w", err)
	}

	// 5. Прочитать PNG
	data, err := os.ReadFile(pngPath)
	if err != nil {
		return nil, err
	}

	return &domain.AppIcon{
		Format: "png",
		Data:   data,
	}, nil
}

// findAppBundle - найти .app bundle в пути
func findAppBundle(path string) string {
	// Ищем .app в пути, идя вверх по директориям
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

// extractFromInfoPlist - fallback: извлечь иконку напрямую из Info.plist
func (e *darwinExtractor) extractFromInfoPlist(appPath string) (*domain.AppIcon, error) {
	// Путь к Resources
	resourcesPath := filepath.Join(appPath, "Contents", "Resources")

	// Попробовать найти .icns файлы
	entries, err := os.ReadDir(resourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Resources: %w", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".icns") {
			icnsPath := filepath.Join(resourcesPath, entry.Name())

			// Конвертировать в PNG через sips
			tmpDir, _ := os.MkdirTemp("", "icons")
			defer os.RemoveAll(tmpDir)

			pngPath := filepath.Join(tmpDir, "icon.png")
			cmd := exec.Command("sips", "-s", "format", "png", icnsPath, "--out", pngPath)
			if err := cmd.Run(); err != nil {
				continue
			}

			// Прочитать PNG
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
