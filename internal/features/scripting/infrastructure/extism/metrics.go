package extism

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics for WASM plugin pooling and execution
var (
	// Pool metrics
	poolHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wasm_plugin_pool_hits_total",
			Help: "Total number of cache hits when acquiring plugin instances",
		},
		[]string{"script_id", "script_name"},
	)

	poolMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wasm_plugin_pool_misses_total",
			Help: "Total number of cache misses (new instances created)",
		},
		[]string{"script_id", "script_name"},
	)

	poolEvictionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wasm_plugin_pool_evictions_total",
			Help: "Total number of instances evicted due to expiry",
		},
		[]string{"script_id", "script_name"},
	)

	poolIdleInstances = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wasm_plugin_pool_idle_instances",
			Help: "Current number of idle instances in pool",
		},
		[]string{"script_id", "script_name"},
	)

	poolActiveInstances = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wasm_plugin_pool_active_instances",
			Help: "Current number of active (executing) instances",
		},
		[]string{"script_id", "script_name"},
	)

	poolBackgroundCloseOps = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wasm_plugin_pool_background_close_ops",
			Help: "Number of pending background close operations",
		},
		[]string{"script_id", "script_name"},
	)

	// Circuit breaker metrics
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "wasm_plugin_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half_open, 2=open)",
		},
		[]string{"script_id", "script_name"},
	)

	circuitBreakerErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wasm_plugin_circuit_breaker_errors_total",
			Help: "Total number of errors tracked by circuit breaker",
		},
		[]string{"script_id", "script_name"},
	)

	circuitBreakerCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wasm_plugin_circuit_breaker_calls_total",
			Help: "Total number of calls tracked by circuit breaker",
		},
		[]string{"script_id", "script_name"},
	)

	// Execution metrics
	executionDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wasm_plugin_execution_duration_seconds",
			Help:    "WASM plugin execution duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to 16s
		},
		[]string{"script_id", "script_name", "status"},
	)

	executionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wasm_plugin_execution_total",
			Help: "Total number of WASM plugin executions",
		},
		[]string{"script_id", "script_name", "status"},
	)

	// Goroutine leak metrics
	runawayGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "wasm_plugin_runaway_goroutines",
			Help: "Current number of goroutines that continue after timeout",
		},
	)

	timeoutsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "wasm_plugin_timeouts_total",
			Help: "Total number of script timeouts",
		},
	)

	// Compilation cache metrics
	compilationCacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "wasm_compilation_cache_hits_total",
			Help: "Total number of WASM compilation cache hits",
		},
	)

	compilationCacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "wasm_compilation_cache_misses_total",
			Help: "Total number of WASM compilation cache misses",
		},
	)

	compilationDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "wasm_compilation_duration_seconds",
			Help:    "WASM compilation duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to 51s
		},
	)
)

// recordPoolMetrics updates Prometheus metrics from PoolMetrics
func recordPoolMetrics(scriptID, scriptName string, metrics PoolMetrics) {
	labels := prometheus.Labels{"script_id": scriptID, "script_name": scriptName}

	poolIdleInstances.With(labels).Set(float64(metrics.IdleCount))
	poolActiveInstances.With(labels).Set(float64(metrics.ActiveCount))
	poolBackgroundCloseOps.With(labels).Set(float64(metrics.BackgroundCloseOps))

	// Circuit breaker state: 0=closed, 1=half_open, 2=open
	circuitBreakerState.With(labels).Set(float64(metrics.CircuitState))
}

// recordPoolCounters updates counter metrics (call once per pool operation)
func recordPoolCounters(scriptID, scriptName string, hit, miss, eviction bool) {
	labels := prometheus.Labels{"script_id": scriptID, "script_name": scriptName}

	if hit {
		poolHitsTotal.With(labels).Inc()
	}
	if miss {
		poolMissesTotal.With(labels).Inc()
	}
	if eviction {
		poolEvictionsTotal.With(labels).Inc()
	}
}

// recordExecution updates execution metrics
func recordExecution(scriptID, scriptName string, durationSeconds float64, status string) {
	labels := prometheus.Labels{"script_id": scriptID, "script_name": scriptName, "status": status}

	executionDurationSeconds.With(labels).Observe(durationSeconds)
	executionTotal.With(labels).Inc()
}

// recordTimeout updates timeout metrics
func recordTimeout() {
	timeoutsTotal.Inc()
}

// recordRunawayGoroutine updates goroutine leak metrics
func recordRunawayGoroutine(count int64) {
	runawayGoroutines.Set(float64(count))
}

// recordCompilation updates compilation metrics
func recordCompilation(durationSeconds float64, cached bool) {
	compilationDurationSeconds.Observe(durationSeconds)
	if cached {
		compilationCacheHits.Inc()
	} else {
		compilationCacheMisses.Inc()
	}
}
