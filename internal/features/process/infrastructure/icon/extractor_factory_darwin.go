//go:build darwin
// +build darwin

package icon

import (
	"network-debugger/internal/features/process/domain"
)

// newPlatformExtractor - создать extractor для macOS
func newPlatformExtractor() (domain.IIconExtractor, error) {
	// Используем darwinExtractor с fallback на Info.plist
	// Работает даже без fileicon утилиты
	return &darwinExtractor{}, nil
}
