//go:build linux
// +build linux

package icon

import (
	"context"
	"testing"
)

// Composer 1.
func TestLinuxExtractor_ExtractByPID(t *testing.T) {
	extractor := &linuxExtractor{}

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
func TestLinuxExtractor_ExtractByPath(t *testing.T) {
	extractor := &linuxExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPath(ctx, "/usr/bin/test")

	if err == nil {
		t.Error("ExtractByPath() should return error (not implemented)")
	}

	if icon != nil {
		t.Error("ExtractByPath() should return nil icon (not implemented)")
	}
}
