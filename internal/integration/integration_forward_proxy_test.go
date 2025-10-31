package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Helper: perform GET via http.Client with proxy and return status and body
func rawProxyGET(t *testing.T, proxyHost string, absoluteURL string, extraHeaders []string) (string, string) {
	t.Helper()
	proxyURL, _ := url.Parse("http://" + proxyHost)
	tr := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	cli := &http.Client{Transport: tr, Timeout: 3 * time.Second}
	req, err := http.NewRequest(http.MethodGet, absoluteURL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for _, h := range extraHeaders {
		if h == "" {
			continue
		}
		if i := strings.Index(h, ":"); i > 0 {
			k := strings.TrimSpace(h[:i])
			v := strings.TrimSpace(h[i+1:])
			req.Header.Set(k, v)
		}
	}
	resp, err := cli.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.Status, string(b)
}

func TestForwardProxy_AbsoluteURI_SetsForwardHeaders(t *testing.T) {
	t.Parallel()
	// upstream echoes forward headers
	mux := http.NewServeMux()
	mux.HandleFunc("/hdr", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"xff": r.Header.Get("X-Forwarded-For"),
			"xfp": r.Header.Get("X-Forwarded-Proto"),
			"via": r.Header.Get("Via"),
		})
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()

	app, deps := startHTTPApp(t)
	defer app.Close()
	proxyHost := ensureForwardProxyAddr(t, app, deps)

	status, body := rawProxyGET(t, proxyHost, upstream.URL+"/hdr", nil)
	if !strings.HasPrefix(status, "HTTP/1.1 200") && !strings.HasPrefix(status, "200 ") {
		t.Fatalf("status: %s", status)
	}
	var got map[string]string
	// body is an HTTP stream; scan for last JSON line
	dec := json.NewDecoder(strings.NewReader(body))
	_ = dec.Decode(&got)
	if got["xfp"] == "" || got["via"] == "" {
		t.Fatalf("forward headers not set: %+v", got)
	}
}

func TestForwardProxy_HopByHopStripped(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"conn": r.Header.Get("Connection"),
			"pc":   r.Header.Get("Proxy-Connection"),
			"te":   r.Header.Get("Te"),
		})
	})
	upstream := httptest.NewServer(mux)
	defer upstream.Close()
	app, deps := startHTTPApp(t)
	defer app.Close()
	proxyHost := ensureForwardProxyAddr(t, app, deps)

	headers := []string{
		"Connection: keep-alive",
		"Proxy-Connection: keep-alive",
		"Te: trailers",
	}
	status, body := rawProxyGET(t, proxyHost, upstream.URL+"/hop", headers)
	if !strings.HasPrefix(status, "HTTP/1.1 200") && !strings.HasPrefix(status, "200 ") {
		t.Fatalf("status: %s", status)
	}
	var got map[string]string
	_ = json.NewDecoder(strings.NewReader(body)).Decode(&got)
	if got["conn"] != "" || got["pc"] != "" {
		t.Fatalf("hop-by-hop must be stripped: %+v", got)
	}
}

func TestForwardProxy_DNSError_Returns502(t *testing.T) {
	t.Parallel()
	app, deps := startHTTPApp(t)
	defer app.Close()
	proxyHost := ensureForwardProxyAddr(t, app, deps)
	status, _ := rawProxyGET(t, proxyHost, "http://nonexistent.invalid/", nil)
	if !strings.HasPrefix(status, "HTTP/1.1 502") && !strings.HasPrefix(status, "502 ") {
		t.Fatalf("expected 502 on DNS error, got %s", status)
	}
}

func TestForwardProxy_ConnectionRefused_Returns502(t *testing.T) {
	t.Parallel()
	app, deps := startHTTPApp(t)
	defer app.Close()
	proxyHost := ensureForwardProxyAddr(t, app, deps)
	status, _ := rawProxyGET(t, proxyHost, "http://127.0.0.1:1/", nil)
	if !strings.HasPrefix(status, "HTTP/1.1 502") && !strings.HasPrefix(status, "502 ") {
		t.Fatalf("expected 502 on refused, got %s", status)
	}
}
