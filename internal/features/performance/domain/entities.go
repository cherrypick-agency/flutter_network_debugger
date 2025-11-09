package domain

import "time"

// TimeRange represents a time range for metrics
type TimeRange struct {
	From time.Time
	To   time.Time
}

// ResponseTimeStats contains response time statistics with percentiles
type ResponseTimeStats struct {
	P50       float64
	P95       float64
	P99       float64
	Min       float64
	Max       float64
	Mean      float64
	Count     int64
	TimeRange TimeRange
}

// EndpointPerformance represents performance metrics for a specific endpoint
type EndpointPerformance struct {
	Endpoint     string
	Method       string
	AvgDuration  float64
	P95Duration  float64
	RequestCount int64
	ErrorCount   int64
	ErrorRate    float64
}

// ThroughputStats contains request throughput statistics
type ThroughputStats struct {
	RequestsPerSecond float64
	TotalRequests     int64
	SuccessRate       float64
	ErrorRate         float64
	Status2xx         int64
	Status4xx         int64
	Status5xx         int64
	TimeRange         TimeRange
}

// BandwidthStats contains bandwidth usage statistics
type BandwidthStats struct {
	TotalBytesUp    int64
	TotalBytesDown  int64
	AvgRequestSize  float64
	AvgResponseSize float64
	TimeRange       TimeRange
}

// TimeSeriesPoint represents a single point in a time series
type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
}

// PerformanceOverview contains all performance metrics
type PerformanceOverview struct {
	ResponseTimeStats ResponseTimeStats
	ThroughputStats   ThroughputStats
	BandwidthStats    BandwidthStats
	TopSlowEndpoints  []EndpointPerformance
	TimeRange         TimeRange
}
