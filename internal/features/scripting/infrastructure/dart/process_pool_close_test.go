package dart

import (
	"context"
	"testing"
	"time"
)

// Composer 1.
func TestProcessPool_Get_AfterClose_ReturnsError(t *testing.T) {
	pool, err := NewProcessPool(0, "")
	if err != nil {
		t.Fatalf("NewProcessPool error: %v", err)
	}

	if err := pool.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err = pool.Get(ctx)
	if err == nil {
		t.Fatal("expected error after Close")
	}
}

// Composer 1.
func TestProcessPool_Close_IsIdempotent(t *testing.T) {
	pool, err := NewProcessPool(0, "")
	if err != nil {
		t.Fatalf("NewProcessPool error: %v", err)
	}

	if err := pool.Close(); err != nil {
		t.Fatalf("first Close error: %v", err)
	}
	if err := pool.Close(); err != nil {
		t.Fatalf("second Close error: %v", err)
	}
}
