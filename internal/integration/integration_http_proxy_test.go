package integration

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
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

func TestForwardProxy_HTTP_HopByHopHeadersStripped(t *testing.T) {
	t.Parallel()
	upstream, upstreamURL := startUpstreamHTTP(t)
	defer upstream.Close()
	app, _ := startHTTPApp(t)
	defer app.Close()

	proxyURL, _ := url.Parse(app.URL)
	conn, err := net.DialTimeout("tcp", proxyURL.Host, 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()

	// upstream /hop echoes Connection,Proxy-Connection,Te
	raw := "GET " + upstreamURL + "/hop HTTP/1.1\r\n" +
		"Host: " + proxyURL.Host + "\r\n" +
		"Connection: keep-alive\r\n" +
		"Proxy-Connection: keep-alive\r\n" +
		"Te: trailers\r\n\r\n"
	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatalf("write: %v", err)
	}
	br := bufio.NewReader(conn)
	// read status line
	if _, err := br.ReadString('\n'); err != nil {
		t.Fatalf("read status: %v", err)
	}
	// skip headers
	for {
		line, _ := br.ReadString('\n')
		if line == "\r\n" || line == "\n" || line == "" {
			break
		}
	}
	// Read small body with deadline to avoid hang on keep-alive
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 64)
	n, _ := br.Read(buf)
	s := string(buf[:n])
	if s != ",," && s != ",,trailers" {
		t.Fatalf("hop-by-hop not stripped (got %q)", s)
	}
}

func TestForwardProxy_HTTP_DNSErrorReturns502(t *testing.T) {
	t.Parallel()
	app, _ := startHTTPApp(t)
	defer app.Close()
	proxyURL, _ := url.Parse(app.URL)
	conn, err := net.DialTimeout("tcp", proxyURL.Host, 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()

	// non-existent domain
	raw := "GET http://nonexistent.invalid/ HTTP/1.1\r\nHost: " + proxyURL.Host + "\r\n\r\n"
	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatalf("write: %v", err)
	}
	br := bufio.NewReader(conn)
	status, _ := br.ReadString('\n')
	if status == "" || status[:12] != "HTTP/1.1 502" {
		t.Fatalf("expected 502, got %q", status)
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
	r, err := app.Client().Get(app.URL + "/api/sessions?limit=1000&q=" + url.QueryEscape(upstreamURL))
	if err != nil {
		t.Fatal(err)
	}
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
	rf1, err := app.Client().Get(app.URL + "/api/sessions/" + sidGet + "/frames?limit=100")
	if err != nil {
		t.Fatal(err)
	}
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
	rf2, err := app.Client().Get(app.URL + "/api/sessions/" + sidGzip + "/frames?limit=100")
	if err != nil {
		t.Fatal(err)
	}
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

func TestHTTPReverseProxy_Cookies_Isolation(t *testing.T) {
	t.Parallel()
	// Upstream A
	muxA := http.NewServeMux()
	muxA.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		// echo received Cookie header back
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Set-Cookie", "sidA=aaa; Path=/; SameSite=None")
		_ = json.NewEncoder(w).Encode(map[string]any{"cookie": r.Header.Get("Cookie")})
	})
	srvA := httptest.NewServer(muxA)
	defer srvA.Close()

	// Upstream B
	muxB := http.NewServeMux()
	muxB.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Set-Cookie", "sidB=bbb; Path=/; SameSite=None")
		_ = json.NewEncoder(w).Encode(map[string]any{"cookie": r.Header.Get("Cookie")})
	})
	srvB := httptest.NewServer(muxB)
	defer srvB.Close()

	// App
	app, deps := startHTTPApp(t)
	defer app.Close()
	// Enable isolate by default
	deps.Cfg.Cookies.Mode = "isolate"
	deps.Cfg.Cookies.PathStrategy = "prefix"
	deps.Cfg.Cookies.DomainStrategy = "hostOnly"

	// Client with cookie jar to persist browser cookies on proxy domain
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Request A (sets sidA on proxy domain)
	uA := app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(srvA.URL)
	rA, err := client.Get(uA)
	if err != nil {
		t.Fatalf("request A: %v", err)
	}
	_ = rA.Body.Close()

	// Request B (sets sidB)
	uB := app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(srvB.URL)
	rB, err := client.Get(uB)
	if err != nil {
		t.Fatalf("request B: %v", err)
	}
	_ = rB.Body.Close()

	// Back to A: only sidA should be forwarded upstream (after unwrapping), sidB must be filtered out
	rA2, err := client.Get(uA)
	if err != nil {
		t.Fatalf("request A2: %v", err)
	}
	defer rA2.Body.Close()
	var got map[string]any
	_ = json.NewDecoder(rA2.Body).Decode(&got)
	cookieHdr, _ := got["cookie"].(string)
	if cookieHdr == "" || !strings.Contains(cookieHdr, "sidA=") || strings.Contains(cookieHdr, "sidB=") {
		t.Fatalf("unexpected upstream cookies (want only sidA): %q", cookieHdr)
	}
}

