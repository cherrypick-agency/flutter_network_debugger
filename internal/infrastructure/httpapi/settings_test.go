package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	sdomain "network-debugger/internal/features/settings/domain"
	settingsuc "network-debugger/internal/features/settings/usecase"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"testing"
)

func TestHandleV1Settings_GET_NoDelay(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp settingsDTO
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.ResponseDelay.Enabled {
		t.Error("ResponseDelay.Enabled should be false when no delay configured")
	}
	if resp.ResponseDelay.Value != "" {
		t.Errorf("ResponseDelay.Value = %s, want empty", resp.ResponseDelay.Value)
	}
}

func TestHandleV1Settings_GET_FixedDelay(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{ResponseDelayMs: 100}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp settingsDTO
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if !resp.ResponseDelay.Enabled {
		t.Error("ResponseDelay.Enabled should be true")
	}
	if resp.ResponseDelay.Value != "100" {
		t.Errorf("ResponseDelay.Value = %s, want 100", resp.ResponseDelay.Value)
	}
}

func TestHandleV1Settings_GET_RangeDelay(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{ResponseDelayMinMs: 50, ResponseDelayMaxMs: 150}}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp settingsDTO
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if !resp.ResponseDelay.Enabled {
		t.Error("ResponseDelay.Enabled should be true")
	}
	if resp.ResponseDelay.Value != "50-150" {
		t.Errorf("ResponseDelay.Value = %s, want 50-150", resp.ResponseDelay.Value)
	}
}

