package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"network-debugger/internal/adapters/storage/memory"
	interceptdomain "network-debugger/internal/features/intercept/domain"
	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

// E2E: reverse proxy with intercept on response, modify body
func TestE2E_ReverseProxy_ResponseInterceptAndModify(t *testing.T) {
	// upstream backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello"))
	}))
	defer backend.Close()

	// app router
	logger := zerolog.New(io.Discard)
	m := obs.NewMetrics()
	cfg := config.Config{CORSAllowOrigin: "*", DefaultTarget: backend.URL, InterceptEnabled: true, InterceptResponses: true, InterceptTimeoutMs: 1000, InterceptBodyMaxBytes: 1024}
	d := &Deps{Cfg: cfg, Logger: &logger, Metrics: m, Monitor: NewMonitorHub(), Live: NewLiveSessions()}
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)
	srv := httptest.NewServer(NewRouterWithoutForwardProxy(d))
	defer srv.Close()

	mgr := d.InterceptSvc.Manager()

	// rule: intercept text/plain responses
	mgr.UpdateRules([]interceptdomain.InterceptRule{{Enabled: true, Priority: 1, Action: "response", When: interceptdomain.InterceptWhen{ContentType: &interceptdomain.RuleStringMatch{Prefix: "text/plain"}}}})

	// start client request (will block on intercept)
	clientDone := make(chan string, 1)
	go func() {
		resp, err := http.Get(srv.URL + "/httpproxy")
		if err != nil {
			t.Errorf("client err: %v", err)
			clientDone <- ""
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		clientDone <- string(b)
	}()

	// wait pending then continue with modified body
	var id string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		items := mgr.ListPending(10)
		if len(items) > 0 {
			id = items[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if id == "" {
		t.Fatal("no pending response item")
	}
	// continue: change body
	mgr.ContinueResponse(id, &interceptdomain.HTTPResponseDecision{Action: "continue", Body: []byte("patched")})
	// assert client got patched body
	select {
	case body := <-clientDone:
		if body != "patched" {
			t.Fatalf("unexpected body: %q", body)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("client timeout")
	}
}

// E2E: forward proxy HTTP absolute-URI with request intercept to drop
func TestE2E_ForwardProxy_RequestInterceptDrop(t *testing.T) {
	// upstream backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer backend.Close()

	// app router
	logger := zerolog.New(io.Discard)
	m := obs.NewMetrics()
	cfg := config.Config{CORSAllowOrigin: "*", InterceptEnabled: true, InterceptRequests: true, InterceptTimeoutMs: 1000}
	d := &Deps{Cfg: cfg, Logger: &logger, Metrics: m, Monitor: NewMonitorHub(), Live: NewLiveSessions()}
	store := memory.NewStore(256, 1024, 5*time.Minute)
	d.Svc = usecase.NewSessionService(store, store, store)
	// For test, start server with forward-proxy wrapper on top of base router
	srv := httptest.NewServer(withForwardProxy(d, NewRouterWithoutForwardProxy(d)))
	defer srv.Close()

	mgr := d.InterceptSvc.Manager()

	// rule: intercept any POST and drop
	mgr.UpdateRules([]interceptdomain.InterceptRule{{Enabled: true, Priority: 1, Action: "request", When: interceptdomain.InterceptWhen{Method: []string{"POST"}}}})

	// http client via proxy (srv)
	proxyURL := srv.URL
	tr := &http.Transport{Proxy: func(r *http.Request) (*url.URL, error) { return url.Parse(proxyURL) }}
	cli := &http.Client{Transport: tr, Timeout: 3 * time.Second}

	// fire POST to backend (absolute URL) — should be dropped by interceptor when we continue with action=drop
	done := make(chan int, 1)
	go func() {
		req, _ := http.NewRequest(http.MethodPost, backend.URL+"/test", nil)
		resp, err := cli.Do(req)
		if err != nil {
			done <- -1
			return
		}
		defer resp.Body.Close()
		done <- resp.StatusCode
	}()

	// wait for pending request intercept
	var id string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		items := mgr.ListPending(10)
		if len(items) > 0 {
			id = items[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if id == "" {
		t.Fatal("no pending request item")
	}
	mgr.ContinueRequest(id, &interceptdomain.HTTPRequestDecision{Action: "drop"})

	// ensure client received non-200 because proxy terminated
	select {
	case code := <-done:
		if code == 200 {
			t.Fatalf("expected non-200 due to drop, got %d", code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("client timeout forward proxy")
	}
}
