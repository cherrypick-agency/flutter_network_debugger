package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sdomain "network-debugger/internal/features/settings/domain"
	settingsuc "network-debugger/internal/features/settings/usecase"
	"network-debugger/internal/infrastructure/config"
)

type mockSettingsRepoForThrottle struct{}

func (m *mockSettingsRepoForThrottle) Load(ctx context.Context) (sdomain.RuntimeSettings, error) {
	return sdomain.RuntimeSettings{}, nil
}

func (m *mockSettingsRepoForThrottle) Save(ctx context.Context, s sdomain.RuntimeSettings) error {
	return nil
}

type mockThrottleProfilesRepo struct {
	listFunc   func(ctx context.Context) ([]sdomain.ThrottleProfile, error)
	upsertFunc func(ctx context.Context, p sdomain.ThrottleProfile) (sdomain.ThrottleProfile, error)
	deleteFunc func(ctx context.Context, id string) error
}

func (m *mockThrottleProfilesRepo) List(ctx context.Context) ([]sdomain.ThrottleProfile, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []sdomain.ThrottleProfile{}, nil
}

func (m *mockThrottleProfilesRepo) Upsert(ctx context.Context, p sdomain.ThrottleProfile) (sdomain.ThrottleProfile, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, p)
	}
	return p, nil
}

func (m *mockThrottleProfilesRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestHandleV1Throttle_GET(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			ThrottleEnabled:       true,
			ThrottleDownKbps:      1000,
			ThrottleUpKbps:        500,
			ThrottlePacketLoss:    2,
			ThrottleLatencyMs:     50,
			ThrottleLatencyJitter: 10,
			ThrottleOffline:       false,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/throttle", nil)
	w := httptest.NewRecorder()

	d.handleV1Throttle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp throttleDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Enabled {
		t.Error("Expected Enabled = true")
	}
	if resp.DownKbps != 1000 {
		t.Errorf("DownKbps = %d, want 1000", resp.DownKbps)
	}
	if resp.UpKbps != 500 {
		t.Errorf("UpKbps = %d, want 500", resp.UpKbps)
	}
	if resp.PacketLossPct != 2 {
		t.Errorf("PacketLossPct = %d, want 2", resp.PacketLossPct)
	}
	if resp.LatencyMs != 50 {
		t.Errorf("LatencyMs = %d, want 50", resp.LatencyMs)
	}
	if resp.LatencyJitter != 10 {
		t.Errorf("LatencyJitter = %d, want 10", resp.LatencyJitter)
	}
	if resp.Offline {
		t.Error("Expected Offline = false")
	}
	if len(resp.Presets) == 0 {
		t.Error("Expected non-empty Presets")
	}
}

func TestHandleV1Throttle_POST(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{},
	}

	payload := throttleDTO{
		Enabled:       true,
		DownKbps:      3000,
		UpKbps:        1500,
		PacketLossPct: 5,
		LatencyMs:     200,
		LatencyJitter: 50,
		Offline:       false,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/throttle", bytes.NewReader(body))
	w := httptest.NewRecorder()

	d.handleV1Throttle(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if !d.Cfg.ThrottleEnabled {
		t.Error("Expected ThrottleEnabled = true")
	}
	if d.Cfg.ThrottleDownKbps != 3000 {
		t.Errorf("ThrottleDownKbps = %d, want 3000", d.Cfg.ThrottleDownKbps)
	}
	if d.Cfg.ThrottleUpKbps != 1500 {
		t.Errorf("ThrottleUpKbps = %d, want 1500", d.Cfg.ThrottleUpKbps)
	}
}

func TestHandleV1Throttle_POST_BadJSON(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{},
	}

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/throttle", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	d.handleV1Throttle(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleV1Throttle_MethodNotAllowed(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{},
	}

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/throttle", nil)
	w := httptest.NewRecorder()

	d.handleV1Throttle(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleV1ThrottleProfiles_GET_NoSettings(t *testing.T) {
	d := &Deps{
		Cfg:      config.Config{},
		Settings: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/throttle/profiles", nil)
	w := httptest.NewRecorder()

	d.handleV1ThrottleProfiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp []any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp) != 0 {
		t.Error("Expected empty array when Settings is nil")
	}
}

func TestHandleV1ThrottleProfiles_POST_Success(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{}
	mockProfilesRepo := &mockThrottleProfilesRepo{
		upsertFunc: func(ctx context.Context, p sdomain.ThrottleProfile) (sdomain.ThrottleProfile, error) {
			return p, nil
		},
	}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, mockProfilesRepo)
	d := &Deps{
		Cfg:      config.Config{},
		Settings: settingsSvc,
	}

	payload := map[string]any{
		"id":            "profile-1",
		"name":          "Test Profile",
		"downKbps":      1000,
		"upKbps":        500,
		"packetLossPct": 2,
		"latencyMs":     50,
		"latencyJitter": 10,
		"offline":       false,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/throttle/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	d.handleV1ThrottleProfiles(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleV1ThrottleProfiles_POST_BadJSON(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{}
	mockProfilesRepo := &mockThrottleProfilesRepo{}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, mockProfilesRepo)
	d := &Deps{
		Cfg:      config.Config{},
		Settings: settingsSvc,
	}

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/throttle/profiles", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	d.handleV1ThrottleProfiles(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleV1ThrottleProfiles_DELETE_Success(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{}
	mockProfilesRepo := &mockThrottleProfilesRepo{
		deleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, mockProfilesRepo)
	d := &Deps{
		Cfg:      config.Config{},
		Settings: settingsSvc,
	}

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/throttle/profiles/profile-1", nil)
	w := httptest.NewRecorder()

	d.handleV1ThrottleProfiles(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestHandleV1ThrottleProfiles_DELETE_EmptyID(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{}
	mockProfilesRepo := &mockThrottleProfilesRepo{}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, mockProfilesRepo)
	d := &Deps{
		Cfg:      config.Config{},
		Settings: settingsSvc,
	}

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/throttle/profiles/", nil)
	w := httptest.NewRecorder()

	d.handleV1ThrottleProfiles(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleV1ThrottleProfiles_MethodNotAllowed(t *testing.T) {
	mockSettingsRepo := &mockSettingsRepoForThrottle{}
	mockProfilesRepo := &mockThrottleProfilesRepo{}
	settingsSvc := settingsuc.NewService(mockSettingsRepo, mockProfilesRepo)
	d := &Deps{
		Cfg:      config.Config{},
		Settings: settingsSvc,
	}

	req := httptest.NewRequest(http.MethodPut, "/_api/v1/throttle/profiles/profile-1", nil)
	w := httptest.NewRecorder()

	d.handleV1ThrottleProfiles(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
