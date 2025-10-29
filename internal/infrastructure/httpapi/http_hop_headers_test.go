package httpapi

import (
    "net/http"
    "testing"
)

func TestRemoveHopHeaders(t *testing.T) {
    h := http.Header{}
    h.Set("Connection", "keep-alive")
    h.Set("Proxy-Connection", "keep-alive")
    h.Set("Keep-Alive", "1")
    h.Set("Proxy-Authenticate", "x")
    h.Set("Proxy-Authorization", "x")
    h.Set("Te", "trailers")
    h.Set("Trailer", "X")
    h.Set("Transfer-Encoding", "chunked")
    h.Set("Upgrade", "websocket")
    h.Set("X", "ok")
    removeHopHeaders(h)
    if h.Get("X") != "ok" { t.Fatalf("non-hop header should remain") }
    if h.Get("Connection")+h.Get("Proxy-Connection")+h.Get("Keep-Alive")+h.Get("Proxy-Authenticate")+h.Get("Proxy-Authorization")+h.Get("Te")+h.Get("Trailer")+h.Get("Transfer-Encoding")+h.Get("Upgrade") != "" {
        t.Fatalf("hop headers must be removed: %+v", h)
    }
}


