package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	registry         *prometheus.Registry
	ActiveSessions   prometheus.Gauge
	FramesTotal      *prometheus.CounterVec
	ProxyErrorsTotal *prometheus.CounterVec
	EvictionsTotal   prometheus.Counter
}

func NewMetrics() *Metrics {
	r := prometheus.NewRegistry()
	m := &Metrics{
		registry: r,
		ActiveSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "network_debugger",
			Name:      "active_sessions",
			Help:      "Number of active sessions",
		}),
		FramesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "network_debugger",
			Name:      "frames_total",
			Help:      "Total frames proxied",
		}, []string{"direction", "opcode"}),
		ProxyErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "network_debugger",
			Name:      "proxy_errors_total",
			Help:      "Total proxy errors by stage",
		}, []string{"stage"}),
		EvictionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "network_debugger",
			Name:      "evictions_total",
			Help:      "Total evicted sessions",
		}),
	}
	r.MustRegister(m.ActiveSessions, m.FramesTotal, m.ProxyErrorsTotal, m.EvictionsTotal)
	return m
}

func (m *Metrics) Registry() *prometheus.Registry { return m.registry }
