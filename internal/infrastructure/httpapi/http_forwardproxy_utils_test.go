package httpapi

import (
    "bytes"
    "net"
    "testing"
    "net/http"
)

type dummyConn struct{ net.Conn }
func (dummyConn) Read(b []byte) (int, error) { copy(b, []byte("B")); return 1, nil }

func TestPrependConn_ReadsPrefixThenConn(t *testing.T) {
    pc := &prependConn{Conn: dummyConn{}, r: bytes.NewReader([]byte("A"))}
    buf := make([]byte, 2)
    n, err := pc.Read(buf)
    if err != nil || n != 1 || buf[0] != 'A' { t.Fatalf("first read A: n=%d err=%v %q", n, err, buf[:n]) }
    n, err = pc.Read(buf)
    if err != nil || n != 1 || buf[0] != 'B' { t.Fatalf("second read B: n=%d err=%v %q", n, err, buf[:n]) }
}

func TestIsAbsoluteURL(t *testing.T) {
    if !isAbsoluteURL("http://example.com/x") || isAbsoluteURL("/x") { t.Fatalf("isAbsoluteURL") }
}

func TestCloneHeader_DeepCopy(t *testing.T) {
    src := http.Header{"A": []string{"1", "2"}}
    cp := cloneHeader(src)
    src.Set("A", "x")
    if len(cp["A"]) != 2 || cp.Get("A") != "1" { t.Fatalf("clone deep copy failed: %+v", cp) }
}


