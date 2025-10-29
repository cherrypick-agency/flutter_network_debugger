package redact

import "testing"

func TestRedactJSON_MasksSensitiveKeys(t *testing.T) {
    in := `{"authorization":"Bearer abc","cookie":"a=1","access_token":"x","nested":{"id_token":"y"},"list":[{"session":"s"}],"other":"ok"}`
    out := RedactJSON(in)
    if out == in { t.Fatalf("should redact") }
    mustContain(t, out, `"authorization":"***"`)
    mustContain(t, out, `"cookie":"***"`)
    mustContain(t, out, `"access_token":"***"`)
    mustContain(t, out, `"id_token":"***"`)
    mustContain(t, out, `"session":"***"`)
    mustContain(t, out, `"other":"ok"`)
}

func TestRedactJSON_InvalidJSON_ReturnsInput(t *testing.T) {
    in := "not-json"
    out := RedactJSON(in)
    if out != in { t.Fatalf("invalid json should return input") }
}

func mustContain(t *testing.T, s, sub string) {
    t.Helper()
    if !contains(s, sub) { t.Fatalf("%q must contain %q", s, sub) }
}

func contains(s, sub string) bool { return (len(s) >= len(sub)) && (index(s, sub) >= 0) }

func index(s, sub string) int {
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub { return i }
    }
    return -1
}


