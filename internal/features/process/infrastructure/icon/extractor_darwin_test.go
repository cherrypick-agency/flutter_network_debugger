//go:build darwin
// +build darwin

package icon

import (
	"context"
	"testing"
)

// Composer 1.
func TestFindAppBundle(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"direct app path", "/Applications/Test.app", "/Applications/Test.app"},
		{"app with executable", "/Applications/Test.app/Contents/MacOS/Test", "/Applications/Test.app"},
		{"nested path", "/Applications/Test.app/Contents/Resources/icon.icns", "/Applications/Test.app"},
		{"no app bundle", "/usr/bin/test", ""},
		{"root path", "/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findAppBundle(tt.path)
			if got != tt.expected {
				t.Errorf("findAppBundle(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

// Composer 1.
func TestDarwinExtractor_ExtractByPID_InvalidPID(t *testing.T) {
	extractor := &darwinExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPID(ctx, 999999)

	if err == nil {
		t.Error("ExtractByPID() should return error for invalid PID")
	}

	if icon != nil {
		t.Error("ExtractByPID() should return nil icon for invalid PID")
	}
}

// Composer 1.
func TestDarwinExtractor_ExtractByPath_NotAppBundle(t *testing.T) {
	extractor := &darwinExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPath(ctx, "/usr/bin/test")

	if err == nil {
		t.Error("ExtractByPath() should return error for non-app bundle")
	}

	if icon != nil {
		t.Error("ExtractByPath() should return nil icon for non-app bundle")
	}
}
