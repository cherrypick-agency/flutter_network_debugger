//go:build windows
// +build windows

package icon

import (
	"context"
	"fmt"

	"network-debugger/internal/features/process/domain"
)

type windowsExtractor struct{}

// ExtractByPID - извлечь иконку по PID (TODO: реализовать через lxn/win)
func (e *windowsExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	// TODO: Реализовать через github.com/lxn/win + ExtractIconEx
	return nil, fmt.Errorf("Windows icon extraction not implemented yet")
}

// ExtractByPath - извлечь иконку по пути (TODO: реализовать через lxn/win)
func (e *windowsExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	// TODO: Реализовать через ExtractIconEx API
	// 1. Использовать win.ExtractIconEx для извлечения HICON
	// 2. Конвертировать HICON в PNG bytes
	// 3. Вернуть AppIcon{Format: "png", Data: bytes}
	return nil, fmt.Errorf("Windows icon extraction not implemented yet")
}
