package httpapi

import "testing"

func FuzzParseSetCookie(f *testing.F) {
	seeds := []string{
		"a=b; Secure; HttpOnly; SameSite=None; Priority=High; Partitioned",
		"n=v; Domain=.example.com; Path=/; Expires=Wed, 21 Oct 2015 07:28:00 GMT",
		"__Host-id=1; Path=/",
		"badtoken",
		"x=y; SameSite=Lax; Foo=Bar; Baz",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		p, ok := parseSetCookieLine(raw)
		if !ok {
			return
		}
		// проверяем, что build не паникует и результат снова парсится без паник
		out := buildSetCookieLine(p)
		if out == "" {
			return
		}
		_, _ = parseSetCookieLine(out)
	})
}
