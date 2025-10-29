package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLiveSessions_RegisterUnregisterCloseAll(t *testing.T) {
	ls := NewLiveSessions()
	ls.Register("", nil, nil) // no-op
	ls.Register("s1", nil, nil)
	ls.Unregister("") // no-op
	ls.Unregister("s1")
	ls.CloseAll() // no panic
}

func TestHandleWSSendText_Errors(t *testing.T) {
	d := &Deps{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/sessions/abc/ws/send", bytes.NewReader([]byte(`{"direction":"client->upstream","payload":"hi"}`)))
	d.handleWSSendText(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("no live -> 503 expected")
	}

	d.Live = NewLiveSessions()
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/sessions/abc/ws/send", bytes.NewReader([]byte(`{"direction":"client->upstream","payload":"hi"}`)))
	d.handleWSSendText(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("unregistered session -> 400 expected, got %d", rr.Code)
	}
}

func TestHandleV1MITMGenerate(t *testing.T) {
	d := &Deps{}
	rr := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{"cn": "dev-ca"})
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/mitm/ca/generate", bytes.NewReader(body))
	d.handleV1MITMGenerate(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("generate status: %d", rr.Code)
	}
	if d.MITM == nil || d.MITM.CA == nil {
		t.Fatalf("CA not set in deps")
	}
}
