package domain

import (
    "encoding/json"
    "testing"
    "time"
)

func TestFrame_JSON_OmitemptyAndEdges(t *testing.T) {
    t.Parallel()

    // zero-value: omitempty для BodyFile
    f0 := Frame{}
    b0, err := json.Marshal(f0)
    if err != nil { t.Fatalf("marshal zero: %v", err) }
    s0 := string(b0)
    if contains(s0, "bodyFile") { t.Fatalf("bodyFile must be omitted on zero-value: %s", s0) }
    if !containsAll(s0, "id", "ts", "direction", "opcode", "size", "preview") {
        t.Fatalf("expected base keys present on zero-value: %s", s0)
    }

    // заполненный: BodyFile должен появиться только если непустой
    f := Frame{
        ID:        "id",
        Ts:        time.Unix(0, 0).UTC(),
        Direction: DirectionClientToUpstream,
        Opcode:    OpcodeText,
        Size:      0,
        Preview:   "p",
        BodyFile:  "file.bin",
    }
    b, err := json.Marshal(f)
    if err != nil { t.Fatalf("marshal: %v", err) }
    s := string(b)
    if !(containsAll(s, `"id"`, `"ts"`, `"direction"`, `"opcode"`, `"size"`, `"preview"`, `"bodyFile"`)) {
        t.Fatalf("missing keys: %s", s)
    }

    // крайние значения Size
    f.Size = 1<<31 - 1
    if err := json.Unmarshal(b, &f); err != nil { t.Fatalf("unmarshal back: %v", err) }
}

func containsAll(s string, subs ...string) bool {
    for _, sub := range subs {
        if !contains(s, sub) { return false }
    }
    return true
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (func() bool { return indexOf(s, sub) >= 0 })() }

func indexOf(s, sub string) int {
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}


