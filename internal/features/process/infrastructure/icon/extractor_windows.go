//go:build windows
// +build windows

package icon

import (
	"context"
	"fmt"

	"network-debugger/internal/features/process/domain"
)

type windowsExtractor struct{}

// ExtractByPID - extract icon by PID (TODO: implement via lxn/win)
func (e *windowsExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	// TODO: Implement via github.com/lxn/win + ExtractIconEx
	return nil, fmt.Errorf("Windows icon extraction not implemented yet")
}

// ExtractByPath - extract icon by path (TODO: implement via lxn/win)
func (e *windowsExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	// TODO: Implement via ExtractIconEx API
	// 1. Use win.ExtractIconEx to extract HICON
	// 2. Convert HICON to PNG bytes
	// 3. Return AppIcon{Format: "png", Data: bytes}
	return nil, fmt.Errorf("Windows icon extraction not implemented yet")
}
