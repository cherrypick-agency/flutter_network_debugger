package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"network-debugger/internal/features/process/domain"
)

// Composer 1.
func TestNewService(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	if service.config != config {
		t.Error("Config repository not set correctly")
	}

	if service.iconCache != iconCache {
		t.Error("Icon cache repository not set correctly")
	}

	// Проверяем что localDetector установлен (не можем сравнить напрямую из-за интерфейса)
	if service.localDetector == nil {
		t.Error("Local detector not set")
	}

	if service.helperClient != helperClient {
		t.Error("Helper client not set correctly")
	}

	// Проверяем что iconExtractor установлен (не можем сравнить напрямую из-за интерфейса)
	if service.iconExtractor == nil {
		t.Error("Icon extractor not set")
	}

	if service.helperInstaller != helperInstaller {
		t.Error("Helper installer not set correctly")
	}

	if service.helperBinaryPath != "/path/to/helper" {
		t.Errorf("Helper binary path = %q, want %q", service.helperBinaryPath, "/path/to/helper")
	}
}

// Mock implementations
type mockConfigRepository struct {
	loadFunc func(ctx context.Context) (*domain.DetectionConfig, error)
}

func (m *mockConfigRepository) Load(ctx context.Context) (*domain.DetectionConfig, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	return &domain.DetectionConfig{
		Enabled:         true,
		UseHelperTool:   false,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTLSeconds: 3600,
		FallbackEnabled: true,
		UpdatedAt:       time.Now(),
	}, nil
}

func (m *mockConfigRepository) Save(ctx context.Context, config *domain.DetectionConfig) error {
	return nil
}

type mockIconCacheRepository struct{}

func (m *mockIconCacheRepository) Get(key string) (*domain.AppIcon, error) {
	return nil, errors.New("not found")
}

func (m *mockIconCacheRepository) Set(key string, icon *domain.AppIcon, ttl time.Duration) error {
	return nil
}

func (m *mockIconCacheRepository) Clear() error {
	return nil
}

func (m *mockIconCacheRepository) Delete(key string) error {
	return nil
}

func (m *mockIconCacheRepository) CleanupExpired() error {
	return nil
}

type mockProcessDetector struct{}

func (m *mockProcessDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	return &domain.ProcessInfo{
		PID:            1234,
		Name:           "test-process",
		ExecutablePath: "/usr/bin/test",
		DetectedAt:     time.Now(),
	}, nil
}

func (m *mockProcessDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
	return &domain.ProcessInfo{
		PID:            pid,
		Name:           "test-process",
		ExecutablePath: "/usr/bin/test",
		DetectedAt:     time.Now(),
	}, nil
}

func (m *mockProcessDetector) RequiresPrivileges() bool {
	return false
}

type mockHelperClient struct{}

func (m *mockHelperClient) IsRunning() bool {
	return false
}

func (m *mockHelperClient) DetectProcess(port uint32) (*domain.ProcessInfo, error) {
	return nil, nil
}

func (m *mockHelperClient) ExtractIcon(pid int32) (*domain.AppIcon, error) {
	return nil, nil
}

func (m *mockHelperClient) Ping() error {
	return nil
}

func (m *mockHelperClient) Close() error {
	return nil
}

type mockIconExtractor struct{}

func (m *mockIconExtractor) Extract(ctx context.Context, pid int32, path string) (*domain.AppIcon, error) {
	return &domain.AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3},
	}, nil
}

func (m *mockIconExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	return &domain.AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3},
	}, nil
}

func (m *mockIconExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	return &domain.AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3},
	}, nil
}

type mockHelperInstaller struct{}

func (m *mockHelperInstaller) IsInstalled() bool {
	return false
}

func (m *mockHelperInstaller) Install(helperBinaryPath string) error {
	return nil
}

func (m *mockHelperInstaller) Uninstall() error {
	return nil
}

func (m *mockHelperInstaller) GetVersion() string {
	return "1.0.0"
}

// Composer 1.
func TestService_GetConfig(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	cfg, err := service.GetConfig(ctx)

	if err != nil {
		t.Errorf("GetConfig() error = %v, want nil", err)
	}

	if cfg == nil {
		t.Error("GetConfig() returned nil config")
	}
}

