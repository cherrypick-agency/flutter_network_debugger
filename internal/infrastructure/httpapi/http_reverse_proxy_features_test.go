package httpapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
)

// upstream that returns gzip+json and Set-Cookie to exercise response preview/cookies
func makeUpstreamGzipJSON() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// reflect query
		_ = r.URL.Query()
		// gzip json body with sensitive fields
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Set-Cookie", "id=1; Path=/; SameSite=None")
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_ = json.NewEncoder(zw).Encode(map[string]any{"authorization": "Bearer abc", "access_token": "x", "ok": true})
		_ = zw.Close()
		w.WriteHeader(200)
		_, _ = w.Write(buf.Bytes())
	}))
}

func TestHTTPProxy_ResponsePreview_Cookies_QueryAndStealth(t *testing.T) {
	upstream := makeUpstreamGzipJSON()
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 256, ExposeSensitiveHeaders: false, PreviewDecompress: true, ResponseDelayMs: 1, Cookies: cfgpkg.CookiesConfig{Mode: "auto", DomainStrategy: "hostOnly", PathStrategy: "prefix"}}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)

	// include suffix path and _stealth toggle to exercise branches
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy/x?_target="+upstream.URL+"&_stealth=0&_cookie_mode=isolate&a=1", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestHTTPProxy_ResponsePreview_NoDecompressAndExpose(t *testing.T) {
	upstream := makeUpstreamGzipJSON()
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 64, ExposeSensitiveHeaders: true, PreviewDecompress: false}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("status: %d", rr.Code)
	}
}

func TestHTTPProxy_V1RouteAlias(t *testing.T) {
	upstream := makeUpstreamGzipJSON()
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 64, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/httpproxy?_target="+upstream.URL, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 httpproxy alias status: %d", rr.Code)
	}
}

func TestUnifiedProxy_HTTPPath(t *testing.T) {
	upstream := makeUpstreamGzipJSON()
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{DefaultTarget: upstream.URL, PreviewMaxBytes: 128, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithDeps(d)
	rr := httptest.NewRecorder()
	// unified /proxy should fall back to HTTP reverse handler
	req := httptest.NewRequest(http.MethodGet, "/proxy/abc", nil)
	// add explicit target via query as well
	q := req.URL.Query()
	q.Set("_target", upstream.URL)
	req.URL.RawQuery = q.Encode()
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("unified proxy status: %d", rr.Code)
	}
}
