package icon

import (
	"network-debugger/internal/features/process/domain"
)

// NewExtractor - создать extractor иконок для текущей платформы
// Реализация зависит от платформы (см. extractor_*.go файлы)
func NewExtractor() (domain.IIconExtractor, error) {
	return newPlatformExtractor()
}
