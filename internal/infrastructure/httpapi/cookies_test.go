package httpapi

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
)

// Тесты придерживаются AAA: Arrange-Act-Assert, F.I.R.S.T.: быстрые, независимые,
// повторяемые, само-утверждающиеся и своевременные.

func Test_computeNamespaceFromURL_variants(t *testing.T) {
	t.Parallel()

	mustURL := func(raw string) *url.URL {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("parse url: %v", err)
		}
		return u
	}

	u1 := mustURL("https://example.com")
	u2 := mustURL("http://example.com")
	u3 := mustURL("https://example.com:443")

	ns1 := computeNamespaceFromURL(u1)
	ns2 := computeNamespaceFromURL(u2)
	ns3 := computeNamespaceFromURL(u3)

	if len(ns1) != 12 || len(ns2) != 12 || len(ns3) != 12 {
		t.Fatalf("namespace length must be 12: %q %q %q", ns1, ns2, ns3)
	}
	if ns1 == ns2 {
		t.Errorf("scheme must affect namespace: %q == %q", ns1, ns2)
	}
	if ns1 == ns3 {
		t.Errorf("host port must affect namespace: %q == %q", ns1, ns3)
	}
}

func Test_sanitizeHost(t *testing.T) {
	t.Parallel()

	if got := sanitizeHost("example.com:8080"); got != "example.com" {
		t.Errorf("sanitizeHost: want example.com, got %q", got)
	}
	if got := sanitizeHost(".example.com"); got != "example.com" {
		t.Errorf("sanitizeHost: want example.com, got %q", got)
	}
	if got := sanitizeHost("example.com"); got != "example.com" {
		t.Errorf("sanitizeHost: want example.com, got %q", got)
	}
}

func Test_domainAttrSafe(t *testing.T) {
	t.Parallel()

	cases := map[string]bool{
		"localhost":     false,
		"127.0.0.1":     false,
		"example.com":   true,
		"sub.domain.io": true,
	}
	for host, want := range cases {
		if got := domainAttrSafe(host); got != want {
			t.Errorf("domainAttrSafe(%q): want %v, got %v", host, want, got)
		}
	}
}

func Test_namespacePrefix(t *testing.T) {
	t.Parallel()

	if got := namespacePrefix(""); got != "gpx__" {
		t.Errorf("namespacePrefix empty: want gpx__, got %q", got)
	}
	if got := namespacePrefix("abcd"); got != "gpx_abcd__" {
		t.Errorf("namespacePrefix abcd: want gpx_abcd__, got %q", got)
	}
}

func Test_parseSetCookieLine_and_buildSetCookieLine_roundtrip(t *testing.T) {
	t.Parallel()

	raw := strings.Join([]string{
		"N=V",
		"Domain=.example.com",
		"Path=/old",
		"Secure",
		"SameSite=none",
		"httponly",
		"Priority=Medium",
		"Partitioned",
		"Foo=Bar",
		"Expires=Wed, 21 Oct 2015 07:28:00 GMT",
		"Max-Age=3600",
	}, "; ")

	p, ok := parseSetCookieLine(raw)
	if !ok {
		t.Fatalf("parseSetCookieLine: failed")
	}

	if p.name != "N" || p.value != "V" {
		t.Fatalf("parsed name/value mismatch: %q=%q", p.name, p.value)
	}
	if !p.flags["secure"] || !p.flags["httponly"] || !p.flags["partitioned"] {
		t.Errorf("flags not parsed correctly: %+v", p.flags)
	}
	if p.attrs["path"] != "/old" || p.attrs["domain"] != ".example.com" {
		t.Errorf("attrs path/domain mismatch: %+v", p.attrs)
	}
	if p.attrs["samesite"] != "none" && p.attrs["samesite"] != "None" {
		t.Errorf("samesite not parsed: %q", p.attrs["samesite"])
	}
	if len(p.extraTokens) == 0 || p.extraTokens[0] != "Foo=Bar" {
		t.Errorf("extra tokens preserved order expected Foo=Bar first, got: %+v", p.extraTokens)
	}

	got := buildSetCookieLine(p)
	// Ожидаемый порядок атрибутов соответствует реализации buildSetCookieLine
	want := strings.Join([]string{
		"N=V",
		"Domain=.example.com",
		"Path=/old",
		"Expires=Wed, 21 Oct 2015 07:28:00 GMT",
		"Max-Age=3600",
		"HttpOnly",
		"Secure",
		"SameSite=None",
		"Priority=Medium",
		"Partitioned",
		"Foo=Bar",
	}, "; ")

	if got != want {
		t.Fatalf("buildSetCookieLine:\nwant: %s\n got: %s", want, got)
	}
}

