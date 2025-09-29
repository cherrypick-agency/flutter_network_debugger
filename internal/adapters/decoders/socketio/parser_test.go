package socketio

import "testing"

func TestParseEventV4NoNSP(t *testing.T) {
    nsp, ev, args, ok := ParseEvent("42[\"chat_message\",{\"text\":\"hi\"}]")
    if !ok || nsp != "" || ev != "chat_message" || len(args) == 0 {
        t.Fatalf("unexpected: ok=%v nsp=%q ev=%q args=%q", ok, nsp, ev, args)
    }
}

func TestParseEventV4WithNSPAndAck(t *testing.T) {
    nsp, ev, args, ok := ParseEvent("42/chat,17[\"message\",{\"text\":\"hi\"}]")
    if !ok || nsp != "/chat" || ev != "message" || len(args) == 0 {
        t.Fatalf("unexpected: ok=%v nsp=%q ev=%q args=%q", ok, nsp, ev, args)
    }
}

func TestParseEventNotEvent(t *testing.T) {
    if _, _, _, ok := ParseEvent("3"); ok {
        t.Fatalf("should not parse non-event packet")
    }
}