func TestHandleV1Settings_POST_DisableDelay(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{ResponseDelayMs: 100}}
	body := []byte(`{"responseDelay":{"enabled":false,"value":""}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if d.Cfg.ResponseDelayMs != 0 {
		t.Errorf("ResponseDelayMs = %d, want 0", d.Cfg.ResponseDelayMs)
	}
}

func TestHandleV1Settings_POST_FixedDelayValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"200"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if d.Cfg.ResponseDelayMs != 200 {
		t.Errorf("ResponseDelayMs = %d, want 200", d.Cfg.ResponseDelayMs)
	}
	if d.Cfg.ResponseDelayMinMs != 0 || d.Cfg.ResponseDelayMaxMs != 0 {
		t.Error("Range values should be 0")
	}
}

func TestHandleV1Settings_POST_RangeDelayValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"100-300"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if d.Cfg.ResponseDelayMinMs != 100 {
		t.Errorf("ResponseDelayMinMs = %d, want 100", d.Cfg.ResponseDelayMinMs)
	}
	if d.Cfg.ResponseDelayMaxMs != 300 {
		t.Errorf("ResponseDelayMaxMs = %d, want 300", d.Cfg.ResponseDelayMaxMs)
	}
	if d.Cfg.ResponseDelayMs != 0 {
		t.Error("Fixed ResponseDelayMs should be 0")
	}
}

func TestHandleV1Settings_POST_RangeDelaySwapped(t *testing.T) {
	// Test that if max < min, they are swapped
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"300-100"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if d.Cfg.ResponseDelayMinMs != 100 {
		t.Errorf("ResponseDelayMinMs = %d, want 100 (should be swapped)", d.Cfg.ResponseDelayMinMs)
	}
	if d.Cfg.ResponseDelayMaxMs != 300 {
		t.Errorf("ResponseDelayMaxMs = %d, want 300 (should be swapped)", d.Cfg.ResponseDelayMaxMs)
	}
}

func TestHandleV1Settings_POST_InvalidJSON(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestHandleV1Settings_POST_InvalidRangeFormat(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"abc-def"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestHandleV1Settings_POST_NegativeValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"-50"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestHandleV1Settings_POST_NegativeRangeValues(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"-50--10"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestHandleV1Settings_POST_ZeroValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{ResponseDelayMs: 100}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"0"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	// Zero value should disable delay
	if d.Cfg.ResponseDelayMs != 0 {
		t.Errorf("ResponseDelayMs = %d, want 0", d.Cfg.ResponseDelayMs)
	}
}

func TestHandleV1Settings_POST_EmptyValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{ResponseDelayMs: 100}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":""}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	// Empty value should disable delay
	if d.Cfg.ResponseDelayMs != 0 {
		t.Errorf("ResponseDelayMs = %d, want 0", d.Cfg.ResponseDelayMs)
	}
}

func TestHandleV1Settings_PUT_MethodNotAllowed(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", nil)
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rr.Code)
	}
}

func TestHandleV1Settings_POST_FontScaleAndTheme(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":false,"value":""},"fontScale":1.5,"highlightTheme":"monokai"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp settingsDTO
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	// FontScale and HighlightTheme are only returned if Settings service is available
	// Without Settings service, they won't be persisted or returned
}

func TestHandleV1Settings_GET_PrefersStoredUISettings(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{
		loadValue: sdomain.RuntimeSettings{
			ResponseDelayMs:     120,
			FontScale:           1.3,
			HighlightTheme:      "monokai",
			HighlightThemeLight: "github",
			HighlightThemeDark:  "dracula",
		},
	}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, &mockThrottleProfilesRepo{})
	d := &Deps{
		Cfg:      cfgpkg.Config{ResponseDelayMs: 10},
		Settings: settingsSvc,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp settingsDTO
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ResponseDelay.Value != "120" || resp.FontScale != 1.3 {
		t.Fatalf("unexpected settings payload: %+v", resp)
	}
	if resp.HighlightTheme != "monokai" || resp.HighlightThemeLight != "github" || resp.HighlightThemeDark != "dracula" {
		t.Fatalf("unexpected theme payload: %+v", resp)
	}
}

func TestHandleV1Settings_POST_PreservesThrottleAndStoresDualThemes(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{
		loadValue: sdomain.RuntimeSettings{
			ThrottleEnabled:       true,
			ThrottleDownKbps:      900,
			ThrottleUpKbps:        450,
			ThrottlePacketLoss:    3,
			ThrottleLatencyMs:     60,
			ThrottleLatencyJitter: 20,
			ThrottleOffline:       false,
		},
	}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, &mockThrottleProfilesRepo{})
	d := &Deps{
		Cfg: cfgpkg.Config{
			ThrottleEnabled:       true,
			ThrottleDownKbps:      900,
			ThrottleUpKbps:        450,
			ThrottlePacketLoss:    3,
			ThrottleLatencyMs:     60,
			ThrottleLatencyJitter: 20,
			ThrottleOffline:       false,
		},
		Settings: settingsSvc,
	}

	body := []byte(`{"responseDelay":{"enabled":true,"value":"200-400"},"fontScale":1.5,"highlightThemeLight":"github","highlightThemeDark":"dracula","highlightTheme":"monokai"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if mockSettingsRepo.saved == nil {
		t.Fatal("expected settings to be saved")
	}
	if mockSettingsRepo.saved.ThrottleDownKbps != 900 || mockSettingsRepo.saved.ThrottleLatencyMs != 60 {
		t.Fatalf("throttle settings were not preserved: %+v", *mockSettingsRepo.saved)
	}
	if mockSettingsRepo.saved.HighlightTheme != "monokai" || mockSettingsRepo.saved.HighlightThemeLight != "github" || mockSettingsRepo.saved.HighlightThemeDark != "dracula" {
		t.Fatalf("theme fields not saved: %+v", *mockSettingsRepo.saved)
	}
	if mockSettingsRepo.saved.ResponseDelayMinMs != 200 || mockSettingsRepo.saved.ResponseDelayMaxMs != 400 || mockSettingsRepo.saved.ResponseDelayMs != 0 {
		t.Fatalf("response delay not saved correctly: %+v", *mockSettingsRepo.saved)
	}
}

func TestHandleV1Settings_POST_ThemeOnly_PreservesResponseDelay(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{
		loadValue: sdomain.RuntimeSettings{
			ResponseDelayMs:     175,
			HighlightTheme:      "old",
			HighlightThemeLight: "old-light",
			HighlightThemeDark:  "old-dark",
		},
	}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, &mockThrottleProfilesRepo{})
	d := &Deps{
		Cfg: cfgpkg.Config{
			ResponseDelayMs: 175,
		},
		Settings: settingsSvc,
	}

	body := []byte(`{"highlightThemeLight":"github","highlightThemeDark":"dracula"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if mockSettingsRepo.saved == nil {
		t.Fatal("expected settings to be saved")
	}
	if mockSettingsRepo.saved.ResponseDelayMs != 175 || mockSettingsRepo.saved.ResponseDelayMinMs != 0 || mockSettingsRepo.saved.ResponseDelayMaxMs != 0 {
		t.Fatalf("response delay changed on theme-only update: %+v", *mockSettingsRepo.saved)
	}
	if mockSettingsRepo.saved.HighlightThemeLight != "github" || mockSettingsRepo.saved.HighlightThemeDark != "dracula" {
		t.Fatalf("dual themes not saved: %+v", *mockSettingsRepo.saved)
	}
}

func TestHandleV1Settings_POST_WhitespaceInValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":" 150 "}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if d.Cfg.ResponseDelayMs != 150 {
		t.Errorf("ResponseDelayMs = %d, want 150 (trimmed)", d.Cfg.ResponseDelayMs)
	}
}

func TestHandleV1Settings_POST_RangeWithWhitespace(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":" 100 - 200 "}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if d.Cfg.ResponseDelayMinMs != 100 {
		t.Errorf("ResponseDelayMinMs = %d, want 100", d.Cfg.ResponseDelayMinMs)
	}
	if d.Cfg.ResponseDelayMaxMs != 200 {
		t.Errorf("ResponseDelayMaxMs = %d, want 200", d.Cfg.ResponseDelayMaxMs)
	}
}

func TestHandleV1Settings_POST_NonIntegerValue(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	body := []byte(`{"responseDelay":{"enabled":true,"value":"abc"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	d.handleV1Settings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}
