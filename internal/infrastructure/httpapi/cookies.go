package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/url"
	"strings"
)

// CookieMode defines behavior of cookie rewriting
const (
	CookieModeIsolate = "isolate"
	CookieModeAuto    = "auto"
	CookieModeOff     = "off"
)

// CookieRewriteOptions controls how Set-Cookie/Cookie are rewritten at the reverse proxy boundary
type CookieRewriteOptions struct {
	Mode            string // isolate|auto|off
	DomainStrategy  string // hostOnly|proxyHost
	PathStrategy    string // prefix|root
	ProxyHost       string // host of proxy (no port)
	ProxyPathPrefix string // "/httpproxy" or "/proxy"
	HTTPS           bool   // true if client→proxy is HTTPS
	Namespace       string // short hash namespace for isolate mode
}

func computeNamespaceFromURL(u *url.URL) string {
	// Stable namespace: scheme://host (including port if present)
	base := u.Scheme + "://" + u.Host
	sum := sha256.Sum256([]byte(base))
	// 12 hex chars are enough to avoid collisions in practice and keep cookie names short
	return hex.EncodeToString(sum[:])[:12]
}

func sanitizeHost(hostport string) string {
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		return h
	}
	// remove leading dot if any
	return strings.TrimPrefix(hostport, ".")
}

func domainAttrSafe(host string) bool {
	h := strings.ToLower(host)
	if h == "localhost" {
		return false
	}
	if net.ParseIP(h) != nil {
		return false
	}
	return true
}

func namespacePrefix(ns string) string {
	if ns == "" {
		return "gpx__"
	}
	return "gpx_" + ns + "__"
}

// rewriteSetCookiesForProxy mutates Set-Cookie headers in h according to options.
// It preserves unknown attributes like Priority/Partitioned by detecting and appending them back.
func rewriteSetCookiesForProxy(h HeaderLike, opts CookieRewriteOptions) {
	if opts.Mode == CookieModeOff {
		return
	}
	orig := h.Values("Set-Cookie")
	if len(orig) == 0 {
		return
	}
	h.Del("Set-Cookie")

	pfx := namespacePrefix(opts.Namespace)
	proxyPath := "/"
	if opts.PathStrategy == "prefix" && opts.ProxyPathPrefix != "" {
		proxyPath = opts.ProxyPathPrefix
	}

	for _, raw := range orig {
		p, ok := parseSetCookieLine(raw)
		if !ok {
			// failed to parse — leave as-is
			h.Add("Set-Cookie", raw)
			continue
		}

		origName := p.name
		isHostPrefix := strings.HasPrefix(origName, "__Host-")
		isSecurePrefix := strings.HasPrefix(origName, "__Secure-")

		// Isolation by renaming cookie name in the browser storage
		if opts.Mode == CookieModeIsolate {
			p.name = pfx + origName
		}

		// Domain strategy
		if isHostPrefix {
			delete(p.attrs, "domain")
		} else if opts.DomainStrategy == "hostOnly" {
			delete(p.attrs, "domain")
		} else if opts.DomainStrategy == "proxyHost" {
			dh := sanitizeHost(opts.ProxyHost)
			if domainAttrSafe(dh) {
				p.attrs["domain"] = dh
			} else {
				delete(p.attrs, "domain")
			}
		}

		// Path strategy
		if isHostPrefix {
			p.attrs["path"] = "/"
		} else if opts.PathStrategy == "prefix" {
			p.attrs["path"] = proxyPath
		} else {
			p.attrs["path"] = "/"
		}

		// Secure handling
		if opts.HTTPS {
			if strings.EqualFold(p.attrs["samesite"], "none") || isSecurePrefix || isHostPrefix {
				p.flags["secure"] = true
			}
		}

		out := buildSetCookieLine(p)
		h.Add("Set-Cookie", out)
	}
}

