package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"network-debugger/internal/domain"
)

func TestMonitorHub_HandleWS_BasicUpgradeAndClose(t *testing.T) {
	hub := NewMonitorHub()
	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWS))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	// wait until client registers in hub
	deadline := time.Now().Add(2 * time.Second)
	for hub.ClientCount() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if hub.ClientCount() == 0 {
		t.Fatal("client not registered in hub")
	}
	// after connecting, send event and verify it arrives
	hub.Broadcast(domain.MonitorEvent{Type: "test", ID: "1"})
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := c.ReadMessage()
	if err != nil || len(msg) == 0 {
		t.Fatalf("expected message from hub, err=%v", err)
	}
	_ = c.Close()
	// give server time to process close and remove client
	time.Sleep(50 * time.Millisecond)
}
