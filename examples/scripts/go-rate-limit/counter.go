package main

import (
	"time"
)

// Simple in-memory counter for tracking request rates
// Note: This is a simplified example. In production, use a proper
// rate limiting algorithm (token bucket, sliding window) and
// persistent storage (Redis, etc.)

type counterEntry struct {
	count     int
	resetTime time.Time
}

var counters = make(map[string]*counterEntry)

// incrementCounter increments the counter for a client IP
// Returns the current count for the rate limit window
func incrementCounter(clientIP string) int {
	now := time.Now()

	entry, exists := counters[clientIP]
	if !exists {
		// First request from this IP
		counters[clientIP] = &counterEntry{
			count:     1,
			resetTime: now.Add(rateLimitWindow),
		}
		return 1
	}

	// Check if the window has expired
	if now.After(entry.resetTime) {
		// Reset counter for new window
		entry.count = 1
		entry.resetTime = now.Add(rateLimitWindow)
		return 1
	}

	// Increment counter within current window
	entry.count++
	return entry.count
}

// cleanupCounters removes expired entries (optional optimization)
// In a real implementation, this would be called periodically
func cleanupCounters() {
	now := time.Now()
	for ip, entry := range counters {
		if now.After(entry.resetTime) {
			delete(counters, ip)
		}
	}
}
