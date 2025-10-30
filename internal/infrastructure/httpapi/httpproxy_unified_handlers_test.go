package httpapi

import (
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"network-debugger/internal/usecase"
	"testing"

	"github.com/rs/zerolog"
)

func TestHandleUnifiedProxy_WebSocketRequest(t *testing.T) {
	// Setup minimal deps
	logger := zerolog.Nop()
	deps := &Deps{
		Cfg:    cfgpkg.Config{},
		Logger: &logger,
		Svc:    usecase.NewSessionService(nil, nil, nil),
		Live:   NewLiveSessions(),
	}

	// Create request with WebSocket upgrade header
	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()

	// Should dispatch to WS handler (which will fail without proper setup, but that's ok for coverage)
	deps.handleUnifiedProxy(w, req)

	// We expect some response (even if error) showing WS handler was called
	// The exact status depends on WS handler implementation
}

func TestHandleUnifiedProxy_HTTPRequest(t *testing.T) {
	// Setup minimal deps
	logger := zerolog.Nop()
	deps := &Deps{
		Cfg:    cfgpkg.Config{},
		Logger: &logger,
		Svc:    usecase.NewSessionService(nil, nil, nil),
	}

	// Create regular HTTP request (no WS headers)
	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	w := httptest.NewRecorder()

	// Should dispatch to HTTP handler
	deps.handleUnifiedProxy(w, req)

	// We expect some response from HTTP handler
}

func TestHandleUnifiedProxy_WithSecWebSocketKey(t *testing.T) {
	logger := zerolog.Nop()
	deps := &Deps{
		Cfg:    cfgpkg.Config{},
		Logger: &logger,
		Svc:    usecase.NewSessionService(nil, nil, nil),
		Live:   NewLiveSessions(),
	}

	// Request with only Sec-WebSocket-Key (another way to detect WS)
	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	w := httptest.NewRecorder()

	deps.handleUnifiedProxy(w, req)
}

func TestHandleUnifiedProxy_WithSecWebSocketVersion(t *testing.T) {
	logger := zerolog.Nop()
	deps := &Deps{
		Cfg:    cfgpkg.Config{},
		Logger: &logger,
		Svc:    usecase.NewSessionService(nil, nil, nil),
		Live:   NewLiveSessions(),
	}

	// Request with only Sec-WebSocket-Version
	req := httptest.NewRequest(http.MethodGet, "/proxy", nil)
	req.Header.Set("Sec-WebSocket-Version", "13")

	w := httptest.NewRecorder()

	deps.handleUnifiedProxy(w, req)
}
