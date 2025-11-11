package httpapi

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func Test_mapToStringMap_and_toString(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"A": "str",
		"B": 123,
		"C": map[string]any{"x": 1},
	}
	out := mapToStringMap(in)
	if out["A"] != "str" {
		t.Fatalf("A mismatch: %q", out["A"])
	}
	if out["B"] != "123" {
		t.Fatalf("B mismatch: %q", out["B"])
	}
	// для объектов — json строка
	var m map[string]int
	if err := json.Unmarshal([]byte(out["C"]), &m); err != nil || m["x"] != 1 {
		t.Fatalf("C not json: %q err=%v", out["C"], err)
	}
}

func Test_parseCacheControl(t *testing.T) {
	t.Parallel()

	cc := "max-age=60, no-cache, stale-while-revalidate=30, private, quoted=\"abc\""
	got := parseCacheControl(cc)
	want := map[string]string{
		"max-age":                "60",
		"no-cache":               "true",
		"stale-while-revalidate": "30",
		"private":                "true",
		"quoted":                 "abc",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("cache-control parse mismatch:\nwant=%v\n got=%v", want, got)
	}
}

func Test_computeCacheMeta(t *testing.T) {
	t.Parallel()

	meta := computeCacheMeta(http.StatusOK, map[string]string{"Cache-Control": "max-age=1", "Age": "10", "ETag": "W/\"123\""})
	if meta.Status != "HIT" || meta.ETag != "W/\"123\"" || meta.Age != 10 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if meta.Directives["max-age"] != "1" {
		t.Fatalf("directives not parsed: %+v", meta.Directives)
	}

	meta = computeCacheMeta(http.StatusNotModified, map[string]string{"Cache-Control": "no-cache"})
	if meta.Status != "REVALIDATED" {
		t.Fatalf("not modified -> REVALIDATED, got: %s", meta.Status)
	}

	meta = computeCacheMeta(http.StatusOK, nil)
	if meta.Status != "UNKNOWN" {
		t.Fatalf("nil headers -> UNKNOWN, got: %s", meta.Status)
	}
}

func Test_computeCORSMeta_preflight_and_simple(t *testing.T) {
	t.Parallel()

	req := map[string]string{
		"Origin":                         "https://a.test",
		"Access-Control-Request-Method":  "POST",
		"Access-Control-Request-Headers": "X-Auth, X-Id",
	}
	resp := map[string]string{
		"Access-Control-Allow-Origin":  "https://a.test",
		"Access-Control-Allow-Methods": "GET, POST",
		"Access-Control-Allow-Headers": "X-Auth, x-id",
		"Vary":                         "Origin",
	}

	cors := computeCORSMeta("OPTIONS", req, resp, true)
	if !cors.Ok || cors.Reason != "" {
		t.Fatalf("preflight should be ok: %+v", cors)
	}

	// simple request path
	cors = computeCORSMeta("POST", map[string]string{"Origin": "https://a.test"}, resp, false)
	if !cors.Ok {
		t.Fatalf("simple should be ok: %+v", cors)
	}

	// origin mismatch
	cors = computeCORSMeta("POST", map[string]string{"Origin": "https://b.test"}, resp, false)
	if cors.Ok || cors.Reason != "origin" {
		t.Fatalf("origin mismatch should fail: %+v", cors)
	}
}

