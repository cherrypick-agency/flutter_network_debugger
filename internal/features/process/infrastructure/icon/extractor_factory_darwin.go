//go:build darwin
// +build darwin

package icon

import (
	"network-debugger/internal/features/process/domain"
)

// newPlatformExtractor - create extractor for macOS
func newPlatformExtractor() (domain.IIconExtractor, error) {
	// Using darwinExtractor with fallback to Info.plist
	// Works even without fileicon utility
	return &darwinExtractor{}, nil
}
