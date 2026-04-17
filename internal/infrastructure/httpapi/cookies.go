package httpapi

import (
	"net/url"

	proxycookies "github.com/777genius/proxykit/cookies"
)

// CookieMode defines behavior of cookie rewriting
const (
	CookieModeIsolate = proxycookies.ModeIsolate
	CookieModeAuto    = proxycookies.ModeAuto
	CookieModeOff     = proxycookies.ModeOff
)

// CookieRewriteOptions controls how Set-Cookie/Cookie are rewritten at the reverse proxy boundary
type CookieRewriteOptions = proxycookies.RewriteOptions

func computeNamespaceFromURL(u *url.URL) string {
	return proxycookies.NamespaceForURL(u)
}

func sanitizeHost(hostport string) string {
	return proxycookies.SanitizeHost(hostport)
}

func domainAttrSafe(host string) bool {
	return proxycookies.DomainAttrSafe(host)
}

func namespacePrefix(ns string) string {
	return proxycookies.NamespacePrefix(ns)
}

// rewriteSetCookiesForProxy mutates Set-Cookie headers in h according to options.
// It preserves unknown attributes like Priority/Partitioned by detecting and appending them back.
func rewriteSetCookiesForProxy(h HeaderLike, opts CookieRewriteOptions) {
	proxycookies.RewriteSetCookies(h, opts)
}

// rewriteOutboundCookieHeaderForUpstream rewrites Cookie header for upstream in isolate mode
func rewriteOutboundCookieHeaderForUpstream(h HeaderLike, opts CookieRewriteOptions) {
	proxycookies.RewriteOutboundCookies(h, opts)
}

// HeaderLike is minimal interface to support http.Header and our wrappers in tests
type HeaderLike = proxycookies.Header

type parsedSetCookie struct {
	name        string
	value       string
	attrs       map[string]string // lower-case keys for recognized attributes
	flags       map[string]bool   // secure, httponly, partitioned
	extraTokens []string          // non-recognized tokens to preserve
}

func parseSetCookieLine(raw string) (parsedSetCookie, bool) {
	parsed, ok := proxycookies.ParseSetCookieLine(raw)
	if !ok {
		return parsedSetCookie{}, false
	}
	return parsedSetCookie{
		name:        parsed.Name,
		value:       parsed.Value,
		attrs:       parsed.Attrs,
		flags:       parsed.Flags,
		extraTokens: parsed.ExtraTokens,
	}, true
}

func buildSetCookieLine(p parsedSetCookie) string {
	return proxycookies.BuildSetCookieLine(proxycookies.ParsedSetCookie{
		Name:        p.name,
		Value:       p.value,
		Attrs:       p.attrs,
		Flags:       p.flags,
		ExtraTokens: p.extraTokens,
	})
}
