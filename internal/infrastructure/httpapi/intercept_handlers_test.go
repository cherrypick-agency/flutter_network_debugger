package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
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

func setupTestRouter(t *testing.T) *Deps {
	t.Helper()
	logger := zerolog.New(io.Discard)
	m := obs.NewMetrics()
	cfg := config.Config{CORSAllowOrigin: "*", InterceptEnabled: true, InterceptRequests: true, InterceptResponses: true, InterceptTimeoutMs: 200, InterceptQueueMax: 10, InterceptBodyMaxBytes: 1024}
	d := &Deps{Cfg: cfg, Logger: &logger, Metrics: m, Monitor: NewMonitorHub(), Live: NewLiveSessions()}
	// minimal in-memory repos for sessions/frames/events
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)
	// init mux (also init Interceptor)
	_ = NewRouterWithoutForwardProxy(d)
	return d
}

func doReq(t *testing.T, h http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var rbody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rbody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rbody)
	if token != "" {
		req.Header.Set("X-Admin-Token", token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestInterceptRulesAndConfigCRUD(t *testing.T) {
	d := setupTestRouter(t)
	// protect with token (since RemoteAddr may not be loopback in tests)
	d.Cfg.AdminToken = "t"
	h := NewRouterWithoutForwardProxy(d)
	// POST rules
	rules := []InterceptRule{{Enabled: true, Priority: 5, Action: "both", When: InterceptWhen{Method: []string{"GET"}}}}
	resp := doReq(t, h, http.MethodPost, "/_api/v1/intercept/rules", rules, "t")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("rules POST status=%d", resp.Code)
	}
	// GET rules
	resp = doReq(t, h, http.MethodGet, "/_api/v1/intercept/rules", nil, "t")
	if resp.Code != http.StatusOK {
		t.Fatalf("rules GET status=%d", resp.Code)
	}
	var got []InterceptRule
	_ = json.Unmarshal(resp.Body.Bytes(), &got)
	if len(got) != 1 || !got[0].Enabled {
		t.Fatalf("unexpected rules: %#v", got)
	}

	// GET config
	resp = doReq(t, h, http.MethodGet, "/_api/v1/intercept/config", nil, "t")
	if resp.Code != http.StatusOK {
		t.Fatalf("config GET status=%d", resp.Code)
	}
	// POST config
	payload := map[string]any{"enabled": true, "requests": true, "responses": true, "timeoutMs": 100}
	resp = doReq(t, h, http.MethodPost, "/_api/v1/intercept/config", payload, "t")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("config POST status=%d", resp.Code)
	}
}

func TestInterceptPendingContinueFlow(t *testing.T) {
	d := setupTestRouter(t)
	d.Cfg.AdminToken = "t"
	h := NewRouterWithoutForwardProxy(d)
	// set rule to intercept GET requests
	doReq(t, h, http.MethodPost, "/_api/v1/intercept/rules", []InterceptRule{{Enabled: true, Priority: 1, Action: "request", When: InterceptWhen{Method: []string{"GET"}}}}, "t")

	// spawn blocking intercept
	req, _ := http.NewRequest(http.MethodGet, "http://example/", nil)
	done := make(chan *HTTPRequestDecision, 1)
	go func() {
		dec, _ := d.Interceptor.InterceptRequest(context.Background(), "sessX", req, "", nil, "")
		done <- dec
	}()

	// wait for pending and then continue
	var id string
	for i := 0; i < 50; i++ {
		w := doReq(t, h, http.MethodGet, "/_api/v1/intercept/pending", nil, "t")
		if w.Code != http.StatusOK {
			t.Fatalf("pending status=%d", w.Code)
		}
		var items []InterceptItem
		_ = json.Unmarshal(w.Body.Bytes(), &items)
		if len(items) > 0 {
			id = items[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if id == "" {
		t.Fatal("no pending item")
	}

	// continue
	resp := doReq(t, h, http.MethodPost, "/_api/v1/intercept/items/"+id+"/continue", map[string]any{"action": "continue", "method": "PUT"}, "t")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("continue status=%d", resp.Code)
	}
	select {
	case dec := <-done:
		if dec == nil || dec.Method != "PUT" {
			t.Fatalf("unexpected decision: %#v", dec)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting decision")
	}
}
