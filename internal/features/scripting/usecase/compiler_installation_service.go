package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"network-debugger/internal/features/scripting/domain"
)

// CompilerInstallationService orchestrates compiler installation process
// This is the UseCase layer - coordinates between domain and infrastructure
type CompilerInstallationService struct {
	cacheManager domain.CacheManager
	downloaders  map[string]domain.CompilerDownloader
	mutex        sync.RWMutex    // Protects concurrent access
	installing   map[string]bool // Track ongoing installations
}

// NewCompilerInstallationService creates a new installation service
func NewCompilerInstallationService(cacheManager domain.CacheManager) *CompilerInstallationService {
	return &CompilerInstallationService{
		cacheManager: cacheManager,
		downloaders:  make(map[string]domain.CompilerDownloader),
		installing:   make(map[string]bool),
	}
}

// RegisterDownloader registers a downloader for a specific language
func (s *CompilerInstallationService) RegisterDownloader(language string, downloader domain.CompilerDownloader) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.downloaders[language] = downloader
}

// GetStatus returns the current status of a compiler
func (s *CompilerInstallationService) GetStatus(language string) (*domain.CompilerInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if downloader exists
	downloader, ok := s.downloaders[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Check if compiler is in cache
	isInstalled := s.cacheManager.IsCompilerCached(language)

	// Check if currently installing
	isInstalling := s.installing[language]

	// Determine status
	var status domain.CompilerStatus
	if isInstalling {
		status = domain.CompilerStatusInstalling
	} else if isInstalled {
		status = domain.CompilerStatusInstalled
	} else {
		status = domain.CompilerStatusNotInstalled
	}

	// Get compiler path and version if installed
	installedPath := ""
	installedSize := int64(0)
	version := "latest" // Default for not installed compilers
	if isInstalled {
		path, err := s.cacheManager.GetCompilerPath(language)
		if err == nil {
			installedPath = path

			// Get size
			size, err := s.cacheManager.GetCompilerSize(language)
			if err == nil {
				installedSize = size
			}

			// Get actual installed version from .version file
			version = getCompilerVersion(path)
		}
	}

	// Get download size if not installed
	downloadSize := int64(0)
	if !isInstalled {
		req := domain.DownloadRequest{
			Language: language,
			Version:  "latest",
			Platform: runtime.GOOS,
			Arch:     runtime.GOARCH,
		}

		metadata, err := downloader.GetMetadata(req)
		if err == nil {
			downloadSize = metadata.Size
		}
	}

	return &domain.CompilerInfo{
		Language:      language,
		Version:       version, // Actual version from .version file or "latest"/"unknown"
		Status:        status,
		InstalledPath: installedPath,
		Size:          installedSize,
		DownloadSize:  downloadSize,
		Error:         "",
	}, nil
}

// ListAll returns status of all supported compilers
func (s *CompilerInstallationService) ListAll() ([]*domain.CompilerInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []*domain.CompilerInfo

	for language := range s.downloaders {
		info, err := s.GetStatus(language)
		if err != nil {
			// Log error but continue
			continue
		}

		result = append(result, info)
	}

	return result, nil
}

// Install downloads and installs a compiler
// progressCb is called periodically with progress updates
func (s *CompilerInstallationService) Install(
	ctx context.Context,
	language string,
	version string,
	progressCb func(domain.DownloadProgress),
) error {
	// Check if language is supported
	s.mutex.RLock()
	downloader, ok := s.downloaders[language]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported language: %s", language)
	}

	// Check if already installed
	if s.cacheManager.IsCompilerCached(language) {
		if progressCb != nil {
			progressCb(domain.DownloadProgress{
				Stage:      "complete",
				Percentage: 100,
				Message:    fmt.Sprintf("%s compiler already installed", language),
			})
		}
		return nil
	}

	// Check if already installing
	s.mutex.Lock()
	if s.installing[language] {
		s.mutex.Unlock()
		return fmt.Errorf("compiler %s is already being installed", language)
	}

	// Mark as installing
	s.installing[language] = true
	s.mutex.Unlock()

	// Ensure we clean up installing flag
	defer func() {
		s.mutex.Lock()
		delete(s.installing, language)
		s.mutex.Unlock()
	}()

	// Ensure cache directory exists
	if err := s.cacheManager.EnsureCacheDir(); err != nil {
		return fmt.Errorf("failed to ensure cache directory: %w", err)
	}

	// Check available disk space before downloading
	// Minimum required: 1 GB free space (compilers typically 100-500 MB + extraction overhead)
	const minRequiredSpace = 1 * 1024 * 1024 * 1024 // 1 GB
	cacheDir := s.cacheManager.GetCacheDir()
	availableSpace, err := getAvailableDiskSpace(cacheDir)
	if err != nil {
		// Log warning but don't fail - disk space check is best-effort
		if progressCb != nil {
			progressCb(domain.DownloadProgress{
				Stage:   "warning",
				Message: fmt.Sprintf("Warning: Could not check disk space: %v", err),
			})
		}
	} else {
		if availableSpace < minRequiredSpace {
			return fmt.Errorf(
				"insufficient disk space: available %d MB, required minimum %d MB",
				availableSpace/(1024*1024),
				minRequiredSpace/(1024*1024),
			)
		}

		if progressCb != nil {
			progressCb(domain.DownloadProgress{
				Stage: "prepare",
				Message: fmt.Sprintf(
					"Disk space check passed: %d MB available (minimum %d MB required)",
					availableSpace/(1024*1024),
					minRequiredSpace/(1024*1024),
				),
			})
		}
	}

	// Get compiler path in cache
	targetDir, err := s.cacheManager.GetCompilerPath(language)
	if err != nil {
		// Compiler not in cache, create directory
		cacheDir := s.cacheManager.GetCacheDir()
		targetDir = fmt.Sprintf("%s/%s", cacheDir, language)
	}

	// Create download request
	req := domain.DownloadRequest{
		Language:  language,
		Version:   version,
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		TargetDir: targetDir,
	}

	// Download and install
	err = downloader.Download(ctx, req, progressCb)
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Save compiler version to .version file for future reference
	if saveErr := saveCompilerVersion(targetDir, version); saveErr != nil {
		// Log warning but don't fail - version tracking is not critical
		if progressCb != nil {
			progressCb(domain.DownloadProgress{
				Stage:   "warning",
				Message: fmt.Sprintf("Warning: Failed to save compiler version: %v", saveErr),
			})
		}
	}

	return nil
}

