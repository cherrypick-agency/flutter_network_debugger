package integration

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Tests forward-proxy for ws:// (without TLS): Upgrade 101 and message exchange
func TestForwardProxy_WSPlain(t *testing.T) {
	t.Parallel()

	// upstream echo ws://
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	// app with forward proxy enabled
	appSrv, deps := startAppServer(t)
	defer appSrv.Close()
	_ = deps

	// Dial via HTTP forward-proxy
	proxyHost := ensureForwardProxyAddr(t, appSrv, deps)
	proxyURL := &url.URL{Scheme: "http", Host: proxyHost}
	// gorilla/websocket Dialer.Proxy supports http-proxy
	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)

	c, _, err := d.Dial(echoWS, nil)
	if err != nil {
		t.Fatalf("forward ws dial failed: %v", err)
	}
	// echo sanity
	if err := c.WriteMessage(websocket.TextMessage, []byte("hello-fwd")); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	_ = c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "hello-fwd" {
		t.Fatalf("unexpected echo: %q", string(data))
	}
	_ = c.Close()
}

// Large frames and backpressure via forward-proxy
func TestForwardProxy_WSLargeAndBackpressure(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, deps := startAppServer(t)
	defer appSrv.Close()

	proxyHost := ensureForwardProxyAddr(t, appSrv, deps)
	proxyURL := &url.URL{Scheme: "http", Host: proxyHost}
	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)

	c, _, err := d.Dial(echoWS, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	// large text ~128KB
	bigText := bytes.Repeat([]byte("A"), 128*1024)
	if err := c.WriteMessage(websocket.TextMessage, bigText); err != nil {
		t.Fatalf("write big text: %v", err)
	}
	// large binary ~150KB
	bigBin := bytes.Repeat([]byte{0xEE}, 150*1024)
	if err := c.WriteMessage(websocket.BinaryMessage, bigBin); err != nil {
		t.Fatalf("write big bin: %v", err)
	}

	// backpressure: series of short messages
	for i := 0; i < 30; i++ {
		if err := c.WriteMessage(websocket.TextMessage, []byte("s")); err != nil {
			t.Fatalf("write small %d: %v", i, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// read at least several echo responses to ensure connection is alive
	reads := 0
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && reads < 3 {
		_ = c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if _, _, err := c.ReadMessage(); err == nil {
			reads++
		}
	}
	if reads == 0 {
		t.Fatalf("no echoes received under load")
	}
	_ = c.Close()
}

// Ping/Pong and proper connection close
func TestForwardProxy_WSPingPongAndClose(t *testing.T) {
	t.Parallel()
	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, deps := startAppServer(t)
	defer appSrv.Close()

	proxyHost := ensureForwardProxyAddr(t, appSrv, deps)
	proxyURL := &url.URL{Scheme: "http", Host: proxyHost}
	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)

	// use "rich" mode so server periodically sends ping
	c, _, err := d.Dial(echoWS+"?server=rich", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	// client ping
	if err := c.WriteControl(websocket.PingMessage, []byte("cli"), time.Now().Add(500*time.Millisecond)); err != nil {
		t.Fatalf("client ping: %v", err)
	}
	// short pause, then proper close
	time.Sleep(200 * time.Millisecond)
	_ = c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(1*time.Second))
	_ = c.Close()
}
