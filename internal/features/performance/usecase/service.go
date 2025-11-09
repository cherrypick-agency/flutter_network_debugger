package usecase

import (
	"context"
	"math"
	"sort"
	"time"

	"network-debugger/internal/domain"
	perf_domain "network-debugger/internal/features/performance/domain"
	"network-debugger/internal/usecase"
)

// Service provides performance analytics use cases
type Service struct {
	sessionSvc *usecase.SessionService
}

// NewService creates a new performance service
func NewService(sessionSvc *usecase.SessionService) *Service {
	return &Service{sessionSvc: sessionSvc}
}

// GetPerformanceOverview calculates all performance metrics for a time range
func (s *Service) GetPerformanceOverview(ctx context.Context, from, to time.Time) (*perf_domain.PerformanceOverview, error) {
	sessions, err := s.getHTTPSessionsInRange(ctx, from, to)
	if err != nil {
		return nil, err
	}

	timeRange := perf_domain.TimeRange{From: from, To: to}

	responseTimeStats := s.calculateResponseTimeStats(sessions, timeRange)
	throughputStats := s.calculateThroughputStats(sessions, timeRange)
	bandwidthStats := s.calculateBandwidthStats(sessions, timeRange)
	slowEndpoints := s.getTopSlowEndpoints(sessions, 10)

	return &perf_domain.PerformanceOverview{
		ResponseTimeStats: responseTimeStats,
		ThroughputStats:   throughputStats,
		BandwidthStats:    bandwidthStats,
		TopSlowEndpoints:  slowEndpoints,
		TimeRange:         timeRange,
	}, nil
}

// getHTTPSessionsInRange fetches HTTP sessions within time range
func (s *Service) getHTTPSessionsInRange(ctx context.Context, from, to time.Time) ([]domain.Session, error) {
	// Use time range filtering in SessionFilter for better performance
	filter := usecase.SessionFilter{
		Limit:    10000, // Large limit for analytics
		TimeFrom: &from, // Filter sessions started on or after 'from'
		TimeTo:   &to,   // Filter sessions started on or before 'to'
	}

	allSessions, _, err := s.sessionSvc.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Filter by HTTP kind (time filtering is now done by repository)
	var filtered []domain.Session
	for _, sess := range allSessions {
		if sess.Kind == "http" {
			filtered = append(filtered, sess)
		}
	}

	return filtered, nil
}

// calculateResponseTimeStats calculates response time statistics
func (s *Service) calculateResponseTimeStats(sessions []domain.Session, timeRange perf_domain.TimeRange) perf_domain.ResponseTimeStats {
	var durations []float64

	// Extract durations (we'll need to parse from httpMeta if available)
	// For MVP, using estimated duration from StartedAt to ClosedAt
	for _, sess := range sessions {
		if sess.ClosedAt != nil {
			duration := sess.ClosedAt.Sub(sess.StartedAt).Milliseconds()
			durations = append(durations, float64(duration))
		}
	}

	if len(durations) == 0 {
		return perf_domain.ResponseTimeStats{TimeRange: timeRange}
	}

	sort.Float64s(durations)

	p50 := percentile(durations, 0.50)
	p95 := percentile(durations, 0.95)
	p99 := percentile(durations, 0.99)

	min := durations[0]
	max := durations[len(durations)-1]
	mean := average(durations)

	return perf_domain.ResponseTimeStats{
		P50:       p50,
		P95:       p95,
		P99:       p99,
		Min:       min,
		Max:       max,
		Mean:      mean,
		Count:     int64(len(durations)),
		TimeRange: timeRange,
	}
}

// calculateThroughputStats calculates throughput and error rate statistics
func (s *Service) calculateThroughputStats(sessions []domain.Session, timeRange perf_domain.TimeRange) perf_domain.ThroughputStats {
	total := int64(len(sessions))
	if total == 0 {
		return perf_domain.ThroughputStats{TimeRange: timeRange}
	}

	var status2xx, status4xx, status5xx int64
	var successCount int64

	// Count status codes (would need httpMeta for real status codes)
	// For MVP, using error field as proxy
	for _, sess := range sessions {
		if sess.Error != nil {
			status5xx++
		} else {
			status2xx++
			successCount++
		}
	}

	durationSec := timeRange.To.Sub(timeRange.From).Seconds()
	if durationSec == 0 {
		durationSec = 1
	}

	rps := float64(total) / durationSec

	var successRate, errorRate float64
	if total > 0 {
		successRate = float64(successCount) / float64(total)
		errorRate = float64(status4xx+status5xx) / float64(total)
	}

	return perf_domain.ThroughputStats{
		RequestsPerSecond: rps,
		TotalRequests:     total,
		SuccessRate:       successRate,
		ErrorRate:         errorRate,
		Status2xx:         status2xx,
		Status4xx:         status4xx,
		Status5xx:         status5xx,
		TimeRange:         timeRange,
	}
}

// calculateBandwidthStats calculates bandwidth statistics
func (s *Service) calculateBandwidthStats(sessions []domain.Session, timeRange perf_domain.TimeRange) perf_domain.BandwidthStats {
	// For MVP, bandwidth data might not be available
	// This would require frame/event size tracking
	return perf_domain.BandwidthStats{
		TimeRange: timeRange,
	}
}

// getTopSlowEndpoints returns top N slowest endpoints
func (s *Service) getTopSlowEndpoints(sessions []domain.Session, limit int) []perf_domain.EndpointPerformance {
	// Group sessions by endpoint (target URL)
	endpointMap := make(map[string][]domain.Session)

	for _, sess := range sessions {
		endpointMap[sess.Target] = append(endpointMap[sess.Target], sess)
	}

	var endpoints []perf_domain.EndpointPerformance

	for endpoint, sessList := range endpointMap {
		var durations []float64
		var errorCount int64

		for _, sess := range sessList {
			if sess.ClosedAt != nil {
				duration := sess.ClosedAt.Sub(sess.StartedAt).Milliseconds()
				durations = append(durations, float64(duration))
			}
			if sess.Error != nil {
				errorCount++
			}
		}

		if len(durations) == 0 {
			continue
		}

		sort.Float64s(durations)

		p95 := percentile(durations, 0.95)
		avg := average(durations)
		count := int64(len(sessList))

		var errorRate float64
		if count > 0 {
			errorRate = float64(errorCount) / float64(count)
		}

		endpoints = append(endpoints, perf_domain.EndpointPerformance{
			Endpoint:     endpoint,
			Method:       "HTTP", // Would need httpMeta for real method
			AvgDuration:  avg,
			P95Duration:  p95,
			RequestCount: count,
			ErrorCount:   errorCount,
			ErrorRate:    errorRate,
		})
	}

	// Sort by P95 duration (slowest first)
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].P95Duration > endpoints[j].P95Duration
	})

	// Return top N
	if len(endpoints) > limit {
		endpoints = endpoints[:limit]
	}

	return endpoints
}

// Helper functions

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	index := p * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	fraction := index - float64(lower)
	return sorted[lower]*(1-fraction) + sorted[upper]*fraction
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
