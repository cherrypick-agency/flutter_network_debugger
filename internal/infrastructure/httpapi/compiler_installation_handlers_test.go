package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

// Composer 1.
// Mock CacheManager для тестирования
type mockCacheManager struct {
	getCacheDirFunc      func() string
	getCompilerPathFunc  func(language string) (string, error)
	isCompilerCachedFunc func(language string) bool
	getCacheSizeFunc     func() (int64, error)
	getCompilerSizeFunc  func(language string) (int64, error)
	clearFunc            func(language string) error
	clearAllFunc         func() error
}

func (m *mockCacheManager) GetCacheDir() string {
	if m.getCacheDirFunc != nil {
		return m.getCacheDirFunc()
	}
	return "/tmp/cache"
}

func (m *mockCacheManager) GetCompilerPath(language string) (string, error) {
	if m.getCompilerPathFunc != nil {
		return m.getCompilerPathFunc(language)
	}
	return "/tmp/cache/" + language, nil
}

func (m *mockCacheManager) EnsureCacheDir() error {
	return nil
}

func (m *mockCacheManager) Clear(language string) error {
	if m.clearFunc != nil {
		return m.clearFunc(language)
	}
	return nil
}

func (m *mockCacheManager) ClearAll() error {
	if m.clearAllFunc != nil {
		return m.clearAllFunc()
	}
	return nil
}

func (m *mockCacheManager) GetCacheSize() (int64, error) {
	if m.getCacheSizeFunc != nil {
		return m.getCacheSizeFunc()
	}
	return 0, nil
}

func (m *mockCacheManager) GetCompilerSize(language string) (int64, error) {
	if m.getCompilerSizeFunc != nil {
		return m.getCompilerSizeFunc(language)
	}
	return 0, nil
}

func (m *mockCacheManager) IsCompilerCached(language string) bool {
	if m.isCompilerCachedFunc != nil {
		return m.isCompilerCachedFunc(language)
	}
	return false
}

// Mock CompilerDownloader для тестирования
type mockCompilerDownloader struct {
	language        string
	getMetadataFunc func(req domain.DownloadRequest) (*domain.CompilerMetadata, error)
	downloadFunc    func(ctx context.Context, req domain.DownloadRequest, progressCb func(domain.DownloadProgress)) error
}

func (m *mockCompilerDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	return "https://example.com/download", nil
}

func (m *mockCompilerDownloader) Download(ctx context.Context, req domain.DownloadRequest, progressCb func(domain.DownloadProgress)) error {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, req, progressCb)
	}
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "downloading",
			Percentage: 50.0,
			Message:    "Downloading...",
		})
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100.0,
			Message:    "Complete",
		})
	}
	return nil
}

func (m *mockCompilerDownloader) Extract(archivePath, targetDir string) error {
	return nil
}

func (m *mockCompilerDownloader) Verify(installPath string) error {
	return nil
}

func (m *mockCompilerDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	if m.getMetadataFunc != nil {
		return m.getMetadataFunc(req)
	}
	return &domain.CompilerMetadata{
		Language: req.Language,
		Version:  req.Version,
		Size:     1000,
	}, nil
}

func setupCompilerInstallationHandlers() (*CompilerInstallationHandlers, *mockCacheManager) {
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := NewCompilerInstallationHandlers(service, logger)
	return handlers, cacheManager
}

func TestCompilerInstallationHandlers_GetCompilersStatus_Success(t *testing.T) {
	// Arrange
	handlers, cacheManager := setupCompilerInstallationHandlers()
	cacheManager.getCacheSizeFunc = func() (int64, error) {
		return 1000, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/status", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.GetCompilersStatus(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompilersListResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.CacheSize != 1000 {
		t.Errorf("Expected cacheSize 1000, got %d", response.CacheSize)
	}
}

