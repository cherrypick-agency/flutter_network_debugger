package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"os"
	"testing"
	"time"
)

// В тестах ниже не используем t.Parallel, т.к. меняем package-level флаги.

func TestHTTPResponsePreview_Gzip_Deflate_MaskAndTruncate(t *testing.T) {
	// Сохраняем и восстанавливаем глобальные настройки
	oldExpose := exposeSensitiveHeaders
	oldDecomp := previewDecompress
	oldMax := previewMaxBytes
	defer func() {
		exposeSensitiveHeaders = oldExpose
		previewDecompress = oldDecomp
		previewMaxBytes = oldMax
	}()

	// Явно включим декомпрессию и маскировку
	t.Setenv("PREVIEW_DECOMPRESS", "1")
	previewDecompress = true
	exposeSensitiveHeaders = false
	previewMaxBytes = 4096

	payload := []byte(`{"authorization":"Bearer abc","access_token":"x","value":123}`)

	// helper: build gzipped body
	mkGzip := func(src []byte) []byte {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, _ = zw.Write(src)
		_ = zw.Close()
		return buf.Bytes()
	}
	// (deflate ветку можно проверить отдельно при необходимости)

	cases := []struct {
		name string
		enc  string
		body []byte
	}{
		{"gzip", "gzip", mkGzip(payload)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Content-Encoding": []string{tc.enc},
					"Content-Type":     []string{"application/json"},
				},
				Body:          io.NopCloser(bytes.NewReader(tc.body)),
				ContentLength: int64(len(tc.body)),
			}
			prev := buildHTTPResponsePreview(resp)
			var m map[string]any
			if err := json.Unmarshal([]byte(prev), &m); err != nil {
				t.Fatalf("invalid json preview: %v", err)
			}
			body, _ := m["body"].(string)
			if body == "" {
				t.Fatalf("body must be present")
			}
			if body == string(payload) {
				t.Fatalf("should be redacted, got raw json")
			}
			if !strContains(body, "\"authorization\":\"***\"") || !strContains(body, "\"access_token\":\"***\"") {
				t.Fatalf("sensitive fields must be masked: %s", body)
			}

			// Усечение
			previewMaxBytes = 8
			resp = &http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Content-Encoding": []string{tc.enc},
					"Content-Type":     []string{"application/json"},
				},
				Body:          io.NopCloser(bytes.NewReader(tc.body)),
				ContentLength: int64(len(tc.body)),
			}
			prev2 := buildHTTPResponsePreview(resp)
			var m2 map[string]any
			_ = json.Unmarshal([]byte(prev2), &m2)
			if b2, _ := m2["body"].(string); !(len(b2) < len(body)) {
				t.Fatalf("body must be shorter after truncation: before=%d after=%d", len(body), len(b2))
			}
		})
	}
}

func TestHTTPRequestPreview_HeadersRedaction_Toggle(t *testing.T) {
	oldExpose := exposeSensitiveHeaders
	defer func() { exposeSensitiveHeaders = oldExpose }()

	// false: только маски, без headersRaw
	t.Setenv("EXPOSE_SENSITIVE_HEADERS", "0")
	exposeSensitiveHeaders = false
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/x", bytes.NewReader([]byte("a=1")))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Cookie", "sid=abc")
	req.Header.Set("X-Api-Key", "k")
	out := buildHTTPRequestPreview(req, []byte("{}"))
	var m map[string]any
	_ = json.Unmarshal([]byte(out), &m)
	hdr := m["headers"].(map[string]any)
	if hdr["Authorization"].(string) != "***" || hdr["Cookie"].(string) != "***" || hdr["X-Api-Key"].(string) != "***" {
		t.Fatalf("headers must be masked: %+v", hdr)
	}
	if _, ok := m["headersRaw"]; ok {
		t.Fatalf("headersRaw must be absent when expose=false")
	}

	// true: присутствует headersRaw с исходными значениями
	t.Setenv("EXPOSE_SENSITIVE_HEADERS", "1")
	exposeSensitiveHeaders = true
	out2 := buildHTTPRequestPreview(req, []byte("{}"))
	var m2 map[string]any
	_ = json.Unmarshal([]byte(out2), &m2)
	if _, ok := m2["headersRaw"]; !ok {
		t.Fatalf("headersRaw must be present when expose=true")
	}
	raw := m2["headersRaw"].(map[string]any)
	if raw["Authorization"].(string) != "Bearer secret" || raw["Cookie"].(string) != "sid=abc" {
		t.Fatalf("raw must contain original values: %+v", raw)
	}
}

