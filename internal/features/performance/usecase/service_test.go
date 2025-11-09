package usecase

import (
	"testing"
	"time"

	"network-debugger/internal/domain"
	perf_domain "network-debugger/internal/features/performance/domain"
)

func TestCalculateResponseTimeStats(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	sessions := []domain.Session{
		{
			Kind:      "http",
			StartedAt: now.Add(-100 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Target:    "http://api.example.com/users",
		},
		{
			Kind:      "http",
			StartedAt: now.Add(-200 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Target:    "http://api.example.com/posts",
		},
		{
			Kind:      "http",
			StartedAt: now.Add(-300 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Target:    "http://api.example.com/comments",
		},
	}

	timeRange := perf_domain.TimeRange{
		From: now.Add(-1 * time.Hour),
		To:   now,
	}

	stats := svc.calculateResponseTimeStats(sessions, timeRange)

	// Check basic stats
	if stats.Count != 3 {
		t.Errorf("expected count 3, got %d", stats.Count)
	}

	// Min should be ~100ms
	if stats.Min < 90 || stats.Min > 110 {
		t.Errorf("expected min ~100ms, got %v", stats.Min)
	}

	// Max should be ~300ms
	if stats.Max < 290 || stats.Max > 310 {
		t.Errorf("expected max ~300ms, got %v", stats.Max)
	}

	// Mean should be ~200ms
	if stats.Mean < 190 || stats.Mean > 210 {
		t.Errorf("expected mean ~200ms, got %v", stats.Mean)
	}

	// P50 should be ~200ms
	if stats.P50 < 190 || stats.P50 > 210 {
		t.Errorf("expected P50 ~200ms, got %v", stats.P50)
	}
}

func TestCalculateResponseTimeStats_EmptySessions(t *testing.T) {
	svc := &Service{}
	timeRange := perf_domain.TimeRange{
		From: time.Now().Add(-1 * time.Hour),
		To:   time.Now(),
	}

	stats := svc.calculateResponseTimeStats([]domain.Session{}, timeRange)

	if stats.Count != 0 {
		t.Errorf("expected count 0, got %d", stats.Count)
	}
	if stats.Mean != 0 {
		t.Errorf("expected mean 0, got %v", stats.Mean)
	}
}

func TestCalculateThroughputStats(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	// Create sessions: 3 successful, 1 with error
	sessions := []domain.Session{
		{Kind: "http", StartedAt: now, Error: nil},
		{Kind: "http", StartedAt: now, Error: nil},
		{Kind: "http", StartedAt: now, Error: nil},
		{Kind: "http", StartedAt: now, Error: ptrString("timeout")},
	}

	timeRange := perf_domain.TimeRange{
		From: now.Add(-10 * time.Second),
		To:   now,
	}

	stats := svc.calculateThroughputStats(sessions, timeRange)

	// Total requests
	if stats.TotalRequests != 4 {
		t.Errorf("expected 4 total requests, got %d", stats.TotalRequests)
	}

	// Success rate: 3/4 = 0.75
	if stats.SuccessRate < 0.74 || stats.SuccessRate > 0.76 {
		t.Errorf("expected success rate ~0.75, got %v", stats.SuccessRate)
	}

	// Error rate: 1/4 = 0.25
	if stats.ErrorRate < 0.24 || stats.ErrorRate > 0.26 {
		t.Errorf("expected error rate ~0.25, got %v", stats.ErrorRate)
	}

	// Status codes
	if stats.Status2xx != 3 {
		t.Errorf("expected 3 2xx responses, got %d", stats.Status2xx)
	}
	if stats.Status5xx != 1 {
		t.Errorf("expected 1 5xx response, got %d", stats.Status5xx)
	}

	// RPS: 4 requests / 10 seconds = 0.4
	if stats.RequestsPerSecond < 0.39 || stats.RequestsPerSecond > 0.41 {
		t.Errorf("expected RPS ~0.4, got %v", stats.RequestsPerSecond)
	}
}

func TestCalculateThroughputStats_EmptySessions(t *testing.T) {
	svc := &Service{}
	timeRange := perf_domain.TimeRange{
		From: time.Now().Add(-1 * time.Hour),
		To:   time.Now(),
	}

	stats := svc.calculateThroughputStats([]domain.Session{}, timeRange)

	if stats.TotalRequests != 0 {
		t.Errorf("expected 0 total requests, got %d", stats.TotalRequests)
	}
	if stats.RequestsPerSecond != 0 {
		t.Errorf("expected 0 RPS, got %v", stats.RequestsPerSecond)
	}
}

func TestGetTopSlowEndpoints(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	// Create sessions with varying durations for different endpoints
	sessions := []domain.Session{
		// Endpoint A: slow (300ms avg)
		{
			Target:    "http://api.example.com/slow",
			StartedAt: now.Add(-300 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Error:     nil,
		},
		{
			Target:    "http://api.example.com/slow",
			StartedAt: now.Add(-350 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Error:     nil,
		},
		// Endpoint B: fast (100ms avg)
		{
			Target:    "http://api.example.com/fast",
			StartedAt: now.Add(-100 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Error:     nil,
		},
		{
			Target:    "http://api.example.com/fast",
			StartedAt: now.Add(-120 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Error:     nil,
		},
	}

	endpoints := svc.getTopSlowEndpoints(sessions, 10)

	// Should have 2 endpoints
	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(endpoints))
	}

	// First endpoint should be the slowest
	if endpoints[0].Endpoint != "http://api.example.com/slow" {
		t.Errorf("expected slowest endpoint first, got %s", endpoints[0].Endpoint)
	}

	// Check request counts
	if endpoints[0].RequestCount != 2 {
		t.Errorf("expected 2 requests for slow endpoint, got %d", endpoints[0].RequestCount)
	}

	// Avg duration for slow endpoint should be ~325ms
	if endpoints[0].AvgDuration < 320 || endpoints[0].AvgDuration > 330 {
		t.Errorf("expected avg ~325ms, got %v", endpoints[0].AvgDuration)
	}
}

func TestGetTopSlowEndpoints_WithLimit(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	// Create sessions for 5 different endpoints
	sessions := make([]domain.Session, 0)
	for i := 0; i < 5; i++ {
		sessions = append(sessions, domain.Session{
			Target:    "http://api.example.com/endpoint" + string(rune('A'+i)),
			StartedAt: now.Add(-time.Duration(i*100) * time.Millisecond),
			ClosedAt:  ptrTime(now),
		})
	}

	endpoints := svc.getTopSlowEndpoints(sessions, 3)

	// Should be limited to 3
	if len(endpoints) != 3 {
		t.Errorf("expected limit of 3 endpoints, got %d", len(endpoints))
	}
}

func TestGetTopSlowEndpoints_ErrorRate(t *testing.T) {
	svc := &Service{}
	now := time.Now()

	sessions := []domain.Session{
		{
			Target:    "http://api.example.com/flaky",
			StartedAt: now.Add(-100 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Error:     nil,
		},
		{
			Target:    "http://api.example.com/flaky",
			StartedAt: now.Add(-100 * time.Millisecond),
			ClosedAt:  ptrTime(now),
			Error:     ptrString("timeout"),
		},
	}

	endpoints := svc.getTopSlowEndpoints(sessions, 10)

	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(endpoints))
	}

	// Error rate: 1/2 = 0.5
	if endpoints[0].ErrorRate < 0.49 || endpoints[0].ErrorRate > 0.51 {
		t.Errorf("expected error rate ~0.5, got %v", endpoints[0].ErrorRate)
	}

	// Error count
	if endpoints[0].ErrorCount != 1 {
		t.Errorf("expected error count 1, got %d", endpoints[0].ErrorCount)
	}
}

func TestPercentile(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}

	tests := []struct {
		name     string
		p        float64
		expected float64
		delta    float64
	}{
		{"P50", 0.50, 55, 5},
		{"P95", 0.95, 95.5, 5},
		{"P99", 0.99, 99.1, 5},
		{"P0", 0.0, 10, 1},
		{"P100", 1.0, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := percentile(values, tt.p)
			if result < tt.expected-tt.delta || result > tt.expected+tt.delta {
				t.Errorf("expected ~%v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPercentile_EmptySlice(t *testing.T) {
	result := percentile([]float64{}, 0.5)
	if result != 0 {
		t.Errorf("expected 0 for empty slice, got %v", result)
	}
}

func TestAverage(t *testing.T) {
	values := []float64{10, 20, 30, 40, 50}
	result := average(values)

	expected := 30.0
	if result != expected {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestAverage_EmptySlice(t *testing.T) {
	result := average([]float64{})
	if result != 0 {
		t.Errorf("expected 0 for empty slice, got %v", result)
	}
}

// Helper functions

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrString(s string) *string {
	return &s
}
