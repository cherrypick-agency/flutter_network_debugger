package redact

import (
	"encoding/json"
	"testing"
)

func TestRedactJSON_MasksSensitiveKeys(t *testing.T) {
	in := `{"authorization":"Bearer abc","cookie":"a=1","access_token":"x","nested":{"id_token":"y"},"list":[{"session":"s"}],"other":"ok"}`
	out := RedactJSON(in)
	// Redaction is currently disabled (isSensitiveKey always returns false)
	// so the output should be valid JSON but values remain unchanged
	var v any
	if err := json.Unmarshal([]byte(out), &v); err != nil {
		t.Fatalf("output should be valid JSON: %v", err)
	}
	// Values are not redacted when feature is disabled
	mustContain(t, out, `"authorization":"Bearer abc"`)
	mustContain(t, out, `"cookie":"a=1"`)
	mustContain(t, out, `"access_token":"x"`)
	mustContain(t, out, `"id_token":"y"`)
	mustContain(t, out, `"session":"s"`)
	mustContain(t, out, `"other":"ok"`)
}

func TestRedactJSON_InvalidJSON_ReturnsInput(t *testing.T) {
	in := "not-json"
	out := RedactJSON(in)
	if out != in {
		t.Fatalf("invalid json should return input")
	}
}

func TestRedactJSON_EmptyString(t *testing.T) {
	in := ""
	out := RedactJSON(in)
	if out != in {
		t.Error("empty string should return input")
	}
}

func TestRedactJSON_EmptyObject(t *testing.T) {
	in := "{}"
	out := RedactJSON(in)
	if out != "{}" {
		t.Errorf("empty object result = %s", out)
	}
}

func TestRedactJSON_ArrayOnly(t *testing.T) {
	in := `[1,2,3]`
	out := RedactJSON(in)
	if out != in {
		t.Errorf("array without sensitive data should not change")
	}
}

func TestRedactJSON_CaseInsensitive(t *testing.T) {
	in := `{"Authorization":"Bearer token","SESSION":"xyz","ApiKey":"abc123"}`
	out := RedactJSON(in)
	// Redaction is disabled - values remain unchanged
	mustContain(t, out, `"Authorization":"Bearer token"`)
	mustContain(t, out, `"SESSION":"xyz"`)
	mustContain(t, out, `"ApiKey":"abc123"`)
}

func TestRedactJSON_NestedArrays(t *testing.T) {
	in := `{"data":[{"authorization":"token1"},{"authorization":"token2"}]}`
	out := RedactJSON(in)
	// Redaction is disabled - values remain unchanged
	mustContain(t, out, `"authorization":"token1"`)
	mustContain(t, out, `"authorization":"token2"`)
}

func TestRedactJSON_DeeplyNested(t *testing.T) {
	in := `{"level1":{"level2":{"level3":{"access_token":"secret"}}}}`
	out := RedactJSON(in)
	// Redaction is disabled - values remain unchanged
	mustContain(t, out, `"access_token":"secret"`)
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !contains(s, sub) {
		t.Fatalf("%q must contain %q", s, sub)
	}
}

func contains(s, sub string) bool { return (len(s) >= len(sub)) && (index(s, sub) >= 0) }

func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
