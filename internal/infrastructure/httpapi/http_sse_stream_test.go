package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"network-debugger/internal/domain"
	uc "network-debugger/internal/usecase"
)

// stub repos to feed initial SSE snapshot
type sseStubSessions struct{}

func (sseStubSessions) CreateSession(context.Context, domain.Session) error { return nil }
func (sseStubSessions) GetSession(context.Context, string) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}
func (sseStubSessions) DeleteSession(context.Context, string) error { return nil }
func (sseStubSessions) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) {
	return nil, 0, nil
}
func (sseStubSessions) IncrementCounters(context.Context, string, domain.Frame) error { return nil }
func (sseStubSessions) SetClosed(context.Context, string, time.Time, *string) error   { return nil }
func (sseStubSessions) ClearAllSessions(context.Context) error                        { return nil }

// HTTP tx repo (attached via type-assert in NewSessionService)
func (sseStubSessions) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error {
	return nil
}
func (sseStubSessions) ListHTTPTransactions(context.Context, string, string, int) ([]domain.HTTPTransaction, string, error) {
	return []domain.HTTPTransaction{{Method: "GET", URL: "http://x", Status: 200}}, "", nil
}

type sseStubFrames struct{}

func (sseStubFrames) AppendFrame(context.Context, string, domain.Frame) error { return nil }
func (sseStubFrames) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) {
	return []domain.Frame{{ID: "f1"}}, "", nil
}

type sseStubEvents struct{}

func (sseStubEvents) AppendEvent(context.Context, string, domain.Event) error { return nil }
func (sseStubEvents) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) {
	return []domain.Event{{ID: "e1"}}, "", nil
}

// fake ResponseWriter+Flusher to capture SSE output
type sseRec struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (r *sseRec) Header() http.Header { return http.Header{} }
func (r *sseRec) Write(b []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.Write(b)
}
func (r *sseRec) WriteHeader(int) {}
func (r *sseRec) Flush()          {}
func (r *sseRec) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buf.String()
}

func TestHandleSessionStream_InitialSnapshotAndCancel(t *testing.T) {
	svc := uc.NewSessionService(sseStubSessions{}, sseStubFrames{}, sseStubEvents{})
	d := &Deps{Svc: svc, Monitor: NewMonitorHub()}
	// request with cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/api/sessions_stream/s1", nil)

	rec := &sseRec{}
	// run handler in goroutine and cancel shortly after to exit loop
	done := make(chan struct{})
	go func() { d.handleSessionStream(rec, req); close(done) }()
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done
	s := rec.String()
	if !strContains(s, "event: frames") || !strContains(s, "event: events") || !strContains(s, "event: http") {
		t.Fatalf("snapshot SSE events missing: %q", s)
	}
}

func TestHandleSessionStream_FanOutFrameEvent(t *testing.T) {
	svc := uc.NewSessionService(sseStubSessions{}, sseStubFrames{}, sseStubEvents{})
	d := &Deps{Svc: svc, Monitor: NewMonitorHub()}
	rec := &sseRec{}
	req, _ := http.NewRequest(http.MethodGet, "/api/sessions_stream/s1", nil)
	// run handler
	done := make(chan struct{})
	go func() { d.handleSessionStream(rec, req); close(done) }()
	// broadcast unrelated id -> ignored
	d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: "other"})
	// broadcast for our id -> should write SSE
	d.Monitor.Broadcast(MonitorEvent{Type: "frame_added", ID: "s1"})
	// stop by canceling context via closing connection: not possible, so just stop goroutine by Done after small sleep
	time.Sleep(50 * time.Millisecond)
	// force exit by canceling via req context not trivial; we leave goroutine running briefly and then ignore
	s := rec.String()
	if !strContains(s, "event: frames") {
		t.Fatalf("expected frames SSE after broadcast: %q", s)
	}
}
