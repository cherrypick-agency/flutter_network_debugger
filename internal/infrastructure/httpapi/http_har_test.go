package httpapi

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "network-debugger/internal/domain"
    uc "network-debugger/internal/usecase"
    "context"
)

type stubRepo struct{}

// SessionRepository
func (stubRepo) CreateSession(context.Context, domain.Session) error { return nil }
func (stubRepo) GetSession(context.Context, string) (domain.Session, bool, error) { return domain.Session{}, false, nil }
func (stubRepo) DeleteSession(context.Context, string) error { return nil }
func (stubRepo) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) { return nil, 0, nil }
func (stubRepo) IncrementCounters(context.Context, string, domain.Frame) error { return nil }
func (stubRepo) SetClosed(context.Context, string, time.Time, *string) error { return nil }
func (stubRepo) ClearAllSessions(context.Context) error { return nil }
// FrameRepository
func (stubRepo) AppendFrame(context.Context, string, domain.Frame) error { return nil }
func (stubRepo) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) { return nil, "", nil }
// EventRepository
func (stubRepo) AppendEvent(context.Context, string, domain.Event) error { return nil }
func (stubRepo) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) { return nil, "", nil }
// HTTPTransactionRepository
func (stubRepo) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error { return nil }
func (stubRepo) ListHTTPTransactions(context.Context, string, string, int) ([]domain.HTTPTransaction, string, error) {
    return []domain.HTTPTransaction{
        {Method: "GET", URL: "http://x/", Status: 200, ReqSize: 1, RespSize: 2, StartedAt: time.Unix(0,0).UTC(), Timings: domain.HTTPTimings{Total: 10}},
    }, "", nil
}

func TestExportHARForSession_Basic(t *testing.T) {
    s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
    d := &Deps{Svc: s}
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/_api/v1/har", nil)
    exportHARForSession(rr, req, d, "sess1")
    if rr.Code != http.StatusOK { t.Fatalf("status") }
    var body map[string]any
    _ = json.Unmarshal(rr.Body.Bytes(), &body)
    log := body["log"].(map[string]any)
    if int(log["version"].(string)[0]) == 0 { t.Fatalf("version missing") }
    entries := log["entries"].([]any)
    if len(entries) != 1 { t.Fatalf("entries") }
}


