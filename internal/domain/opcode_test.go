package domain

import (
    "encoding/json"
    "testing"
)

func TestOpcode_ConstantsAndJSON(t *testing.T) {
    t.Parallel()

    want := map[Opcode]string{
        OpcodeText:   "text",
        OpcodeBinary: "binary",
        OpcodePing:   "ping",
        OpcodePong:   "pong",
        OpcodeClose:  "close",
    }
    for k, v := range want {
        if string(k) != v { t.Fatalf("opcode value changed: %q != %q", k, v) }
    }

    type wrap struct{ Op Opcode `json:"op"` }
    for k := range want {
        b, err := json.Marshal(wrap{Op: k})
        if err != nil { t.Fatalf("marshal: %v", err) }
        var w wrap
        if err := json.Unmarshal(b, &w); err != nil { t.Fatalf("unmarshal: %v", err) }
        if w.Op != k { t.Fatalf("roundtrip mismatch: %q", k) }
    }

    // произвольное значение остаётся строкой
    b, _ := json.Marshal(wrap{Op: Opcode("x")})
    var w wrap
    _ = json.Unmarshal(b, &w)
    if w.Op != Opcode("x") { t.Fatalf("custom opcode lost") }
}


