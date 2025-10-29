package httpapi

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    cfgpkg "network-debugger/internal/infrastructure/config"
    obs "network-debugger/internal/infrastructure/observability"
    uc "network-debugger/internal/usecase"
    "network-debugger/internal/domain"
    "context"
)

type apiStubRepo struct{}

// sessions
func (apiStubRepo) CreateSession(context.Context, domain.Session) error { return nil }
func (apiStubRepo) GetSession(context.Context, string) (domain.Session, bool, error) { return domain.Session{ID: "s1", StartedAt: time.Unix(0,0).UTC()}, true, nil }
func (apiStubRepo) DeleteSession(context.Context, string) error { return nil }
func (apiStubRepo) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) { return []domain.Session{{ID: "s1"},{ID: "s2"}}, 2, nil }
func (apiStubRepo) IncrementCounters(context.Context, string, domain.Frame) error { return nil }
func (apiStubRepo) SetClosed(context.Context, string, time.Time, *string) error { return nil }
func (apiStubRepo) ClearAllSessions(context.Context) error { return nil }
// frames
func (apiStubRepo) AppendFrame(context.Context, string, domain.Frame) error { return nil }
func (apiStubRepo) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) { return []domain.Frame{{ID: "f1"}}, "", nil }
// events
func (apiStubRepo) AppendEvent(context.Context, string, domain.Event) error { return nil }
func (apiStubRepo) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) { return []domain.Event{{ID: "e1", Name: "ev"}}, "", nil }
// http tx
func (apiStubRepo) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error { return nil }
func (apiStubRepo) ListHTTPTransactions(context.Context, string, string, int) ([]domain.HTTPTransaction, string, error) { return []domain.HTTPTransaction{{Method: "GET", URL: "u", Status: 200}}, "", nil }

func makeDepsWithAPIStub() *Deps {
    s := uc.NewSessionService(apiStubRepo{}, apiStubRepo{}, apiStubRepo{})
    return &Deps{Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
}

func TestHandleListSessions_GetAndDelete(t *testing.T) {
    d := makeDepsWithAPIStub()
    h := NewRouterWithoutForwardProxy(d)
    // GET
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=10", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("list get") }
    var body map[string]any
    _ = json.Unmarshal(rr.Body.Bytes(), &body)
    if int(body["total"].(float64)) != 2 { t.Fatalf("total") }
    // DELETE
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodDelete, "/api/sessions", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusNoContent { t.Fatalf("list delete") }
}

func TestV1ListSessions_Filters(t *testing.T) {
    d := makeDepsWithAPIStub()
    h := NewRouterWithoutForwardProxy(d)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?captures=all&includeUnassigned=1&limit=10&offset=0&q=x&_target=t", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("list v1 with filters") }
}

func TestHandleSessionByID_GetFramesEventsHTTP(t *testing.T) {
    d := makeDepsWithAPIStub()
    h := NewRouterWithoutForwardProxy(d)
    // session json
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/sessions/s1", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("get session") }

    // frames
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/api/sessions/s1/frames", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("frames") }
    // events
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/api/sessions/s1/events", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("events") }
    // http txs
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/api/sessions/s1/http", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("http txs") }
}

func TestHandleSessionByID_ExportGzipAndHar(t *testing.T) {
    d := makeDepsWithAPIStub()
    h := NewRouterWithoutForwardProxy(d)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/sessions/s1/export?gzip=1", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("export gzip") }
    if rr.Header().Get("Content-Encoding") != "gzip" { t.Fatalf("expect gzip header") }

    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/api/sessions/s1/har", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("har export") }

    // Accept-Encoding gzip path
    rr = httptest.NewRecorder()
    req = httptest.NewRequest(http.MethodGet, "/api/sessions/s1/export", nil)
    req.Header.Set("Accept-Encoding", "gzip")
    h.ServeHTTP(rr, req)
    if rr.Header().Get("Content-Encoding") != "gzip" { t.Fatalf("accept-encoding gzip path") }
}

func TestV1Session_BodyNotImplemented(t *testing.T) {
    d := makeDepsWithAPIStub()
    h := NewRouterWithoutForwardProxy(d)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/s1/body", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusNotFound { t.Fatalf("body should be 404 not_implemented") }
}

func TestLegacySession_UnknownSubresource_404(t *testing.T) {
    d := makeDepsWithAPIStub()
    h := NewRouterWithoutForwardProxy(d)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/sessions/s1/unknown", nil)
    h.ServeHTTP(rr, req)
    if rr.Code != http.StatusNotFound { t.Fatalf("legacy unknown subresource should be 404") }
}


