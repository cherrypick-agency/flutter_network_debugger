package httpapi

import (
	"testing"
)

func TestRewriteOutboundCookieHeaderForUpstream_ModeAuto_NoChange(t *testing.T) {
	t.Parallel()
	h := newHeaderFake()
	h.Add("Cookie", "a=1")
	h.Add("Cookie", "b=2")

	rewriteOutboundCookieHeaderForUpstream(h, CookieRewriteOptions{Mode: CookieModeAuto, Namespace: "ns"})

	vals := h.Values("Cookie")
	if len(vals) != 2 {
		t.Fatalf("for Mode!=isolate Cookie should not change, got: %v", vals)
	}
}

func TestRewriteOutboundCookieHeaderForUpstream_Isolate_FilterAndUnprefix(t *testing.T) {
	t.Parallel()
	h := newHeaderFake()
	// multiple headers, some pairs are malformed
	h.Add("Cookie", "gpx_ns__a=1; x; c=3")
	h.Add("Cookie", "gpx_ns__b=2; bad; d")

	rewriteOutboundCookieHeaderForUpstream(h, CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: "ns"})

	vals := h.Values("Cookie")
	if len(vals) == 0 {
		t.Fatalf("should have one Cookie left after merging")
	}

	out := vals[0]
	// only pairs from current namespace without prefix should remain
	if out != "a=1; b=2" {
		t.Fatalf("expected only a=1; b=2, got: %q", out)
	}
}

func TestRewriteOutboundCookieHeaderForUpstream_Isolate_EmptyAfterFilter(t *testing.T) {
	t.Parallel()
	h := newHeaderFake()
	h.Add("Cookie", "c=3") // no namespace

	rewriteOutboundCookieHeaderForUpstream(h, CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: "ns"})

	vals := h.Values("Cookie")
	if len(vals) != 0 {
		t.Fatalf("if empty after filtering — Cookie should not be set: %v", vals)
	}
}
