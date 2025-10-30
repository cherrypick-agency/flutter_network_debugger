package httpapi

import (
	uc "network-debugger/internal/usecase"
	"testing"
)

func TestRecordSIOIfAny_EventAndAck(t *testing.T) {
	// minimal svc with events repo
	// reuse NewSessionService with stub repos
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}
	// 42 event
	if !d.recordSIOIfAny("s1", "42[\"ev\",{}]", "f1") {
		t.Fatalf("42 should record")
	}
	// 43 ack
	if !d.recordSIOIfAny("s1", "43,5[]", "f2") {
		t.Fatalf("43 should record ack")
	}
}

func TestRecordSIOIfAny_EmptyString(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	if d.recordSIOIfAny("s1", "", "f1") {
		t.Error("empty string should return false")
	}
}

func TestRecordSIOIfAny_EventWithNamespace(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Event with namespace
	if !d.recordSIOIfAny("s1", "42/chat,[\"message\",{\"text\":\"hello\"}]", "f1") {
		t.Error("event with namespace should be recorded")
	}
}

func TestRecordSIOIfAny_EventWithAckID(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Event with ack ID
	if !d.recordSIOIfAny("s1", "42/chat,17[\"event\",{}]", "f1") {
		t.Error("event with ackID should be recorded")
	}
}

func TestRecordSIOIfAny_InvalidFormat(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Invalid formats should return false
	if d.recordSIOIfAny("s1", "invalid", "f1") {
		t.Error("invalid format should return false")
	}
	if d.recordSIOIfAny("s1", "41", "f1") {
		t.Error("non-42/43 code should return false")
	}
	if d.recordSIOIfAny("s1", "4", "f1") {
		t.Error("too short string should return false")
	}
}

func TestRecordSIOIfAny_Fallback42(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Try a 42 packet that might trigger fallback logic
	if !d.recordSIOIfAny("s1", "42[\"test\",1,2,3]", "f1") {
		t.Error("valid 42 event should be recorded")
	}
}

func TestRecordSIOIfAny_Fallback43WithAck(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// 43 packet with ack ID that triggers fallback
	if !d.recordSIOIfAny("s1", "43,10[]", "f1") {
		t.Error("43 ack packet should be recorded via fallback")
	}
}

func TestRecordSIOIfAny_Fallback43WithoutAck(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// 43 packet without ack ID - should not record
	if d.recordSIOIfAny("s1", "43invalid", "f1") {
		t.Error("43 without ack should return false")
	}
}

func TestRecordSIOIfAny_BinaryEvent(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Binary event (packet type 45)
	if !d.recordSIOIfAny("s1", "451-[\"binary_event\",{}]", "f1") {
		t.Error("binary event should be recorded")
	}
}

func TestRecordSIOIfAny_BinaryAck(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Binary ack (packet type 46)
	if !d.recordSIOIfAny("s1", "462-[1,2]", "f1") {
		t.Error("binary ack should be recorded")
	}
}

func TestRecordSIOIfAny_NotSocketIO(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Svc: s, Monitor: NewMonitorHub()}

	// Non-Socket.IO messages
	if d.recordSIOIfAny("s1", "0", "f1") {
		t.Error("packet type 0 should not be recorded as event")
	}
	if d.recordSIOIfAny("s1", "1", "f1") {
		t.Error("packet type 1 should not be recorded as event")
	}
	if d.recordSIOIfAny("s1", "2", "f1") {
		t.Error("packet type 2 should not be recorded as event")
	}
}
