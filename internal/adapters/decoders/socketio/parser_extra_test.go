package socketio

import (
	"os"
	"strings"
	"testing"
)

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

func TestParseEvent_MoreVariants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   string
		ns   string
		ev   string
		ok   bool
	}{
		{"event-basic", "42[\"login\",{\"u\":\"john\"}]", "", "login", true},
		{"event-with-ackid", "42,17[\"ev\"]", "", "ev", true},
		{"event-with-nsp", "42/chat,3[\"ev\"]", "/chat", "ev", true},
		{"ack-basic", "43[\"ok\"]", "", "ack", true},
		{"ack-with-nsp-ackid", "43/x,5[\"ok\"]", "/x", "ack", true},
		{"binary-event", "451-/room,2[\"msg\",{\"_placeholder\":true,\"num\":0}]", "/room", "msg", true},
		{"binary-ack", "461-/x,7[\"ignored\"]", "/x", "ack", true},
		{"invalid-short", "4", "", "", false},
		{"invalid-badjson", "42[not-json]", "", "", false},
		{"invalid-binary-no-dash", "451/room[\"x\"]", "", "", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ns, ev, _, ok := ParseEvent(tc.in)
			if ok != tc.ok || ns != tc.ns || ev != tc.ev {
				t.Fatalf("got ns=%q ev=%q ok=%v", ns, ev, ok)
			}
		})
	}
}

func TestParseEvent_LargeInputsFromTestdata(t *testing.T) {
	t.Parallel()
	files := []string{"testdata/large_event.txt", "testdata/binary_ack.txt"}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		in := strings.TrimSpace(string(data))
		if _, _, _, ok := ParseEvent(in); !ok {
			t.Fatalf("ParseEvent failed for %s", f)
		}
	}
}
