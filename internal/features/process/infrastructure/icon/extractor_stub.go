package icon

import (
	"context"
	"fmt"

	"network-debugger/internal/features/process/domain"
)

// stubExtractor - fallback extractor для платформ где нет реализации
type stubExtractor struct{}

// ExtractByPID - заглушка
func (e *stubExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	return nil, fmt.Errorf("icon extraction not supported on this platform")
}

// ExtractByPath - заглушка
func (e *stubExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	return nil, fmt.Errorf("icon extraction not supported on this platform")
}
