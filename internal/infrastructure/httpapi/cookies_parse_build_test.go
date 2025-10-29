package httpapi

import (
	"strings"
	"testing"
)

func TestParseBuildCookies_FlagsSameSitePriorityExtra(t *testing.T) {
	t.Parallel()

	// валидный сложный случай
	raw := "sid=abc; Secure; HttpOnly; Partitioned; SameSite=None; Priority=High; Domain=.example.com; Path=/api; Foo=Bar; Baz"
	p, ok := parseSetCookieLine(raw)
	if !ok {
		t.Fatalf("parseSetCookieLine: ожидался ok=true")
	}
	if !p.flags["secure"] || !p.flags["httponly"] || !p.flags["partitioned"] {
		t.Fatalf("флаги Secure/HttpOnly/Partitioned не распознаны: %+v", p.flags)
	}
	if p.attrs["domain"] != ".example.com" || p.attrs["path"] != "/api" {
		t.Fatalf("ожидались Domain=.example.com и Path=/api, получили: %+v", p.attrs)
	}
	if p.attrs["samesite"] != "None" {
		t.Fatalf("ожидался SameSite=None, получили: %q", p.attrs["samesite"])
	}
	if p.attrs["priority"] != "High" {
		t.Fatalf("ожидался Priority=High, получили: %q", p.attrs["priority"])
	}
	if len(p.extraTokens) == 0 || !containsAll(p.extraTokens, []string{"Foo=Bar", "Baz"}) {
		t.Fatalf("ожидались extraTokens с Foo=Bar и Baz, получили: %#v", p.extraTokens)
	}

	out := buildSetCookieLine(p)
	if !strings.Contains(out, "; Priority=High") || !strings.Contains(out, "; Foo=Bar") || !strings.Contains(out, "; Baz") {
		t.Fatalf("buildSetCookieLine потерял Priority/extraTokens: %q", out)
	}
	if !strings.Contains(out, "SameSite=None") {
		t.Fatalf("ожидался нормализованный SameSite=None в build: %q", out)
	}

	// повторный parse результата build
	p2, ok := parseSetCookieLine(out)
	if !ok {
		t.Fatalf("parseSetCookieLine(build(..)) должен быть ok=true")
	}
	if !p2.flags["secure"] || !p2.flags["httponly"] || !p2.flags["partitioned"] {
		t.Fatalf("флаги потеряны после round-trip: %+v", p2.flags)
	}
	if p2.attrs["samesite"] != "None" {
		t.Fatalf("SameSite должен сохраниться после round-trip: %q", p2.attrs["samesite"])
	}

	// невалидная строка
	if _, ok := parseSetCookieLine("badtoken"); ok {
		t.Fatalf("невалидная строка должна давать ok=false")
	}

	// неизвестное значение SameSite должно сохраняться как есть в build
	rawWeird := "n=v; SameSite=Weird"
	p3, ok := parseSetCookieLine(rawWeird)
	if !ok {
		t.Fatalf("ожидался ok для строки с неизвестным SameSite")
	}
	out3 := buildSetCookieLine(p3)
	if !strings.Contains(out3, "SameSite=Weird") {
		t.Fatalf("ожидалось сохранить неизвестный SameSite как есть, получили: %q", out3)
	}
}

func containsAll(ss []string, need []string) bool {
	for _, n := range need {
		found := false
		for _, s := range ss {
			if s == n {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}


