package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	pruntime "network-debugger/internal/infrastructure/proxyruntime"
	"network-debugger/internal/usecase"
)

// End-to-end: enable forward-proxy on dynamic port and make request through it
// to local upstream, expecting 200 and body.
func TestForwardProxy_E2E_HTTP(t *testing.T) {
	// Upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer upstream.Close()

	// Deps (minimally sufficient)
	logger := obs.NewLogger("warn")
	deps := &Deps{Cfg: defaultTestConfig(), Logger: logger, Metrics: obs.NewMetrics(), Svc: newTestSessionSvc(), Monitor: NewMonitorHub()}

	// Runtime manager
	rt := pruntime.New(logger)
	deps.ProxyRt = rt

	// Start forward proxy on dynamic port
	if err := rt.Apply(contextWithNoCancel(), pruntime.ApplyConfig{ForwardEnabled: true, ForwardAddr: "127.0.0.1:0"}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deps.handleForwardOrNotFound(w, r)
	})); err != nil {
		t.Fatalf("apply: %v", err)
	}
	proxyAddr := rt.ForwardAddr()
	if proxyAddr == "" {
		t.Fatalf("no proxy addr")
	}

	// Client with proxy
	proxyURL, _ := url.Parse("http://" + proxyAddr)
	cli := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	// Real request to upstream through proxy (absolute URL)
	req, _ := http.NewRequest(http.MethodGet, upstream.URL, nil)
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "ok" {
		t.Fatalf("body=%q", string(b))
	}
}

// --- helpers ---
func defaultTestConfig() config.Config {
	// base values; UI port is not used
	return config.Config{Addr: ":0", PreviewMaxBytes: 1024, ExposeSensitiveHeaders: true, PreviewDecompress: true}
}

func newTestSessionSvc() *usecase.SessionService {
	store := memory.NewStore(100, 1000, time.Minute)
	return usecase.NewSessionService(store, store, store)
}
