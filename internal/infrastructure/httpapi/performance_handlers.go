package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	perf_domain "network-debugger/internal/features/performance/domain"
)

// Performance DTOs

type performanceOverviewDTO struct {
	ResponseTimeStats responseTimeStatsDTO     `json:"responseTimeStats"`
	ThroughputStats   throughputStatsDTO       `json:"throughputStats"`
	BandwidthStats    bandwidthStatsDTO        `json:"bandwidthStats"`
	TopSlowEndpoints  []endpointPerformanceDTO `json:"topSlowEndpoints"`
	TimeRange         timeRangeDTO             `json:"timeRange"`
}

type responseTimeStatsDTO struct {
	P50       float64      `json:"p50"`
	P95       float64      `json:"p95"`
	P99       float64      `json:"p99"`
	Min       float64      `json:"min"`
	Max       float64      `json:"max"`
	Mean      float64      `json:"mean"`
	Count     int64        `json:"count"`
	TimeRange timeRangeDTO `json:"timeRange"`
}

type throughputStatsDTO struct {
	RequestsPerSecond float64      `json:"requestsPerSecond"`
	TotalRequests     int64        `json:"totalRequests"`
	SuccessRate       float64      `json:"successRate"`
	ErrorRate         float64      `json:"errorRate"`
	Status2xx         int64        `json:"status2xx"`
	Status4xx         int64        `json:"status4xx"`
	Status5xx         int64        `json:"status5xx"`
	TimeRange         timeRangeDTO `json:"timeRange"`
}

type bandwidthStatsDTO struct {
	TotalBytesUp    int64        `json:"totalBytesUp"`
	TotalBytesDown  int64        `json:"totalBytesDown"`
	AvgRequestSize  float64      `json:"avgRequestSize"`
	AvgResponseSize float64      `json:"avgResponseSize"`
	TimeRange       timeRangeDTO `json:"timeRange"`
}

type endpointPerformanceDTO struct {
	Endpoint     string  `json:"endpoint"`
	Method       string  `json:"method"`
	AvgDuration  float64 `json:"avgDuration"`
	P95Duration  float64 `json:"p95Duration"`
	RequestCount int64   `json:"requestCount"`
	ErrorCount   int64   `json:"errorCount"`
	ErrorRate    float64 `json:"errorRate"`
}

type timeRangeDTO struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// Converters

func performanceOverviewToDTO(d *perf_domain.PerformanceOverview) performanceOverviewDTO {
	slowEndpoints := make([]endpointPerformanceDTO, 0, len(d.TopSlowEndpoints))
	for _, ep := range d.TopSlowEndpoints {
		slowEndpoints = append(slowEndpoints, endpointPerformanceDTO{
			Endpoint:     ep.Endpoint,
			Method:       ep.Method,
			AvgDuration:  ep.AvgDuration,
			P95Duration:  ep.P95Duration,
			RequestCount: ep.RequestCount,
			ErrorCount:   ep.ErrorCount,
			ErrorRate:    ep.ErrorRate,
		})
	}

	return performanceOverviewDTO{
		ResponseTimeStats: responseTimeStatsDTO{
			P50:   d.ResponseTimeStats.P50,
			P95:   d.ResponseTimeStats.P95,
			P99:   d.ResponseTimeStats.P99,
			Min:   d.ResponseTimeStats.Min,
			Max:   d.ResponseTimeStats.Max,
			Mean:  d.ResponseTimeStats.Mean,
			Count: d.ResponseTimeStats.Count,
			TimeRange: timeRangeDTO{
				From: d.ResponseTimeStats.TimeRange.From,
				To:   d.ResponseTimeStats.TimeRange.To,
			},
		},
		ThroughputStats: throughputStatsDTO{
			RequestsPerSecond: d.ThroughputStats.RequestsPerSecond,
			TotalRequests:     d.ThroughputStats.TotalRequests,
			SuccessRate:       d.ThroughputStats.SuccessRate,
			ErrorRate:         d.ThroughputStats.ErrorRate,
			Status2xx:         d.ThroughputStats.Status2xx,
			Status4xx:         d.ThroughputStats.Status4xx,
			Status5xx:         d.ThroughputStats.Status5xx,
			TimeRange: timeRangeDTO{
				From: d.ThroughputStats.TimeRange.From,
				To:   d.ThroughputStats.TimeRange.To,
			},
		},
		BandwidthStats: bandwidthStatsDTO{
			TotalBytesUp:    d.BandwidthStats.TotalBytesUp,
			TotalBytesDown:  d.BandwidthStats.TotalBytesDown,
			AvgRequestSize:  d.BandwidthStats.AvgRequestSize,
			AvgResponseSize: d.BandwidthStats.AvgResponseSize,
			TimeRange: timeRangeDTO{
				From: d.BandwidthStats.TimeRange.From,
				To:   d.BandwidthStats.TimeRange.To,
			},
		},
		TopSlowEndpoints: slowEndpoints,
		TimeRange: timeRangeDTO{
			From: d.TimeRange.From,
			To:   d.TimeRange.To,
		},
	}
}

// Handlers

func (d *Deps) handlePerformanceOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Parse time range from query params
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_FROM", "from must be RFC3339 format", nil)
			return
		}
	} else {
		// Default: last hour
		from = time.Now().Add(-1 * time.Hour)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_TO", "to must be RFC3339 format", nil)
			return
		}
	} else {
		// Default: now
		to = time.Now()
	}

	// Validate time range
	if to.Before(from) {
		writeError(w, http.StatusBadRequest, "INVALID_RANGE", "to must be after from", nil)
		return
	}

	if d.PerformanceSvc == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Performance service not available", nil)
		return
	}

	ctx := r.Context()
	overview, err := d.PerformanceSvc.GetPerformanceOverview(ctx, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PERFORMANCE_FETCH_FAILED", err.Error(), nil)
		return
	}

	dto := performanceOverviewToDTO(overview)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto)
}
