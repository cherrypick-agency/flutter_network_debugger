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

func TestParseSetCookieLine_EdgeCases(t *testing.T) {
	t.Parallel()

	// Empty string
	if _, ok := parseSetCookieLine(""); ok {
		t.Error("empty string should return ok=false")
	}

	// Only "="
	if _, ok := parseSetCookieLine("="); ok {
		t.Error("only '=' should return ok=false")
	}

	// "=value" without name
	if _, ok := parseSetCookieLine("=value"); ok {
		t.Error("'=value' without name should return ok=false")
	}

	// "name=" without value (should be ok)
	p, ok := parseSetCookieLine("name=")
	if !ok {
		t.Error("'name=' should be valid")
	}
	if p.name != "name" || p.value != "" {
		t.Errorf("'name=': got name=%q value=%q", p.name, p.value)
	}

	// Cookie with spaces
	p, ok = parseSetCookieLine("  id  =  123  ")
	if !ok {
		t.Error("cookie with spaces should be valid")
	}
	if p.name != "id" || p.value != "123" {
		t.Errorf("spaces: got name=%q value=%q", p.name, p.value)
	}

	// Multiple empty tokens
	p, ok = parseSetCookieLine("a=1; ; ; Secure")
	if !ok {
		t.Error("empty tokens should be skipped")
	}
	if !p.flags["secure"] {
		t.Error("secure flag should be set")
	}

	// Attribute without '=' is extraToken
	p, ok = parseSetCookieLine("n=v; RandomFlag")
	if !ok {
		t.Error("should parse with random flag")
	}
	if len(p.extraTokens) == 0 || p.extraTokens[0] != "RandomFlag" {
		t.Errorf("RandomFlag should be in extraTokens, got %v", p.extraTokens)
	}

	// Unknown attribute with '=' goes to extraTokens
	p, ok = parseSetCookieLine("n=v; Custom=Value")
	if !ok {
		t.Error("should parse with custom attribute")
	}
	if len(p.extraTokens) == 0 || p.extraTokens[0] != "Custom=Value" {
		t.Errorf("Custom=Value should be in extraTokens, got %v", p.extraTokens)
	}

	// expires attribute
	p, ok = parseSetCookieLine("n=v; Expires=Wed, 21 Oct 2015 07:28:00 GMT")
	if !ok {
		t.Error("should parse with expires")
	}
	if p.attrs["expires"] != "Wed, 21 Oct 2015 07:28:00 GMT" {
		t.Errorf("expires: got %q", p.attrs["expires"])
	}

	// max-age attribute
	p, ok = parseSetCookieLine("n=v; Max-Age=3600")
	if !ok {
		t.Error("should parse with max-age")
	}
	if p.attrs["max-age"] != "3600" {
		t.Errorf("max-age: got %q", p.attrs["max-age"])
	}
}
