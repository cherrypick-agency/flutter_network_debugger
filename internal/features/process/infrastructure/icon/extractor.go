package icon

import (
	"network-debugger/internal/features/process/domain"
)

// NewExtractor - create icon extractor for the current platform
// Implementation depends on the platform (see extractor_*.go files)
func NewExtractor() (domain.IIconExtractor, error) {
	return newPlatformExtractor()
}
