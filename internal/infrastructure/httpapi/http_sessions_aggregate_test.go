package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	mem "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestV1SessionsAggregate(t *testing.T) {
	store := mem.NewStore(100, 100, 0)
	s := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	d := &Deps{Logger: &logger, Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	_ = s.Create(contextWithNoCancel(), domain.Session{ID: "a", Target: "http://a.example/x", StartedAt: time.Unix(0, 0).UTC()})
	_ = s.Create(contextWithNoCancel(), domain.Session{ID: "b", Target: "http://b.example/x", StartedAt: time.Unix(0, 0).UTC()})
	_ = s.Create(contextWithNoCancel(), domain.Session{ID: "a2", Target: "http://a.example/y", StartedAt: time.Unix(0, 0).UTC()})
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/aggregate", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("aggregate")
	}
}

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "http with default port 80",
			input:    "http://google.com:80/path",
			expected: "google.com",
		},
		{
			name:     "https with default port 443",
			input:    "https://google.com:443/path",
			expected: "google.com",
		},
		{
			name:     "http without port",
			input:    "http://google.com/path",
			expected: "google.com",
		},
		{
			name:     "https without port",
			input:    "https://google.com/path",
			expected: "google.com",
		},
		{
			name:     "http with non-default port",
			input:    "http://google.com:8080/path",
			expected: "google.com:8080",
		},
		{
			name:     "https with non-default port",
			input:    "https://google.com:8443/path",
			expected: "google.com:8443",
		},
		{
			name:     "with userinfo",
			input:    "http://user:pass@google.com/path",
			expected: "google.com",
		},
		{
			name:     "with userinfo and port",
			input:    "https://user:pass@google.com:443/path",
			expected: "google.com",
		},
		{
			name:     "no scheme",
			input:    "google.com/path",
			expected: "google.com",
		},
		{
			name:     "ws with default http port",
			input:    "ws://google.com:80/path",
			expected: "google.com:80",
		},
		{
			name:     "wss with default https port",
			input:    "wss://google.com:443/path",
			expected: "google.com:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeHost(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeHost(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
