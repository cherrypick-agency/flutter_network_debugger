package httpapi

import (
	"network-debugger/internal/infrastructure/config"
	"testing"
	"time"
)

func TestSleepResponseDelay_NoDelay(t *testing.T) {
	cfg := config.Config{
		ResponseDelayMs:    0,
		ResponseDelayMinMs: 0,
		ResponseDelayMaxMs: 0,
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Errorf("no delay expected, but took %v", elapsed)
	}
}

func TestSleepResponseDelay_FixedDelay(t *testing.T) {
	cfg := config.Config{
		ResponseDelayMs: 50,
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	if elapsed < 45*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("expected ~50ms delay, got %v", elapsed)
	}
}

func TestSleepResponseDelay_RandomRange(t *testing.T) {
	cfg := config.Config{
		ResponseDelayMinMs: 10,
		ResponseDelayMaxMs: 50,
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	if elapsed < 9*time.Millisecond || elapsed > 70*time.Millisecond {
		t.Errorf("expected 10-50ms delay range, got %v", elapsed)
	}
}

func TestSleepResponseDelay_RandomRangeNegativeDelta(t *testing.T) {
	// Test case where max < min (delta would be negative)
	cfg := config.Config{
		ResponseDelayMinMs: 50,
		ResponseDelayMaxMs: 10,
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	// Should use min delay when delta is negative
	if elapsed < 45*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("expected ~50ms delay, got %v", elapsed)
	}
}

func TestSleepResponseDelay_OnlyMin(t *testing.T) {
	// Test with only min set (max is 0)
	cfg := config.Config{
		ResponseDelayMinMs: 30,
		ResponseDelayMaxMs: 0,
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	// Should not delay when max is not positive
	if elapsed > 10*time.Millisecond {
		t.Errorf("expected no delay when max=0, got %v", elapsed)
	}
}

func TestSleepResponseDelay_OnlyMax(t *testing.T) {
	// Test with only max set (min is 0)
	cfg := config.Config{
		ResponseDelayMinMs: 0,
		ResponseDelayMaxMs: 50,
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	// Should not delay when min is not positive
	if elapsed > 10*time.Millisecond {
		t.Errorf("expected no delay when min=0, got %v", elapsed)
	}
}

func TestSleepResponseDelay_PreferRangeOverFixed(t *testing.T) {
	// Test that range takes precedence over fixed delay
	cfg := config.Config{
		ResponseDelayMs:    100, // fixed
		ResponseDelayMinMs: 10,  // range min
		ResponseDelayMaxMs: 30,  // range max
	}

	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)

	// Should use range (10-30ms), not fixed (100ms)
	if elapsed > 70*time.Millisecond {
		t.Errorf("expected range delay (10-30ms), not fixed delay, got %v", elapsed)
	}
}
