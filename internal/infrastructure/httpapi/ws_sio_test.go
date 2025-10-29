package httpapi

import (
    "testing"
    uc "network-debugger/internal/usecase"
)

func TestRecordSIOIfAny_EventAndAck(t *testing.T) {
    // minimal svc with events repo
    // reuse NewSessionService with stub repos
    s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
    d := &Deps{Svc: s, Monitor: NewMonitorHub()}
    // 42 event
    if !d.recordSIOIfAny("s1", "42[\"ev\",{}]", "f1") { t.Fatalf("42 should record") }
    // 43 ack
    if !d.recordSIOIfAny("s1", "43,5[]", "f2") { t.Fatalf("43 should record ack") }
}


