package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/rs/zerolog"
	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

// CompilerInstallationHandlers handles HTTP requests for compiler installation
// This is the adapter layer for HTTP (Clean Architecture)
type CompilerInstallationHandlers struct {
	service          *usecase.CompilerInstallationService
	logger           zerolog.Logger
	progressChannels sync.Map // language -> chan domain.DownloadProgress
}

// NewCompilerInstallationHandlers creates a new handlers instance
func NewCompilerInstallationHandlers(
	service *usecase.CompilerInstallationService,
	logger zerolog.Logger,
) *CompilerInstallationHandlers {
	return &CompilerInstallationHandlers{
		service: service,
		logger:  logger,
	}
}

// CompilerStatusResponse represents a single compiler's status
type CompilerStatusResponse struct {
	Language      string `json:"language"`
	Version       string `json:"version"`
	Status        string `json:"status"` // "not_installed", "installing", "installed", "error"
	InstalledPath string `json:"installedPath,omitempty"`
	Size          int64  `json:"size,omitempty"`         // Installed size in bytes
	DownloadSize  int64  `json:"downloadSize,omitempty"` // Size to download in bytes
	Error         string `json:"error,omitempty"`
}

// CompilersListResponse represents list of all compilers
type CompilersListResponse struct {
	Compilers []CompilerStatusResponse `json:"compilers"`
	CacheSize int64                    `json:"cacheSize"` // Total cache size in bytes
}

// InstallCompilerRequest is the request body for installing a compiler
type InstallCompilerRequest struct {
	Version string `json:"version,omitempty"` // "latest" or specific version
}

// InstallCompilerResponse is the response body for installation
type InstallCompilerResponse struct {
	Status  string `json:"status"` // "success", "error", "already_installed"
	Message string `json:"message"`
}

