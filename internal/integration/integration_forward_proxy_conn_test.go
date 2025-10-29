package integration

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// helper: dial to proxy with rw deadlines
func dialProxy(t *testing.T, host string, rd, wd time.Duration) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(rd))
	_ = conn.SetWriteDeadline(time.Now().Add(wd))
	return conn
}

func TestForwardProxy_AbsoluteURI_StealthOff_AndHopByHopStripped(t *testing.T) {

	// upstream that echoes forwarding and hop-by-hop headers
	mux := http.NewServeMux()
	mux.HandleFunc("/hdr", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"xff": r.Header.Get("X-Forwarded-For"),
			"xfp": r.Header.Get("X-Forwarded-Proto"),
			"via": r.Header.Get("Via"),
			"conn": r.Header.Get("Connection"),
			"pc": r.Header.Get("Proxy-Connection"),
			"te": r.Header.Get("Te"),
		})
	})
	up := httptest.NewServer(mux)
	defer up.Close()

	app, deps := startHTTPApp(t)
	defer app.Close()
	// Явно выключаем stealth, чтобы проверять, что заголовки выставляются
	deps.Cfg.StealthHeaders = false

	pURL, _ := url.Parse(app.URL)
	conn := dialProxy(t, pURL.Host, 2*time.Second, 2*time.Second)
	defer conn.Close()

	raw := strings.Join([]string{
		"GET " + up.URL + "/hdr HTTP/1.1",
		"Host: " + pURL.Host,
		"Connection: keep-alive",
		"Proxy-Connection: keep-alive",
		"Te: trailers",
		"",
		"",
	}, "\r\n")
	if _, err := conn.Write([]byte(raw)); err != nil {
		t.Fatalf("write: %v", err)
	}
	br := bufio.NewReader(conn)
	status, _ := br.ReadString('\n')
	if !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("status: %s", status)
	}
	// skip headers
	for {
		line, _ := br.ReadString('\n')
		if line == "\r\n" || line == "\n" || line == "" {
			break
		}
	}
	// small body
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	body, _ := br.ReadBytes('}')
	var got map[string]string
	_ = json.Unmarshal(body, &got)
	if got["xfp"] == "" || got["via"] == "" {
		t.Fatalf("stealth off — ожидаем заполненные Via/X-Forwarded-Proto: %+v", got)
	}
	if got["conn"] != "" || got["pc"] != "" {
		t.Fatalf("hop-by-hop должны быть стёрты: %+v", got)
	}
}

func TestForwardProxy_Errors_DNS_Refused_Timeout(t *testing.T) {
	app, _ := startHTTPApp(t)
	defer app.Close()
	pURL, _ := url.Parse(app.URL)

	// DNS error -> 502
	conn1 := dialProxy(t, pURL.Host, 1*time.Second, 1*time.Second)
	_, _ = conn1.Write([]byte("GET http://nonexistent.invalid/ HTTP/1.1\r\nHost: " + pURL.Host + "\r\n\r\n"))
	br1 := bufio.NewReader(conn1)
	line1, _ := br1.ReadString('\n')
	_ = conn1.Close()
	if !strings.HasPrefix(line1, "HTTP/1.1 502") {
		t.Fatalf("DNS error must be 502, got %q", line1)
	}

	// Connection refused -> 502
	conn2 := dialProxy(t, pURL.Host, 1*time.Second, 1*time.Second)
	_, _ = conn2.Write([]byte("GET http://127.0.0.1:1/ HTTP/1.1\r\nHost: " + pURL.Host + "\r\n\r\n"))
	br2 := bufio.NewReader(conn2)
	line2, _ := br2.ReadString('\n')
	_ = conn2.Close()
	if !strings.HasPrefix(line2, "HTTP/1.1 502") {
		t.Fatalf("refused must be 502, got %q", line2)
	}

	// Timeout: создаём upstream, который принимает, но не отвечает
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	go func() {
		c, _ := ln.Accept()
		if c != nil {
			defer c.Close()
			// просто молчим достаточно долго
			time.Sleep(2 * time.Second)
		}
	}()
	conn3 := dialProxy(t, pURL.Host, 300*time.Millisecond, 1*time.Second)
	_, _ = conn3.Write([]byte("GET http://" + ln.Addr().String() + "/ HTTP/1.1\r\nHost: " + pURL.Host + "\r\n\r\n"))
	br3 := bufio.NewReader(conn3)
	_ = conn3.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	if _, err := br3.ReadString('\n'); err == nil {
		// Если вдруг что-то ответили — это тоже ок, но ожидаем таймаут
		_ = conn3.Close()
	} else if ne, ok := err.(net.Error); !ok || !ne.Timeout() {
		_ = conn3.Close()
		t.Fatalf("ожидался timeout, получили: %v", err)
	}
}

func TestForwardProxy_CONNECT_Tunnel_HTTPGet(t *testing.T) {
	up, upURL := startUpstreamHTTP(t)
	defer up.Close()
	app, _ := startHTTPApp(t)
	defer app.Close()

	u, _ := url.Parse(upURL)
	target := u.Host
	pURL, _ := url.Parse(app.URL)

	conn := dialProxy(t, pURL.Host, 2*time.Second, 2*time.Second)
	defer conn.Close()

	// CONNECT
	req := "CONNECT " + target + " HTTP/1.1\r\nHost: " + target + "\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write connect: %v", err)
	}
	br := bufio.NewReader(conn)
	status, _ := br.ReadString('\n')
	if !strings.HasPrefix(status, "HTTP/1.1 200") {
		t.Fatalf("connect status: %s", status)
	}
	// consume the rest of CONNECT response headers until blank line
	for {
		line, _ := br.ReadString('\n')
		if line == "\r\n" || line == "\n" || line == "" {
			break
		}
	}

	// now HTTP GET through the tunnel
	get := "GET /get?q=ok HTTP/1.1\r\nHost: " + target + "\r\n\r\n"
	if _, err := conn.Write([]byte(get)); err != nil {
		t.Fatalf("tunnel write: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	line, err := br.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "HTTP/1.1 ") {
		t.Fatalf("tunnel read: %v (%q)", err, line)
	}
}


