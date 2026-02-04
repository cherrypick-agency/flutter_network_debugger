//go:build !darwin
// +build !darwin

package icon

import (
	"network-debugger/internal/features/process/domain"
)

// newPlatformExtractor - create extractor for Windows/Linux
func newPlatformExtractor() (domain.IIconExtractor, error) {
	// For now using stub for Windows/Linux
	// TODO: implement extractors for these platforms
	return &stubExtractor{}, nil
}