// Uninstall removes a compiler from cache
func (s *CompilerInstallationService) Uninstall(language string) error {
	// Check if language is supported
	s.mutex.RLock()
	_, ok := s.downloaders[language]
	s.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("unsupported language: %s", language)
	}

	// Check if installed
	if !s.cacheManager.IsCompilerCached(language) {
		return fmt.Errorf("compiler %s is not installed", language)
	}

	// Remove from cache
	err := s.cacheManager.Clear(language)
	if err != nil {
		return fmt.Errorf("failed to uninstall: %w", err)
	}

	return nil
}

// UninstallAll removes all compilers from cache
func (s *CompilerInstallationService) UninstallAll() error {
	return s.cacheManager.ClearAll()
}

// GetCacheSize returns total size of compiler cache
func (s *CompilerInstallationService) GetCacheSize() (int64, error) {
	return s.cacheManager.GetCacheSize()
}

// IsCompilerInstalled checks if a specific compiler is installed
func (s *CompilerInstallationService) IsCompilerInstalled(language string) bool {
	return s.cacheManager.IsCompilerCached(language)
}

// GetCompilerPath returns the path to installed compiler
func (s *CompilerInstallationService) GetCompilerPath(language string) (string, error) {
	if !s.cacheManager.IsCompilerCached(language) {
		return "", fmt.Errorf("compiler %s is not installed", language)
	}

	return s.cacheManager.GetCompilerPath(language)
}

// GetSupportedLanguages returns list of supported languages
func (s *CompilerInstallationService) GetSupportedLanguages() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	languages := make([]string, 0, len(s.downloaders))
	for lang := range s.downloaders {
		languages = append(languages, lang)
	}

	return languages
}

// saveCompilerVersion saves the compiler version to a .version file in the compiler directory
func saveCompilerVersion(compilerDir, version string) error {
	versionFile := filepath.Join(compilerDir, ".version")
	return os.WriteFile(versionFile, []byte(version), 0644)
}

// getCompilerVersion reads the saved compiler version from .version file
// Returns "unknown" if file doesn't exist or cannot be read
func getCompilerVersion(compilerDir string) string {
	versionFile := filepath.Join(compilerDir, ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(data))
}

// getAvailableDiskSpace returns available disk space in bytes for the given path
// Returns error if unable to determine disk space (platform-specific implementation)
func getAvailableDiskSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, fmt.Errorf("failed to stat filesystem: %w", err)
	}

	// Available space = available blocks * block size
	// Use Bavail (blocks available to unprivileged users) instead of Bfree
	availableBytes := int64(stat.Bavail) * int64(stat.Bsize)

	return availableBytes, nil
}