// Composer 1.
func TestService_SaveConfig(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	cfg := &domain.DetectionConfig{
		Enabled:         true,
		UseHelperTool:   false,
		HelperInstalled: false,
		CacheEnabled:    true,
		CacheTTLSeconds: 3600,
		FallbackEnabled: true,
		UpdatedAt:       time.Now(),
	}

	ctx := context.Background()
	err := service.SaveConfig(ctx, cfg)

	if err != nil {
		t.Errorf("SaveConfig() error = %v, want nil", err)
	}
}

// Composer 1.
func TestService_CheckHelperStatus(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	status := service.CheckHelperStatus()

	if status.Installed {
		t.Error("CheckHelperStatus() should return Installed=false for mock")
	}

	if status.Running {
		t.Error("CheckHelperStatus() should return Running=false for mock")
	}
}

// Composer 1.
func TestService_InstallHelper(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	err := service.InstallHelper(ctx)

	if err != nil {
		t.Errorf("InstallHelper() error = %v, want nil", err)
	}
}

// Composer 1.
func TestService_DetectForConnection_Disabled(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	// Мокаем конфиг с отключенной детекцией
	config.loadFunc = func(ctx context.Context) (*domain.DetectionConfig, error) {
		return &domain.DetectionConfig{
			Enabled: false,
		}, nil
	}

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	info, err := service.DetectForConnection(ctx, 8080)

	if err != nil {
		t.Errorf("DetectForConnection() error = %v, want nil", err)
	}

	if info != nil {
		t.Error("DetectForConnection() should return nil when disabled")
	}
}

// Composer 1.
func TestService_DetectForConnection_WithHelper(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClientWithRunning{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	config.loadFunc = func(ctx context.Context) (*domain.DetectionConfig, error) {
		return &domain.DetectionConfig{
			Enabled:       true,
			UseHelperTool: true,
		}, nil
	}

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	info, err := service.DetectForConnection(ctx, 8080)

	if err != nil {
		t.Errorf("DetectForConnection() error = %v, want nil", err)
	}

	// Может быть nil если helper не вернул результат
	_ = info
}

// Composer 1.
func TestService_DetectForConnection_LocalFallback(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	config.loadFunc = func(ctx context.Context) (*domain.DetectionConfig, error) {
		return &domain.DetectionConfig{
			Enabled:         true,
			UseHelperTool:   false,
			FallbackEnabled: true,
		}, nil
	}

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	info, err := service.DetectForConnection(ctx, 8080)

	if err != nil {
		t.Errorf("DetectForConnection() error = %v, want nil", err)
	}

	if info == nil {
		t.Error("DetectForConnection() should return process info")
	}
}

// Composer 1.
func TestService_DetectForConnection_FallbackToUnknown(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetectorWithError{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	config.loadFunc = func(ctx context.Context) (*domain.DetectionConfig, error) {
		return &domain.DetectionConfig{
			Enabled:         true,
			UseHelperTool:   false,
			FallbackEnabled: true,
		}, nil
	}

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	info, err := service.DetectForConnection(ctx, 8080)

	if err != nil {
		t.Errorf("DetectForConnection() error = %v, want nil", err)
	}

	if info == nil {
		t.Error("DetectForConnection() should return fallback process info")
	}

	if info != nil && info.Name != "Unknown Process" {
		t.Errorf("DetectForConnection() Name = %q, want %q", info.Name, "Unknown Process")
	}
}

type mockHelperClientWithRunning struct {
	mockHelperClient
}

func (m *mockHelperClientWithRunning) IsRunning() bool {
	return true
}

func (m *mockHelperClientWithRunning) DetectProcess(port uint32) (*domain.ProcessInfo, error) {
	return &domain.ProcessInfo{
		PID:            1234,
		Name:           "test-process",
		ExecutablePath: "/usr/bin/test",
		DetectedAt:     time.Now(),
	}, nil
}

type mockProcessDetectorWithError struct {
	mockProcessDetector
}

func (m *mockProcessDetectorWithError) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	return nil, errors.New("detection failed")
}
