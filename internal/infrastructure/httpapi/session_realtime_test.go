package httpapi

import (
	"testing"

	"network-debugger/internal/domain"
)

type fakeRealtimeClient struct {
	id      string
	emitted []realtimeEmission
	joined  []string
	left    []string
}

type realtimeEmission struct {
	event   string
	payload any
}

func (c *fakeRealtimeClient) ID() string { return c.id }

func (c *fakeRealtimeClient) Emit(event string, payload any) {
	c.emitted = append(c.emitted, realtimeEmission{event: event, payload: payload})
}

func (c *fakeRealtimeClient) JoinSession(sessionID string) {
	c.joined = append(c.joined, sessionID)
}

func (c *fakeRealtimeClient) LeaveSession(sessionID string) {
	c.left = append(c.left, sessionID)
}

func (c *fakeRealtimeClient) emissionsFor(event string) []realtimeEmission {
	out := make([]realtimeEmission, 0, len(c.emitted))
	for _, emission := range c.emitted {
		if emission.event == event {
			out = append(out, emission)
		}
	}
	return out
}

type fakeSessionRoomEmitter struct {
	emitted []roomEmission
}

type roomEmission struct {
	sessionID string
	event     string
	payload   any
}

func (e *fakeSessionRoomEmitter) EmitToSessionRoom(sessionID, event string, payload any) {
	e.emitted = append(e.emitted, roomEmission{sessionID: sessionID, event: event, payload: payload})
}

func TestSessionRealtimeServiceSendInitEmitsProjectedSessions(t *testing.T) {
	deps := makeDepsWithAPIStub()
	deps.Monitor = nil
	rooms := &fakeSessionRoomEmitter{}
	svc := newSessionRealtimeService(deps, rooms)
	client := &fakeRealtimeClient{id: "c1"}

	svc.sendInit(client, sioFilters{Limit: 1000, GroupBy: "domain", IncludeUnassigned: true})

	initEvents := client.emissionsFor("sessions:init")
	if len(initEvents) != 1 {
		t.Fatalf("expected exactly one sessions:init, got %d", len(initEvents))
	}
	payload, ok := initEvents[0].payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", initEvents[0].payload)
	}
	items, ok := payload["items"].([]sessionV1)
	if !ok {
		t.Fatalf("expected []sessionV1 items, got %T", payload["items"])
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 projected sessions, got %d", len(items))
	}
	groups, ok := payload["aggregate"].([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any aggregate, got %T", payload["aggregate"])
	}
	if len(groups) == 0 {
		t.Fatalf("expected non-empty aggregate groups")
	}
}

func TestSessionRealtimeServiceHandleMonitorEventSessionsCleared(t *testing.T) {
	deps := makeDepsWithAPIStub()
	deps.Monitor = nil
	svc := newSessionRealtimeService(deps, &fakeSessionRoomEmitter{})
	client := &fakeRealtimeClient{id: "c1"}
	svc.addSubscription(client, sioFilters{Limit: 1000, GroupBy: "domain", IncludeUnassigned: true})

	svc.handleMonitorEvent(domain.MonitorEvent{Type: "sessions_cleared", ID: "*"})

	initEvents := client.emissionsFor("sessions:init")
	if len(initEvents) != 1 {
		t.Fatalf("expected one sessions:init after clear, got %d", len(initEvents))
	}
	payload, ok := initEvents[0].payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", initEvents[0].payload)
	}
	items, ok := payload["items"].([]any)
	if !ok {
		t.Fatalf("expected []any items, got %T", payload["items"])
	}
	if len(items) != 0 {
		t.Fatalf("expected empty items after clear, got %d", len(items))
	}
	groups, ok := payload["aggregate"].([]any)
	if !ok {
		t.Fatalf("expected []any aggregate, got %T", payload["aggregate"])
	}
	if len(groups) != 0 {
		t.Fatalf("expected empty aggregate after clear, got %d", len(groups))
	}
}

func TestSessionRealtimeServiceSubscribeSessionEmitsCatchUpAndRoomFanout(t *testing.T) {
	deps := makeDepsWithAPIStub()
	deps.Monitor = nil
	rooms := &fakeSessionRoomEmitter{}
	svc := newSessionRealtimeService(deps, rooms)
	client := &fakeRealtimeClient{id: "c1"}

	svc.subscribeSession(client, "s1")
	svc.unsubscribeSession(client, "s1")

	if len(client.joined) != 1 || client.joined[0] != "s1" {
		t.Fatalf("expected join to session s1, got %#v", client.joined)
	}
	if len(client.left) != 1 || client.left[0] != "s1" {
		t.Fatalf("expected leave from session s1, got %#v", client.left)
	}
	if len(client.emissionsFor("session:frames")) != 1 {
		t.Fatalf("expected session:frames catch-up")
	}
	if len(client.emissionsFor("session:events")) != 1 {
		t.Fatalf("expected session:events catch-up")
	}
	if len(client.emissionsFor("session:http")) != 1 {
		t.Fatalf("expected session:http catch-up")
	}

	svc.handleMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: "s1"})
	svc.handleMonitorEvent(domain.MonitorEvent{Type: "event_added", ID: "s1"})
	svc.handleMonitorEvent(domain.MonitorEvent{Type: "http_tx_added", ID: "s1"})
	svc.handleMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: "s1"})
	svc.handleMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: "s1"})

	if len(rooms.emitted) != 5 {
		t.Fatalf("expected 5 room emissions, got %d", len(rooms.emitted))
	}
	wantEvents := []string{"session:frames", "session:events", "session:http", "session:started", "session:ended"}
	for i, want := range wantEvents {
		if rooms.emitted[i].sessionID != "s1" || rooms.emitted[i].event != want {
			t.Fatalf("unexpected room emission %d: %#v", i, rooms.emitted[i])
		}
	}
}

func TestSessionRealtimeServiceApplyEventEmitsUpsertAndAggregate(t *testing.T) {
	deps := makeDepsWithAPIStub()
	deps.Monitor = nil
	svc := newSessionRealtimeService(deps, &fakeSessionRoomEmitter{})
	client := &fakeRealtimeClient{id: "c1"}
	svc.addSubscription(client, sioFilters{Limit: 1000, GroupBy: "domain", IncludeUnassigned: true})

	sub, ok := svc.currentSubscription(client.ID())
	if !ok {
		t.Fatalf("expected subscription")
	}

	svc.applyEventToSubscription("session_started", "s1", sub)
	svc.emitAggregate(sub)

	if len(client.emissionsFor("sessions:upsert")) != 1 {
		t.Fatalf("expected sessions:upsert emission")
	}
	if len(client.emissionsFor("aggregate:update")) != 1 {
		t.Fatalf("expected aggregate:update emission")
	}
}
