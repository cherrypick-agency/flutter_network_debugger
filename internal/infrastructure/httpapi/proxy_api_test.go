package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mappingp "network-debugger/internal/features/mapping/infrastructure/persistence"
	processp "network-debugger/internal/features/process/infrastructure/persistence"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	proxyuc "network-debugger/internal/features/proxy/usecase"
	settingsp "network-debugger/internal/features/settings/infrastructure/persistence"
	dbinfra "network-debugger/internal/infrastructure/db"
	obs "network-debugger/internal/infrastructure/observability"
	pruntime "network-debugger/internal/infrastructure/proxyruntime"
)

func setupProxyDeps(t *testing.T) *Deps {
	t.Helper()
	gdb, err := dbinfra.NewSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := gdb.AutoMigrate(
		&proxyp.ProxyConfigModel{},
		&settingsp.RuntimeSettingsModel{},
		&settingsp.ThrottleProfileModel{},
		&mappingp.MapRuleModel{},
		&processp.ProcessDetectionConfigModel{},
		&processp.IconCacheModel{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	d := &Deps{}
	d.DB = gdb
	d.ProxySvc = proxyuc.NewService(proxyp.NewRepo(gdb))
	l := obs.NewLogger("warn")
	d.Logger = l
	d.ProxyRt = pruntime.New(d.Logger)
	return d
}

func TestProxyAPI_GetPost(t *testing.T) {
	d := setupProxyDeps(t)

	// GET
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/proxy/config", nil)
	d.handleV1ProxyConfig(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("get code=%d", rr.Code)
	}

	// POST
	body := map[string]any{
		"forward": map[string]any{"enabled": true, "port": 0, "addr": "127.0.0.1:0"},
		"socks":   map[string]any{"enabled": false, "port": 0},
	}
	b, _ := json.Marshal(body)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader(b))
	d.handleV1ProxyConfig(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("post code=%d", rr.Code)
	}
}

func TestHandleV1ProxyConfig_FeatureDisabled(t *testing.T) {
	d := &Deps{ProxySvc: nil, ProxyRt: nil}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/proxy/config", nil)
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
}

func TestHandleV1ProxyConfig_POST_BadJSON(t *testing.T) {
	d := setupProxyDeps(t)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader([]byte("{invalid")))
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleV1ProxyConfig_POST_InvalidPort(t *testing.T) {
	d := setupProxyDeps(t)

	tests := []struct {
		name    string
		payload map[string]any
	}{
		{
			name: "forward port negative",
			payload: map[string]any{
				"forward": map[string]any{"enabled": true, "port": -1},
				"socks":   map[string]any{"enabled": false, "port": 0},
			},
		},
		{
			name: "forward port too large",
			payload: map[string]any{
				"forward": map[string]any{"enabled": true, "port": 99999},
				"socks":   map[string]any{"enabled": false, "port": 0},
			},
		},
		{
			name: "socks port negative",
			payload: map[string]any{
				"forward": map[string]any{"enabled": false, "port": 0},
				"socks":   map[string]any{"enabled": true, "port": -1},
			},
		},
		{
			name: "socks port too large",
			payload: map[string]any{
				"forward": map[string]any{"enabled": false, "port": 0},
				"socks":   map[string]any{"enabled": true, "port": 70000},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader(body))
			w := httptest.NewRecorder()

			d.handleV1ProxyConfig(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestHandleV1ProxyConfig_POST_InvalidAuthMode(t *testing.T) {
	d := setupProxyDeps(t)

	payload := map[string]any{
		"forward": map[string]any{"enabled": false, "port": 0},
		"socks":   map[string]any{"enabled": true, "port": 1080, "authMode": "invalid"},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader(body))
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleV1ProxyConfig_POST_WithPort(t *testing.T) {
	d := setupProxyDeps(t)

	payload := map[string]any{
		"forward": map[string]any{"enabled": true, "port": 8080},
		"socks":   map[string]any{"enabled": true, "port": 1080, "authMode": "none"},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader(body))
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp proxyCfgDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Forward.Port != 8080 {
		t.Errorf("Forward.Port = %d, want 8080", resp.Forward.Port)
	}
	if resp.Socks.Port != 1080 {
		t.Errorf("Socks.Port = %d, want 1080", resp.Socks.Port)
	}
}

func TestHandleV1ProxyConfig_POST_WithAddr(t *testing.T) {
	d := setupProxyDeps(t)

	payload := map[string]any{
		"forward": map[string]any{"enabled": true, "port": 0, "addr": "127.0.0.1:9090"},
		"socks":   map[string]any{"enabled": true, "port": 0, "addr": "localhost:2080"},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader(body))
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp proxyCfgDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Forward.Addr != "127.0.0.1:9090" {
		t.Errorf("Forward.Addr = %s, want 127.0.0.1:9090", resp.Forward.Addr)
	}
	if resp.Socks.Addr != "localhost:2080" {
		t.Errorf("Socks.Addr = %s, want localhost:2080", resp.Socks.Addr)
	}
}

func TestHandleV1ProxyConfig_POST_WithAuth(t *testing.T) {
	d := setupProxyDeps(t)

	payload := map[string]any{
		"forward": map[string]any{"enabled": false, "port": 0},
		"socks": map[string]any{
			"enabled":  true,
			"port":     1080,
			"authMode": "userpass",
			"user":     "testuser",
			"pass":     "testpass",
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/proxy/config", bytes.NewReader(body))
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp proxyCfgDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Socks.AuthMode != "userpass" {
		t.Errorf("Socks.AuthMode = %s, want userpass", resp.Socks.AuthMode)
	}
}

func TestHandleV1ProxyConfig_MethodNotAllowed(t *testing.T) {
	d := setupProxyDeps(t)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/proxy/config", nil)
	w := httptest.NewRecorder()

	d.handleV1ProxyConfig(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