func TestHTTPResponsePreview_HeadersRedaction_Toggle(t *testing.T) {
	oldExpose := exposeSensitiveHeaders
	defer func() { exposeSensitiveHeaders = oldExpose }()

	resp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Set-Cookie":    []string{"id=1; Secure"},
			"Authorization": []string{"Bearer xyz"},
			"Content-Type":  []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader([]byte("{}"))),
	}

	exposeSensitiveHeaders = false
	prev := buildHTTPResponsePreview(resp)
	var m map[string]any
	_ = json.Unmarshal([]byte(prev), &m)
	hdr := m["headers"].(map[string]any)
	if hdr["Set-Cookie"].(string) != "***" || hdr["Authorization"].(string) != "***" {
		t.Fatalf("response sensitive headers must be masked: %+v", hdr)
	}
	if _, ok := m["headersRaw"]; ok {
		t.Fatalf("headersRaw must be absent when expose=false")
	}

	exposeSensitiveHeaders = true
	prev2 := buildHTTPResponsePreview(resp)
	var m2 map[string]any
	_ = json.Unmarshal([]byte(prev2), &m2)
	if _, ok := m2["headersRaw"]; !ok {
		t.Fatalf("headersRaw must be present when expose=true")
	}
	raw := m2["headersRaw"].(map[string]any)
	if raw["Authorization"].(string) != "Bearer xyz" {
		t.Fatalf("raw authorization expected")
	}
}

