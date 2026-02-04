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

	// Check that localDetector is set (can't compare directly due to interface)
	if service.localDetector == nil {
		t.Error("Local detector not set")
	}

	if service.helperClient != helperClient {
		t.Error("Helper client not set correctly")
	}

	// Check that iconExtractor is set (can't compare directly due to interface)
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

	// Mock config with detection disabled
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

	// May be nil if helper didn't return a result
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

func TestService_DetectForConnection_ConfigLoadError(t *testing.T) {
	config := &mockConfigRepository{
		loadFunc: func(ctx context.Context) (*domain.DetectionConfig, error) {
			return nil, errors.New("config load failed")
		},
	}
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
	info, err := service.DetectForConnection(ctx, 8080)

	if err == nil {
		t.Error("DetectForConnection() should return error when config load fails")
	}

	if info != nil {
		t.Error("DetectForConnection() should return nil info when config load fails")
	}
}

func TestService_DetectForConnection_HelperErrorButNoInfo(t *testing.T) {
	config := &mockConfigRepository{
		loadFunc: func(ctx context.Context) (*domain.DetectionConfig, error) {
			return &domain.DetectionConfig{
				Enabled:       true,
				UseHelperTool: true,
			}, nil
		},
	}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClientWithError{}
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
	info, err := service.DetectForConnection(ctx, 8080)

	if err != nil {
		t.Errorf("DetectForConnection() should not return error when helper fails but local succeeds, got: %v", err)
	}

	if info == nil {
		t.Error("DetectForConnection() should return info from local detector when helper fails")
	}
}

func TestService_DetectForConnection_FallbackDisabled(t *testing.T) {
	config := &mockConfigRepository{
		loadFunc: func(ctx context.Context) (*domain.DetectionConfig, error) {
			return &domain.DetectionConfig{
				Enabled:         true,
				UseHelperTool:   false,
				FallbackEnabled: false,
			}, nil
		},
	}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetectorWithError{}
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
	info, err := service.DetectForConnection(ctx, 8080)

	if err == nil {
		t.Error("DetectForConnection() should return error when fallback is disabled and detection fails")
	}

	if info != nil {
		t.Error("DetectForConnection() should return nil info when fallback is disabled and detection fails")
	}
}

func TestService_DetectForConnection_IconError(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractorWithError{}
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
	info, err := service.DetectForConnection(ctx, 8080)

	if err != nil {
		t.Errorf("DetectForConnection() should not return error when icon extraction fails, got: %v", err)
	}

	if info == nil {
		t.Error("DetectForConnection() should return info even when icon extraction fails")
	}

	if info != nil && info.Icon != nil {
		t.Error("DetectForConnection() should not set icon when extraction fails")
	}
}

type mockHelperClientWithError struct {
	mockHelperClient
}

func (m *mockHelperClientWithError) IsRunning() bool {
	return true
}

func (m *mockHelperClientWithError) DetectProcess(port uint32) (*domain.ProcessInfo, error) {
	return nil, errors.New("helper detection failed")
}

type mockIconExtractorWithError struct {
	mockIconExtractor
}

func (m *mockIconExtractorWithError) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	return nil, errors.New("icon extraction failed")
}

func TestService_getIcon_WithCache(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepositoryWithCache{}
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
	icon, err := service.getIcon(ctx, 1234, "/usr/bin/test")

	if err != nil {
		t.Errorf("getIcon() error = %v, want nil", err)
	}

	if icon == nil {
		t.Error("getIcon() should return cached icon")
	}
}

func TestService_getIcon_WithoutCache(t *testing.T) {
	config := &mockConfigRepository{
		loadFunc: func(ctx context.Context) (*domain.DetectionConfig, error) {
			return &domain.DetectionConfig{
				CacheEnabled: false,
			}, nil
		},
	}
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
	icon, err := service.getIcon(ctx, 1234, "/usr/bin/test")

	if err != nil {
		t.Errorf("getIcon() error = %v, want nil", err)
	}

	if icon == nil {
		t.Error("getIcon() should return icon")
	}
}

func TestService_getIcon_CacheSetError(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepositoryWithSetError{}
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
	icon, err := service.getIcon(ctx, 1234, "/usr/bin/test")

	if err != nil {
		t.Errorf("getIcon() should not return error even if cache set fails, got: %v", err)
	}

	if icon == nil {
		t.Error("getIcon() should return icon even if cache set fails")
	}
}

func TestService_extractIcon_WithHelper(t *testing.T) {
	config := &mockConfigRepository{
		loadFunc: func(ctx context.Context) (*domain.DetectionConfig, error) {
			return &domain.DetectionConfig{
				UseHelperTool: true,
			}, nil
		},
	}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClientWithIcon{}
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
	icon, err := service.extractIcon(ctx, 1234, "/usr/bin/test")

	if err != nil {
		t.Errorf("extractIcon() error = %v, want nil", err)
	}

	if icon == nil {
		t.Error("extractIcon() should return icon from helper")
	}
}

