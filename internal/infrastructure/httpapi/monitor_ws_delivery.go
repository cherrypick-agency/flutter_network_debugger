package httpapi

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"network-debugger/internal/domain"

	"github.com/gorilla/websocket"
)

type monitorWSDelivery struct {
	mu       sync.RWMutex
	clients  map[*websocket.Conn]chan []byte
	upgrader websocket.Upgrader
}

func newMonitorWSDelivery() *monitorWSDelivery {
	return &monitorWSDelivery{
		clients:  make(map[*websocket.Conn]chan []byte),
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
}

func (d *monitorWSDelivery) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := d.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	out := make(chan []byte, 256)
	d.mu.Lock()
	d.clients[conn] = out
	d.mu.Unlock()

	go d.writeLoop(conn, out)

	_ = conn.SetReadDeadline(time.Time{})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	d.mu.Lock()
	if out, ok := d.clients[conn]; ok {
		delete(d.clients, conn)
		close(out)
	}
	d.mu.Unlock()
}

func (d *monitorWSDelivery) Broadcast(ev domain.MonitorEvent) {
	payload, _ := json.Marshal(ev)

	d.mu.RLock()
	for _, out := range d.clients {
		select {
		case out <- payload:
		default:
		}
	}
	d.mu.RUnlock()
}

func (d *monitorWSDelivery) ClientCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.clients)
}

func (d *monitorWSDelivery) writeLoop(conn *websocket.Conn, out <-chan []byte) {
	for payload := range out {
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			break
		}
	}
	_ = conn.Close()
}
