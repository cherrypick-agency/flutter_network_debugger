package httpapi

import (
	"strings"
	"testing"
)

func TestRewriteSetCookiesForProxy(t *testing.T) {
	t.Parallel()

	// Mode=off — no changes
	{
		h := newHeaderFake()
		raw := "a=b; Path=/"
		h.Add("Set-Cookie", raw)
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{Mode: CookieModeOff})
		vals := h.Values("Set-Cookie")
		if len(vals) != 1 || vals[0] != raw {
			t.Fatalf("expected no changes, got: %#v", vals)
		}
	}

	// isolate + hostOnly — rename and remove Domain
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "id=1; Domain=example.com; Path=/")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:           CookieModeIsolate,
			DomainStrategy: "hostOnly",
			PathStrategy:   "root",
			Namespace:      "ns",
		})
		vals := h.Values("Set-Cookie")
		if len(vals) != 1 {
			t.Fatalf("expected 1 header, got: %d", len(vals))
		}
		if !strings.HasPrefix(vals[0], "gpx_ns__id=") {
			t.Fatalf("expected rename with namespace, got: %q", vals[0])
		}
		if strings.Contains(vals[0], "Domain=") {
			t.Fatalf("with hostOnly domain should be removed: %q", vals[0])
		}
	}

	// proxyHost with safe domain
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "s=v; Path=/")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:           CookieModeIsolate,
			DomainStrategy: "proxyHost",
			ProxyHost:      "proxy.example",
			PathStrategy:   "root",
			Namespace:      "n",
		})
		out := h.Values("Set-Cookie")[0]
		if !strings.Contains(out, "; Domain=proxy.example") {
			t.Fatalf("expected domain proxy.example to be set: %q", out)
		}
	}

	// proxyHost with IP/localhost — domain is removed
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "s=v; Domain=example.com; Path=/")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:           CookieModeIsolate,
			DomainStrategy: "proxyHost",
			ProxyHost:      "127.0.0.1",
			PathStrategy:   "root",
			Namespace:      "n",
		})
		out := h.Values("Set-Cookie")[0]
		if strings.Contains(out, "Domain=") {
			t.Fatalf("for IP domain should be removed: %q", out)
		}
	}

	// PathStrategy=prefix — path prefix
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "x=1")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:            CookieModeIsolate,
			DomainStrategy:  "hostOnly",
			PathStrategy:    "prefix",
			ProxyPathPrefix: "/proxy",
			Namespace:       "n",
		})
		out := h.Values("Set-Cookie")[0]
		if !strings.Contains(out, "; Path=/proxy") {
			t.Fatalf("expected Path=/proxy: %q", out)
		}
	}

	// __Host- — forced Path=/ and no Domain
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "__Host-sid=1; Domain=example.com; Path=/x")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:            CookieModeIsolate,
			DomainStrategy:  "proxyHost",
			ProxyHost:       "proxy.example",
			PathStrategy:    "prefix",
			ProxyPathPrefix: "/p",
			Namespace:       "n",
			HTTPS:           true,
		})
		out := h.Values("Set-Cookie")[0]
		if strings.Contains(out, "Domain=") {
			t.Fatalf("for __Host- domain should not exist: %q", out)
		}
		if !strings.Contains(out, "; Path=/") {
			t.Fatalf("for __Host- path should be /: %q", out)
		}
		if !strings.Contains(out, "; Secure") {
			t.Fatalf("for __Host- under HTTPS Secure should be set: %q", out)
		}
	}

	// HTTPS + SameSite=None — Secure is required
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "a=b; SameSite=None")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:           CookieModeIsolate,
			DomainStrategy: "hostOnly",
			PathStrategy:   "root",
			Namespace:      "n",
			HTTPS:          true,
		})
		out := h.Values("Set-Cookie")[0]
		if !strings.Contains(out, "; Secure") {
			t.Fatalf("SameSite=None under HTTPS should include Secure: %q", out)
		}
	}

	// multiple + unparseable string
	{
		h := newHeaderFake()
		h.Add("Set-Cookie", "ok=1; Path=/")
		h.Add("Set-Cookie", "badtoken")
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{
			Mode:           CookieModeIsolate,
			DomainStrategy: "hostOnly",
			PathStrategy:   "root",
			Namespace:      "n",
		})
		vals := h.Values("Set-Cookie")
		if len(vals) != 2 {
			t.Fatalf("expected 2 Set-Cookie strings, got: %d", len(vals))
		}
		if vals[1] != "badtoken" {
			t.Fatalf("unparseable string should pass as-is: %q", vals[1])
		}
	}
}
