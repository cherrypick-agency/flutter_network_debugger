package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	processdomain "network-debugger/internal/features/process/domain"
	processuc "network-debugger/internal/features/process/usecase"
)

// Composer 1.
// Mock dependencies для тестирования process handlers
type mockConfigRepoForProcess struct {
	loadFunc func(ctx context.Context) (*processdomain.DetectionConfig, error)
	saveFunc func(ctx context.Context, cfg *processdomain.DetectionConfig) error
}

func (m *mockConfigRepoForProcess) Load(ctx context.Context) (*processdomain.DetectionConfig, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	return &processdomain.DetectionConfig{
		ID:              1,
		Enabled:         true,
		UseHelperTool:   false,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTLSeconds: 3600,
		FallbackEnabled: true,
		UpdatedAt:       time.Now(),
	}, nil
}

func (m *mockConfigRepoForProcess) Save(ctx context.Context, cfg *processdomain.DetectionConfig) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, cfg)
	}
	return nil
}

type mockIconCacheRepoForProcess struct{}

func (m *mockIconCacheRepoForProcess) Get(key string) (*processdomain.AppIcon, error) {
	return nil, errors.New("not found")
}

func (m *mockIconCacheRepoForProcess) Set(key string, icon *processdomain.AppIcon, ttl time.Duration) error {
	return nil
}

func (m *mockIconCacheRepoForProcess) Delete(key string) error {
	return nil
}

func (m *mockIconCacheRepoForProcess) Clear() error {
	return nil
}

func (m *mockIconCacheRepoForProcess) CleanupExpired() error {
	return nil
}

type mockProcessDetectorForProcess struct{}

func (m *mockProcessDetectorForProcess) DetectByPort(ctx context.Context, port uint32) (*processdomain.ProcessInfo, error) {
	return nil, nil
}

func (m *mockProcessDetectorForProcess) DetectByPID(ctx context.Context, pid int32) (*processdomain.ProcessInfo, error) {
	return nil, nil
}

func (m *mockProcessDetectorForProcess) RequiresPrivileges() bool {
	return false
}

type mockHelperClientForProcess struct{}

func (m *mockHelperClientForProcess) IsRunning() bool {
	return false
}

func (m *mockHelperClientForProcess) DetectProcess(port uint32) (*processdomain.ProcessInfo, error) {
	return nil, nil
}

func (m *mockHelperClientForProcess) ExtractIcon(pid int32) (*processdomain.AppIcon, error) {
	return nil, nil
}

func (m *mockHelperClientForProcess) Ping() error {
	return nil
}

func (m *mockHelperClientForProcess) Close() error {
	return nil
}

type mockIconExtractorForProcess struct{}

func (m *mockIconExtractorForProcess) ExtractByPID(ctx context.Context, pid int32) (*processdomain.AppIcon, error) {
	return nil, nil
}

func (m *mockIconExtractorForProcess) ExtractByPath(ctx context.Context, path string) (*processdomain.AppIcon, error) {
	return nil, nil
}

type mockHelperInstallerForProcess struct {
	isInstalledFunc func() bool
	installFunc     func(binaryPath string) error
	getVersionFunc  func() string
}

func (m *mockHelperInstallerForProcess) IsInstalled() bool {
	if m.isInstalledFunc != nil {
		return m.isInstalledFunc()
	}
	return false
}

func (m *mockHelperInstallerForProcess) Install(binaryPath string) error {
	if m.installFunc != nil {
		return m.installFunc(binaryPath)
	}
	return nil
}

func (m *mockHelperInstallerForProcess) Uninstall() error {
	return nil
}

func (m *mockHelperInstallerForProcess) GetVersion() string {
	if m.getVersionFunc != nil {
		return m.getVersionFunc()
	}
	return ""
}

func setupProcessDeps(configRepo *mockConfigRepoForProcess, installer *mockHelperInstallerForProcess) *Deps {
	iconCache := &mockIconCacheRepoForProcess{}
	localDetector := &mockProcessDetectorForProcess{}
	helperClient := &mockHelperClientForProcess{}
	iconExtractor := &mockIconExtractorForProcess{}
	if installer == nil {
		installer = &mockHelperInstallerForProcess{}
	}

	logger := zerolog.Nop()
	service := processuc.NewService(
		configRepo,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		installer,
		"/path/to/helper",
		&logger,
	)

	return &Deps{
		ProcessSvc: service,
	}
}

// Composer 1.
// Тесты для handleV1ProcessConfig

func TestHandleV1ProcessConfig_GET_Success(t *testing.T) {
	expectedConfig := &processdomain.DetectionConfig{
		ID:              1,
		Enabled:         true,
		UseHelperTool:   true,
		HelperInstalled: true,
		CacheEnabled:    true,
		CacheTTLSeconds: 7200,
		FallbackEnabled: false,
		UpdatedAt:       time.Now(),
	}

	configRepo := &mockConfigRepoForProcess{
		loadFunc: func(ctx context.Context) (*processdomain.DetectionConfig, error) {
			return expectedConfig, nil
		},
	}
	deps := setupProcessDeps(configRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/process/config", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp processConfigDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Enabled != expectedConfig.Enabled {
		t.Errorf("Enabled = %v, want %v", resp.Enabled, expectedConfig.Enabled)
	}
	if resp.UseHelperTool != expectedConfig.UseHelperTool {
		t.Errorf("UseHelperTool = %v, want %v", resp.UseHelperTool, expectedConfig.UseHelperTool)
	}
	if resp.CacheTTL != expectedConfig.CacheTTLSeconds {
		t.Errorf("CacheTTL = %d, want %d", resp.CacheTTL, expectedConfig.CacheTTLSeconds)
	}
}