func Test_rewriteSetCookiesForProxy_isolate_proxyHost_prefix_https(t *testing.T) {
	t.Parallel()

	h := http.Header{}
	h.Add("Set-Cookie", "id=1; Path=/; Domain=example.com; SameSite=None")
	h.Add("Set-Cookie", "__Host-strict=1; Path=/; Secure; HttpOnly")

	opts := CookieRewriteOptions{
		Mode:            CookieModeIsolate,
		DomainStrategy:  "proxyHost",
		PathStrategy:    "prefix",
		ProxyHost:       "proxy.local:8443",
		ProxyPathPrefix: "/proxy",
		HTTPS:           true,
		Namespace:       "abcd",
	}

	rewriteSetCookiesForProxy(h, opts)

	got := h.Values("Set-Cookie")
	sort.Strings(got)

	want := []string{
		"gpx_abcd__id=1; Domain=proxy.local; Path=/proxy; Secure; SameSite=None",
		"gpx_abcd____Host-strict=1; Path=/; HttpOnly; Secure",
	}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("Set-Cookie count mismatch: want %d got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Set-Cookie[%d]:\nwant: %s\n got: %s", i, want[i], got[i])
		}
	}
}

func Test_rewriteSetCookiesForProxy_modeOff_noChange(t *testing.T) {
	t.Parallel()

	h := http.Header{}
	original := []string{
		"a=1; Path=/",
		"b=2; Path=/; HttpOnly",
	}
	for _, s := range original {
		h.Add("Set-Cookie", s)
	}

	rewriteSetCookiesForProxy(h, CookieRewriteOptions{Mode: CookieModeOff})
	got := h.Values("Set-Cookie")
	if len(got) != len(original) {
		t.Fatalf("should keep original count: want %d got %d", len(original), len(got))
	}
	for i, s := range original {
		if got[i] != s {
			t.Errorf("value mismatch at %d: want %q got %q", i, s, got[i])
		}
	}
}

func Test_rewriteOutboundCookieHeaderForUpstream_isolate_filtersByNamespace(t *testing.T) {
	t.Parallel()

	h := http.Header{}
	h.Add("Cookie", "gpx_abcd__a=1; other=2")
	h.Add("Cookie", "gpx_efgh__b=3; gpx_abcd__c=4")

	opts := CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: "abcd"}
	rewriteOutboundCookieHeaderForUpstream(h, opts)

	gotVals := h.Values("Cookie")
	if len(gotVals) != 1 {
		t.Fatalf("exactly one Cookie header expected, got %d", len(gotVals))
	}
	got := gotVals[0]
	// Порядок сохраняется как при обходе пар (a=1 затем c=4)
	if got != "a=1; c=4" {
		t.Fatalf("Cookie rewrite expected 'a=1; c=4', got %q", got)
	}
}

func Test_rewriteOutboundCookieHeaderForUpstream_nonIsolate_noChange(t *testing.T) {
	t.Parallel()

	h := http.Header{}
	h.Add("Cookie", "x=1")
	rewriteOutboundCookieHeaderForUpstream(h, CookieRewriteOptions{Mode: CookieModeAuto})

	got := h.Values("Cookie")
	if len(got) != 1 || got[0] != "x=1" {
		t.Fatalf("non-isolate should not change Cookie, got: %+v", got)
	}
}

// (продолжение тестов) — используем те же импорты, что и выше

type headerAdapter http.Header

func (h headerAdapter) Values(key string) []string { return http.Header(h).Values(key) }
func (h headerAdapter) Set(key, value string)      { http.Header(h).Set(key, value) }
func (h headerAdapter) Add(key, value string)      { http.Header(h).Add(key, value) }
func (h headerAdapter) Del(key string)             { http.Header(h).Del(key) }

func TestParseAndBuildSetCookie_Roundtrip(t *testing.T) {
	raw := "sess=abc; Path=/; Domain=api.example.com; HttpOnly; Secure; SameSite=Lax; Priority=High; Partitioned"
	p, ok := parseSetCookieLine(raw)
	if !ok {
		t.Fatalf("failed to parse")
	}
	out := buildSetCookieLine(p)
	if !strings.Contains(out, "sess=abc") || !strings.Contains(out, "; Path=/") || !strings.Contains(out, "; HttpOnly") {
		t.Fatalf("unexpected roundtrip: %s", out)
	}
	if !strings.Contains(out, "; Priority=High") || !strings.Contains(out, "; Partitioned") {
		t.Fatalf("lost attributes: %s", out)
	}
}