// rewriteOutboundCookieHeaderForUpstream rewrites Cookie header for upstream in isolate mode
func rewriteOutboundCookieHeaderForUpstream(h HeaderLike, opts CookieRewriteOptions) {
	if opts.Mode != CookieModeIsolate {
		return
	}
	cookies := h.Values("Cookie")
	if len(cookies) == 0 {
		return
	}
	// Merge into one logical header and parse pairs
	combined := strings.Join(cookies, "; ")
	pairs := strings.Split(combined, ";")
	pfx := namespacePrefix(opts.Namespace)
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.IndexByte(p, '=')
		if eq <= 0 {
			continue
		}
		name := strings.TrimSpace(p[:eq])
		val := strings.TrimSpace(p[eq+1:])
		if strings.HasPrefix(name, pfx) {
			origName := strings.TrimPrefix(name, pfx)
			// only cookies from current namespace will be sent upstream
			out = append(out, origName+"="+val)
		}
	}
	h.Del("Cookie")
	if len(out) > 0 {
		h.Set("Cookie", strings.Join(out, "; "))
	}
}

// HeaderLike is minimal interface to support http.Header and our wrappers in tests
type HeaderLike interface {
	Values(key string) []string
	Set(key, value string)
	Add(key, value string)
	Del(key string)
}

type parsedSetCookie struct {
	name        string
	value       string
	attrs       map[string]string // lower-case keys for recognized attributes
	flags       map[string]bool   // secure, httponly, partitioned
	extraTokens []string          // non-recognized tokens to preserve
}

func parseSetCookieLine(raw string) (parsedSetCookie, bool) {
	parts := strings.Split(raw, ";")
	if len(parts) == 0 {
		return parsedSetCookie{}, false
	}
	first := strings.TrimSpace(parts[0])
	eq := strings.Index(first, "=")
	if eq <= 0 {
		return parsedSetCookie{}, false
	}
	p := parsedSetCookie{
		name:  strings.TrimSpace(first[:eq]),
		value: strings.TrimSpace(first[eq+1:]),
		attrs: map[string]string{},
		flags: map[string]bool{},
	}
	for i := 1; i < len(parts); i++ {
		tok := strings.TrimSpace(parts[i])
		if tok == "" {
			continue
		}
		lower := strings.ToLower(tok)
		switch lower {
		case "secure":
			p.flags["secure"] = true
			continue
		case "httponly":
			p.flags["httponly"] = true
			continue
		case "partitioned":
			p.flags["partitioned"] = true
			continue
		}
		if kv := strings.Index(tok, "="); kv > 0 {
			key := strings.ToLower(strings.TrimSpace(tok[:kv]))
			val := strings.TrimSpace(tok[kv+1:])
			switch key {
			case "domain", "path", "expires", "max-age", "samesite", "priority":
				p.attrs[key] = val
			default:
				p.extraTokens = append(p.extraTokens, tok)
			}
		} else {
			p.extraTokens = append(p.extraTokens, tok)
		}
	}
	return p, true
}

func buildSetCookieLine(p parsedSetCookie) string {
	var b strings.Builder
	b.WriteString(p.name)
	b.WriteByte('=')
	b.WriteString(p.value)
	if v := p.attrs["domain"]; v != "" {
		b.WriteString("; Domain=")
		b.WriteString(v)
	}
	if v := p.attrs["path"]; v != "" {
		b.WriteString("; Path=")
		b.WriteString(v)
	}
	if v := p.attrs["expires"]; v != "" {
		b.WriteString("; Expires=")
		b.WriteString(v)
	}
	if v := p.attrs["max-age"]; v != "" {
		b.WriteString("; Max-Age=")
		b.WriteString(v)
	}
	if p.flags["httponly"] {
		b.WriteString("; HttpOnly")
	}
	if p.flags["secure"] {
		b.WriteString("; Secure")
	}
	if v := p.attrs["samesite"]; v != "" {
		// normalize capitalization
		lv := strings.ToLower(v)
		switch lv {
		case "lax":
			b.WriteString("; SameSite=Lax")
		case "strict":
			b.WriteString("; SameSite=Strict")
		case "none":
			b.WriteString("; SameSite=None")
		default:
			b.WriteString("; SameSite=")
			b.WriteString(v)
		}
	}
	if v := p.attrs["priority"]; v != "" {
		b.WriteString("; Priority=")
		b.WriteString(v)
	}
	if p.flags["partitioned"] {
		b.WriteString("; Partitioned")
	}
	for _, t := range p.extraTokens {
		b.WriteString("; ")
		b.WriteString(t)
	}
	return b.String()
}
