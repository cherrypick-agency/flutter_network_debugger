package httpapi

import (
	"bytes"
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

type mockCacheManagerForInstallation struct {
	cacheDir  string
	cached    map[string]bool
	paths     map[string]string
	sizes     map[string]int64
	cacheSize int64
	errors    map[string]error
}

func (m *mockCacheManagerForInstallation) GetCacheDir() string {
	if m.cacheDir != "" {
		return m.cacheDir
	}
	return "/tmp/cache"
}

func (m *mockCacheManagerForInstallation) GetCompilerPath(language string) (string, error) {
	if m.errors != nil && m.errors["GetCompilerPath"] != nil {
		return "", m.errors["GetCompilerPath"]
	}
	if path, ok := m.paths[language]; ok {
		return path, nil
	}
	return "", errors.New("compiler not found")
}

func (m *mockCacheManagerForInstallation) EnsureCacheDir() error {
	if m.errors != nil && m.errors["EnsureCacheDir"] != nil {
		return m.errors["EnsureCacheDir"]
	}
	return nil
}

func (m *mockCacheManagerForInstallation) Clear(language string) error {
	if m.errors != nil && m.errors["Clear"] != nil {
		return m.errors["Clear"]
	}
	if m.cached != nil {
		delete(m.cached, language)
	}
	return nil
}

func (m *mockCacheManagerForInstallation) ClearAll() error {
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

func (m *mockCacheManagerForInstallation) GetCacheSize() (int64, error) {
	if m.errors != nil && m.errors["GetCacheSize"] != nil {
		return 0, m.errors["GetCacheSize"]
	}
	return m.cacheSize, nil
}

func (m *mockCacheManagerForInstallation) GetCompilerSize(language string) (int64, error) {
	if m.errors != nil && m.errors["GetCompilerSize"] != nil {
		return 0, m.errors["GetCompilerSize"]
	}
	if size, ok := m.sizes[language]; ok {
		return size, nil
	}
	return 0, nil
}

func (m *mockCacheManagerForInstallation) IsCompilerCached(language string) bool {
	if m.cached == nil {
		return false
	}
	return m.cached[language]
}

type mockCompilerDownloaderForInstallation struct {
	language      string
	metadata      *domain.CompilerMetadata
	downloadError error
}

func (m *mockCompilerDownloaderForInstallation) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	return "https://example.com/download", nil
}

func (m *mockCompilerDownloaderForInstallation) Download(ctx context.Context, req domain.DownloadRequest, progressCb func(domain.DownloadProgress)) error {
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

func (m *mockCompilerDownloaderForInstallation) Extract(archivePath, targetDir string) error {
	return nil
}

func (m *mockCompilerDownloaderForInstallation) Verify(installPath string) error {
	return nil
}

func (m *mockCompilerDownloaderForInstallation) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	if m.metadata != nil {
		return m.metadata, nil
	}
	return &domain.CompilerMetadata{
		Language: m.language,
		Size:     100 * 1024 * 1024,
	}, nil
}

func setupCompilerInstallationHandlers(service *usecase.CompilerInstallationService) *CompilerInstallationHandlers {
	logger := zerolog.Nop()
	return NewCompilerInstallationHandlers(service, logger)
}

func TestCompilerInstallationHandlers_GetCompilersStatus_Success(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached:    make(map[string]bool),
		paths:     make(map[string]string),
		sizes:     make(map[string]int64),
		cacheSize: 200 * 1024 * 1024,
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader1 := &mockCompilerDownloaderForInstallation{language: "rust"}
	downloader2 := &mockCompilerDownloaderForInstallation{language: "go"}

	service.RegisterDownloader("rust", downloader1)
	service.RegisterDownloader("go", downloader2)

	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/status", nil)
	w := httptest.NewRecorder()

	handlers.GetCompilersStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response CompilersListResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response.Compilers) != 2 {
		t.Errorf("Expected 2 compilers, got %d", len(response.Compilers))
	}

	if response.CacheSize != 200*1024*1024 {
		t.Errorf("Expected cache size %d, got %d", 200*1024*1024, response.CacheSize)
	}
}

