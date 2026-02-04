//go:build linux
// +build linux

package icon

import (
	"context"
	"fmt"

	"network-debugger/internal/features/process/domain"
)

type linuxExtractor struct{}

// ExtractByPID - extract icon by PID (TODO: implement via .desktop files)
func (e *linuxExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	// TODO: Implement via:
	// 1. Get process executable path
	// 2. Find corresponding .desktop file in /usr/share/applications/
	// 3. Extract Icon= field
	// 4. Find icon in /usr/share/icons/ or ~/.icons/
	// 5. Return PNG/SVG data
	return nil, fmt.Errorf("Linux icon extraction not implemented yet")
}

// ExtractByPath - extract icon by path (TODO: implement via freedesktop spec)
func (e *linuxExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	// TODO: Implement via freedesktop.org icon theme specification
	return nil, fmt.Errorf("Linux icon extraction not implemented yet")
}
