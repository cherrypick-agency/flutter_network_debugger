package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"network-debugger/internal/adapters/storage/memory"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"

	"github.com/rs/zerolog"
)

func TestHTTPProxy_SessionsVisible_WhenCapturePaused_IncludeUnassigned(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	store := memory.NewStore(500, 10_000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(d)

	// 1) Create a session while recording (assigned to current capture).
	store.StartCapture()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("proxy status (assigned): %d", rr.Code)
	}

	// 2) Pause capture and create another session (unassigned captureId=nil).
	store.StopCapture()
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("proxy status (unassigned): %d", rr.Code)
	}

	// 3) List sessions for current capture and include unassigned.
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captureId=current&includeUnassigned=true&limit=1000", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sessions list status: %d", rr.Code)
	}

	var out struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	if len(out.Items) != 2 {
		t.Fatalf("items=%d want=2", len(out.Items))
	}

	// Ensure we got both: one assigned captureId and one unassigned (nil).
	var hasAssigned, hasUnassigned bool
	for _, it := range out.Items {
		if it["captureId"] == nil {
			hasUnassigned = true
		} else {
			hasAssigned = true
		}
	}
	if !hasAssigned || !hasUnassigned {
		t.Fatalf("assigned=%v unassigned=%v items=%v", hasAssigned, hasUnassigned, out.Items)
	}
}

func TestHTTPProxy_SessionsFilterByStatusGroup_3xxVs2xx(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/uk/")
		w.WriteHeader(http.StatusFound) // 302
	}))
	defer upstream.Close()

	store := memory.NewStore(500, 10_000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(d)

	// Make one reverse-proxy request that results in 302.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("proxy status: %d", rr.Code)
	}

	// sessions?status=3xx should include it
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captures=all&includeUnassigned=true&limit=1000&status=3xx", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sessions list 3xx status: %d", rr.Code)
	}
	var out3 struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out3); err != nil {
		t.Fatalf("decode sessions 3xx: %v", err)
	}
	if len(out3.Items) != 1 {
		t.Fatalf("3xx items=%d want=1", len(out3.Items))
	}

	// sessions?status=2xx should not include it
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captures=all&includeUnassigned=true&limit=1000&status=2xx", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sessions list 2xx status: %d", rr.Code)
	}
	var out2 struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out2); err != nil {
		t.Fatalf("decode sessions 2xx: %v", err)
	}
	if len(out2.Items) != 0 {
		t.Fatalf("2xx items=%d want=0", len(out2.Items))
	}

	// sanity: verify status code is present in httpMeta (helps explain UI filters)
	meta, _ := out3.Items[0]["httpMeta"].(map[string]any)
	if meta == nil {
		t.Fatalf("expected httpMeta in session: %v", out3.Items[0])
	}
	st, _ := meta["status"].(float64)
	if int(st) != http.StatusFound {
		t.Fatalf("httpMeta.status=%v want=302", meta["status"])
	}
}

func TestHTTPProxy_SessionsFilterByTypeTag_Http(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer upstream.Close()

	store := memory.NewStore(500, 10_000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	d := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(d)

	// One HTTP reverse-proxy session.
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("proxy status: %d", rr.Code)
	}

	// types=http should keep it
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captures=all&includeUnassigned=true&limit=1000&types=http", bytes.NewReader(nil))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("sessions list types=http status: %d", rr.Code)
	}
	var out struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode sessions: %v", err)
	}
	if len(out.Items) != 1 {
		t.Fatalf("items=%d want=1", len(out.Items))
	}
}
