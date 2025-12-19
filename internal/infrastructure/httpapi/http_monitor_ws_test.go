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
	// после подключения пошлём событие и убедимся, что оно доезжает
	hub.Broadcast(domain.MonitorEvent{Type: "test", ID: "1"})
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := c.ReadMessage()
	if err != nil || len(msg) == 0 {
		t.Fatalf("expected message from hub, err=%v", err)
	}
	_ = c.Close()
	// дать серверу время обработать закрытие и убрать клиента
	time.Sleep(50 * time.Millisecond)
}
