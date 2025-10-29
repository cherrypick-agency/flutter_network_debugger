package httpapi

import (
	"bytes"
	"encoding/json"
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

func makeDepsWithMemory() *Deps {
	store := mem.NewStore(100, 1000, 0)
	s := uc.NewSessionService(store, store, store)
	return &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
}

func TestV1Capture_GetStartStopAndList(t *testing.T) {
	d := makeDepsWithMemory()
	h := NewRouterWithoutForwardProxy(d)
	// GET
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/capture", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("capture get")
	}
	// POST start
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/capture", bytes.NewReader([]byte(`{"action":"start"}`)))
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("capture start")
	}
	// create session to get capture id used
	_ = d.Svc.Create(contextWithNoCancel(), domain.Session{ID: "s1", StartedAt: time.Unix(0, 0).UTC(), Target: "http://x"})
	// list captures
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/captures", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("captures list")
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if _, ok := body["items"].([]any); !ok {
		t.Fatalf("captures items")
	}
	// POST stop
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/capture", bytes.NewReader([]byte(`{"action":"stop"}`)))
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("capture stop")
	}
}

func TestV1Capture_BadAction(t *testing.T) {
	d := makeDepsWithMemory()
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/capture", bytes.NewReader([]byte(`{"action":"bad"}`)))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad action -> 400")
	}
}
