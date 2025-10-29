package httpapi

import (
	"net/url"
	"regexp"
	"testing"
)

func TestCookies_Utils(t *testing.T) {
	t.Parallel()

	// computeNamespaceFromURL
	{
		re := regexp.MustCompile(`^[0-9a-f]{12}$`)
		u1, _ := url.Parse("http://example.com")
		u2, _ := url.Parse("https://example.com")
		u3, _ := url.Parse("http://example.com:8080")

		n1 := computeNamespaceFromURL(u1)
		n1b := computeNamespaceFromURL(u1)
		n2 := computeNamespaceFromURL(u2)
		n3 := computeNamespaceFromURL(u3)

		if !re.MatchString(n1) || !re.MatchString(n2) || !re.MatchString(n3) {
			t.Fatalf("namespace должен быть 12-значным hex: %q %q %q", n1, n2, n3)
		}
		if n1 != n1b {
			t.Fatalf("namespace должен быть детерминированным: %q vs %q", n1, n1b)
		}
		if n1 == n2 {
			t.Fatalf("разные scheme должны давать разные namespace: %q == %q", n1, n2)
		}
		if n1 == n3 {
			t.Fatalf("разные host:port должны давать разные namespace: %q == %q", n1, n3)
		}
	}

	// sanitizeHost
	{
		cases := []struct{
			in  string
			out string
		}{
			{"example.com:443", "example.com"},
			{".example.com", "example.com"},
			{"[::1]:8080", "::1"},
			{"localhost:123", "localhost"},
		}
		for _, c := range cases {
			if got := sanitizeHost(c.in); got != c.out {
				t.Fatalf("sanitizeHost(%q)=%q, ожидалось %q", c.in, got, c.out)
			}
		}
	}

	// domainAttrSafe
	{
		cases := []struct{
			in   string
			want bool
		}{
			{"localhost", false},
			{"127.0.0.1", false},
			{"::1", false},
			{"Example.com", true},
		}
		for _, c := range cases {
			if got := domainAttrSafe(c.in); got != c.want {
				t.Fatalf("domainAttrSafe(%q)=%v, ожидалось %v", c.in, got, c.want)
			}
		}
	}

	// namespacePrefix
	{
		if got := namespacePrefix(""); got != "gpx__" {
			t.Fatalf("namespacePrefix(\"\")=%q, ожидалось %q", got, "gpx__")
		}
		if got := namespacePrefix("abcdef"); got != "gpx_abcdef__" {
			t.Fatalf("namespacePrefix(\"abcdef\")=%q, ожидалось %q", got, "gpx_abcdef__")
		}
	}
}
