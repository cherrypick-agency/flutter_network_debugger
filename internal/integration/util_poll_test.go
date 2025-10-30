package integration

import (
	"testing"
	"time"
)

func TestPollUntil_Success(t *testing.T) {
	count := 0
	result := pollUntil(1*time.Second, 10*time.Millisecond, 100*time.Millisecond, func() bool {
		count++
		return count >= 3
	})

	if !result {
		t.Error("pollUntil should return true when fn returns true")
	}
	if count != 3 {
		t.Errorf("fn should be called 3 times, got %d", count)
	}
}

func TestPollUntil_Timeout(t *testing.T) {
	count := 0
	result := pollUntil(100*time.Millisecond, 20*time.Millisecond, 50*time.Millisecond, func() bool {
		count++
		return false
	})

	if result {
		t.Error("pollUntil should return false on timeout")
	}
	if count == 0 {
		t.Error("fn should be called at least once")
	}
}

func TestPollUntil_ImmediateSuccess(t *testing.T) {
	count := 0
	result := pollUntil(1*time.Second, 10*time.Millisecond, 100*time.Millisecond, func() bool {
		count++
		return true
	})

	if !result {
		t.Error("pollUntil should return true immediately")
	}
	if count != 1 {
		t.Errorf("fn should be called exactly once, got %d", count)
	}
}

func TestPollUntil_DefaultDelayValues(t *testing.T) {
	count := 0
	// Test with zero/negative delay values to verify defaults are used
	result := pollUntil(500*time.Millisecond, 0, 0, func() bool {
		count++
		return count >= 2
	})

	if !result {
		t.Error("pollUntil should work with default delay values")
	}
	if count < 2 {
		t.Errorf("fn should be called at least 2 times, got %d", count)
	}
}

func TestPollUntil_ExponentialBackoff(t *testing.T) {
	calls := []time.Time{}
	start := time.Now()

	pollUntil(500*time.Millisecond, 10*time.Millisecond, 100*time.Millisecond, func() bool {
		calls = append(calls, time.Now())
		// Проверяем экспоненциальный рост задержки
		if len(calls) >= 3 {
			return true
		}
		return false
	})

	if len(calls) < 3 {
		t.Fatalf("expected at least 3 calls, got %d", len(calls))
	}

	// Verify that delays are increasing
	delay1 := calls[1].Sub(calls[0])
	delay2 := calls[2].Sub(calls[1])

	// delay2 should be roughly 2x delay1 (exponential backoff)
	// Allow some tolerance for timing variations
	if delay2 < delay1 {
		t.Errorf("delays should increase exponentially: delay1=%v, delay2=%v", delay1, delay2)
	}

	totalDuration := time.Since(start)
	if totalDuration > 600*time.Millisecond {
		t.Errorf("test took too long: %v", totalDuration)
	}
}

func TestPollUntil_MaxDelayClamp(t *testing.T) {
	count := 0
	maxDelay := 50 * time.Millisecond

	pollUntil(1*time.Second, 30*time.Millisecond, maxDelay, func() bool {
		count++
		// Run enough times to ensure delay hits the max
		return count >= 5
	})

	if count < 5 {
		t.Errorf("fn should be called at least 5 times, got %d", count)
	}
}

func TestPollUntil_NegativeInitialDelay(t *testing.T) {
	count := 0
	result := pollUntil(500*time.Millisecond, -10*time.Millisecond, 100*time.Millisecond, func() bool {
		count++
		return count >= 2
	})

	if !result {
		t.Error("pollUntil should work with negative initial delay (uses default)")
	}
}

func TestPollUntil_NegativeMaxDelay(t *testing.T) {
	count := 0
	result := pollUntil(500*time.Millisecond, 10*time.Millisecond, -100*time.Millisecond, func() bool {
		count++
		return count >= 2
	})

	if !result {
		t.Error("pollUntil should work with negative max delay (uses default)")
	}
}
