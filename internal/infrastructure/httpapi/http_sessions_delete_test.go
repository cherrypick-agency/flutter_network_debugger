package httpapi

import (
	"net/http"
	"net/http/httptest"
	mem "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
	"time"
)

func TestLegacyAndV1_DeleteSessionsAndByID(t *testing.T) {
	store := mem.NewStore(100, 100, 0)
	s := uc.NewSessionService(store, store, store)
	d := &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	_ = s.Create(contextWithNoCancel(), domain.Session{ID: "s1", StartedAt: time.Unix(0, 0).UTC()})
	h := NewRouterWithoutForwardProxy(d)

	// legacy delete all
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/sessions", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("legacy delete all")
	}

	_ = s.Create(contextWithNoCancel(), domain.Session{ID: "s2", StartedAt: time.Unix(0, 0).UTC()})
	// v1 delete all
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/_api/v1/sessions", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("v1 delete all")
	}

	_ = s.Create(contextWithNoCancel(), domain.Session{ID: "s3", StartedAt: time.Unix(0, 0).UTC()})
	// delete by id (legacy)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/sessions/s3", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete by id")
	}
}
