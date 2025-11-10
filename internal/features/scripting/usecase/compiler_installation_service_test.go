package usecase

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewCompilerInstallationService(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	if service == nil {
		t.Fatal("NewCompilerInstallationService returned nil")
	}

	if service.cacheManager != cacheManager {
		t.Error("CacheManager not set correctly")
	}

	if service.downloaders == nil {
		t.Error("Downloaders map not initialized")
	}

	if service.installing == nil {
		t.Error("Installing map not initialized")
	}
}

// Composer 1.
func TestCompilerInstallationService_RegisterDownloader(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{language: "zig"}
	service.RegisterDownloader("zig", downloader)

	if len(service.downloaders) != 1 {
		t.Errorf("Expected 1 downloader, got %d", len(service.downloaders))
	}

	if service.downloaders["zig"] != downloader {
		t.Error("Downloader not registered correctly")
	}
}

// Composer 1.
func TestCompilerInstallationService_GetStatus_NotInstalled(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{
		language: "zig",
		metadata: &domain.CompilerMetadata{
			Size: 100 * 1024 * 1024,
		},
	}

	service.RegisterDownloader("zig", downloader)

	info, err := service.GetStatus("zig")

	if err != nil {
		t.Errorf("GetStatus() error = %v, want nil", err)
	}

	if info == nil {
		t.Fatal("GetStatus() returned nil info")
	}

	if info.Status != domain.CompilerStatusNotInstalled {
		t.Errorf("Status = %v, want %v", info.Status, domain.CompilerStatusNotInstalled)
	}

	if info.DownloadSize != 100*1024*1024 {
		t.Errorf("DownloadSize = %d, want %d", info.DownloadSize, 100*1024*1024)
	}
}

// Composer 1.
func TestCompilerInstallationService_GetStatus_Installed(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: map[string]bool{
			"zig": true,
		},
		paths: map[string]string{
			"zig": "/cache/zig",
		},
		sizes: map[string]int64{
			"zig": 50 * 1024 * 1024,
		},
	}

	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{language: "zig"}
	service.RegisterDownloader("zig", downloader)

	info, err := service.GetStatus("zig")

	if err != nil {
		t.Errorf("GetStatus() error = %v, want nil", err)
	}

	if info.Status != domain.CompilerStatusInstalled {
		t.Errorf("Status = %v, want %v", info.Status, domain.CompilerStatusInstalled)
	}

	if info.InstalledPath != "/cache/zig" {
		t.Errorf("InstalledPath = %q, want %q", info.InstalledPath, "/cache/zig")
	}
}

// Composer 1.
func TestCompilerInstallationService_GetStatus_UnsupportedLanguage(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	_, err := service.GetStatus("nonexistent")

	if err == nil {
		t.Error("GetStatus() should return error for unsupported language")
	}
}

// Composer 1.
func TestCompilerInstallationService_ListAll(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	downloader1 := &mockCompilerDownloader{language: "zig"}
	downloader2 := &mockCompilerDownloader{language: "kotlin"}

	service.RegisterDownloader("zig", downloader1)
	service.RegisterDownloader("kotlin", downloader2)

	list, err := service.ListAll()

	if err != nil {
		t.Errorf("ListAll() error = %v, want nil", err)
	}

	if len(list) != 2 {
		t.Errorf("ListAll() returned %d compilers, want 2", len(list))
	}
}

// Composer 1.
func TestCompilerInstallationService_Install_AlreadyInstalled(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: map[string]bool{
			"zig": true,
		},
	}

	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{language: "zig"}
	service.RegisterDownloader("zig", downloader)

	progressCalled := false
	progressCb := func(progress domain.DownloadProgress) {
		progressCalled = true
		if progress.Stage != "complete" {
			t.Errorf("Progress stage = %q, want %q", progress.Stage, "complete")
		}
	}

	ctx := context.Background()
	err := service.Install(ctx, "zig", "latest", progressCb)

	if err != nil {
		t.Errorf("Install() error = %v, want nil", err)
	}

	if !progressCalled {
		t.Error("Progress callback should be called when already installed")
	}
}

// Composer 1.
func TestCompilerInstallationService_Install_UnsupportedLanguage(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	ctx := context.Background()
	err := service.Install(ctx, "nonexistent", "latest", nil)

	if err == nil {
		t.Error("Install() should return error for unsupported language")
	}
}

// Composer 1.
func TestCompilerInstallationService_Install_AlreadyInstalling(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{language: "zig"}
	service.RegisterDownloader("zig", downloader)

	service.mutex.Lock()
	service.installing["zig"] = true
	service.mutex.Unlock()

	ctx := context.Background()
	err := service.Install(ctx, "zig", "latest", nil)

	if err == nil {
		t.Error("Install() should return error when already installing")
	}
}

