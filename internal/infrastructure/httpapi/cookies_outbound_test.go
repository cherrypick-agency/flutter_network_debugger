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
		t.Fatalf("для Mode!=isolate Cookie не должен меняться, получили: %v", vals)
	}
}

func TestRewriteOutboundCookieHeaderForUpstream_Isolate_FilterAndUnprefix(t *testing.T) {
	t.Parallel()
	h := newHeaderFake()
	// несколько заголовков, часть пар битые
	h.Add("Cookie", "gpx_ns__a=1; x; c=3")
	h.Add("Cookie", "gpx_ns__b=2; bad; d")

	rewriteOutboundCookieHeaderForUpstream(h, CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: "ns"})

	vals := h.Values("Cookie")
	if len(vals) == 0 {
		t.Fatalf("должен остаться один Cookie после склейки")
	}

	out := vals[0]
	// должны остаться только пары из текущего namespace без префикса
	if out != "a=1; b=2" {
		t.Fatalf("ожидались только a=1; b=2, получили: %q", out)
	}
}

func TestRewriteOutboundCookieHeaderForUpstream_Isolate_EmptyAfterFilter(t *testing.T) {
	t.Parallel()
	h := newHeaderFake()
	h.Add("Cookie", "c=3") // нет namespace

	rewriteOutboundCookieHeaderForUpstream(h, CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: "ns"})

	vals := h.Values("Cookie")
	if len(vals) != 0 {
		t.Fatalf("если после фильтра пусто — Cookie не должен быть установлен: %v", vals)
	}
}
