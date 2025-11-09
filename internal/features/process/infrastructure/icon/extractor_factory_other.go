//go:build !darwin
// +build !darwin

package icon

import (
	"network-debugger/internal/features/process/domain"
)

// newPlatformExtractor - создать extractor для Windows/Linux
func newPlatformExtractor() (domain.IIconExtractor, error) {
	// Пока используем stub для Windows/Linux
	// TODO: реализовать extractors для этих платформ
	return &stubExtractor{}, nil
}