// Composer 1.
func TestCompilerInstallationService_Install_Success(t *testing.T) {
	tmpDir := t.TempDir()
	cacheManager := &mockCacheManager{
		cacheDir: tmpDir,
		cached:   make(map[string]bool),
		paths:    make(map[string]string),
	}

	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{
		language:      "zig",
		downloadError: nil,
	}

	service.RegisterDownloader("zig", downloader)

	progressCalls := []domain.DownloadProgress{}
	progressCb := func(progress domain.DownloadProgress) {
		progressCalls = append(progressCalls, progress)
	}

	ctx := context.Background()
	err := service.Install(ctx, "zig", "latest", progressCb)

	if err != nil {
		t.Errorf("Install() error = %v, want nil", err)
	}

	if len(progressCalls) == 0 {
		t.Error("Progress callback should be called during installation")
	}
}

// Composer 1.
func TestCompilerInstallationService_Install_DownloadError(t *testing.T) {
	tmpDir := t.TempDir()
	cacheManager := &mockCacheManager{
		cacheDir: tmpDir,
		cached:   make(map[string]bool),
		paths:    make(map[string]string),
	}

	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{
		language:      "zig",
		downloadError: errors.New("download failed"),
	}

	service.RegisterDownloader("zig", downloader)

	ctx := context.Background()
	err := service.Install(ctx, "zig", "latest", nil)

	if err == nil {
		t.Error("Install() should return error when download fails")
	}
}

// Composer 1.
func TestCompilerInstallationService_Uninstall_Success(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: map[string]bool{
			"zig": true,
		},
	}

	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{language: "zig"}
	service.RegisterDownloader("zig", downloader)

	err := service.Uninstall("zig")

	if err != nil {
		t.Errorf("Uninstall() error = %v, want nil", err)
	}
}

// Composer 1.
func TestCompilerInstallationService_Uninstall_NotInstalled(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: make(map[string]bool),
	}

	service := NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloader{language: "zig"}
	service.RegisterDownloader("zig", downloader)

	err := service.Uninstall("zig")

	if err == nil {
		t.Error("Uninstall() should return error when compiler not installed")
	}
}

// Composer 1.
func TestCompilerInstallationService_Uninstall_UnsupportedLanguage(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	err := service.Uninstall("nonexistent")

	if err == nil {
		t.Error("Uninstall() should return error for unsupported language")
	}
}

// Composer 1.
func TestCompilerInstallationService_UninstallAll(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: map[string]bool{
			"zig":    true,
			"kotlin": true,
		},
	}

	service := NewCompilerInstallationService(cacheManager)

	err := service.UninstallAll()

	if err != nil {
		t.Errorf("UninstallAll() error = %v, want nil", err)
	}
}

// Composer 1.
func TestCompilerInstallationService_GetCacheSize(t *testing.T) {
	cacheManager := &mockCacheManager{
		cacheSize: 1024 * 1024,
	}

	service := NewCompilerInstallationService(cacheManager)

	size, err := service.GetCacheSize()

	if err != nil {
		t.Errorf("GetCacheSize() error = %v, want nil", err)
	}

	if size != 1024*1024 {
		t.Errorf("GetCacheSize() = %d, want %d", size, 1024*1024)
	}
}

// Composer 1.
func TestCompilerInstallationService_IsCompilerInstalled(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: map[string]bool{
			"zig": true,
		},
	}

	service := NewCompilerInstallationService(cacheManager)

	if !service.IsCompilerInstalled("zig") {
		t.Error("IsCompilerInstalled('zig') should return true")
	}

	if service.IsCompilerInstalled("kotlin") {
		t.Error("IsCompilerInstalled('kotlin') should return false")
	}
}

// Composer 1.
func TestCompilerInstallationService_GetCompilerPath_Success(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: map[string]bool{
			"zig": true,
		},
		paths: map[string]string{
			"zig": "/cache/zig",
		},
	}

	service := NewCompilerInstallationService(cacheManager)

	path, err := service.GetCompilerPath("zig")

	if err != nil {
		t.Errorf("GetCompilerPath() error = %v, want nil", err)
	}

	if path != "/cache/zig" {
		t.Errorf("GetCompilerPath() = %q, want %q", path, "/cache/zig")
	}
}

// Composer 1.
func TestCompilerInstallationService_GetCompilerPath_NotInstalled(t *testing.T) {
	cacheManager := &mockCacheManager{
		cached: make(map[string]bool),
	}

	service := NewCompilerInstallationService(cacheManager)

	_, err := service.GetCompilerPath("zig")

	if err == nil {
		t.Error("GetCompilerPath() should return error when compiler not installed")
	}
}

// Composer 1.
func TestCompilerInstallationService_GetSupportedLanguages(t *testing.T) {
	cacheManager := &mockCacheManager{}
	service := NewCompilerInstallationService(cacheManager)

	downloader1 := &mockCompilerDownloader{language: "zig"}
	downloader2 := &mockCompilerDownloader{language: "kotlin"}

	service.RegisterDownloader("zig", downloader1)
	service.RegisterDownloader("kotlin", downloader2)

	languages := service.GetSupportedLanguages()

	if len(languages) != 2 {
		t.Errorf("GetSupportedLanguages() returned %d languages, want 2", len(languages))
	}

	hasZig := false
	hasKotlin := false
	for _, lang := range languages {
		if lang == "zig" {
			hasZig = true
		}
		if lang == "kotlin" {
			hasKotlin = true
		}
	}

	if !hasZig {
		t.Error("GetSupportedLanguages() should include zig")
	}

	if !hasKotlin {
		t.Error("GetSupportedLanguages() should include kotlin")
	}
}

