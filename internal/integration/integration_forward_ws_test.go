package integration

import (
    "net/http"
    "net/url"
    "testing"
    "time"

    "github.com/gorilla/websocket"
)

// Тестирует forward‑proxy для ws:// (без TLS): Upgrade 101 и обмен сообщениями
func TestForwardProxy_WSPlain(t *testing.T) {
    t.Parallel()

    // upstream echo ws://
    echoSrv, echoWS := startEchoWSServer(t)
    defer echoSrv.Close()

    // app with forward proxy enabled
    appSrv, deps := startAppServer(t)
    defer appSrv.Close()
    _ = deps

    // Dial через HTTP forward‑proxy
    proxyURL, _ := url.Parse(appSrv.URL)
    // httptest.Server URL: http://127.0.0.1:xxxxx
    // gorilla/websocket Dialer.Proxy поддерживает http‑прокси
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


