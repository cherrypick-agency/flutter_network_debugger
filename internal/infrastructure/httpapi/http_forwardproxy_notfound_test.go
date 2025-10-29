package httpapi

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandleForwardOrNotFound_404(t *testing.T) {
    d := &Deps{}
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/nope", nil)
    d.handleForwardOrNotFound(rr, req)
    if rr.Code != http.StatusNotFound {
        t.Fatalf("expected 404, got %d", rr.Code)
    }
    if !strContains(rr.Body.String(), "NOT_FOUND") {
        t.Fatalf("expected NOT_FOUND code: %s", rr.Body.String())
    }
}


