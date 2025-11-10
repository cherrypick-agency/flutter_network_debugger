package icon

import (
	"context"
	"testing"
)

// Composer 1.
func TestStubExtractor_ExtractByPID(t *testing.T) {
	extractor := &stubExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPID(ctx, 1234)

	if err == nil {
		t.Error("ExtractByPID() should return error (not supported)")
	}

	if icon != nil {
		t.Error("ExtractByPID() should return nil icon (not supported)")
	}
}

// Composer 1.
func TestStubExtractor_ExtractByPath(t *testing.T) {
	extractor := &stubExtractor{}

	ctx := context.Background()
	icon, err := extractor.ExtractByPath(ctx, "/usr/bin/test")

	if err == nil {
		t.Error("ExtractByPath() should return error (not supported)")
	}

	if icon != nil {
		t.Error("ExtractByPath() should return nil icon (not supported)")
	}
}
