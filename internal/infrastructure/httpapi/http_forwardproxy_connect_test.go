package httpapi

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type hijackRW struct {
	buf *bytes.Buffer
}

func (h *hijackRW) Read(p []byte) (int, error)  { return 0, nil }
func (h *hijackRW) Write(p []byte) (int, error) { return h.buf.Write(p) }
func (h *hijackRW) Flush()                      {}

// fake ResponseWriter that supports Hijacker
type hijackableWriter struct{ out *bytes.Buffer }

func (w *hijackableWriter) Header() http.Header         { return http.Header{} }
func (w *hijackableWriter) Write(b []byte) (int, error) { return w.out.Write(b) }
func (w *hijackableWriter) WriteHeader(statusCode int)  {}
func (w *hijackableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// Return a non-nil conn and a bufio RW that writes into buffer
	c1, _ := net.Pipe()
	rw := bufio.NewReadWriter(bufio.NewReader(&hijackRW{w.out}), bufio.NewWriter(&hijackRW{w.out}))
	return c1, rw, nil
}

func TestHandleConnectTunnel_Error502(t *testing.T) {
	d := &Deps{}
	w := &hijackableWriter{out: &bytes.Buffer{}}
	req := httptest.NewRequest(http.MethodConnect, "/", nil)
	// invalid/unreachable upstream -> 502
	req.Host = "127.0.0.1:1"
	d.handleConnectTunnel(w, req)
	if !bytes.Contains(w.out.Bytes(), []byte("502 Bad Gateway")) {
		t.Fatalf("expected 502 in raw response: %q", w.out.String())
	}
}
