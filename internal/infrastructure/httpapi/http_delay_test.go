package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"testing"
	"time"
)

// Без t.Parallel — измеряем время сна.

func TestSleepResponseDelay_Fixed(t *testing.T) {
	cfg := cfgpkg.Config{ResponseDelayMs: 15}
	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)
	if elapsed < 15*time.Millisecond {
		t.Fatalf("delay too small: %v", elapsed)
	}
}

func TestSleepResponseDelay_Range(t *testing.T) {
	// Узкий диапазон, чтобы тест был быстрым, и без флаков
	cfg := cfgpkg.Config{ResponseDelayMinMs: 5, ResponseDelayMaxMs: 10}
	start := time.Now()
	sleepResponseDelay(cfg)
	elapsed := time.Since(start)
	if elapsed < 5*time.Millisecond {
		t.Fatalf("delay below min: %v", elapsed)
	}
	// Верхнюю границу не проверяем строго из‑за планировщика, важно что >= min
}

func TestHandleV1Settings_GetAndPost(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	// GET default
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	d.handleV1Settings(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET status")
	}

	// POST bad json
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader([]byte("{")))
	d.handleV1Settings(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("POST bad json status")
	}

	// POST range
	rr = httptest.NewRecorder()
	body := []byte(`{"responseDelay":{"enabled":true,"value":"10-20"}}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	d.handleV1Settings(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("POST ok status")
	}
	if d.Cfg.ResponseDelayMinMs != 10 || d.Cfg.ResponseDelayMaxMs != 20 {
		t.Fatalf("cfg not set: %+v", d.Cfg)
	}
}

func TestHandleV1Settings_MethodNotAllowed(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/settings", nil)
	d.handleV1Settings(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405")
	}
}

func TestHandleV1Settings_BadValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader([]byte(`{"responseDelay":{"enabled":true,"value":"x-y"}}`)))
	d.handleV1Settings(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad value must be 400")
	}
}

func TestHandleV1Settings_RangeSwap(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader([]byte(`{"responseDelay":{"enabled":true,"value":"20-10"}}`)))
	d.handleV1Settings(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("range swap should be ok")
	}
}
