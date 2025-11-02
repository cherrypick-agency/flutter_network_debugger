package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	mem "network-debugger/internal/adapters/storage/memory"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
	"testing"

	"github.com/rs/zerolog"
)

func TestUnifiedProxy_HTTP_HappyPath(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK-UP"))
	}))
	defer upstream.Close()

	logger := zerolog.New(io.Discard)
	d := &Deps{Logger: &logger, Monitor: NewMonitorHub(), Metrics: obs.NewMetrics()}
	store := mem.NewStore(8, 8, 0)
	d.Svc = usecase.NewSessionService(store, store, store)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/proxy?_target="+upstream.URL+"/a", nil)
	d.handleUnifiedProxy(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: %d", rr.Code)
	}
	b, _ := io.ReadAll(rr.Body)
	if string(b) != "OK-UP" {
		t.Fatalf("body: %s", string(b))
	}
}