// GetCompilersStatus returns status of all supported compilers
// GET /_api/v1/compilers/status
func (h *CompilerInstallationHandlers) GetCompilersStatus(w http.ResponseWriter, r *http.Request) {
	// Get all compiler statuses
	compilers, err := h.service.ListAll()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to list compilers")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get total cache size
	cacheSize, err := h.service.GetCacheSize()
	if err != nil {
		h.logger.Warn().Err(err).Msg("failed to get cache size")
		cacheSize = 0
	}

	// Convert to response format
	response := CompilersListResponse{
		Compilers: make([]CompilerStatusResponse, 0, len(compilers)),
		CacheSize: cacheSize,
	}

	for _, compiler := range compilers {
		response.Compilers = append(response.Compilers, CompilerStatusResponse{
			Language:      compiler.Language,
			Version:       compiler.Version,
			Status:        string(compiler.Status),
			InstalledPath: compiler.InstalledPath,
			Size:          compiler.Size,
			DownloadSize:  compiler.DownloadSize,
			Error:         compiler.Error,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCompilerStatus returns status of a specific compiler
// GET /_api/v1/compilers/:language/status
func (h *CompilerInstallationHandlers) GetCompilerStatus(w http.ResponseWriter, r *http.Request) {
	language := r.PathValue("language")
	if language == "" {
		http.Error(w, "language required", http.StatusBadRequest)
		return
	}

	// Get compiler status
	info, err := h.service.GetStatus(language)
	if err != nil {
		h.logger.Error().Err(err).Str("language", language).Msg("failed to get compiler status")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := CompilerStatusResponse{
		Language:      info.Language,
		Version:       info.Version,
		Status:        string(info.Status),
		InstalledPath: info.InstalledPath,
		Size:          info.Size,
		DownloadSize:  info.DownloadSize,
		Error:         info.Error,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// InstallCompiler downloads and installs a compiler
// POST /_api/v1/compilers/:language/install
func (h *CompilerInstallationHandlers) InstallCompiler(w http.ResponseWriter, r *http.Request) {
	language := r.PathValue("language")
	if language == "" {
		http.Error(w, "language required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req InstallCompilerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to latest version
		req.Version = "latest"
	}

	if req.Version == "" {
		req.Version = "latest"
	}

	h.logger.Info().Str("language", language).Str("version", req.Version).Msg("installing compiler")

	// Check if already installed
	if h.service.IsCompilerInstalled(language) {
		response := InstallCompilerResponse{
			Status:  "already_installed",
			Message: "Compiler is already installed",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create progress channel for this installation
	progressChan := make(chan domain.DownloadProgress, 10)
	h.progressChannels.Store(language, progressChan)
	defer h.progressChannels.Delete(language)

	// Progress callback
	progressCb := func(progress domain.DownloadProgress) {
		select {
		case progressChan <- progress:
		default:
			// Channel full, skip this progress update
		}

		// Log progress
		h.logger.Debug().
			Str("language", language).
			Str("stage", progress.Stage).
			Float64("percentage", progress.Percentage).
			Str("message", progress.Message).
			Msg("installation progress")
	}

	// Install in goroutine to allow progress updates
	errChan := make(chan error, 1)
	go func() {
		err := h.service.Install(r.Context(), language, req.Version, progressCb)
		errChan <- err
		close(progressChan)
	}()

	// Wait for completion
	err := <-errChan
	if err != nil {
		h.logger.Error().Err(err).Str("language", language).Msg("installation failed")
		response := InstallCompilerResponse{
			Status:  "error",
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Success
	response := InstallCompilerResponse{
		Status:  "success",
		Message: "Compiler installed successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UninstallCompiler removes a compiler from cache
// DELETE /_api/v1/compilers/:language
func (h *CompilerInstallationHandlers) UninstallCompiler(w http.ResponseWriter, r *http.Request) {
	language := r.PathValue("language")
	if language == "" {
		http.Error(w, "language required", http.StatusBadRequest)
		return
	}

	h.logger.Info().Str("language", language).Msg("uninstalling compiler")

	err := h.service.Uninstall(language)
	if err != nil {
		h.logger.Error().Err(err).Str("language", language).Msg("uninstall failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := InstallCompilerResponse{
		Status:  "success",
		Message: "Compiler uninstalled successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetInstallationProgress streams installation progress via SSE
// GET /_api/v1/compilers/:language/progress
func (h *CompilerInstallationHandlers) GetInstallationProgress(w http.ResponseWriter, r *http.Request) {
	language := r.PathValue("language")
	if language == "" {
		http.Error(w, "language required", http.StatusBadRequest)
		return
	}

	// Check if installation is in progress
	progressChanInterface, ok := h.progressChannels.Load(language)
	if !ok {
		http.Error(w, "no installation in progress", http.StatusNotFound)
		return
	}

	progressChan := progressChanInterface.(chan domain.DownloadProgress)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Stream progress updates
	for progress := range progressChan {
		data, err := json.Marshal(progress)
		if err != nil {
			log.Printf("failed to marshal progress: %v", err)
			continue
		}

		// Send SSE message
		_, err = w.Write([]byte("data: "))
		if err != nil {
			return
		}

		_, err = w.Write(data)
		if err != nil {
			return
		}

		_, err = w.Write([]byte("\n\n"))
		if err != nil {
			return
		}

		flusher.Flush()

		// Check if client disconnected
		select {
		case <-r.Context().Done():
			return
		default:
		}
	}
}

// GetSupportedLanguages returns list of supported languages
// GET /_api/v1/compilers/supported
func (h *CompilerInstallationHandlers) GetSupportedLanguages(w http.ResponseWriter, r *http.Request) {
	languages := h.service.GetSupportedLanguages()

	response := map[string][]string{
		"languages": languages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCacheSize returns total size of compiler cache
// GET /_api/v1/compilers/cache/size
func (h *CompilerInstallationHandlers) GetCacheSize(w http.ResponseWriter, r *http.Request) {
	cacheSize, err := h.service.GetCacheSize()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get cache size")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int64{
		"size": cacheSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ClearCache removes all installed compilers from cache
// DELETE /_api/v1/compilers/cache
func (h *CompilerInstallationHandlers) ClearCache(w http.ResponseWriter, r *http.Request) {
	h.logger.Info().Msg("clearing compiler cache")

	err := h.service.UninstallAll()
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to clear cache")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := InstallCompilerResponse{
		Status:  "success",
		Message: "Compiler cache cleared successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