func TestHandleV1ProcessConfig_GET_ServiceError(t *testing.T) {
	configRepo := &mockConfigRepoForProcess{
		loadFunc: func(ctx context.Context) (*processdomain.DetectionConfig, error) {
			return nil, errors.New("database error")
		},
	}
	deps := setupProcessDeps(configRepo, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/process/config", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleV1ProcessConfig_POST_Success(t *testing.T) {
	var savedConfig *processdomain.DetectionConfig
	configRepo := &mockConfigRepoForProcess{
		saveFunc: func(ctx context.Context, cfg *processdomain.DetectionConfig) error {
			savedConfig = cfg
			return nil
		},
		loadFunc: func(ctx context.Context) (*processdomain.DetectionConfig, error) {
			if savedConfig != nil {
				return savedConfig, nil
			}
			return &processdomain.DetectionConfig{ID: 1, UpdatedAt: time.Now()}, nil
		},
	}
	deps := setupProcessDeps(configRepo, nil)

	body := processConfigDTO{
		Enabled:         false,
		UseHelperTool:   true,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTL:        1800,
		FallbackEnabled: true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/process/config", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleV1ProcessConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if savedConfig == nil {
		t.Fatal("Config was not saved")
	}

	if savedConfig.Enabled != body.Enabled {
		t.Errorf("Saved Enabled = %v, want %v", savedConfig.Enabled, body.Enabled)
	}
	if savedConfig.CacheTTLSeconds != body.CacheTTL {
		t.Errorf("Saved CacheTTL = %d, want %d", savedConfig.CacheTTLSeconds, body.CacheTTL)
	}

	var resp processConfigDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Enabled != body.Enabled {
		t.Errorf("Response Enabled = %v, want %v", resp.Enabled, body.Enabled)
	}
}

func TestHandleV1ProcessConfig_POST_InvalidJSON(t *testing.T) {
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/process/config", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleV1ProcessConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleV1ProcessConfig_POST_ServiceError(t *testing.T) {
	configRepo := &mockConfigRepoForProcess{
		saveFunc: func(ctx context.Context, cfg *processdomain.DetectionConfig) error {
			return errors.New("save failed")
		},
	}
	deps := setupProcessDeps(configRepo, nil)

	body := processConfigDTO{
		Enabled: true,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/process/config", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	deps.handleV1ProcessConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleV1ProcessConfig_InvalidMethod(t *testing.T) {
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, nil)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/process/config", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessConfig(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// Composer 1.
// Тесты для handleV1ProcessHelperStatus

func TestHandleV1ProcessHelperStatus_GET_Success(t *testing.T) {
	installer := &mockHelperInstallerForProcess{
		isInstalledFunc: func() bool {
			return true
		},
		getVersionFunc: func() string {
			return "1.0.0"
		},
	}
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, installer)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/process/helper/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessHelperStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["installed"] != true {
		t.Errorf("Installed = %v, want true", resp["installed"])
	}
	if resp["version"] != "1.0.0" {
		t.Errorf("Version = %v, want '1.0.0'", resp["version"])
	}
}

func TestHandleV1ProcessHelperStatus_GET_NotInstalled(t *testing.T) {
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/process/helper/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessHelperStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["installed"] != false {
		t.Errorf("Installed = %v, want false", resp["installed"])
	}
}

func TestHandleV1ProcessHelperStatus_InvalidMethod(t *testing.T) {
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/process/helper/status", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessHelperStatus(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}

// Composer 1.
// Тесты для handleV1ProcessHelperInstall

func TestHandleV1ProcessHelperInstall_POST_Success(t *testing.T) {
	installCalled := false
	installer := &mockHelperInstallerForProcess{
		installFunc: func(binaryPath string) error {
			installCalled = true
			return nil
		},
	}
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, installer)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/process/helper/install", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessHelperInstall(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	if !installCalled {
		t.Error("InstallHelper was not called")
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "installed" {
		t.Errorf("Status = %q, want 'installed'", resp["status"])
	}
}

func TestHandleV1ProcessHelperInstall_POST_ServiceError(t *testing.T) {
	installer := &mockHelperInstallerForProcess{
		installFunc: func(binaryPath string) error {
			return errors.New("installation failed")
		},
	}
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, installer)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/process/helper/install", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessHelperInstall(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleV1ProcessHelperInstall_InvalidMethod(t *testing.T) {
	deps := setupProcessDeps(&mockConfigRepoForProcess{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/process/helper/install", nil)
	w := httptest.NewRecorder()

	deps.handleV1ProcessHelperInstall(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