func TestCompilerInstallationHandlers_GetCompilersStatus_ServiceError(t *testing.T) {
	// Arrange
	handlers, cacheManager := setupCompilerInstallationHandlers()
	cacheManager.getCacheSizeFunc = func() (int64, error) {
		return 0, errors.New("cache error")
	}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/status", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.GetCompilersStatus(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompilerInstallationHandlers_GetCompilerStatus_Success(t *testing.T) {
	// Arrange
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{
		isCompilerCachedFunc: func(language string) bool {
			return language == "rust"
		},
		getCompilerPathFunc: func(language string) (string, error) {
			if language == "rust" {
				return "/tmp/cache/rust", nil
			}
			return "", errors.New("not found")
		},
		getCompilerSizeFunc: func(language string) (int64, error) {
			return 500, nil
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	downloader := &mockCompilerDownloader{language: "rust"}
	service.RegisterDownloader("rust", downloader)
	handlers := NewCompilerInstallationHandlers(service, logger)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/rust/status", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	// Act
	handlers.GetCompilerStatus(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompilerStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Language != "rust" {
		t.Errorf("Expected language 'rust', got '%s'", response.Language)
	}
}

func TestCompilerInstallationHandlers_GetCompilerStatus_NotFound(t *testing.T) {
	// Arrange
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := NewCompilerInstallationHandlers(service, logger)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/unknown/status", nil)
	req.SetPathValue("language", "unknown")
	w := httptest.NewRecorder()

	// Act
	handlers.GetCompilerStatus(w, req)

	// Assert
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCompilerInstallationHandlers_GetCompilerStatus_MissingLanguage(t *testing.T) {
	// Arrange
	handlers, _ := setupCompilerInstallationHandlers()

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers//status", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.GetCompilerStatus(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_InstallCompiler_AlreadyInstalled(t *testing.T) {
	// Arrange
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{
		isCompilerCachedFunc: func(language string) bool {
			return language == "rust"
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	downloader := &mockCompilerDownloader{language: "rust"}
	service.RegisterDownloader("rust", downloader)
	handlers := NewCompilerInstallationHandlers(service, logger)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/compilers/rust/install", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	// Act
	handlers.InstallCompiler(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response InstallCompilerResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "already_installed" {
		t.Errorf("Expected status 'already_installed', got '%s'", response.Status)
	}
}

func TestCompilerInstallationHandlers_InstallCompiler_MissingLanguage(t *testing.T) {
	// Arrange
	handlers, _ := setupCompilerInstallationHandlers()

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/compilers//install", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.InstallCompiler(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_UninstallCompiler_Success(t *testing.T) {
	// Arrange
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{
		clearFunc: func(language string) error {
			return nil
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := NewCompilerInstallationHandlers(service, logger)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/rust", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	// Act
	handlers.UninstallCompiler(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response InstallCompilerResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}
}

func TestCompilerInstallationHandlers_UninstallCompiler_MissingLanguage(t *testing.T) {
	// Arrange
	handlers, _ := setupCompilerInstallationHandlers()

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.UninstallCompiler(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_GetSupportedLanguages_Success(t *testing.T) {
	// Arrange
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{}
	service := usecase.NewCompilerInstallationService(cacheManager)
	downloader1 := &mockCompilerDownloader{language: "rust"}
	downloader2 := &mockCompilerDownloader{language: "go"}
	service.RegisterDownloader("rust", downloader1)
	service.RegisterDownloader("go", downloader2)
	handlers := NewCompilerInstallationHandlers(service, logger)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/supported", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.GetSupportedLanguages(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string][]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	languages, ok := response["languages"]
	if !ok {
		t.Error("Expected 'languages' key in response")
	}
	if len(languages) == 0 {
		t.Error("Expected at least one language")
	}
}

func TestCompilerInstallationHandlers_GetCacheSize_Success(t *testing.T) {
	// Arrange
	handlers, cacheManager := setupCompilerInstallationHandlers()
	cacheManager.getCacheSizeFunc = func() (int64, error) {
		return 5000, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/cache/size", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.GetCacheSize(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]int64
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if size, ok := response["size"]; !ok || size != 5000 {
		t.Errorf("Expected size 5000, got %d", size)
	}
}

func TestCompilerInstallationHandlers_ClearCache_Success(t *testing.T) {
	// Arrange
	logger := zerolog.Nop()
	cacheManager := &mockCacheManager{
		clearAllFunc: func() error {
			return nil
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := NewCompilerInstallationHandlers(service, logger)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/cache", nil)
	w := httptest.NewRecorder()

	// Act
	handlers.ClearCache(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response InstallCompilerResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", response.Status)
	}
}