func TestHTTPReverseProxy_StealthHeaders_OnOff(t *testing.T) {
	t.Parallel()
	// Upstream that echoes proxy-related headers
	mux := http.NewServeMux()
	mux.HandleFunc("/hdr", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"xff":  r.Header.Get("X-Forwarded-For"),
			"xfp":  r.Header.Get("X-Forwarded-Proto"),
			"via":  r.Header.Get("Via"),
			"host": r.Host,
		})
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, deps := startHTTPApp(t)
	defer app.Close()
	deps.Cfg.StealthHeaders = true

	// Stealth ON (default): headers must be empty
	uOn := app.URL + "/httpproxy/hdr?_target=" + url.QueryEscape(upstream.URL)
	rOn, err := app.Client().Get(uOn)
	if err != nil {
		t.Fatalf("stealth on: %v", err)
	}
	defer rOn.Body.Close()
	var gotOn map[string]string
	_ = json.NewDecoder(rOn.Body).Decode(&gotOn)
	if gotOn["xfp"] != "" || gotOn["via"] != "" {
		t.Fatalf("stealth on should suppress Via and X-Forwarded-Proto: %+v", gotOn)
	}

	// Stealth OFF via query override
	uOff := app.URL + "/httpproxy/hdr?_target=" + url.QueryEscape(upstream.URL) + "&_stealth=0"
	rOff, err := app.Client().Get(uOff)
	if err != nil {
		t.Fatalf("stealth off: %v", err)
	}
	defer rOff.Body.Close()
	var gotOff map[string]string
	_ = json.NewDecoder(rOff.Body).Decode(&gotOff)
	if gotOff["xfp"] == "" || gotOff["via"] == "" { // Via и X-Forwarded-Proto должны быть выставлены
		t.Fatalf("stealth off should set Via and X-Forwarded-Proto: %+v", gotOff)
	}
}

func TestHTTPReverseProxy_Cookies_AutoMode_NoIsolation(t *testing.T) {
	t.Parallel()
	// Upstream A echoes Cookie header
	muxA := http.NewServeMux()
	muxA.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "sidA=aaa; Path=/; SameSite=None")
		_ = json.NewEncoder(w).Encode(map[string]any{"cookie": r.Header.Get("Cookie")})
	})
	srvA := httptest.NewServer(muxA)
	defer srvA.Close()

	// Upstream B sets another cookie
	muxB := http.NewServeMux()
	muxB.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "sidB=bbb; Path=/; SameSite=None")
		_ = json.NewEncoder(w).Encode(map[string]any{"cookie": r.Header.Get("Cookie")})
	})
	srvB := httptest.NewServer(muxB)
	defer srvB.Close()

	app, deps := startHTTPApp(t)
	defer app.Close()
	deps.Cfg.Cookies.Mode = "auto"
	deps.Cfg.Cookies.PathStrategy = "prefix"

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// set both cookies under proxy domain
	_, _ = client.Get(app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(srvA.URL) + "&_cookie_mode=auto")
	_, _ = client.Get(app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(srvB.URL) + "&_cookie_mode=auto")

	// request to A should carry both sidA and sidB since auto не изолирует имена
	rA, err := client.Get(app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(srvA.URL) + "&_cookie_mode=auto")
	if err != nil {
		t.Fatalf("auto A: %v", err)
	}
	defer rA.Body.Close()
	var got map[string]any
	_ = json.NewDecoder(rA.Body).Decode(&got)
	cookieHdr, _ := got["cookie"].(string)
	if !(strings.Contains(cookieHdr, "sidA=") && strings.Contains(cookieHdr, "sidB=")) {
		t.Fatalf("auto mode should forward both cookies: %q", cookieHdr)
	}
}

func TestHTTPReverseProxy_Cookies_HeaderDomainForLocalhostIP_Omitted(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "x=1; Path=/; SameSite=Lax")
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, deps := startHTTPApp(t)
	defer app.Close()
	deps.Cfg.Cookies.DomainStrategy = "proxyHost"

	resp, err := app.Client().Get(app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(upstream.URL))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	sc := resp.Header.Get("Set-Cookie")
	if strings.Contains(strings.ToLower(sc), "domain=") {
		t.Fatalf("Domain attribute must be omitted for localhost/IP proxy host: %s", sc)
	}
}

func TestHTTPReverseProxy_Cookies_ModeOff_PassThrough(t *testing.T) {
	t.Parallel()
	// Upstream sets Domain and specific Path
	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "sid=abc; Domain=api.example.com; Path=/api; SameSite=None")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	// cookie_mode=off — прокси не должен переписывать Set-Cookie
	resp, err := app.Client().Get(app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(upstream.URL) + "&_cookie_mode=off")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	sc := resp.Header.Get("Set-Cookie")
	if !strings.Contains(sc, "Domain=api.example.com") || !strings.Contains(sc, "; Path=/api") {
		t.Fatalf("expected passthrough Set-Cookie with Domain and Path intact: %s", sc)
	}
	if strings.Contains(sc, "/httpproxy") {
		t.Fatalf("path must not be rewritten under mode=off: %s", sc)
	}
}
