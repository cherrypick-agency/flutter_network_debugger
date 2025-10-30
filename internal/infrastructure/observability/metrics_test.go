package observability

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()

	if m == nil {
		t.Fatal("NewMetrics returned nil")
	}

	if m.ActiveSessions == nil {
		t.Error("ActiveSessions should not be nil")
	}
	if m.FramesTotal == nil {
		t.Error("FramesTotal should not be nil")
	}
	if m.ProxyErrorsTotal == nil {
		t.Error("ProxyErrorsTotal should not be nil")
	}
	if m.EvictionsTotal == nil {
		t.Error("EvictionsTotal should not be nil")
	}
}

func TestMetrics_Registry(t *testing.T) {
	m := NewMetrics()

	registry := m.Registry()
	if registry == nil {
		t.Fatal("Registry() returned nil")
	}

	// Initialize CounterVec metrics with labels so they appear in the registry
	m.FramesTotal.WithLabelValues("inbound", "text").Add(0)
	m.ProxyErrorsTotal.WithLabelValues("dial").Add(0)

	// Verify metrics are registered by gathering
	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	if len(families) == 0 {
		t.Error("No metrics found in registry")
	}

	// Verify expected metric names exist
	expectedNames := map[string]bool{
		"network_debugger_active_sessions":    false,
		"network_debugger_frames_total":       false,
		"network_debugger_proxy_errors_total": false,
		"network_debugger_evictions_total":    false,
	}

	for _, family := range families {
		if _, exists := expectedNames[family.GetName()]; exists {
			expectedNames[family.GetName()] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("Expected metric %s not found in registry", name)
		}
	}
}

func TestMetrics_ActiveSessions(t *testing.T) {
	m := NewMetrics()

	// Test setting gauge value
	m.ActiveSessions.Set(10)

	// Gather metrics to verify
	families, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, family := range families {
		if family.GetName() == "network_debugger_active_sessions" {
			found = true
			if len(family.GetMetric()) > 0 {
				val := family.GetMetric()[0].GetGauge().GetValue()
				if val != 10 {
					t.Errorf("ActiveSessions value = %v, want 10", val)
				}
			}
		}
	}

	if !found {
		t.Error("active_sessions metric not found")
	}
}

func TestMetrics_FramesTotal(t *testing.T) {
	m := NewMetrics()

	// Test incrementing counter with labels
	m.FramesTotal.WithLabelValues("inbound", "text").Inc()
	m.FramesTotal.WithLabelValues("outbound", "binary").Add(5)

	// Gather metrics to verify
	families, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, family := range families {
		if family.GetName() == "network_debugger_frames_total" {
			found = true
			// We should have 2 metric entries (one for each label combination)
			if len(family.GetMetric()) != 2 {
				t.Errorf("Expected 2 metrics, got %d", len(family.GetMetric()))
			}
		}
	}

	if !found {
		t.Error("frames_total metric not found")
	}
}

func TestMetrics_ProxyErrorsTotal(t *testing.T) {
	m := NewMetrics()

	// Test incrementing counter with labels
	m.ProxyErrorsTotal.WithLabelValues("dial").Inc()
	m.ProxyErrorsTotal.WithLabelValues("upgrade").Add(3)

	// Gather metrics to verify
	families, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, family := range families {
		if family.GetName() == "network_debugger_proxy_errors_total" {
			found = true
			if len(family.GetMetric()) != 2 {
				t.Errorf("Expected 2 metrics, got %d", len(family.GetMetric()))
			}
		}
	}

	if !found {
		t.Error("proxy_errors_total metric not found")
	}
}

func TestMetrics_EvictionsTotal(t *testing.T) {
	m := NewMetrics()

	// Test incrementing counter
	m.EvictionsTotal.Inc()
	m.EvictionsTotal.Add(4)

	// Gather metrics to verify
	families, err := m.Registry().Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool
	for _, family := range families {
		if family.GetName() == "network_debugger_evictions_total" {
			found = true
			if len(family.GetMetric()) > 0 {
				val := family.GetMetric()[0].GetCounter().GetValue()
				if val != 5 {
					t.Errorf("EvictionsTotal value = %v, want 5", val)
				}
			}
		}
	}

	if !found {
		t.Error("evictions_total metric not found")
	}
}

func TestMetrics_NoCollisions(t *testing.T) {
	// Create two separate metric instances to ensure they use separate registries
	m1 := NewMetrics()
	m2 := NewMetrics()

	// Each should have its own registry
	if m1.Registry() == m2.Registry() {
		t.Error("Expected different registry instances")
	}

	// Test that they can be used independently
	m1.ActiveSessions.Set(10)
	m2.ActiveSessions.Set(20)

	families1, _ := m1.Registry().Gather()
	families2, _ := m2.Registry().Gather()

	// Both should have metrics
	if len(families1) == 0 || len(families2) == 0 {
		t.Error("Both registries should have metrics")
	}
}

func TestMetrics_InterfaceCompliance(t *testing.T) {
	// Verify that metrics implement prometheus interfaces
	m := NewMetrics()

	// Test that metrics can be used as prometheus types
	var _ prometheus.Gauge = m.ActiveSessions
	var _ prometheus.Counter = m.EvictionsTotal
	var _ *prometheus.CounterVec = m.FramesTotal
	var _ *prometheus.CounterVec = m.ProxyErrorsTotal
}