func TestService_extractIcon_HelperError(t *testing.T) {
	config := &mockConfigRepository{
		loadFunc: func(ctx context.Context) (*domain.DetectionConfig, error) {
			return &domain.DetectionConfig{
				UseHelperTool: true,
			}, nil
		},
	}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClientWithIconError{}
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
	icon, err := service.extractIcon(ctx, 1234, "/usr/bin/test")

	if err != nil {
		t.Errorf("extractIcon() error = %v, want nil (should fallback to local)", err)
	}

	if icon == nil {
		t.Error("extractIcon() should return icon from local extractor")
	}
}

func TestService_SaveConfig_ValidationError(t *testing.T) {
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
		CacheTTLSeconds: -1,
	}

	ctx := context.Background()
	err := service.SaveConfig(ctx, cfg)

	if err == nil {
		t.Error("SaveConfig() should return error when validation fails")
	}
}

func TestService_CheckHelperStatus_WithInstalled(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClientWithRunning{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstallerInstalled{}
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

	if !status.Installed {
		t.Error("CheckHelperStatus() should return Installed=true")
	}

	if !status.Running {
		t.Error("CheckHelperStatus() should return Running=true")
	}

	if status.Version == "" {
		t.Error("CheckHelperStatus() should return Version when installed")
	}
}

func TestService_CheckHelperStatus_NilClient(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstaller{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		nil,
		iconExtractor,
		helperInstaller,
		"/path/to/helper",
		&logger,
	)

	status := service.CheckHelperStatus()

	if status.Installed {
		t.Error("CheckHelperStatus() should return Installed=false when client is nil")
	}

	if status.Running {
		t.Error("CheckHelperStatus() should return Running=false when client is nil")
	}
}

func TestService_InstallHelper_NoInstaller(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	logger := zerolog.Nop()

	service := NewService(
		config,
		iconCache,
		localDetector,
		helperClient,
		iconExtractor,
		nil,
		"/path/to/helper",
		&logger,
	)

	ctx := context.Background()
	err := service.InstallHelper(ctx)

	if err == nil {
		t.Error("InstallHelper() should return error when installer is nil")
	}
}

func TestService_InstallHelper_NoBinaryPath(t *testing.T) {
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
		"",
		&logger,
	)

	ctx := context.Background()
	err := service.InstallHelper(ctx)

	if err == nil {
		t.Error("InstallHelper() should return error when binary path is empty")
	}
}

func TestService_InstallHelper_InstallError(t *testing.T) {
	config := &mockConfigRepository{}
	iconCache := &mockIconCacheRepository{}
	localDetector := &mockProcessDetector{}
	helperClient := &mockHelperClient{}
	iconExtractor := &mockIconExtractor{}
	helperInstaller := &mockHelperInstallerWithError{}
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

	if err == nil {
		t.Error("InstallHelper() should return error when installation fails")
	}
}

type mockIconCacheRepositoryWithCache struct {
	mockIconCacheRepository
}

func (m *mockIconCacheRepositoryWithCache) Get(key string) (*domain.AppIcon, error) {
	return &domain.AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3},
	}, nil
}

type mockIconCacheRepositoryWithSetError struct {
	mockIconCacheRepository
}

func (m *mockIconCacheRepositoryWithSetError) Set(key string, icon *domain.AppIcon, ttl time.Duration) error {
	return errors.New("cache set failed")
}

type mockHelperClientWithIcon struct {
	mockHelperClient
}

func (m *mockHelperClientWithIcon) IsRunning() bool {
	return true
}

func (m *mockHelperClientWithIcon) ExtractIcon(pid int32) (*domain.AppIcon, error) {
	return &domain.AppIcon{
		Format: "png",
		Data:   []byte{4, 5, 6},
	}, nil
}

type mockHelperClientWithIconError struct {
	mockHelperClient
}

func (m *mockHelperClientWithIconError) IsRunning() bool {
	return true
}

func (m *mockHelperClientWithIconError) ExtractIcon(pid int32) (*domain.AppIcon, error) {
	return nil, errors.New("helper icon extraction failed")
}

type mockHelperInstallerInstalled struct {
	mockHelperInstaller
}

func (m *mockHelperInstallerInstalled) IsInstalled() bool {
	return true
}

type mockHelperInstallerWithError struct {
	mockHelperInstaller
}

func (m *mockHelperInstallerWithError) Install(helperBinaryPath string) error {
	return errors.New("installation failed")
}
