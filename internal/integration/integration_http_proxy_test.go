package integration

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/infrastructure/config"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

// startUpstreamHTTP spins up a small HTTP server that echoes request info
func startUpstreamHTTP(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "sid=supersecret")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"q":  r.URL.Query().Get("q"),
			"ua": r.Header.Get("User-Agent"),
		})
	})
	mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"len": len(b)})
	})
	mux.HandleFunc("/gzip", func(w http.ResponseWriter, r *http.Request) {
		// return large JSON to trigger preview truncation (proxy side doesn't need to decompress)
		w.Header().Set("Content-Type", "application/json")
		big := make([]byte, 10000)
		for i := range big {
			big[i] = 'a'
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"big": string(big)})
	})
	mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
		// echo connection-related headers presence to ensure proxy strips hop-by-hop
		_, _ = w.Write([]byte(r.Header.Get("Connection") + "," + r.Header.Get("Proxy-Connection") + "," + r.Header.Get("Te")))
	})
	srv := httptest.NewServer(mux)
	return srv, srv.URL
}

func startHTTPApp(t *testing.T) (*httptest.Server, *httpapi.Deps) {
	t.Helper()
	logger := obs.NewLogger("error")
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*"}, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	srv := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	return srv, deps
}

func TestHTTPReverseProxy_BasicGetAndPost(t *testing.T) {
	upstream, upstreamURL := startUpstreamHTTP(t)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	client := app.Client()

	// GET via /httpproxy with path join and query pass-through
	u, _ := url.Parse(app.URL + "/httpproxy/get?_target=" + url.QueryEscape(upstreamURL) + "&q=42")
	resp, err := client.Get(u.String())
	if err != nil {
		t.Fatalf("get via reverse: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	var got map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&got)
	if got["q"].(string) != "42" {
		t.Fatalf("query not forwarded: %v", got)
	}

	// POST body
	resp2, err := client.Post(app.URL+"/httpproxy/post?_target="+url.QueryEscape(upstreamURL), "application/json", io.NopCloser(io.LimitReader(io.MultiReader(), 0)))
	if err != nil {
		t.Fatalf("post via reverse: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp2.StatusCode)
	}

	// hop-by-hop headers should be stripped; upstream /hop will echo empties
	u2 := app.URL + "/httpproxy/hop?_target=" + url.QueryEscape(upstreamURL)
	req2, _ := http.NewRequest(http.MethodGet, u2, nil)
	req2.Header.Set("Connection", "keep-alive")
	req2.Header.Set("Proxy-Connection", "keep-alive")
	req2.Header.Set("Te", "trailers")
	r2, _ := client.Do(req2)
	b2, _ := io.ReadAll(r2.Body)
	r2.Body.Close()
	parts := []byte(b2)
	// Expect first two values (Connection, Proxy-Connection) to be empty. TE may be set by stack to 'trailers'.
	if string(parts) != ",," && string(parts) != ",,trailers" {
		t.Fatalf("hop-by-hop not stripped (got %q)", string(b2))
	}
}

func TestUnifiedProxy_DefaultTarget(t *testing.T) {
	t.Parallel()
	upstream, upstreamURL := startUpstreamHTTP(t)
	defer upstream.Close()

	logger := obs.NewLogger("error")
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*", DefaultTarget: upstreamURL}, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	app := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	defer app.Close()

	resp, err := app.Client().Get(app.URL + "/proxy/get?q=ok")
	if err != nil {
		t.Fatalf("unified proxy request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestForwardProxy_HTTP_AbsoluteURI(t *testing.T) {
	t.Parallel()
	upstream, upstreamURL := startUpstreamHTTP(t)
	defer upstream.Close()
	app, _ := startHTTPApp(t)
	defer app.Close()

	// Build a client that targets proxy root and sends absolute-URI
	proxyURL, _ := url.Parse(app.URL)
	// extract host:port
	hostPort := proxyURL.Host

	// manual request with absolute-URI to proxy
	req, _ := http.NewRequest(http.MethodGet, upstreamURL+"/get?q=1", nil)
	// override URL to proxy
	pURL := *proxyURL
	pURL.Path = "/"
	req.URL = &pURL
	req.Host = hostPort
	// Raw absolute-URI in RequestURI — httptest.Client doesn't expose directly. Use net.Dial and write raw HTTP.
	conn, err := net.DialTimeout("tcp", hostPort, 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()
	raw := "GET " + upstreamURL + "/get?q=1 HTTP/1.1\r\nHost: " + hostPort + "\r\n\r\n"
	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatalf("write: %v", err)
	}
	br := bufio.NewReader(conn)
	line, err := br.ReadString('\n')
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if line == "" {
		t.Fatalf("no response from proxy forward handler")
	}
}

func TestForwardProxy_CONNECT_TunnelToHTTP(t *testing.T) {
	t.Parallel()
	upstream, upstreamURL := startUpstreamHTTP(t)
	defer upstream.Close()
	app, _ := startHTTPApp(t)
	defer app.Close()

	// parse upstream host:port
	u, _ := url.Parse(upstreamURL)
	target := u.Host

	// CONNECT to proxy
	proxyURL, _ := url.Parse(app.URL)
	conn, err := net.DialTimeout("tcp", proxyURL.Host, 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()
	req := "CONNECT " + target + " HTTP/1.1\r\nHost: " + target + "\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write connect: %v", err)
	}
	br := bufio.NewReader(conn)
	status, _ := br.ReadString('\n')
	if status == "" || status[:12] != "HTTP/1.1 200" {
		t.Fatalf("connect status: %q", status)
	}

	// now send plain HTTP GET through the tunnel to upstream
	get := "GET /get?q=77 HTTP/1.1\r\nHost: " + target + "\r\n\r\n"
	if _, err := conn.Write([]byte(get)); err != nil {
		t.Fatalf("tunnel write: %v", err)
	}
	line, _ := br.ReadString('\n')
	if line == "" {
		t.Fatalf("no response through tunnel")
	}
}

func TestHTTPReverseProxy_RedactionAndFrames(t *testing.T) {
	upstream, upstreamURL := startUpstreamHTTP(t)
	defer upstream.Close()
	app, deps := startHTTPApp(t)
	defer app.Close()

	// perform request with sensitive headers and ensure redaction in frames
	req, _ := http.NewRequest(http.MethodGet, app.URL+"/httpproxy/get?_target="+url.QueryEscape(upstreamURL)+"&q=ok", nil)
	req.Header.Set("Authorization", "Bearer topsecret")
	req.Header.Set("Cookie", "sid=clientsecret")
	resp, err := app.Client().Do(req)
	if err != nil {
		t.Fatalf("reverse get: %v", err)
	}
	resp.Body.Close()

	// large body path to exercise preview truncation path
	_, _ = app.Client().Get(app.URL + "/httpproxy/gzip?_target=" + url.QueryEscape(upstreamURL))

	// list sessions filtered by upstream URL substring (q filter is contains)
	r, _ := app.Client().Get(app.URL + "/api/sessions?limit=1000&q=" + url.QueryEscape(upstreamURL))
	defer r.Body.Close()
	var list struct {
		Items []struct {
			ID     string
			Target string
		} `json:"items"`
	}
	_ = json.NewDecoder(r.Body).Decode(&list)
	if len(list.Items) == 0 {
		t.Fatalf("no sessions")
	}
	// choose sessions explicitly by path
	var sidGet, sidGzip string
	for _, it := range list.Items {
		if strings.Contains(it.Target, "/get") {
			sidGet = it.ID
		}
		if strings.Contains(it.Target, "/gzip") {
			sidGzip = it.ID
		}
	}
	if sidGet == "" || sidGzip == "" {
		t.Fatalf("expected sessions for /get and /gzip; got %+v", list.Items)
	}

	// Check redaction on /get session (Authorization and Set-Cookie)
	rf1, _ := app.Client().Get(app.URL + "/api/sessions/" + sidGet + "/frames?limit=100")
	defer rf1.Body.Close()
	var frames1 struct {
		Items []struct{ Preview string } `json:"items"`
	}
	_ = json.NewDecoder(rf1.Body).Decode(&frames1)
	sawReqRedacted := false
	sawRespRedacted := false
	for _, f := range frames1.Items {
		var js map[string]any
		_ = json.Unmarshal([]byte(f.Preview), &js)
		if js["type"] == "http_request" {
			if hdr, ok := js["headers"].(map[string]any); ok {
				if hdr["Authorization"] == "***" || hdr["authorization"] == "***" {
					sawReqRedacted = true
				}
			}
		}
		if js["type"] == "http_response" {
			if hdr, ok := js["headers"].(map[string]any); ok {
				if hdr["Set-Cookie"] == "***" || hdr["set-cookie"] == "***" {
					sawRespRedacted = true
				}
			}
		}
	}
	if !sawReqRedacted || !sawRespRedacted {
		t.Fatalf("expected header redaction: req=%v resp=%v", sawReqRedacted, sawRespRedacted)
	}

	// Check big body preview on /gzip session
	rf2, _ := app.Client().Get(app.URL + "/api/sessions/" + sidGzip + "/frames?limit=100")
	defer rf2.Body.Close()
	var frames2 struct {
		Items []struct{ Preview string } `json:"items"`
	}
	_ = json.NewDecoder(rf2.Body).Decode(&frames2)
	sawBig := false
	for _, f := range frames2.Items {
		var js map[string]any
		_ = json.Unmarshal([]byte(f.Preview), &js)
		if js["type"] == "http_response" {
			if body, ok := js["body"].(string); ok && len(body) >= 100 {
				sawBig = true
			}
		}
	}
	if !sawBig {
		t.Fatalf("expected big response body preview in /gzip session")
	}
	_ = deps
}