func TestBuildHTTPRequestPreview_FormURLEncoded_Truncation(t *testing.T) {
	oldMax := previewMaxBytes
	defer func() { previewMaxBytes = oldMax }()
	previewMaxBytes = 8
	body := []byte("a=1234567890&b=2")
	req, _ := http.NewRequest(http.MethodPost, "http://example/form", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	out := buildHTTPRequestPreview(req, body)
	var m map[string]any
	_ = json.Unmarshal([]byte(out), &m)
	if len(m["body"].(string)) > 8 {
		t.Fatalf("truncation expected")
	}
}

// локальные утилиты
func strContains(s, sub string) bool { return len(s) >= len(sub) && strIndex(s, sub) >= 0 }
func strIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestBuildHTTPRequestPreview_FormParsingAndBody(t *testing.T) {
	oldExpose := exposeSensitiveHeaders
	oldDecomp := previewDecompress
	oldMax := previewMaxBytes
	defer func() { exposeSensitiveHeaders = oldExpose; previewDecompress = oldDecomp; previewMaxBytes = oldMax }()
	exposeSensitiveHeaders = false
	previewDecompress = true
	previewMaxBytes = 1024

	// urlencoded
	req1, _ := http.NewRequest(http.MethodPost, "http://example.com/x", bytes.NewReader([]byte("a=1&b=2")))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	out1 := buildHTTPRequestPreview(req1, []byte("a=1&b=2"))
	var m1 map[string]any
	_ = json.Unmarshal([]byte(out1), &m1)
	if _, ok := m1["form"].(map[string]any); !ok {
		t.Fatalf("expected form section for urlencoded")
	}

	// multipart
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	_ = mw.WriteField("field", "value")
	fw, _ := mw.CreateFormFile("file", "a.txt")
	_, _ = fw.Write([]byte("hello"))
	_ = mw.Close()
	req2, _ := http.NewRequest(http.MethodPost, "http://example.com/upload", bytes.NewReader(buf.Bytes()))
	req2.Header.Set("Content-Type", mw.FormDataContentType())
	out2 := buildHTTPRequestPreview(req2, buf.Bytes())
	var m2 map[string]any
	_ = json.Unmarshal([]byte(out2), &m2)
	if _, ok := m2["form"].(map[string]any); !ok {
		t.Fatalf("expected form section for multipart")
	}
}

func TestAugmentPreviewWithTimings_andHelpers(t *testing.T) {
	// augment timings
	prev := `{"type":"http_response"}`
	out := augmentPreviewWithTimings(prev, 10, 20)
	var m map[string]any
	_ = json.Unmarshal([]byte(out), &m)
	tm := m["timings"].(map[string]any)
	if int(tm["ttfbMs"].(float64)) != 10 || int(tm["totalMs"].(float64)) != 20 {
		t.Fatalf("timings not injected: %+v", tm)
	}

	// binary preview limit
	big := bytes.Repeat([]byte{0xAB}, 300)
	s := formatBinaryPreview(big, 300)
	if len(s) > 3*256 {
		t.Fatalf("binary preview should cap at 256 bytes")
	}

	// tls mappers
	if tlsVersionString(tls.VersionTLS12) == "" {
		t.Fatalf("tls version mapping")
	}
	if cipherSuiteString(tls.TLS_AES_128_GCM_SHA256) == "" {
		t.Fatalf("cipher mapping")
	}

	// time helpers
	now := time.Now()
	if durationMs(time.Time{}, now) != 0 {
		t.Fatalf("durationMs zero handling")
	}
	if useOrFallback(time.Time{}, now).IsZero() {
		t.Fatalf("useOrFallback")
	}
	if !timeFromUnixNanoOrZero(0).IsZero() {
		t.Fatalf("timeFromUnixNanoOrZero")
	}
	if int64ToInt(-1) != 0 {
		t.Fatalf("int64ToInt negative to zero")
	}
}

func TestRemoveHopHeaders_AndAbsoluteURL(t *testing.T) {
	h := http.Header{}
	h.Set("Connection", "keep-alive")
	h.Set("Proxy-Connection", "keep-alive")
	h.Set("Te", "trailers")
	h.Set("Upgrade", "websocket")
	h.Set("X-Ok", "1")
	removeHopHeaders(h)
	if h.Get("X-Ok") != "1" {
		t.Fatalf("non-hop header should remain")
	}
	if h.Get("Connection") != "" || h.Get("Proxy-Connection") != "" || h.Get("Te") != "" || h.Get("Upgrade") != "" {
		t.Fatalf("hop headers must be removed: %+v", h)
	}

	if !isAbsoluteURL("http://example.com/x") || isAbsoluteURL("/x") {
		t.Fatalf("isAbsoluteURL failed")
	}
	_, _ = url.Parse("") // keep import used
}

func TestCertsSummary_AndTryCompactJSON(t *testing.T) {
	// certsSummary empty
	if certsSummary(nil) != nil {
		t.Fatalf("nil -> nil")
	}
	c := &x509.Certificate{Subject: pkix.Name{CommonName: "cn"}}
	s := certsSummary([]*x509.Certificate{c})
	if len(s) != 1 {
		t.Fatalf("one cert expected")
	}

	if tryCompactJSON([]byte("not-json")) != "" {
		t.Fatalf("invalid json must return empty")
	}
	if tryCompactJSON([]byte("{\n \"a\" : 1 }")) != "{\"a\":1}" {
		t.Fatalf("should compact")
	}
}

func TestSanitizeForUpgrade_AndCookieSummary(t *testing.T) {
	// sanitizeForUpgrade keeps Upgrade/Connection
	h := http.Header{}
	h.Set("Proxy-Connection", "keep-alive")
	h.Set("Keep-Alive", "1")
	h.Set("Te", "trailers")
	h.Set("Upgrade", "websocket")
	h.Set("Connection", "Upgrade")
	sanitizeForUpgrade(h)
	if h.Get("Upgrade") == "" || h.Get("Connection") == "" {
		t.Fatalf("must keep upgrade/connection")
	}
	if h.Get("Proxy-Connection") != "" || h.Get("Keep-Alive") != "" || h.Get("Te") != "" {
		t.Fatalf("must strip hop headers: %+v", h)
	}

	// cookieSummary should count flags without exposing values
	resp := &http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Set-Cookie": []string{
				"a=1; Secure; HttpOnly; SameSite=None",
				"b=2; SameSite=Lax",
				"c=3; SameSite=Strict",
			},
		},
		Body: io.NopCloser(bytes.NewReader([]byte(""))),
	}
	exposeSensitiveHeaders = false
	prev := buildHTTPResponsePreview(resp)
	var m map[string]any
	_ = json.Unmarshal([]byte(prev), &m)
	cs := m["cookieSummary"].(map[string]any)
	if int(cs["setCookieCount"].(float64)) != 3 || int(cs["secure"].(float64)) < 1 || int(cs["httpOnly"].(float64)) < 1 {
		t.Fatalf("cookie summary mismatch: %+v", cs)
	}
}

