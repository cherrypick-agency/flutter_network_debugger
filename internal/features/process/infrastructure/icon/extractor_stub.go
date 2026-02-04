package icon

import (
	"context"
	"fmt"

	"network-debugger/internal/features/process/domain"
)

// stubExtractor - fallback extractor for platforms without implementation
type stubExtractor struct{}

// ExtractByPID - stub
func (e *stubExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	return nil, fmt.Errorf("icon extraction not supported on this platform")
}

// ExtractByPath - stub
func (e *stubExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	return nil, fmt.Errorf("icon extraction not supported on this platform")
}
