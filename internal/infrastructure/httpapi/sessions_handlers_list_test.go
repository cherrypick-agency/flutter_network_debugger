package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

func setupSessionsHandlerDeps(t *testing.T) *Deps {
	t.Helper()
	logger := zerolog.New(io.Discard)
	m := obs.NewMetrics()
	cfg := config.Config{CORSAllowOrigin: "*"}
	d := &Deps{
		Cfg:     cfg,
		Logger:  &logger,
		Metrics: m,
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
	}
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)
	return d
}

func TestHandleV1ListSessions_GET_DefaultParams(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithQuery(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?q=test&limit=50", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithTypes(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?types=http,ws&status=200,400", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithLimit(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name  string
		limit string
	}{
		{"default limit", ""},
		{"custom limit", "50"},
		{"too large limit", "2000"}, // Should cap at 1000
		{"zero limit", "0"},         // Should use default 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/_api/v1/sessions"
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			d.handleV1ListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleV1ListSessions_GET_WithOffset(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?offset=10", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithCaptureID(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name      string
		captureID string
	}{
		{"current capture", "current"},
		{"specific capture", "5"},
		{"invalid capture", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captureId="+tt.captureID, nil)
			w := httptest.NewRecorder()

			d.handleV1ListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleV1ListSessions_GET_WithIncludeUnassigned(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	tests := []struct {
		name  string
		value string
	}{
		{"true", "true"},
		{"1", "1"},
		{"false", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?includeUnassigned="+tt.value, nil)
			w := httptest.NewRecorder()

			d.handleV1ListSessions(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
			}
		})
	}
}

func TestHandleV1ListSessions_GET_WithCapturesAll(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captures=all", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithScanGraphQL(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?scan=graphql", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_GET_WithTarget(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?_target=example.com", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ListSessions_DELETE(t *testing.T) {
	d := setupSessionsHandlerDeps(t)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleV1ListSessions_DELETE_WithMonitor(t *testing.T) {
	d := setupSessionsHandlerDeps(t)
	d.Monitor = NewMonitorHub()

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions", nil)
	w := httptest.NewRecorder()

	d.handleV1ListSessions(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}
