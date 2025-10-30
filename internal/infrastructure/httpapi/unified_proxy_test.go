package httpapi

import (
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"testing"
	"time"
)

func TestIsWebSocketRequest_UpgradeHeader(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Upgrade", "websocket")

	if !isWebSocketRequest(r) {
		t.Error("should return true for Upgrade: websocket header")
	}
}

func TestIsWebSocketRequest_SecWebSocketKey(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	if !isWebSocketRequest(r) {
		t.Error("should return true for Sec-WebSocket-Key header")
	}
}

func TestIsWebSocketRequest_SecWebSocketVersion(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Sec-WebSocket-Version", "13")

	if !isWebSocketRequest(r) {
		t.Error("should return true for Sec-WebSocket-Version header")
	}
}

func TestIsWebSocketRequest_NoWSHeaders(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	if isWebSocketRequest(r) {
		t.Error("should return false when no WebSocket headers present")
	}
}

func TestIsWebSocketRequest_RegularHeaders(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Accept", "application/json")

	if isWebSocketRequest(r) {
		t.Error("should return false for regular HTTP headers")
	}
}

func TestNewTransport_DefaultSettings(t *testing.T) {
	tr := newTransport(cfgpkg.Config{})

	if tr == nil {
		t.Fatal("newTransport returned nil")
	}

	// Check timeout settings
	if tr.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want 10s", tr.TLSHandshakeTimeout)
	}
	if tr.ResponseHeaderTimeout != 25*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 25s", tr.ResponseHeaderTimeout)
	}
	if tr.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want 90s", tr.IdleConnTimeout)
	}
	if tr.ExpectContinueTimeout != 1*time.Second {
		t.Errorf("ExpectContinueTimeout = %v, want 1s", tr.ExpectContinueTimeout)
	}

	// MaxIdleConns
	if tr.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", tr.MaxIdleConns)
	}

	// TLS config should be nil when InsecureTLS is false
	if tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("TLSClientConfig should not skip verify when InsecureTLS is false")
	}
}

func TestNewTransport_InsecureTLS(t *testing.T) {
	tr := newTransport(cfgpkg.Config{InsecureTLS: true})

	if tr.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig should not be nil when InsecureTLS is true")
	}
	if !tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true when InsecureTLS is true")
	}
}

func TestNewTransport_SecureTLS(t *testing.T) {
	tr := newTransport(cfgpkg.Config{InsecureTLS: false})

	// Should either be nil or have InsecureSkipVerify set to false
	if tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false when InsecureTLS is false")
	}
}

func TestNewTransport_ProxyFromEnvironment(t *testing.T) {
	tr := newTransport(cfgpkg.Config{})

	if tr.Proxy == nil {
		t.Error("Proxy function should be set to ProxyFromEnvironment")
	}
}

func TestNewTransport_CanMakeRequests(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create transport and client
	tr := newTransport(cfgpkg.Config{})
	client := &http.Client{Transport: tr}

	// Make a request
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}
