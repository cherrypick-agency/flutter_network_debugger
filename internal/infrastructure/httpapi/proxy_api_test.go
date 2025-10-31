package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	proxyuc "network-debugger/internal/features/proxy/usecase"
	dbinfra "network-debugger/internal/infrastructure/db"
	obs "network-debugger/internal/infrastructure/observability"
	pruntime "network-debugger/internal/infrastructure/proxyruntime"
)

func TestProxyAPI_GetPost(t *testing.T) {
	// in-memory DB
	gdb, err := dbinfra.NewSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	if err := gdb.AutoMigrate(&proxyp.ProxyConfigModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	d := &Deps{}
	d.DB = gdb
	d.ProxySvc = proxyuc.NewService(proxyp.NewRepo(gdb))
	l := obs.NewLogger("warn")
	d.Logger = l
	d.ProxyRt = pruntime.New(d.Logger)

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
