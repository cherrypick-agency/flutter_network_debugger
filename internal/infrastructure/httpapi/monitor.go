package httpapi

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"network-debugger/internal/domain"
	"sync"
	"time"
)

// MonitorHub manages WebSocket connections and in-process listeners for monitor events.
type MonitorHub struct {
	mu       sync.RWMutex
	clients  map[*websocket.Conn]chan []byte
	upgrader websocket.Upgrader
	// listeners are in-process subscribers (e.g., SSE forwarders)
	lmu       sync.RWMutex
	listeners map[chan domain.MonitorEvent]struct{}
}

func NewMonitorHub() *MonitorHub {
	return &MonitorHub{
		clients:   make(map[*websocket.Conn]chan []byte),
		upgrader:  websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		listeners: make(map[chan domain.MonitorEvent]struct{}),
	}
}

func (h *MonitorHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	ch := make(chan []byte, 256)
	h.mu.Lock()
	h.clients[c] = ch
	h.mu.Unlock()
	// writer goroutine
	go func(conn *websocket.Conn, out <-chan []byte) {
		for msg := range out {
			_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				break
			}
		}
		_ = conn.Close()
	}(c, ch)
	_ = c.SetReadDeadline(time.Time{})
	for {
		// keepalive reads to detect client close
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
	h.mu.Lock()
	if ch, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *MonitorHub) Broadcast(ev domain.MonitorEvent) {
	data, _ := json.Marshal(ev)
	// Отправляем под RLock, чтобы не пересечься с close(ch) при удалении клиента.
	// Отправка неблокирующая, поэтому держать RLock безопасно и коротко.
	h.mu.RLock()
	for _, ch := range h.clients {
		select {
		case ch <- data:
		default: /* drop if slow */
		}
	}
	h.mu.RUnlock()

	// Аналогично для in-process listeners
	h.lmu.RLock()
	for ch := range h.listeners {
		select {
		case ch <- ev:
		default: /* drop if slow */
		}
	}
	h.lmu.RUnlock()
}

// Subscribe returns a channel receiving monitor events. Caller must Unsubscribe.
func (h *MonitorHub) Subscribe() chan domain.MonitorEvent {
	ch := make(chan domain.MonitorEvent, 256)
	h.lmu.Lock()
	h.listeners[ch] = struct{}{}
	h.lmu.Unlock()
	return ch
}

// Unsubscribe removes a listener channel.
func (h *MonitorHub) Unsubscribe(ch chan domain.MonitorEvent) {
	h.lmu.Lock()
	if _, ok := h.listeners[ch]; ok {
		delete(h.listeners, ch)
		close(ch)
	}
	h.lmu.Unlock()
}

// ClientCount returns the number of connected WebSocket clients.
func (h *MonitorHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
