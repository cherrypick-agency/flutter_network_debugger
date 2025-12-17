package httpapi

import (
	"context"
	"io"
	"net/http/httptest"
	"time"

	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"

	"github.com/rs/zerolog"
)

// stubRepoNotFound always returns not found for GetSession
type stubRepoNotFound struct{}

func (stubRepoNotFound) CreateSession(context.Context, domain.Session) error { return nil }
func (stubRepoNotFound) GetSession(context.Context, string) (domain.Session, bool, error) {
	return domain.Session{}, false, nil // Always not found
}
func (stubRepoNotFound) DeleteSession(context.Context, string) error { return nil }
func (stubRepoNotFound) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) {
	return nil, 0, nil
}
func (stubRepoNotFound) IncrementCounters(context.Context, string, domain.Frame) error { return nil }
func (stubRepoNotFound) SetClosed(context.Context, string, time.Time, *string) error   { return nil }
func (stubRepoNotFound) ClearAllSessions(context.Context) error                        { return nil }
func (stubRepoNotFound) DeleteImportedSessions(context.Context) error                  { return nil }
func (stubRepoNotFound) AppendFrame(context.Context, string, domain.Frame) error       { return nil }
func (stubRepoNotFound) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) {
	return nil, "", nil
}
func (stubRepoNotFound) GetFrameByID(context.Context, string, string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
}
func (stubRepoNotFound) UpdateFrameBodyFile(context.Context, string, string, string) error {
	return nil
}
func (stubRepoNotFound) AppendEvent(context.Context, string, domain.Event) error { return nil }
func (stubRepoNotFound) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) {
	return nil, "", nil
}
func (stubRepoNotFound) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error {
	return nil
}
func (stubRepoNotFound) ListHTTPTransactions(context.Context, string, string, int) ([]domain.HTTPTransaction, string, error) {
	return nil, "", nil
}

func TestRouter_BasicEndpoints(t *testing.T) {
	// deps with stubs
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	logger := zerolog.New(io.Discard)
	d := &Deps{Logger: &logger, Cfg: cfgpkg.Config{CORSAllowOrigin: "*", PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)

	// healthz
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("healthz")
	}

	// readyz
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/readyz", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("readyz")
	}

	// api version
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/version", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("version")
	}

	// settings get
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/_api/v1/settings", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("settings get")
	}

	// mitm status
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/_api/v1/mitm/status", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("mitm status")
	}
}

func TestRouter_SessionsEndpointsAndMetrics(t *testing.T) {
	s := uc.NewSessionService(stubRepoNotFound{}, stubRepoNotFound{}, stubRepoNotFound{})
	logger := zerolog.New(io.Discard)
	d := &Deps{Logger: &logger, Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)

	// list sessions
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/_api/v1/sessions", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("list sessions")
	}

	// get session by id (not found)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/_api/v1/sessions/unknown", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 404 {
		t.Fatalf("session by id 404 expected")
	}

	// metrics
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/metrics", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("metrics")
	}
}
