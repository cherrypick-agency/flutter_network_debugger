package httpapi

import (
	"io"
	"testing"
	"time"

	mem "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"

	"github.com/rs/zerolog"
)

func TestSessionDetailServiceListLegacyEventsKeepsAliasFields(t *testing.T) {
	details := newSessionDetailService(makeDepsWithAPIStub())

	page, apiErr := details.listLegacyEvents(contextWithNoCancel(), "s1", "", 10)
	if apiErr != nil {
		t.Fatalf("unexpected apiErr: %+v", apiErr)
	}
	if len(page.Items) != 1 {
		t.Fatalf("len(items)=%d want=1", len(page.Items))
	}
	item := page.Items[0]
	if item.Event != "ev" || item.Name != "ev" {
		t.Fatalf("event/name mismatch: %+v", item)
	}
}

func TestSessionDetailServiceFrameBodyPreviewFallback(t *testing.T) {
	store := mem.NewStore(100, 100, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*"},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}

	const sessionID = "preview-session"
	const frameID = "preview-frame"
	_ = svc.Create(contextWithNoCancel(), domain.Session{ID: sessionID, StartedAt: time.Now().UTC()})
	_ = svc.AddFrame(contextWithNoCancel(), sessionID, domain.Frame{
		ID:        frameID,
		Ts:        time.Now().UTC(),
		Direction: domain.DirectionClientToUpstream,
		Opcode:    domain.OpcodeText,
		Preview:   `{"hello":"preview"}`,
	})

	body, apiErr := newSessionDetailService(deps).frameBody(contextWithNoCancel(), sessionID, frameID)
	if apiErr != nil {
		t.Fatalf("unexpected apiErr: %+v", apiErr)
	}
	if body.Source != "preview" {
		t.Fatalf("source=%q want preview", body.Source)
	}
	if body.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("contentType=%q", body.ContentType)
	}
	if string(body.Data) != `{"hello":"preview"}` {
		t.Fatalf("body=%q", string(body.Data))
	}
}
