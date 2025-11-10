//go:build windows
// +build windows

package icon

import (
	"context"
	"testing"

	"network-debugger/internal/features/process/domain"
)

// Composer 1.
func TestWindowsExtractor_ExtractByPID(t *testing.T) {
	extractor := &windowsExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPID(ctx, 1234)

	if err == nil {
		t.Error("ExtractByPID() should return error (not implemented)")
	}

	if icon != nil {
		t.Error("ExtractByPID() should return nil icon (not implemented)")
	}
}

// Composer 1.
func TestWindowsExtractor_ExtractByPath(t *testing.T) {
	extractor := &windowsExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPath(ctx, "C:\\Windows\\System32\\notepad.exe")

	if err == nil {
		t.Error("ExtractByPath() should return error (not implemented)")
	}

	if icon != nil {
		t.Error("ExtractByPath() should return nil icon (not implemented)")
	}
}
