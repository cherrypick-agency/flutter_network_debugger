package httpapi

import (
	"bytes"
	"github.com/rs/zerolog"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
)

func TestHTTPProxy_POST_SpoolAndFormPreview(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer upstream.Close()
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 1024, ExposeSensitiveHeaders: false, PreviewDecompress: true, CaptureBodies: true, BodyMaxBytes: 1024}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	h := NewRouterWithoutForwardProxy(d)

	// multipart body
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("a", "1")
	fw, _ := mw.CreateFormFile("file", "f.txt")
	_, _ = fw.Write([]byte("hello"))
	_ = mw.Close()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/httpproxy/upload?_target="+upstream.URL, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("post status: %d", rr.Code)
	}
}

func TestHTTPProxy_ErrorHandler_ConnectionRefused(t *testing.T) {
	s := uc.NewSessionService(stubRepo{}, stubRepo{}, stubRepo{})
	lg := zerolog.New(io.Discard)
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 128, ExposeSensitiveHeaders: false, PreviewDecompress: true}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s, Logger: &lg}
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	// impossible target 127.0.0.1:1
	req := httptest.NewRequest(http.MethodGet, "/httpproxy?_target=http://127.0.0.1:1", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rr.Code)
	}
}
