package id

import "testing"

func TestNew_GeneratesHex24(t *testing.T) {
	got := New()
	if len(got) != 24 {
		t.Fatalf("expected 24 hex chars, got %d: %q", len(got), got)
	}
	for i := 0; i < len(got); i++ {
		c := got[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("non-hex char at %d: %q", i, got)
		}
	}
}

func TestNew_Uniqueness(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		s := New()
		if _, ok := seen[s]; ok {
			t.Fatalf("duplicate id: %s", s)
		}
		seen[s] = struct{}{}
	}
}
