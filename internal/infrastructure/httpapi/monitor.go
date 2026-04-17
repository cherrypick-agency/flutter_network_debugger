package httpapi

import (
	"net/http"
	"network-debugger/internal/domain"
	"sync"
)

// MonitorHub manages WebSocket connections and in-process listeners for monitor events.
type MonitorHub struct {
	ws *monitorWSDelivery
	// listeners are in-process subscribers (e.g., SSE forwarders)
	lmu       sync.RWMutex
	listeners map[chan domain.MonitorEvent]struct{}
}

func NewMonitorHub() *MonitorHub {
	return &MonitorHub{
		ws:        newMonitorWSDelivery(),
		listeners: make(map[chan domain.MonitorEvent]struct{}),
	}
}

func (h *MonitorHub) HandleWS(w http.ResponseWriter, r *http.Request) {
	h.ws.HandleWS(w, r)
}

func (h *MonitorHub) Broadcast(ev domain.MonitorEvent) {
	h.ws.Broadcast(ev)

	// Similarly for in-process listeners
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
	return h.ws.ClientCount()
}
