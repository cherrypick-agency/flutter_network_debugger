package httpapi

import (
	"testing"
	"time"

	"network-debugger/internal/domain"
)

func TestMonitorHub_SubscribeBroadcastUnsubscribe(t *testing.T) {
	h := NewMonitorHub()
	ch := h.Subscribe()
	h.Broadcast(domain.MonitorEvent{Type: "t", ID: "1"})
	select {
	case ev := <-ch:
		if ev.Type != "t" || ev.ID != "1" {
			t.Fatalf("unexpected ev: %+v", ev)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timeout waiting for event")
	}
	h.Unsubscribe(ch)
}
