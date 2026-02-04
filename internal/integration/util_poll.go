package integration

import (
	"time"
)

// pollUntil executes fn until it returns true or timeout expires.
// Exponential delay from initial to maxDelay is used between attempts.
func pollUntil(timeout time.Duration, initial time.Duration, maxDelay time.Duration, fn func() bool) bool {
	deadline := time.Now().Add(timeout)
	delay := initial
	if delay <= 0 {
		delay = 50 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 300 * time.Millisecond
	}
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(delay)
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}
	return false
}
