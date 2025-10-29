package httpapi

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type MonitorEvent struct {
	Type  string        `json:"type"`
	ID    string        `json:"id"`
	Ref   string        `json:"ref,omitempty"`
	Error *ErrorDetails `json:"error,omitempty"`
}

type ErrorDetails struct {
	Code    string `json:"code"`             // Error code: CONNECTION_CLOSED, SERVER_UNAVAILABLE, etc.
	Message string `json:"message"`          // Human-readable error message
	Raw     string `json:"raw,omitempty"`    // Original technical error (for debugging)
	Target  string `json:"target,omitempty"` // Target URL
	Method  string `json:"method,omitempty"` // HTTP method
}

type MonitorHub struct {
	mu       sync.RWMutex
    clients  map[*websocket.Conn]chan []byte
	upgrader websocket.Upgrader
	// listeners are in-process subscribers (e.g., SSE forwarders)
	lmu       sync.RWMutex
	listeners map[chan MonitorEvent]struct{}
}

func NewMonitorHub() *MonitorHub {
	return &MonitorHub{
        clients:   make(map[*websocket.Conn]chan []byte),
		upgrader:  websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		listeners: make(map[chan MonitorEvent]struct{}),
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

func (h *MonitorHub) Broadcast(ev MonitorEvent) {
	data, _ := json.Marshal(ev)
	// snapshot clients to avoid holding read lock during writes
	h.mu.RLock()
    outs := make([]chan []byte, 0, len(h.clients))
    for _, ch := range h.clients {
        outs = append(outs, ch)
	}
	h.mu.RUnlock()
	// snapshot listeners
	h.lmu.RLock()
	subs := make([]chan MonitorEvent, 0, len(h.listeners))
	for ch := range h.listeners {
		subs = append(subs, ch)
	}
	h.lmu.RUnlock()
    // non-blocking fan-out
    for _, ch := range outs {
        select { case ch <- data: default: /* drop if slow */ }
    }
	// non-blocking notify listeners
	for _, ch := range subs {
		select {
		case ch <- ev:
		default: /* drop if slow */
		}
	}
}

// Subscribe returns a channel receiving monitor events. Caller must Unsubscribe.
func (h *MonitorHub) Subscribe() chan MonitorEvent {
	ch := make(chan MonitorEvent, 256)
	h.lmu.Lock()
	h.listeners[ch] = struct{}{}
	h.lmu.Unlock()
	return ch
}

// Unsubscribe removes a listener channel.
func (h *MonitorHub) Unsubscribe(ch chan MonitorEvent) {
	h.lmu.Lock()
	if _, ok := h.listeners[ch]; ok {
		delete(h.listeners, ch)
		close(ch)
	}
	h.lmu.Unlock()
}
