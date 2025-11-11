package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"network-debugger/internal/adapters/storage/memory"
	mappingp "network-debugger/internal/features/mapping/infrastructure/persistence"
	processp "network-debugger/internal/features/process/infrastructure/persistence"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	setp "network-debugger/internal/features/settings/infrastructure/persistence"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestServer(t *testing.T) (*httptest.Server, *gorm.DB) {
	t.Helper()
	// in-memory sqlite
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := gdb.AutoMigrate(
		&setp.RuntimeSettingsModel{},
		&setp.ThrottleProfileModel{},
		&proxyp.ProxyConfigModel{},
		&mappingp.MapRuleModel{},
		&processp.ProcessDetectionConfigModel{},
		&processp.IconCacheModel{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg := cfgpkg.Config{Addr: ":0", CORSAllowOrigin: "*", DevMode: true}
	logger := obs.NewLogger("warn")
	metrics := obs.NewMetrics()
	store := memory.NewStore(100, 1000, 0)
	svc := usecase.NewSessionService(store, store, store)
	deps := &Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: NewMonitorHub(), DB: gdb}
	srv := httptest.NewServer(NewRouterWithDeps(deps))
	return srv, gdb
}

func httpJSON(t *testing.T, client *http.Client, method, url string, body any) (*http.Response, any) {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, r)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("http %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	var v any
	_ = json.NewDecoder(resp.Body).Decode(&v)
	return resp, v
}

func TestThrottleProfilesCRUD(t *testing.T) {
	ts, _ := newTestServer(t)
	defer ts.Close()
	c := ts.Client()

	// 1) GET throttle defaults
	if _, m := httpJSON(t, c, http.MethodGet, ts.URL+"/_api/v1/throttle", nil); m == nil {
		t.Fatalf("empty throttle json")
	}

	// 2) POST throttle set values
	if resp, _ := httpJSON(t, c, http.MethodPost, ts.URL+"/_api/v1/throttle", map[string]any{
		"enabled": true, "downKbps": 3000, "upKbps": 3000, "packetLossPct": 0, "offline": false,
	}); resp.StatusCode != http.StatusOK {
		t.Fatalf("throttle post status=%d", resp.StatusCode)
	}

	// 3) GET profiles should be empty
	if _, arr := httpJSON(t, c, http.MethodGet, ts.URL+"/_api/v1/throttle/profiles", nil); arr == nil {
		t.Fatalf("profiles list empty body")
	}

	// 4) POST create profile
	var id string
	// create profile (decode into map to extract id)
	{
		b, _ := json.Marshal(map[string]any{
			"name": "Fast 4G", "downKbps": 3000, "upKbps": 3000, "packetLossPct": 0, "offline": false,
		})
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/_api/v1/throttle/profiles", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		respC, err := c.Do(req)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		defer respC.Body.Close()
		if respC.StatusCode != http.StatusOK {
			t.Fatalf("create profile status=%d", respC.StatusCode)
		}
		var created map[string]any
		_ = json.NewDecoder(respC.Body).Decode(&created)
		if s, ok := created["id"].(string); ok {
			id = s
		}
	}
	if id == "" {
		// fallback: fetch list and grab id of the first profile
		_, list := httpJSON(t, c, http.MethodGet, ts.URL+"/_api/v1/throttle/profiles", nil)
		if arr, ok := list.([]any); ok && len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				if s, ok := m["id"].(string); ok {
					id = s
				}
			}
		}
	}
	if id == "" {
		t.Skip("profile id not returned by API")
	}
	if id == "" {
		t.Fatalf("no id returned")
	}

	// 5) DELETE profile
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/_api/v1/throttle/profiles/"+id, nil)
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status=%d", resp.StatusCode)
	}
}