func TestHTTPResponsePreview_TLS_Summary(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		TLS: &tls.ConnectionState{
			Version:            tls.VersionTLS13,
			CipherSuite:        tls.TLS_AES_128_GCM_SHA256,
			NegotiatedProtocol: "h2",
			ServerName:         "example.com",
		},
		Body: io.NopCloser(bytes.NewReader([]byte("ok"))),
	}
	prev := buildHTTPResponsePreview(resp)
	var m map[string]any
	_ = json.Unmarshal([]byte(prev), &m)
	if _, ok := m["tls"].(map[string]any); !ok {
		t.Fatalf("tls summary expected")
	}
}

func TestHTTPRequestPreview_GzipBody_Decompress(t *testing.T) {
	oldDecomp, oldMax := previewDecompress, previewMaxBytes
	defer func() { previewDecompress = oldDecomp; previewMaxBytes = oldMax }()
	previewDecompress = true
	previewMaxBytes = 64
	// gzip body
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("{\"a\":1}"))
	_ = zw.Close()
	req, _ := http.NewRequest(http.MethodPost, "http://example/x", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	out := buildHTTPRequestPreview(req, buf.Bytes())
	var m map[string]any
	_ = json.Unmarshal([]byte(out), &m)
	if m["body"].(string) != "{\"a\":1}" {
		t.Fatalf("expected decompressed json body: %v", m["body"])
	}
}

func TestHTTPRequestPreview_NoDecompress_ShowsRaw(t *testing.T) {
	oldDecomp, oldMax := previewDecompress, previewMaxBytes
	defer func() { previewDecompress = oldDecomp; previewMaxBytes = oldMax }()
	previewDecompress = false
	previewMaxBytes = 64
	// gzip body
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write([]byte("{\"a\":1}"))
	_ = zw.Close()
	req, _ := http.NewRequest(http.MethodPost, "http://example/x", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	out := buildHTTPRequestPreview(req, buf.Bytes())
	var m map[string]any
	_ = json.Unmarshal([]byte(out), &m)
	// body should not equal decompressed json when decompress disabled
	if m["body"].(string) == "{\"a\":1}" {
		t.Fatalf("expected raw when previewDecompress=false")
	}
}

func TestHTTPRequestPreview_DeflateBody_Decompress(t *testing.T) {
	oldDecomp, oldMax := previewDecompress, previewMaxBytes
	defer func() { previewDecompress = oldDecomp; previewMaxBytes = oldMax }()
	previewDecompress = true
	previewMaxBytes = 64
	// deflate body (raw DEFLATE)
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	_, _ = zw.Write([]byte("{\"a\":1}"))
	_ = zw.Close()
	req, _ := http.NewRequest(http.MethodPost, "http://example/x", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "deflate")
	out := buildHTTPRequestPreview(req, buf.Bytes())
	var m map[string]any
	_ = json.Unmarshal([]byte(out), &m)
	if m["body"].(string) != "{\"a\":1}" {
		t.Fatalf("expected decompressed json body: %v", m["body"])
	}
}

func TestCipherSuiteString_DefaultAndInt64ToIntOverflow(t *testing.T) {
	if cipherSuiteString(0xFFFF) == "" {
		t.Fatalf("default cipher mapping expected non-empty")
	}
	if int64ToInt((1 << 62)) <= 0 {
		t.Fatalf("overflow clamp expected >0")
	}
}

func TestTLSVersionString_DefaultEmpty(t *testing.T) {
	if tlsVersionString(0) != "" {
		t.Fatalf("unknown should map to empty string")
	}
	if tlsVersionString(tls.VersionTLS13) == "" {
		t.Fatalf("tls1.3 mapping expected")
	}
	if cipherSuiteString(tls.TLS_CHACHA20_POLY1305_SHA256) == "" {
		t.Fatalf("cipher mapping branch")
	}
}

func TestSpoolBody_WritesLimitedBytes(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{BodySpoolDir: "", BodyMaxBytes: 0}}
	data := bytes.Repeat([]byte("A"), 1024)
	path, err := d.spoolBody(bytes.NewReader(data), 100, "resp")
	if err != nil || path == "" {
		t.Fatalf("spool error: %v path=%q", err, path)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if st.Size() != 100 {
		t.Fatalf("expected 100 bytes spooled, got %d", st.Size())
	}
	_ = os.Remove(path)
}

func TestCloneAndCopyHeader(t *testing.T) {
	src := http.Header{"A": []string{"1", "2"}}
	dst := http.Header{}
	copyHeader(dst, src)
	if len(dst["A"]) != 2 {
		t.Fatalf("copyHeader failed: %+v", dst)
	}
	c := cloneHeader(dst)
	dst.Set("A", "x")
	if c.Get("A") != "1" {
		t.Fatalf("clone must be deep copy, got: %v", c["A"])
	}
}

func TestPrependConn_Read(t *testing.T) {
	prefix := []byte("HELLO ")
	a, b := net.Pipe()
	defer a.Close()
	defer b.Close()
	// Запишем в базовое соединение продолжение
	go func() { _, _ = b.Write([]byte("WORLD")) }()
	pc := &prependConn{Conn: a, r: bytes.NewReader(prefix)}
	buf := make([]byte, 11)
	n, err := io.ReadFull(pc, buf)
	if err != nil || string(buf[:n]) != "HELLO WORLD" {
		t.Fatalf("unexpected read: n=%d err=%v data=%q", n, err, buf[:n])
	}
}

func TestWriteError_JSON(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusBadRequest, "BAD", "oops", map[string]any{"x": 1})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status")
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content-type")
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["error"].(map[string]any)["code"].(string) != "BAD" {
		t.Fatalf("code")
	}
}

