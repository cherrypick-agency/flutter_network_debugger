package downloaders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// ZigDownloader implements CompilerDownloader for Zig language
// Zig compiles to WASM with excellent size and performance
type ZigDownloader struct {
	*BaseDownloader
	indexURL string
}

// zigIndex represents the structure of Zig's download index.json
type zigIndex struct {
	Master zigRelease            `json:"master"`
	Stable map[string]zigRelease `json:",inline"` // Other versions are inline
}

type zigRelease struct {
	Version string `json:"version"`
	Targets map[string]zigTarget
}

type zigTarget struct {
	Tarball string `json:"tarball"`
	Shasum  string `json:"shasum"`
	Size    int64  `json:"size,string"`
}

// NewZigDownloader creates a new Zig downloader
func NewZigDownloader() *ZigDownloader {
	return &ZigDownloader{
		BaseDownloader: NewBaseDownloader(),
		indexURL:       "https://ziglang.org/download/index.json",
	}
}

// GetDownloadURL returns the download URL for Zig compiler
func (z *ZigDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	// Fetch index.json
	index, err := z.fetchIndex()
	if err != nil {
		return "", fmt.Errorf("failed to fetch Zig index: %w", err)
	}

	// Get the requested version
	var release zigRelease
	if req.Version == "latest" || req.Version == "" {
		// Use master (latest development build)
		release = index.Master
	} else {
		// Try to find specific version
		found := false
		for version, rel := range index.Stable {
			if version == req.Version || strings.Contains(version, req.Version) {
				release = rel
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("version %s not found", req.Version)
		}
	}

	// Map platform and architecture to Zig's naming
	targetKey := z.mapPlatformToZigTarget(req.Platform, req.Arch)

	// Get target info
	target, ok := release.Targets[targetKey]
	if !ok {
		return "", fmt.Errorf("platform %s not supported by Zig", targetKey)
	}

	return target.Tarball, nil
}

// Download downloads and installs Zig compiler
func (z *ZigDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Get download URL and metadata
	metadata, err := z.GetMetadata(req)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "zig-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download archive
	archivePath := filepath.Join(tmpDir, "zig.tar.xz")

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "downloading",
			Message:    fmt.Sprintf("Downloading Zig %s...", metadata.Version),
			TotalBytes: metadata.Size,
		})
	}

	err = z.DownloadFile(ctx, metadata.DownloadURL, archivePath, progressCb)
	if err != nil {
		return fmt.Errorf("failed to download Zig: %w", err)
	}

	// Verify checksum
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying checksum...",
		})
	}

	if metadata.Checksum != "" {
		err = z.VerifyChecksum(archivePath, metadata.Checksum)
		if err != nil {
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}

	// Extract archive
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Extracting Zig compiler...",
		})
	}

	err = z.Extract(archivePath, req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to extract Zig: %w", err)
	}

	// Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = z.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    fmt.Sprintf("Zig %s installed successfully!", metadata.Version),
		})
	}

	return nil
}

// Extract extracts Zig archive to target directory
func (z *ZigDownloader) Extract(archivePath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract tar.xz
	if err := z.ExtractTarXz(archivePath, targetDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Zig extracts to a versioned directory (e.g., zig-macos-aarch64-0.14.0)
	// Find the extracted directory and move its contents to target dir
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	// Find the zig-* directory
	var zigDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "zig-") {
			zigDir = filepath.Join(targetDir, entry.Name())
			break
		}
	}

	if zigDir == "" {
		return fmt.Errorf("zig directory not found in archive")
	}

	// Move contents up one level if needed
	if zigDir != targetDir {
		// Read contents of zig directory
		zigEntries, err := os.ReadDir(zigDir)
		if err != nil {
			return fmt.Errorf("failed to read zig directory: %w", err)
		}

		// Move each entry to parent directory
		for _, entry := range zigEntries {
			src := filepath.Join(zigDir, entry.Name())
			dst := filepath.Join(targetDir, entry.Name())

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
		}

		// Remove the now-empty zig directory
		if err := os.Remove(zigDir); err != nil {
			// Ignore error if directory not empty
		}
	}

	return nil
}

// Verify checks that Zig compiler works correctly
func (z *ZigDownloader) Verify(installPath string) error {
	zigBinary := filepath.Join(installPath, "zig")
	if runtime.GOOS == "windows" {
		zigBinary += ".exe"
	}

	// Check if binary exists
	if _, err := os.Stat(zigBinary); err != nil {
		return fmt.Errorf("zig binary not found at %s: %w", zigBinary, err)
	}

	// Make binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(zigBinary, 0755); err != nil {
			return fmt.Errorf("failed to make zig executable: %w", err)
		}
	}

	// Run 'zig version' to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, zigBinary, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zig version check failed: %w (output: %s)", err, string(output))
	}

	// Check output contains version
	if !strings.Contains(string(output), ".") {
		return fmt.Errorf("unexpected zig version output: %s", string(output))
	}

	return nil
}

// GetMetadata returns metadata about Zig compiler download
func (z *ZigDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	// Fetch index.json
	index, err := z.fetchIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Zig index: %w", err)
	}

	// Get the requested version
	var release zigRelease
	if req.Version == "latest" || req.Version == "" {
		release = index.Master
	} else {
		found := false
		for version, rel := range index.Stable {
			if version == req.Version || strings.Contains(version, req.Version) {
				release = rel
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("version %s not found", req.Version)
		}
	}

	// Map platform and architecture
	targetKey := z.mapPlatformToZigTarget(req.Platform, req.Arch)

	// Get target info
	target, ok := release.Targets[targetKey]
	if !ok {
		return nil, fmt.Errorf("platform %s not supported", targetKey)
	}

	return &domain.CompilerMetadata{
		Language:    "zig",
		Version:     release.Version,
		DownloadURL: target.Tarball,
		Size:        target.Size,
		Checksum:    target.Shasum,
		ReleaseDate: time.Time{}, // Not available in index.json
	}, nil
}

// fetchIndex fetches and parses Zig's download index.json
func (z *ZigDownloader) fetchIndex() (*zigIndex, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, z.indexURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index fetch failed with status: %d", resp.StatusCode)
	}

	// Read and parse JSON
	body, err := readAllLimited(resp.Body, 15*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index zigIndex
	if err := json.Unmarshal(body, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &index, nil
}

// mapPlatformToZigTarget maps Go's GOOS/GOARCH to Zig's target naming
// Examples:
// - darwin + amd64 -> x86_64-macos
// - darwin + arm64 -> aarch64-macos
// - linux + amd64 -> x86_64-linux
// - windows + amd64 -> x86_64-windows
func (z *ZigDownloader) mapPlatformToZigTarget(platform, arch string) string {
	// Map architecture
	var zigArch string
	switch arch {
	case "amd64":
		zigArch = "x86_64"
	case "arm64":
		zigArch = "aarch64"
	case "386":
		zigArch = "i386"
	case "arm":
		zigArch = "armv7a"
	default:
		zigArch = arch
	}

	// Map platform
	var zigPlatform string
	switch platform {
	case "darwin":
		zigPlatform = "macos"
	case "linux":
		zigPlatform = "linux"
	case "windows":
		zigPlatform = "windows"
	default:
		zigPlatform = platform
	}

	return fmt.Sprintf("%s-%s", zigArch, zigPlatform)
}