func TestRewriteSetCookie_IsolateAndPrefix(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "token=val; Path=/; Domain=api.example.com; HttpOnly")
	u, _ := url.Parse("https://api.example.com")
	ns := computeNamespaceFromURL(u)
	opts := CookieRewriteOptions{
		Mode:            CookieModeIsolate,
		DomainStrategy:  "hostOnly",
		PathStrategy:    "prefix",
		ProxyHost:       "proxy.local",
		ProxyPathPrefix: "/httpproxy",
		HTTPS:           true,
		Namespace:       ns,
	}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Values("Set-Cookie")
	if len(out) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(out))
	}
	if !strings.HasPrefix(out[0], "gpx_"+ns+"__token=") {
		t.Fatalf("name not isolated: %s", out[0])
	}
	if !strings.Contains(out[0], "; Path=/httpproxy") {
		t.Fatalf("path not rewritten: %s", out[0])
	}
	if strings.Contains(strings.ToLower(out[0]), "; domain=") {
		t.Fatalf("domain should be host-only: %s", out[0])
	}
}

func TestRewriteOutboundCookie_IsolateFiltersAndUnwraps(t *testing.T) {
	h := http.Header{}
	nsA := "aaaaaaaaaaaa"
	nsB := "bbbbbbbbbbbb"
	h.Add("Cookie", "gpx_"+nsA+"__token=1; gpx_"+nsB+"__token=2; other=3")
	opts := CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: nsA}
	rewriteOutboundCookieHeaderForUpstream(headerAdapter(h), opts)
	got := h.Get("Cookie")
	if got != "token=1" {
		t.Fatalf("unexpected cookie forwarded: %q", got)
	}
}

func TestRewriteSetCookie_SameSiteNoneAddsSecureOnHTTPS(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "sid=xyz; Path=/; SameSite=None")
	opts := CookieRewriteOptions{
		Mode:            CookieModeAuto,
		DomainStrategy:  "hostOnly",
		PathStrategy:    "prefix",
		ProxyHost:       "proxy.local",
		ProxyPathPrefix: "/httpproxy",
		HTTPS:           true,
	}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if !strings.Contains(out, "; Secure") {
		t.Fatalf("expected Secure when SameSite=None on HTTPS: %s", out)
	}
}

func TestRewriteSetCookie_SameSiteNoneNoSecureOnHTTP(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "sid=xyz; Path=/; SameSite=None")
	opts := CookieRewriteOptions{
		Mode:            CookieModeAuto,
		DomainStrategy:  "hostOnly",
		PathStrategy:    "prefix",
		ProxyHost:       "proxy.local",
		ProxyPathPrefix: "/httpproxy",
		HTTPS:           false,
	}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if strings.Contains(out, "; Secure") {
		t.Fatalf("must not add Secure on HTTP: %s", out)
	}
}

func TestRewriteSetCookie_HostPrefixRules(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "__Host-sid=abc; Path=/; SameSite=Lax")
	opts := CookieRewriteOptions{
		Mode:            CookieModeAuto,
		DomainStrategy:  "proxyHost",
		PathStrategy:    "root",
		ProxyHost:       "proxy.example.com",
		ProxyPathPrefix: "/httpproxy",
		HTTPS:           true,
	}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if !strings.HasPrefix(out, "__Host-sid=") {
		t.Fatalf("__Host- name must remain: %s", out)
	}
	if !strings.Contains(out, "; Path=/") {
		t.Fatalf("__Host- requires Path=/ : %s", out)
	}
	if strings.Contains(strings.ToLower(out), "domain=") {
		t.Fatalf("__Host- forbids Domain: %s", out)
	}
	if !strings.Contains(out, "; Secure") {
		t.Fatalf("__Host- should be Secure on HTTPS: %s", out)
	}
}

func TestRewriteSetCookie_SecurePrefixOnHTTPS(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "__Secure-id=1; Path=/")
	opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "prefix", HTTPS: true, ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if !strings.Contains(out, "; Secure") {
		t.Fatalf("__Secure- must be Secure on HTTPS: %s", out)
	}
}

