package httpapi

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
)

func Test_copyHeaderIfPresent(t *testing.T) {
	t.Parallel()

	src := http.Header{}
	dst := http.Header{}
	src.Set("X-Test", "v")
	copyHeaderIfPresent(&dst, src, "X-Test")
	if dst.Get("X-Test") != "v" {
		t.Fatalf("expected header copied")
	}
	copyHeaderIfPresent(&dst, src, "X-None")
	if dst.Get("X-None") != "" {
		t.Fatalf("should not set absent header")
	}
}

func Test_clientHost(t *testing.T) {
	t.Parallel()

	if clientHost("1.2.3.4:5678") != "1.2.3.4" {
		t.Fatalf("strip port failed")
	}
	if clientHost("example.com") != "example.com" {
		t.Fatalf("no port stays same")
	}
}

func Test_tryExtractAckID(t *testing.T) {
	t.Parallel()

	cases := map[string]int64{
		"42/chat,17[\"ev\",{}]": 17,
		"43,5[1,2]":             5,
		"45/chat,10[\"ev\",{}]": 10,
		"46/chat,99[1]":         99,
		"42[\"ev\",{}]":         -1,
		"4":                     -1,
		"42/chat,abc[1]":        -1,
		"42/chat,17x[1]":        -1,
		"42":                    -1, // exactly 2 chars
		"":                      -1, // empty string
		"1":                     -1, // single char
		"42/chat":               -1, // namespace without ack
		"43,":                   -1, // comma but no ack
		"42/chat,":              -1, // namespace with comma but no ack
	}
	for s, want := range cases {
		if got := tryExtractAckID(s); got != want {
			t.Errorf("%q -> want %d got %d", s, want, got)
		}
	}
}

func Test_shouldSuppressWSCloseError(t *testing.T) {
	t.Parallel()

	if !shouldSuppressWSCloseError(&websocket.CloseError{Code: websocket.CloseNormalClosure, Text: "done"}) {
		t.Fatalf("normal close should be suppressed")
	}
	if !shouldSuppressWSCloseError(&websocket.CloseError{Code: websocket.CloseNoStatusReceived}) {
		t.Fatalf("no-status close should be suppressed")
	}
	if shouldSuppressWSCloseError(&websocket.CloseError{Code: websocket.CloseAbnormalClosure}) {
		t.Fatalf("abnormal close should not be suppressed")
	}
	if shouldSuppressWSCloseError(errors.New("boom")) {
		t.Fatalf("generic error should not be suppressed")
	}
}
