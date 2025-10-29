package httpapi

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"
)

type fakeRec struct{ bytes.Buffer }
func (f *fakeRec) Header() http.Header { return http.Header{} }
func (f *fakeRec) Write(b []byte) (int, error) { return f.Buffer.Write(b) }
func (f *fakeRec) WriteHeader(statusCode int) {}
func (f *fakeRec) Flush() {}

func TestWriteSSE(t *testing.T) {
    fr := &fakeRec{}
    enc := json.NewEncoder(fr)
    if err := writeSSE(fr, fr, "ev", map[string]int{"x":1}, enc); err != nil { t.Fatalf("writeSSE: %v", err) }
    s := fr.String()
    if !strContains(s, "event: ev") || !strContains(s, "\n\n") { t.Fatalf("bad sse format: %q", s) }
}