func TestCompilerInstallationHandlers_GetCompilersStatus_ServiceError(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		errors: map[string]error{
			"GetCacheSize": errors.New("cache error"),
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/status", nil)
	w := httptest.NewRecorder()

	handlers.GetCompilersStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompilerInstallationHandlers_GetCompilerStatus_Success(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: map[string]bool{
			"rust": true,
		},
		paths: map[string]string{
			"rust": "/cache/rust",
		},
		sizes: map[string]int64{
			"rust": 50 * 1024 * 1024,
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloaderForInstallation{language: "rust"}
	service.RegisterDownloader("rust", downloader)

	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/rust/status", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	handlers.GetCompilerStatus(w, req)

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
	if response.Status != string(domain.CompilerStatusInstalled) {
		t.Errorf("Expected status '%s', got '%s'", domain.CompilerStatusInstalled, response.Status)
	}
}

func TestCompilerInstallationHandlers_GetCompilerStatus_MissingLanguage(t *testing.T) {
	service := usecase.NewCompilerInstallationService(&mockCacheManagerForInstallation{})
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers//status", nil)
	w := httptest.NewRecorder()

	handlers.GetCompilerStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_GetCompilerStatus_NotFound(t *testing.T) {
	service := usecase.NewCompilerInstallationService(&mockCacheManagerForInstallation{})
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/nonexistent/status", nil)
	req.SetPathValue("language", "nonexistent")
	w := httptest.NewRecorder()

	handlers.GetCompilerStatus(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCompilerInstallationHandlers_InstallCompiler_Success(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: make(map[string]bool),
		paths:  make(map[string]string),
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloaderForInstallation{
		language:      "rust",
		downloadError: nil,
	}
	service.RegisterDownloader("rust", downloader)

	handlers := setupCompilerInstallationHandlers(service)

	reqBody := map[string]interface{}{
		"version": "latest",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/compilers/rust/install", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	handlers.InstallCompiler(w, req)

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

func TestCompilerInstallationHandlers_InstallCompiler_AlreadyInstalled(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: map[string]bool{
			"rust": true,
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloaderForInstallation{language: "rust"}
	service.RegisterDownloader("rust", downloader)

	handlers := setupCompilerInstallationHandlers(service)

	reqBody := map[string]interface{}{
		"version": "latest",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/compilers/rust/install", bytes.NewReader(body))
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	handlers.InstallCompiler(w, req)

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
	service := usecase.NewCompilerInstallationService(&mockCacheManagerForInstallation{})
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/compilers//install", nil)
	w := httptest.NewRecorder()

	handlers.InstallCompiler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_InstallCompiler_DefaultVersion(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: make(map[string]bool),
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloaderForInstallation{language: "rust"}
	service.RegisterDownloader("rust", downloader)

	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodPost, "/_api/v1/compilers/rust/install", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	handlers.InstallCompiler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCompilerInstallationHandlers_UninstallCompiler_Success(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: map[string]bool{
			"rust": true,
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloaderForInstallation{language: "rust"}
	service.RegisterDownloader("rust", downloader)

	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/rust", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	handlers.UninstallCompiler(w, req)

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
	service := usecase.NewCompilerInstallationService(&mockCacheManagerForInstallation{})
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/", nil)
	w := httptest.NewRecorder()

	handlers.UninstallCompiler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_UninstallCompiler_Error(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: map[string]bool{
			"rust": true,
		},
		errors: map[string]error{
			"Clear": errors.New("uninstall failed"),
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)

	downloader := &mockCompilerDownloaderForInstallation{language: "rust"}
	service.RegisterDownloader("rust", downloader)

	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/rust", nil)
	req.SetPathValue("language", "rust")
	w := httptest.NewRecorder()

	handlers.UninstallCompiler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCompilerInstallationHandlers_GetSupportedLanguages_Success(t *testing.T) {
	service := usecase.NewCompilerInstallationService(&mockCacheManagerForInstallation{})

	downloader1 := &mockCompilerDownloaderForInstallation{language: "rust"}
	downloader2 := &mockCompilerDownloaderForInstallation{language: "go"}

	service.RegisterDownloader("rust", downloader1)
	service.RegisterDownloader("go", downloader2)

	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/supported", nil)
	w := httptest.NewRecorder()

	handlers.GetSupportedLanguages(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string][]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(response["languages"]) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(response["languages"]))
	}
}

func TestCompilerInstallationHandlers_GetCacheSize_Success(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cacheSize: 500 * 1024 * 1024,
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/cache/size", nil)
	w := httptest.NewRecorder()

	handlers.GetCacheSize(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]int64
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["size"] != 500*1024*1024 {
		t.Errorf("Expected cache size %d, got %d", 500*1024*1024, response["size"])
	}
}

func TestCompilerInstallationHandlers_GetCacheSize_Error(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		errors: map[string]error{
			"GetCacheSize": errors.New("cache error"),
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodGet, "/_api/v1/compilers/cache/size", nil)
	w := httptest.NewRecorder()

	handlers.GetCacheSize(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCompilerInstallationHandlers_ClearCache_Success(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		cached: map[string]bool{
			"rust": true,
			"go":   true,
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/cache", nil)
	w := httptest.NewRecorder()

	handlers.ClearCache(w, req)

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

func TestCompilerInstallationHandlers_ClearCache_Error(t *testing.T) {
	cacheManager := &mockCacheManagerForInstallation{
		errors: map[string]error{
			"ClearAll": errors.New("clear failed"),
		},
	}
	service := usecase.NewCompilerInstallationService(cacheManager)
	handlers := setupCompilerInstallationHandlers(service)

	req := httptest.NewRequest(http.MethodDelete, "/_api/v1/compilers/cache", nil)
	w := httptest.NewRecorder()

	handlers.ClearCache(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
