package adapters

import (
	"context"
	"network-debugger/internal/features/sessions_cli/ports"
	httpapi "network-debugger/internal/infrastructure/httpapi"
)

// MonitorEventStream — adapter MonitorHub → ports.EventStream.
type MonitorEventStream struct {
	hub *httpapi.MonitorHub
}

func NewMonitorEventStream(h *httpapi.MonitorHub) *MonitorEventStream {
	return &MonitorEventStream{hub: h}
}

func (m *MonitorEventStream) Subscribe(ctx context.Context) (<-chan ports.Event, func(), error) {
	src := m.hub.Subscribe()
	out := make(chan ports.Event, 256)

	// Proxy events and map types.
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-src:
				if !ok {
					return
				}
				var t ports.EventType
				switch ev.Type {
				case "session_started":
					t = ports.EventSessionStarted
				case "session_ended":
					t = ports.EventSessionEnded
				case "http_tx_added":
					t = ports.EventHTTPTxAdded
				default:
					// ignore others
					continue
				}
				out <- ports.Event{Type: t, SessionID: ev.ID, Ref: ev.Ref}
			}
		}
	}()

	cancel := func() { m.hub.Unsubscribe(src) }
	return out, cancel, nil
}
