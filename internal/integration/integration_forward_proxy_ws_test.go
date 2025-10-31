package integration

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Минимальный тест ws:// через forward‑proxy: Upgrade 101 и один фрейм эхо
func TestForwardProxy_WS_UpgradeAndEcho(t *testing.T) {

	echoSrv, echoWS := startEchoWSServer(t)
	defer echoSrv.Close()

	appSrv, deps := startAppServer(t)
	defer appSrv.Close()

	proxyHost := ensureForwardProxyAddr(t, appSrv, deps)
	proxyURL := &url.URL{Scheme: "http", Host: proxyHost}
	d := *websocket.DefaultDialer
	d.Proxy = http.ProxyURL(proxyURL)

	c, resp, err := d.Dial(echoWS, nil)
	if err != nil {
		t.Fatalf("dial via proxy failed: %v", err)
	}
	if resp == nil || resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101 Switching Protocols, got %v", resp)
	}
	defer c.Close()

	if err := c.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = c.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, data, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != "ping" {
		t.Fatalf("unexpected echo: %q", string(data))
	}
}
