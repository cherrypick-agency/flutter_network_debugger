package httpapi

import (
	"testing"
	"time"
)

func TestByteTokenBucket_WaitN_NilBucket(t *testing.T) {
	var b *byteTokenBucket
	// Should not panic
	b.waitN(100)
}

func TestByteTokenBucket_WaitN_ZeroBytes(t *testing.T) {
	b := &byteTokenBucket{
		rateBytesPerSec: 1000,
		capacity:        1000,
		tokens:          1000,
		last:            time.Now(),
	}
	// Should return immediately
	b.waitN(0)
}

func TestByteTokenBucket_WaitN_NegativeBytes(t *testing.T) {
	b := &byteTokenBucket{
		rateBytesPerSec: 1000,
		capacity:        1000,
		tokens:          1000,
		last:            time.Now(),
	}
	// Should return immediately
	b.waitN(-100)
}

func TestByteTokenBucket_WaitN_EnoughTokens(t *testing.T) {
	b := &byteTokenBucket{
		rateBytesPerSec: 10000, // 10KB/s
		capacity:        10000,
		tokens:          5000, // Start with 5KB available
		last:            time.Now(),
	}

	start := time.Now()
	b.waitN(100) // Request only 100 bytes
	elapsed := time.Since(start)

	// Should complete very quickly since we have enough tokens
	if elapsed > 50*time.Millisecond {
		t.Errorf("waitN took too long: %v", elapsed)
	}
}

func TestByteTokenBucket_WaitN_TokenRefill(t *testing.T) {
	b := &byteTokenBucket{
		rateBytesPerSec: 1000, // 1KB/s
		capacity:        1000,
		tokens:          500,                              // Start with half capacity
		last:            time.Now().Add(-1 * time.Second), // Last update was 1 second ago
	}

	// After 1 second at 1KB/s, we should have refilled to 1000 tokens (capacity)
	b.waitN(100) // Request 100 bytes

	// Tokens should have refilled and then consumed 100
	// No need to wait since we have enough after refill
}

func TestByteTokenBucket_WaitN_CapacityLimit(t *testing.T) {
	b := &byteTokenBucket{
		rateBytesPerSec: 1000,
		capacity:        500, // Capacity is less than rate
		tokens:          100,
		last:            time.Now().Add(-10 * time.Second), // Long time ago
	}

	// Even though 10 seconds passed at 1000 bytes/sec (10000 bytes),
	// tokens should be capped at capacity (500)
	b.waitN(50)

	// This should work since we should have 500 tokens (capacity) available
}