func TestWriteError_DefaultCode(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusTeapot, "", "", nil)
	if rr.Code != http.StatusTeapot {
		t.Fatalf("status mismatch")
	}
	if !strContains(rr.Body.String(), "I'm a teapot") {
		t.Fatalf("expected default status text in body: %s", rr.Body.String())
	}
}

func TestMultipartReaderFrom_NoBoundary(t *testing.T) {
	if mr := multipartReaderFrom("multipart/form-data", []byte("")); mr != nil {
		t.Fatalf("expected nil without boundary")
	}
}

func TestWithCORS_WritesHeadersAndHandlesOPTIONS(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := withCORS(cfgpkg.Config{CORSAllowOrigin: "*"}, base)

	// OPTIONS
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/x", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("options should return 204")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatalf("cors headers missing")
	}

	// GET
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("wrapped handler should be called")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("cors methods header missing")
	}
}

func TestHumanizeProxyError_Mapping(t *testing.T) {
	if c, _ := humanizeProxyError(io.EOF); c != "CONNECTION_CLOSED" {
		t.Fatalf("EOF mapping")
	}
	if c, _ := humanizeProxyError(context.DeadlineExceeded); c != "TIMEOUT" {
		t.Fatalf("timeout mapping")
	}
	if c, _ := humanizeProxyError(errors.New("connection refused")); c != "SERVER_UNAVAILABLE" {
		t.Fatalf("refused")
	}
	if c, _ := humanizeProxyError(errors.New("no such host")); c != "DNS_ERROR" {
		t.Fatalf("dns")
	}
	if c, _ := humanizeProxyError(errors.New("tls: bad cert")); c != "TLS_ERROR" {
		t.Fatalf("tls")
	}
}

func TestTryDecompress_Invalid(t *testing.T) {
	if _, ok := tryDecompress([]byte("junk"), "gzip"); ok {
		t.Fatalf("invalid gzip should fail")
	}
	if _, ok := tryDecompress([]byte("junk"), "deflate"); ok {
		t.Fatalf("invalid deflate should fail")
	}
}

func TestIsWebSocketRequest(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	if isWebSocketRequest(r) {
		t.Fatalf("no ws headers -> false")
	}
	r.Header.Set("Sec-WebSocket-Key", "k")
	if !isWebSocketRequest(r) {
		t.Fatalf("sec headers -> true")
	}
	r2, _ := http.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("Upgrade", "websocket")
	if !isWebSocketRequest(r2) {
		t.Fatalf("upgrade -> true")
	}
}
