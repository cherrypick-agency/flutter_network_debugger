package httpapi

import (
	"bytes"
	"testing"
)

func Test_opcodeFromWS(t *testing.T) {
	t.Parallel()

	cases := map[byte]string{
		0x1: "text",
		0x2: "binary",
		0x8: "close",
		0x9: "ping",
		0xA: "pong",
		0x3: "binary", // fallback
	}
	for code, want := range cases {
		if got := string(opcodeFromWS(code)); got != want {
			t.Errorf("opcodeFromWS(%x): want %q, got %q", code, want, got)
		}
	}
}

func Test_hexPreview(t *testing.T) {
	t.Parallel()

	if hexPreview(nil) != "" {
		t.Fatalf("nil -> empty string expected")
	}
	if hexPreview([]byte{}) != "" {
		t.Fatalf("empty -> empty string expected")
	}
	// Small buffer
	small := []byte{0x00, 0x01, 0xFA}
	if got := hexPreview(small); got != "0001fa" {
		t.Fatalf("unexpected hex: %q", got)
	}
	// 256 byte limit
	big := bytes.Repeat([]byte{0xAB}, 300)
	got := hexPreview(big)
	// expect encoding length 256 bytes -> 512 hex characters
	if len(got) != 512 {
		t.Fatalf("expected 512 hex chars, got %d", len(got))
	}
	// and that all characters correspond to 0xAB => "ab"
	for i := 0; i < len(got); i += 2 {
		if got[i:i+2] != "ab" {
			t.Fatalf("unexpected pair at %d: %q", i, got[i:i+2])
		}
	}
}

func Test_min(t *testing.T) {
	t.Parallel()

	if min(1, 2) != 1 || min(2, 1) != 1 || min(2, 2) != 2 {
		t.Fatalf("min broken")
	}
}
