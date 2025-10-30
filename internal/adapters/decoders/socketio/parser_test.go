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

func TestParseEvent_Ack(t *testing.T) {
	_, ev, args, ok := ParseEvent("43,5[]")
	if !ok {
		t.Fatal("should parse ack")
	}
	if ev != "ack" {
		t.Errorf("event = %s, want ack", ev)
	}
	if args != "[]" {
		t.Errorf("args = %s, want []", args)
	}
}

func TestParseEvent_AckWithNamespace(t *testing.T) {
	nsp, ev, _, ok := ParseEvent("43/chat,5[1,2]")
	if !ok {
		t.Fatal("should parse ack with namespace")
	}
	if nsp != "/chat" {
		t.Errorf("namespace = %s, want /chat", nsp)
	}
	if ev != "ack" {
		t.Errorf("event = %s, want ack", ev)
	}
}

func TestParseEvent_BinaryEvent(t *testing.T) {
	_, ev, _, ok := ParseEvent("451-[\"event\",{\"_placeholder\":true,\"num\":0}]")
	if !ok {
		t.Fatal("should parse binary event")
	}
	if ev != "event" {
		t.Errorf("event = %s, want event", ev)
	}
}

func TestParseEvent_BinaryEventWithNamespace(t *testing.T) {
	nsp, ev, _, ok := ParseEvent("451-/chat,[\"message\",{}]")
	if !ok {
		t.Fatal("should parse binary event with namespace")
	}
	if nsp != "/chat" {
		t.Errorf("namespace = %s, want /chat", nsp)
	}
	if ev != "message" {
		t.Errorf("event = %s, want message", ev)
	}
}

func TestParseEvent_BinaryAck(t *testing.T) {
	_, ev, _, ok := ParseEvent("462-[1,2]")
	if !ok {
		t.Fatal("should parse binary ack")
	}
	if ev != "ack" {
		t.Errorf("event = %s, want ack", ev)
	}
}

func TestParseEvent_EmptyString(t *testing.T) {
	_, _, _, ok := ParseEvent("")
	if ok {
		t.Error("empty string should not parse")
	}
}

func TestParseEvent_TooShort(t *testing.T) {
	_, _, _, ok := ParseEvent("4")
	if ok {
		t.Error("too short string should not parse")
	}
}

func TestParseEvent_InvalidJSON(t *testing.T) {
	_, _, _, ok := ParseEvent("42[invalid]")
	if ok {
		t.Error("invalid JSON should not parse")
	}
}

func TestParseEvent_EmptyArray(t *testing.T) {
	_, _, _, ok := ParseEvent("42[]")
	if ok {
		t.Error("empty array should not parse")
	}
}

func TestParseEvent_NonStringEvent(t *testing.T) {
	_, _, _, ok := ParseEvent("42[123]")
	if ok {
		t.Error("non-string event name should not parse")
	}
}

func TestParseEvent_NamespaceWithoutComma(t *testing.T) {
	_, _, _, ok := ParseEvent("42/chat[\"event\"]")
	if ok {
		t.Error("namespace without comma should not parse")
	}
}

func TestParseEvent_AckNamespaceWithoutComma(t *testing.T) {
	_, _, _, ok := ParseEvent("43/chat[\"event\"]")
	if ok {
		t.Error("ack namespace without comma should not parse")
	}
}

func TestParseEvent_AckWithoutBracket(t *testing.T) {
	_, _, _, ok := ParseEvent("43,5")
	if ok {
		t.Error("ack without bracket should not parse")
	}
}

func TestParseBinaryLike_NoAttachments(t *testing.T) {
	_, _, _, ok := ParseEvent("45-[\"event\"]")
	if ok {
		t.Error("missing attachments number should not parse")
	}
}

func TestParseBinaryLike_NoDash(t *testing.T) {
	_, _, _, ok := ParseEvent("451[\"event\"]")
	if ok {
		t.Error("missing dash should not parse")
	}
}

func TestParseBinaryLike_NamespaceWithoutComma(t *testing.T) {
	_, _, _, ok := ParseEvent("451-/chat[\"event\",{}]")
	if ok {
		t.Error("binary event namespace without comma should not parse")
	}
}

func TestParseBinaryLike_WithAckID(t *testing.T) {
	_, ev, _, ok := ParseEvent("451-17[\"event\",{}]")
	if !ok {
		t.Fatal("should parse binary event with ack ID")
	}
	if ev != "event" {
		t.Errorf("event = %s, want event", ev)
	}
}

func TestParseBinaryLike_InvalidJSON(t *testing.T) {
	_, _, _, ok := ParseEvent("451-[invalid]")
	if ok {
		t.Error("binary event with invalid JSON should not parse")
	}
}

func TestParseBinaryLike_EmptyArray(t *testing.T) {
	_, _, _, ok := ParseEvent("451-[]")
	if ok {
		t.Error("binary event with empty array should not parse")
	}
}

func TestParseBinaryLike_NonStringEvent(t *testing.T) {
	_, _, _, ok := ParseEvent("451-[123]")
	if ok {
		t.Error("binary event with non-string event name should not parse")
	}
}

func TestIsDigits_Valid(t *testing.T) {
	if !isDigits("123") {
		t.Error("123 should be digits")
	}
	if !isDigits("0") {
		t.Error("0 should be digits")
	}
}

func TestIsDigits_Invalid(t *testing.T) {
	if isDigits("") {
		t.Error("empty string should not be digits")
	}
	if isDigits("12a") {
		t.Error("12a should not be digits")
	}
	if isDigits("abc") {
		t.Error("abc should not be digits")
	}
}
