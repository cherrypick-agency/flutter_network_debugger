package httpapi

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

type testHijackableWriter struct {
	out *bytes.Buffer
}

func (w *testHijackableWriter) Header() http.Header         { return http.Header{} }
func (w *testHijackableWriter) Write(b []byte) (int, error) { return w.out.Write(b) }
func (w *testHijackableWriter) WriteHeader(statusCode int)  {}
func (w *testHijackableWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	c1, c2 := net.Pipe()
	rw := bufio.NewReadWriter(bufio.NewReader(c2), bufio.NewWriter(&testHijackRW{w.out}))
	return c1, rw, nil
}

type testHijackRW struct {
	buf *bytes.Buffer
}

func (h *testHijackRW) Read(p []byte) (int, error)  { return 0, nil }
func (h *testHijackRW) Write(p []byte) (int, error) { return h.buf.Write(p) }
func (h *testHijackRW) Flush()                      {}
