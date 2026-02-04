package domain

import "context"

// IIconExtractor - abstraction for application icon extraction
type IIconExtractor interface {
	// ExtractByPID - extract application icon by process PID
	ExtractByPID(ctx context.Context, pid int32) (*AppIcon, error)

	// ExtractByPath - extract icon by path to application/executable
	ExtractByPath(ctx context.Context, path string) (*AppIcon, error)
}
