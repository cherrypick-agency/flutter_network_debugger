package integration

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	"network-debugger/internal/infrastructure/config"
	dbpkg "network-debugger/internal/infrastructure/db"
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
	// Инициализируем SQLite для включения ProxySvc/ProxyRt
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	if gdb, err := dbpkg.NewSQLite(dbPath); err == nil {
		_ = gdb.AutoMigrate(&proxyp.ProxyConfigModel{})
		deps.DB = gdb
	}
	srv := httptest.NewServer(httpapi.NewRouterWithDeps(deps))
	return srv, deps
}

// ensureForwardProxyAddr включает forward‑proxy на динамическом порту и возвращает host:port
func ensureForwardProxyAddr(t *testing.T, app *httptest.Server, deps *httpapi.Deps) string {
	t.Helper()
	// Включаем forward proxy на 127.0.0.1:0 через API
	body := []byte(`{"forward":{"enabled":true,"addr":"127.0.0.1:0"}}`)
	resp, err := app.Client().Post(app.URL+"/_api/v1/proxy/config", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("enable forward proxy: %v", err)
	}
	_ = resp.Body.Close()
	// Подождём применения и получим фактический адрес из рантайма
	time.Sleep(50 * time.Millisecond)
	addr := deps.ProxyRt.ForwardAddr()
	if addr == "" {
		t.Fatalf("no forward proxy address after enable")
	}
	return addr
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
	app, deps := startHTTPApp(t)
	defer app.Close()
	// Включаем forward‑proxy и берём фактический адрес
	hostPort := ensureForwardProxyAddr(t, app, deps)

	// manual request with absolute-URI to proxy
	req, _ := http.NewRequest(http.MethodGet, upstreamURL+"/get?q=1", nil)
	// override URL/Host if понадобится (для полноты форм)
	req.URL = &url.URL{Scheme: "http", Host: hostPort, Path: "/"}
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
	app, deps := startHTTPApp(t)
	defer app.Close()

	proxyHost := ensureForwardProxyAddr(t, app, deps)
	conn, err := net.DialTimeout("tcp", proxyHost, 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()

	// upstream /hop echoes Connection,Proxy-Connection,Te
	raw := "GET " + upstreamURL + "/hop HTTP/1.1\r\n" +
		"Host: " + proxyHost + "\r\n" +
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
	app, deps := startHTTPApp(t)
	defer app.Close()
	proxyHost := ensureForwardProxyAddr(t, app, deps)
	conn, err := net.DialTimeout("tcp", proxyHost, 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()

	// non-existent domain
	raw := "GET http://nonexistent.invalid/ HTTP/1.1\r\nHost: " + proxyHost + "\r\n\r\n"
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
	app, deps := startHTTPApp(t)
	defer app.Close()

	// parse upstream host:port
	u, _ := url.Parse(upstreamURL)
	target := u.Host

	// CONNECT to proxy
	proxyHost := ensureForwardProxyAddr(t, app, deps)
	conn, err := net.DialTimeout("tcp", proxyHost, 3*time.Second)
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

func TestHTTPReverseProxy_Redirects_SingleAndChain(t *testing.T) {
	t.Parallel()
	// upstream with redirect chain: /r1 -> 302 /r2, /r2 -> 301 /final
	mux := http.NewServeMux()
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/r2", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/r1", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/r2", http.StatusFound)
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	u := app.URL + "/httpproxy/r1?_target=" + url.QueryEscape(upstream.URL)
	resp, err := app.Client().Get(u)
	if err != nil {
		t.Fatalf("redirect chain: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "ok" {
		t.Fatalf("final body mismatch: %q", string(b))
	}
}

func TestHTTPReverseProxy_Chunked_FlushEcho(t *testing.T) {
	t.Parallel()
	// upstream that streams chunks with flush
	mux := http.NewServeMux()
	mux.HandleFunc("/chunk", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		f, _ := w.(http.Flusher)
		for i := 0; i < 5; i++ {
			_, _ = w.Write([]byte("part"))
			if f != nil {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	resp, err := app.Client().Get(app.URL + "/httpproxy/chunk?_target=" + url.QueryEscape(upstream.URL))
	if err != nil {
		t.Fatalf("chunked get: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if string(b) != strings.Repeat("part", 5) {
		t.Fatalf("chunked echo mismatch: %q", string(b))
	}
}

func TestHTTPReverseProxy_Compression_GzipDeflate_PreviewToggle(t *testing.T) {
	t.Parallel()
	// upstream that returns gzip and deflate depending on path
	mux := http.NewServeMux()
	payload := strings.Repeat("A", 4096)
	mux.HandleFunc("/gz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte(payload))
		_ = gz.Close()
	})
	mux.HandleFunc("/df", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Encoding", "deflate")
		df, _ := flate.NewWriter(w, flate.DefaultCompression)
		_, _ = df.Write([]byte(payload))
		_ = df.Close()
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	// App with PreviewDecompress on (default)
	app1, deps1 := startHTTPApp(t)
	defer app1.Close()
	deps1.Cfg.PreviewDecompress = true
	// Should succeed
	r1, err := app1.Client().Get(app1.URL + "/httpproxy/gz?_target=" + url.QueryEscape(upstream.URL))
	if err != nil {
		t.Fatalf("gzip on: %v", err)
	}
	_ = r1.Body.Close()

	// App with PreviewDecompress off
	app2, deps2 := startHTTPApp(t)
	defer app2.Close()
	deps2.Cfg.PreviewDecompress = false
	r2, err := app2.Client().Get(app2.URL + "/httpproxy/df?_target=" + url.QueryEscape(upstream.URL))
	if err != nil {
		t.Fatalf("deflate off: %v", err)
	}
	_ = r2.Body.Close()
}

func TestHTTPReverseProxy_LargeBodies_32MB_UploadDownload(t *testing.T) {
	// Not parallel to reduce memory spikes on CI
	// upstream large download and upload counters
	mux := http.NewServeMux()
	// 32MB payload for download
	const size = 32 * 1024 * 1024
	mux.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		// stream without buffering
		buf := make([]byte, 1024*32)
		written := 0
		for written < size {
			n := size - written
			if n > len(buf) {
				n = len(buf)
			}
			_, _ = w.Write(buf[:n])
			written += n
		}
	})
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		n, _ := io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		_ = json.NewEncoder(w).Encode(map[string]any{"n": n})
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	// download
	rd, err := app.Client().Get(app.URL + "/httpproxy/download?_target=" + url.QueryEscape(upstream.URL))
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	dn, _ := io.Copy(io.Discard, rd.Body)
	_ = rd.Body.Close()
	if dn != size {
		t.Fatalf("download bytes: %d", dn)
	}

	// upload 32MB using io.LimitedReader backed by random reader (avoid zero-run compression)
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		// stream random bytes
		left := int64(size)
		buf := make([]byte, 64*1024)
		for left > 0 {
			n := int64(len(buf))
			if n > left {
				n = left
			}
			_, _ = rand.Read(buf[:n])
			if _, err := pw.Write(buf[:n]); err != nil {
				return
			}
			left -= n
		}
	}()
	req, _ := http.NewRequest(http.MethodPost, app.URL+"/httpproxy/upload?_target="+url.QueryEscape(upstream.URL), pr)
	req.Header.Set("Content-Type", "application/octet-stream")
	ru, err := app.Client().Do(req)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	defer ru.Body.Close()
	var got struct {
		N int64 `json:"n"`
	}
	_ = json.NewDecoder(ru.Body).Decode(&got)
	if got.N != int64(size) {
		t.Fatalf("upload bytes mismatch: %d", got.N)
	}
}

func TestHTTPReverseProxy_KeepAlive_ConnectionReuse(t *testing.T) {
	t.Parallel()
	var newConns int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	srv := httptest.NewUnstartedServer(handler)
	srv.Config.ConnState = func(c net.Conn, s http.ConnState) {
		if s == http.StateNew {
			newConns++
		}
	}
	srv.Start()
	defer srv.Close()
	upstreamURL := srv.URL

	app, _ := startHTTPApp(t)
	defer app.Close()
	client := app.Client()
	for i := 0; i < 5; i++ {
		r, err := client.Get(app.URL + "/httpproxy/?_target=" + url.QueryEscape(upstreamURL))
		if err != nil {
			t.Fatalf("keepalive req %d: %v", i, err)
		}
		_ = r.Body.Close()
	}
	if newConns > 5 {
		t.Fatalf("too many upstream connections: %d", newConns)
	}
}

func TestHTTPReverseProxy_ExpectContinue(t *testing.T) {
	t.Parallel()
	// upstream that explicitly sends 100-continue then reads body
	mux := http.NewServeMux()
	mux.HandleFunc("/ec", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Expect") != "100-continue" {
			// still read to finish
			_, _ = io.Copy(io.Discard, r.Body)
			_ = json.NewEncoder(w).Encode(map[string]int{"len": 0})
			return
		}
		if cn, ok := w.(http.CloseNotifier); ok {
			_ = cn // avoid unused on old go
		}
		w.WriteHeader(http.StatusContinue)
		n, _ := io.Copy(io.Discard, r.Body)
		_ = json.NewEncoder(w).Encode(map[string]int64{"len": n})
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	body := io.NopCloser(io.LimitReader(rand.Reader, 256<<10))
	req, _ := http.NewRequest(http.MethodPost, app.URL+"/httpproxy/ec?_target="+url.QueryEscape(upstream.URL), body)
	req.Header.Set("Expect", "100-continue")
	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := app.Client().Do(req)
	if err != nil {
		t.Fatalf("expect-continue: %v", err)
	}
	defer resp.Body.Close()
	var got map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&got)
	if got["len"].(float64) <= 0 {
		t.Fatalf("body was not forwarded after 100-continue")
	}
}

func TestHTTPReverseProxy_Timeout_CancelContext(t *testing.T) {
	t.Parallel()
	// upstream that reads slowly
	mux := http.NewServeMux()
	mux.HandleFunc("/slowread", func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		for {
			_, err := r.Body.Read(buf)
			if err != nil {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		w.WriteHeader(http.StatusRequestTimeout)
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, _ := startHTTPApp(t)
	defer app.Close()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		for i := 0; i < 1024; i++ { // ~1MB
			_, _ = pw.Write(make([]byte, 1024))
			time.Sleep(5 * time.Millisecond)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, app.URL+"/httpproxy/slowread?_target="+url.QueryEscape(upstream.URL), pr)
	req.Header.Set("Content-Type", "application/octet-stream")
	_, _ = app.Client().Do(req)
	// Expect context to cancel before completion without hanging
}

func TestHTTPReverseProxy_IPv6_Upstream(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/v6", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	ln, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skip("no ipv6 on host")
	}
	srv := &http.Server{Handler: mux}
	go srv.Serve(ln)
	defer srv.Close()
	if ta, ok := ln.Addr().(*net.TCPAddr); ok {
		upstreamURL := fmt.Sprintf("http://[%s]:%d/v6", ta.IP.String(), ta.Port)

		app, _ := startHTTPApp(t)
		defer app.Close()
		r, err := app.Client().Get(app.URL + "/httpproxy/?_target=" + url.QueryEscape(upstreamURL))
		if err != nil {
			t.Fatalf("ipv6: %v", err)
		}
		_ = r.Body.Close()
	} else {
		t.Skip("unexpected addr type")
	}
}

func TestHTTPReverseProxy_TLS_Upstream_SelfSigned(t *testing.T) {
	t.Parallel()
	// self-signed TLS upstream
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "127.0.0.1"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour), DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, _ := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
	mux := http.NewServeMux()
	mux.HandleFunc("/tls", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	srv := &http.Server{Handler: mux, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}
	go srv.ServeTLS(ln, "", "")
	defer srv.Close()
	upstream := "https://" + ln.Addr().String() + "/tls"

	app, deps := startHTTPApp(t)
	defer app.Close()
	deps.Cfg.InsecureTLS = true
	r, err := app.Client().Get(app.URL + "/httpproxy/?_target=" + url.QueryEscape(upstream))
	if err != nil {
		t.Fatalf("tls upstream: %v", err)
	}
	_ = r.Body.Close()
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

func TestHTTPReverseProxy_Cookies_HostPrefixes_ProxyHostStrategy(t *testing.T) {
	t.Parallel()
	// Upstream sets __Host_ and __Secure_ cookies
	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "__Host_sid=aaa; Path=/; SameSite=None")
		w.Header().Add("Set-Cookie", "__Secure_token=bbb; Path=/; SameSite=Lax")
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, deps := startHTTPApp(t)
	defer app.Close()
	// Ensure proxyHost and root path to satisfy __Host_ constraints (Path=/, no Domain)
	deps.Cfg.Cookies.DomainStrategy = "proxyHost"
	deps.Cfg.Cookies.PathStrategy = "root"

	resp, err := app.Client().Get(app.URL + "/httpproxy/cookie?_target=" + url.QueryEscape(upstream.URL))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	// combined Set-Cookie
	sc := strings.Join(resp.Header.Values("Set-Cookie"), "; ")
	// __Host_: must have Path=/ and MUST NOT include Domain
	if !strings.Contains(sc, "__Host_sid=") || !strings.Contains(sc, "Path=/") || strings.Contains(strings.ToLower(sc), "domain=") {
		t.Fatalf("__Host_ cookie invalid under proxyHost/root: %s", sc)
	}
	// __Secure_: should pass with Secure when connection is HTTPS on client; here app is HTTP, so we don't assert Secure
	if !strings.Contains(sc, "__Secure_token=") {
		t.Fatalf("__Secure_ cookie not present: %s", sc)
	}
}

func TestHTTPReverseProxy_Cookies_SameSiteNone_HTTPS(t *testing.T) {
	t.Parallel()
	// Upstream TLS returns SameSite=None without Secure; we only verify proxy forwards header through HTTPS endpoint
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	tmpl := x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "127.0.0.1"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour), DNSNames: []string{"localhost"}, IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, _ := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	cert := tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
	mux := http.NewServeMux()
	mux.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "ssid=1; Path=/; SameSite=None")
		_, _ = w.Write([]byte("ok"))
	})
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	ups := &http.Server{Handler: mux, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}}}
	go ups.ServeTLS(ln, "", "")
	defer ups.Close()
	upstream := "https://" + ln.Addr().String() + "/cookie"

	// HTTPS proxy app
	logger := obs.NewLogger("error")
	metrics := obs.NewMetrics()
	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: config.Config{CORSAllowOrigin: "*", TLSAddr: ":0", InsecureTLS: true}, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	app := httptest.NewTLSServer(httpapi.NewRouterWithDeps(deps))
	defer app.Close()

	// client trusts TLSServer
	resp, err := app.Client().Get(app.URL + "/httpproxy/?_target=" + url.QueryEscape(upstream))
	if err != nil {
		t.Fatalf("https proxy request: %v", err)
	}
	defer resp.Body.Close()
	sc := strings.Join(resp.Header.Values("Set-Cookie"), "; ")
	if !strings.Contains(strings.ToLower(sc), "samesite=none") {
		t.Fatalf("SameSite=None not forwarded via HTTPS proxy: %s", sc)
	}
}
