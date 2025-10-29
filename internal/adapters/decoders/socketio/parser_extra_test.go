package socketio

import "testing"

func TestParseEvent_Ack43Variants(t *testing.T) {
    t.Parallel()

    // Вариант с запятой перед ack id
    nsp, ev, args, ok := ParseEvent("43,5[1,2]")
    if !ok || nsp != "" || ev != "ack" || args == "" {
        t.Fatalf("unexpected: ok=%v nsp=%q ev=%q args=%q", ok, nsp, ev, args)
    }

    // Вариант с namespace и ack id
    nsp, ev, args, ok = ParseEvent("43/chat,7[\"ok\"]")
    if !ok || nsp != "/chat" || ev != "ack" || args == "" {
        t.Fatalf("unexpected: ok=%v nsp=%q ev=%q args=%q", ok, nsp, ev, args)
    }

    // Некорректный: отсутствует args массив
    if _, _, _, ok := ParseEvent("43/chat,5"); ok {
        t.Fatalf("should be invalid without args array")
    }
}

func TestParseEvent_Binary45_Event(t *testing.T) {
    t.Parallel()

    // 45<attachments>-[/nsp][,ack][args]
    nsp, ev, args, ok := ParseEvent("451-/room,99[\"greet\",{\"x\":1}]")
    if !ok || nsp != "/room" || ev != "greet" || args == "" {
        t.Fatalf("unexpected: ok=%v nsp=%q ev=%q args=%q", ok, nsp, ev, args)
    }
}

func TestParseEvent_Binary46_Ack(t *testing.T) {
    t.Parallel()

    nsp, ev, args, ok := ParseEvent("460-/nsp,3[1,2,3]")
    if !ok || nsp != "/nsp" || ev != "ack" || args == "" {
        t.Fatalf("unexpected: ok=%v nsp=%q ev=%q args=%q", ok, nsp, ev, args)
    }
}

func TestParseEvent_Binary_InvalidAttachments(t *testing.T) {
    t.Parallel()

    // Нет цифр перед '-'
    if _, _, _, ok := ParseEvent("45-/[\"x\"]"); ok {
        t.Fatalf("should be invalid without attachment digits")
    }
}

func Test_isDigits(t *testing.T) {
    t.Parallel()

    if !isDigits("123") {
        t.Fatalf("expected true for digits")
    }
    if isDigits("") {
        t.Fatalf("empty is not digits")
    }
    if isDigits("12a") {
        t.Fatalf("letters should fail")
    }
}