// Composer 1.
func TestSaveCompilerVersion(t *testing.T) {
	tmpDir := t.TempDir()
	compilerDir := filepath.Join(tmpDir, "zig")

	err := os.MkdirAll(compilerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler dir: %v", err)
	}

	err = saveCompilerVersion(compilerDir, "0.14.0")
	if err != nil {
		t.Errorf("saveCompilerVersion() error = %v, want nil", err)
	}

	versionFile := filepath.Join(compilerDir, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("Failed to read version file: %v", err)
	}

	if string(data) != "0.14.0" {
		t.Errorf("Version file content = %q, want %q", string(data), "0.14.0")
	}
}

// Composer 1.
func TestGetCompilerVersion(t *testing.T) {
	tmpDir := t.TempDir()
	compilerDir := filepath.Join(tmpDir, "zig")

	err := os.MkdirAll(compilerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler dir: %v", err)
	}

	versionFile := filepath.Join(compilerDir, ".version")
	err = os.WriteFile(versionFile, []byte("0.14.0\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write version file: %v", err)
	}

	version := getCompilerVersion(compilerDir)
	if version != "0.14.0" {
		t.Errorf("getCompilerVersion() = %q, want %q", version, "0.14.0")
	}
}

// Composer 1.
func TestGetCompilerVersion_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	compilerDir := filepath.Join(tmpDir, "nonexistent")

	version := getCompilerVersion(compilerDir)
	if version != "unknown" {
		t.Errorf("getCompilerVersion() = %q, want %q", version, "unknown")
	}
}

// Mock implementations
type mockCacheManager struct {
	cacheDir  string
	cached    map[string]bool
	paths     map[string]string
	sizes     map[string]int64
	cacheSize int64
	errors    map[string]error
}

func (m *mockCacheManager) GetCacheDir() string {
	if m.cacheDir != "" {
		return m.cacheDir
	}
	return "/tmp/cache"
}

func (m *mockCacheManager) GetCompilerPath(language string) (string, error) {
	if m.errors != nil && m.errors["GetCompilerPath"] != nil {
		return "", m.errors["GetCompilerPath"]
	}
	if path, ok := m.paths[language]; ok {
		return path, nil
	}
	return "", errors.New("compiler not found")
}

func (m *mockCacheManager) EnsureCacheDir() error {
	if m.errors != nil && m.errors["EnsureCacheDir"] != nil {
		return m.errors["EnsureCacheDir"]
	}
	return nil
}

func (m *mockCacheManager) Clear(language string) error {
	if m.errors != nil && m.errors["Clear"] != nil {
		return m.errors["Clear"]
	}
	if m.cached != nil {
		delete(m.cached, language)
	}
	return nil
}

func (m *mockCacheManager) ClearAll() error {
	if m.errors != nil && m.errors["ClearAll"] != nil {
		return m.errors["ClearAll"]
	}
	if m.cached != nil {
		for k := range m.cached {
			delete(m.cached, k)
		}
	}
	return nil
}

func (m *mockCacheManager) GetCacheSize() (int64, error) {
	if m.errors != nil && m.errors["GetCacheSize"] != nil {
		return 0, m.errors["GetCacheSize"]
	}
	return m.cacheSize, nil
}

func (m *mockCacheManager) GetCompilerSize(language string) (int64, error) {
	if m.errors != nil && m.errors["GetCompilerSize"] != nil {
		return 0, m.errors["GetCompilerSize"]
	}
	if size, ok := m.sizes[language]; ok {
		return size, nil
	}
	return 0, nil
}

func (m *mockCacheManager) IsCompilerCached(language string) bool {
	if m.cached == nil {
		return false
	}
	return m.cached[language]
}

type mockCompilerDownloader struct {
	language      string
	metadata      *domain.CompilerMetadata
	downloadError error
}

func (m *mockCompilerDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	return "https://example.com/download", nil
}

func (m *mockCompilerDownloader) Download(ctx context.Context, req domain.DownloadRequest, progressCb func(domain.DownloadProgress)) error {
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "downloading",
			Percentage: 50,
			Message:    "Downloading...",
		})
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    "Download complete",
		})
	}
	return m.downloadError
}

func (m *mockCompilerDownloader) Extract(archivePath, targetDir string) error {
	return nil
}

func (m *mockCompilerDownloader) Verify(installPath string) error {
	return nil
}

func (m *mockCompilerDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	if m.metadata != nil {
		return m.metadata, nil
	}
	return &domain.CompilerMetadata{
		Language: m.language,
		Size:     100 * 1024 * 1024,
	}, nil
}
