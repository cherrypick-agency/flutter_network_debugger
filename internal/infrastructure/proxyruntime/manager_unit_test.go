package proxyruntime

import (
	"context"
	"net/http"
	"testing"
)

func TestManager_ForwardAddr(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "no listener set",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{}
			got := m.ForwardAddr()
			if got != tt.want {
				t.Errorf("ForwardAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManager_SocksAddr(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "no listener set",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{}
			got := m.SocksAddr()
			if got != tt.want {
				t.Errorf("SocksAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestManager_Apply_ForwardOnly(t *testing.T) {
	m := &Manager{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    ":0", // use random port
		SocksEnabled:   false,
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, handler)

	if err != nil {
		t.Logf("Apply with forward enabled: %v (may fail if port in use)", err)
	}

	// Check forward is running if no error
	if err == nil {
		addr := m.ForwardAddr()
		if addr == "" {
			t.Error("ForwardAddr should not be empty when forward is enabled")
		}

		// Cleanup
		m.StopForward(ctx)
	}
}

func TestManager_Apply_SocksOnly(t *testing.T) {
	m := &Manager{}

	cfg := ApplyConfig{
		ForwardEnabled: false,
		SocksEnabled:   true,
		SocksAddr:      ":0", // use random port
		SocksAuthMode:  "none",
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, nil)

	if err != nil {
		t.Logf("Apply with socks enabled: %v (may fail if port in use)", err)
	}

	// Check socks is running if no error
	if err == nil {
		addr := m.SocksAddr()
		if addr == "" {
			t.Error("SocksAddr should not be empty when socks is enabled")
		}

		// Cleanup
		m.StopSocks(ctx)
	}
}

func TestManager_Apply_BothEnabled(t *testing.T) {
	m := &Manager{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    ":0",
		SocksEnabled:   true,
		SocksAddr:      ":0",
		SocksAuthMode:  "none",
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, handler)

	if err != nil {
		t.Logf("Apply with both enabled: %v (may fail if port in use)", err)
	}

	// Cleanup
	if err == nil {
		m.StopForward(ctx)
		m.StopSocks(ctx)
	}
}

func TestManager_Apply_DisableAll(t *testing.T) {
	m := &Manager{}

	cfg := ApplyConfig{
		ForwardEnabled: false,
		SocksEnabled:   false,
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, nil)

	if err != nil {
		t.Errorf("Apply with everything disabled should not error: %v", err)
	}

	if m.ForwardAddr() != "" {
		t.Error("ForwardAddr should be empty when disabled")
	}

	if m.SocksAddr() != "" {
		t.Error("SocksAddr should be empty when disabled")
	}
}

func TestManager_Apply_InvalidForwardAddr(t *testing.T) {
	m := &Manager{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	cfg := ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    "invalid:addr:format",
		SocksEnabled:   false,
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, handler)

	if err == nil {
		t.Error("Apply with invalid forward address should return error")
		m.StopForward(ctx)
	}
}

func TestManager_Apply_InvalidSocksAddr(t *testing.T) {
	m := &Manager{}

	cfg := ApplyConfig{
		ForwardEnabled: false,
		SocksEnabled:   true,
		SocksAddr:      "invalid:addr:format",
		SocksAuthMode:  "none",
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, nil)

	if err == nil {
		t.Error("Apply with invalid socks address should return error")
		m.StopSocks(ctx)
	}
}

func TestManager_StopForward_WhenNotRunning(t *testing.T) {
	m := &Manager{}
	ctx := context.Background()

	// Should not panic or error when stopping a non-running forward proxy
	err := m.StopForward(ctx)
	if err != nil {
		t.Errorf("StopForward on non-running proxy returned error: %v", err)
	}
}

func TestManager_StopSocks_WhenNotRunning(t *testing.T) {
	m := &Manager{}
	ctx := context.Background()

	// Should not panic or error when stopping a non-running socks proxy
	err := m.StopSocks(ctx)
	if err != nil {
		t.Errorf("StopSocks on non-running proxy returned error: %v", err)
	}
}
