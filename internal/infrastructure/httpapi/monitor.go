package httpapi

import (
    "encoding/json"
    "net/http"
    "sync"
    "time"
    "github.com/gorilla/websocket"
)

type MonitorEvent struct {
    Type string `json:"type"`
    ID   string `json:"id"`
    Ref  string `json:"ref,omitempty"`
}

type MonitorHub struct {
    mu      sync.RWMutex
    clients map[*websocket.Conn]struct{}
    upgrader websocket.Upgrader
    wmu     sync.Mutex
    // listeners are in-process subscribers (e.g., SSE forwarders)
    lmu       sync.RWMutex
    listeners map[chan MonitorEvent]struct{}
}

func NewMonitorHub() *MonitorHub {
    return &MonitorHub{
        clients: make(map[*websocket.Conn]struct{}),
        upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
        listeners: make(map[chan MonitorEvent]struct{}),
    }
}

func (h *MonitorHub) HandleWS(w http.ResponseWriter, r *http.Request) {
    c, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil { return }
    h.mu.Lock()
    h.clients[c] = struct{}{}
    h.mu.Unlock()
    _ = c.SetReadDeadline(time.Time{})
    for {
        // keepalive reads to detect client close
        if _, _, err := c.ReadMessage(); err != nil {
            break
        }
    }
    h.mu.Lock()
    delete(h.clients, c)
    h.mu.Unlock()
    _ = c.Close()
}

func (h *MonitorHub) Broadcast(ev MonitorEvent) {
    data, _ := json.Marshal(ev)
    // snapshot clients to avoid holding read lock during writes
    h.mu.RLock()
    clients := make([]*websocket.Conn, 0, len(h.clients))
    for c := range h.clients { clients = append(clients, c) }
    h.mu.RUnlock()
    // snapshot listeners
    h.lmu.RLock()
    subs := make([]chan MonitorEvent, 0, len(h.listeners))
    for ch := range h.listeners { subs = append(subs, ch) }
    h.lmu.RUnlock()
    // serialize writes to prevent concurrent writes to same conn
    h.wmu.Lock()
    for _, c := range clients {
        _ = c.SetWriteDeadline(time.Now().Add(2 * time.Second))
        _ = c.WriteMessage(websocket.TextMessage, data)
    }
    h.wmu.Unlock()
    // non-blocking notify listeners
    for _, ch := range subs {
        select { case ch <- ev: default: /* drop if slow */ }
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