func TestRewriteSetCookie_ProxyHostDomainFallbackLocalhostAndIP(t *testing.T) {
	// localhost
	h1 := http.Header{}
	h1.Add("Set-Cookie", "a=1")
	opts1 := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "proxyHost", PathStrategy: "prefix", ProxyHost: "localhost:9091", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h1), opts1)
	if strings.Contains(strings.ToLower(h1.Get("Set-Cookie")), "domain=") {
		t.Fatalf("domain must be omitted for localhost")
	}
	// IP
	h2 := http.Header{}
	h2.Add("Set-Cookie", "b=1")
	opts2 := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "proxyHost", PathStrategy: "prefix", ProxyHost: "127.0.0.1", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h2), opts2)
	if strings.Contains(strings.ToLower(h2.Get("Set-Cookie")), "domain=") {
		t.Fatalf("domain must be omitted for IP")
	}
	// normal host
	h3 := http.Header{}
	h3.Add("Set-Cookie", "c=1")
	opts3 := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "proxyHost", PathStrategy: "prefix", ProxyHost: "proxy.example.com", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h3), opts3)
	if !strings.Contains(h3.Get("Set-Cookie"), "; Domain=proxy.example.com") {
		t.Fatalf("domain must be set for normal host: %s", h3.Get("Set-Cookie"))
	}
}

func TestRewriteSetCookie_MultipleCookiesHandled(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "a=1; Path=/")
	h.Add("Set-Cookie", "b=2; Path=/; Priority=High")
	opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "prefix", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	vals := h.Values("Set-Cookie")
	if len(vals) != 2 {
		t.Fatalf("expected 2 cookies after rewrite, got %d", len(vals))
	}
	if !strings.Contains(vals[0], "; Path=/httpproxy") || !strings.Contains(vals[1], "; Path=/httpproxy") {
		t.Fatalf("path not rewritten for both: %+v", vals)
	}
	if !strings.Contains(vals[1], "; Priority=") {
		t.Fatalf("priority must be preserved: %s", vals[1])
	}
}

func TestRewriteSetCookie_PreserveExpiresWithComma(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "sess=1; Expires=Wed, 21 Oct 2015 07:28:00 GMT; Path=/")
	opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "prefix", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if !strings.Contains(out, "Expires=Wed, 21 Oct 2015 07:28:00 GMT") {
		t.Fatalf("expires must be preserved intact: %s", out)
	}
}

func TestRewriteSetCookie_HostOnlyRemovesDomainFromUpstream(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "a=1; Domain=api.example.com; Path=/")
	opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "prefix", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := strings.ToLower(h.Get("Set-Cookie"))
	if strings.Contains(out, "domain=") {
		t.Fatalf("hostOnly must strip Domain: %s", out)
	}
}

func TestRewriteSetCookie_PathRootStrategy(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "a=1; Path=/api")
	opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "root", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if !strings.Contains(out, "; Path=/") || strings.Contains(out, "; Path=/httpproxy") {
		t.Fatalf("root strategy should set Path=/ : %s", out)
	}
}

func TestRewriteOutboundCookie_MultipleHeadersJoinAndFilter(t *testing.T) {
	h := http.Header{}
	ns := "zzzzzzzzzzzz"
	pfx := "gpx_" + ns + "__"
	h.Add("Cookie", pfx+"a=1")
	h.Add("Cookie", "b=2; ")
	opts := CookieRewriteOptions{Mode: CookieModeIsolate, Namespace: ns}
	rewriteOutboundCookieHeaderForUpstream(headerAdapter(h), opts)
	got := h.Get("Cookie")
	if got != "a=1" {
		t.Fatalf("should unwrap only namespaced cookie and drop others: %q", got)
	}
}

func TestRewriteSetCookie_SameSite_CaseInsensitiveNormalization(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{
		{"a=1; SameSite=NoNe", "SameSite=None"},
		{"b=1; SameSite=STRICT", "SameSite=Strict"},
		{"c=1; SameSite=LAX", "SameSite=Lax"},
	}
	for _, tc := range cases {
		h := http.Header{}
		h.Add("Set-Cookie", tc.in)
		opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "prefix", ProxyPathPrefix: "/httpproxy"}
		rewriteSetCookiesForProxy(headerAdapter(h), opts)
		out := h.Get("Set-Cookie")
		if !strings.Contains(out, tc.exp) {
			t.Fatalf("expected normalized %q in %q", tc.exp, out)
		}
	}
}

func TestRewriteSetCookie_PreserveUnknownAttributesAtEnd(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "a=1; Foo=Bar; Path=/")
	opts := CookieRewriteOptions{Mode: CookieModeAuto, DomainStrategy: "hostOnly", PathStrategy: "prefix", ProxyPathPrefix: "/httpproxy"}
	rewriteSetCookiesForProxy(headerAdapter(h), opts)
	out := h.Get("Set-Cookie")
	if !strings.Contains(out, "; Foo=Bar") {
		t.Fatalf("unknown attribute must be preserved: %s", out)
	}
	// должен быть в конце (после стандартных атрибутов)
	if !strings.HasSuffix(out, "; Foo=Bar") {
		t.Fatalf("unknown attribute should be appended at the end: %s", out)
	}
}
