package domain

import (
    "encoding/json"
    "testing"
    "time"
)

func TestEvent_JSON_AckID_Omitempty_AndFrameIDs(t *testing.T) {
    t.Parallel()

    ts := time.Unix(0, 0).UTC()
    e := Event{
        ID:        "e1",
        Ts:        ts,
        Namespace: "/nsp",
        Name:      "ev",
        ArgsPreview:"p",
        FrameIDs:  []string{"f1", "f2"},
    }
    b, err := json.Marshal(e)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    if !containsAll(s, `"id"`, `"ts"`, `"namespace"`, `"event"`, `"argsPreview"`, `"frameIds"`) {
        t.Fatalf("missing keys: %s", s)
    }
    if contains(s, "ackId") { t.Fatalf("ackId must be omitted when nil: %s", s) }

    ack := int64(7)
    e.AckID = &ack
    b2, _ := json.Marshal(e)
    if !contains(string(b2), "ackId") { t.Fatalf("ackId must be present when non-nil") }
}