func Test_headerGetCaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string]any
		key     string
		wantVal string
		wantOk  bool
	}{
		{
			name:    "string value",
			headers: map[string]any{"Content-Type": "application/json"},
			key:     "content-type",
			wantVal: "application/json",
			wantOk:  true,
		},
		{
			name:    "case insensitive match",
			headers: map[string]any{"CONTENT-TYPE": "text/html"},
			key:     "content-type",
			wantVal: "text/html",
			wantOk:  true,
		},
		{
			name:    "array of any with string",
			headers: map[string]any{"Content-Type": []any{"application/json"}},
			key:     "content-type",
			wantVal: "application/json",
			wantOk:  true,
		},
		{
			name:    "array of any with non-string",
			headers: map[string]any{"Content-Type": []any{123}},
			key:     "content-type",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "array of any empty",
			headers: map[string]any{"Content-Type": []any{}},
			key:     "content-type",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "array of string",
			headers: map[string]any{"Content-Type": []string{"text/plain"}},
			key:     "content-type",
			wantVal: "text/plain",
			wantOk:  true,
		},
		{
			name:    "array of string empty",
			headers: map[string]any{"Content-Type": []string{}},
			key:     "content-type",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "exact match lowercase",
			headers: map[string]any{"content-type": "application/xml"},
			key:     "content-type",
			wantVal: "application/xml",
			wantOk:  true,
		},
		{
			name:    "exact match mixed case",
			headers: map[string]any{"Content-Type": "text/plain"},
			key:     "content-type",
			wantVal: "text/plain",
			wantOk:  true,
		},
		{
			name:    "exact match non-string value",
			headers: map[string]any{"content-type": 123},
			key:     "content-type",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "not found",
			headers: map[string]any{"Other-Header": "value"},
			key:     "content-type",
			wantVal: "",
			wantOk:  false,
		},
		{
			name:    "empty map",
			headers: map[string]any{},
			key:     "content-type",
			wantVal: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOk := headerGetCaseInsensitive(tt.headers, tt.key)
			if gotOk != tt.wantOk {
				t.Errorf("headerGetCaseInsensitive() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotVal != tt.wantVal {
				t.Errorf("headerGetCaseInsensitive() val = %q, want %q", gotVal, tt.wantVal)
			}
		})
	}
}

func Test_classifyNetError(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"context deadline exceeded": "TIMEOUT",
		"timeout":                   "TIMEOUT",
		"no such host":              "DNS",
		"server misbehaving":        "DNS",
		"x509: cert":                "TLS",
		"certificate":               "TLS",
		"tls":                       "TLS",
		"connection refused":        "CONNECT",
		"cannot assign":             "CONNECT",
		"reset by peer":             "RST",
		"unexpected EOF":            "EOF",
		"before full header":        "EOF",
		"early eof":                 "EOF",
		"request canceled":          "CANCEL",
		"client canceled":           "CANCEL",
		"something else":            "ERROR",
	}
	for msg, want := range cases {
		if got := classifyNetError(msg); got != want {
			t.Errorf("%q -> want %s got %s", msg, want, got)
		}
	}
}

func Test_fold_utils(t *testing.T) {
	t.Parallel()

	h := map[string]string{"Content-Type": "text/plain", "x-id": "42"}
	if v := getFold(h, "content-type"); v != "text/plain" {
		t.Fatalf("getFold failed: %q", v)
	}
	if !hasHeaderFold(h, "X-ID") {
		t.Fatalf("hasHeaderFold failed")
	}
	if hasHeaderFold(nil, "x") {
		t.Fatalf("nil map should be false")
	}

	if !containsFoldSlice([]string{"A", "b"}, "a") {
		t.Fatalf("containsFoldSlice should be true")
	}
	if containsFoldSlice([]string{"A", "b"}, "c") {
		t.Fatalf("containsFoldSlice should be false")
	}

	if !allAllowedFold([]string{"X-Auth", "X-Id"}, []string{"x-auth"}) {
		t.Fatalf("allAllowedFold single should be true")
	}
	if allAllowedFold([]string{"X-Auth"}, []string{"x-auth", "x-id"}) {
		t.Fatalf("allAllowedFold should be false when missing header")
	}

	if got := csvToSlice("A, b , ,C"); len(got) != 3 || got[1] != "b" {
		t.Fatalf("csvToSlice unexpected: %+v", got)
	}
	if csvToSlice("") != nil {
		t.Fatalf("csvToSlice empty -> nil expected")
	}
}
