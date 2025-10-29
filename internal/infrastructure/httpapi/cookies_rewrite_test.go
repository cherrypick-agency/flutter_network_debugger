package httpapi

import (
	"strings"
	"testing"
)

func TestRewriteSetCookiesForProxy(t *testing.T) {
	t.Parallel()

	// Mode=off — без изменений
	{
		h := newHeaderFake()
		raw := "a=b; Path=/"
		h.Add("Set-Cookie", raw)
		rewriteSetCookiesForProxy(h, CookieRewriteOptions{Mode: CookieModeOff})
		vals := h.Values("Set-Cookie")
		if len(vals) != 1 || vals[0] != raw {
			t.Fatalf("ожидалось без изменений, получили: %#v", vals)
		}
	}

	// isolate + hostOnly — переименование и удаление Domain
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
			t.Fatalf("ожидался 1 заголовок, получили: %d", len(vals))
		}
		if !strings.HasPrefix(vals[0], "gpx_ns__id=") {
			t.Fatalf("ожидалось переименование с namespace, получили: %q", vals[0])
		}
		if strings.Contains(vals[0], "Domain=") {
			t.Fatalf("при hostOnly домен должен удаляться: %q", vals[0])
		}
	}

	// proxyHost c безопасным доменом
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
			t.Fatalf("ожидался выставленный домен proxy.example: %q", out)
		}
	}

	// proxyHost с IP/localhost — домен убирается
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
			t.Fatalf("для IP домен должен быть удалён: %q", out)
		}
	}

	// PathStrategy=prefix — путь префикс
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
			t.Fatalf("ожидался Path=/proxy: %q", out)
		}
	}

	// __Host- — принудительный Path=/ и без Domain
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
			t.Fatalf("для __Host- домена быть не должно: %q", out)
		}
		if !strings.Contains(out, "; Path=/") {
			t.Fatalf("для __Host- путь должен быть /: %q", out)
		}
		if !strings.Contains(out, "; Secure") {
			t.Fatalf("для __Host- под HTTPS должен быть Secure: %q", out)
		}
	}

	// HTTPS + SameSite=None — Secure обязателен
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
			t.Fatalf("SameSite=None под HTTPS должен включать Secure: %q", out)
		}
	}

	// множественные + непарсибельная строка
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
			t.Fatalf("ожидались 2 строки Set-Cookie, получили: %d", len(vals))
		}
		if vals[1] != "badtoken" {
			t.Fatalf("непарсибельная строка должна пройти как есть: %q", vals[1])
		}
	}
}
