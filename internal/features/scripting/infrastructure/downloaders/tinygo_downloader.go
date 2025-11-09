package downloaders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// TinyGoDownloader implements CompilerDownloader for TinyGo
// TinyGo is a Go compiler for small places (microcontrollers, WebAssembly)
type TinyGoDownloader struct {
	*BaseDownloader
}

// tinygoRelease represents a GitHub release
type tinygoRelease struct {
	TagName    string        `json:"tag_name"`
	Name       string        `json:"name"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	CreatedAt  time.Time     `json:"created_at"`
	Assets     []tinygoAsset `json:"assets"`
}

type tinygoAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// NewTinyGoDownloader creates a new TinyGo downloader
func NewTinyGoDownloader() *TinyGoDownloader {
	return &TinyGoDownloader{
		BaseDownloader: NewBaseDownloader(),
	}
}

// GetDownloadURL returns the download URL for TinyGo compiler
func (t *TinyGoDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	// Get latest release from GitHub API
	release, err := t.getLatestRelease(req.Version)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}

	// Map platform and arch to TinyGo asset naming
	assetName := t.getAssetName(req.Platform, req.Arch, release.TagName)

	// Find matching asset
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset.DownloadURL, nil
		}
	}

	return "", fmt.Errorf("no asset found for platform %s/%s", req.Platform, req.Arch)
}

// Download downloads and installs TinyGo compiler
func (t *TinyGoDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Get metadata
	metadata, err := t.GetMetadata(req)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "tinygo-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download archive
	archivePath := filepath.Join(tmpDir, "tinygo.tar.gz")

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "downloading",
			Message:    fmt.Sprintf("Downloading TinyGo %s...", metadata.Version),
			TotalBytes: metadata.Size,
		})
	}

	err = t.DownloadFile(ctx, metadata.DownloadURL, archivePath, progressCb)
	if err != nil {
		return fmt.Errorf("failed to download TinyGo: %w", err)
	}

	// Extract archive
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Extracting TinyGo compiler...",
		})
	}

	err = t.Extract(archivePath, req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to extract TinyGo: %w", err)
	}

	// Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = t.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    fmt.Sprintf("TinyGo %s installed successfully!", metadata.Version),
		})
	}

	return nil
}

// Extract extracts TinyGo archive to target directory
func (t *TinyGoDownloader) Extract(archivePath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract tar.gz
	if err := t.ExtractTarGz(archivePath, targetDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// TinyGo extracts to a versioned directory (e.g., tinygo0.33.0)
	// Find the extracted directory and move its contents to target dir
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	// Find the tinygo* directory
	var tinygoDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "tinygo") {
			tinygoDir = filepath.Join(targetDir, entry.Name())
			break
		}
	}

	if tinygoDir == "" {
		return fmt.Errorf("tinygo directory not found in archive")
	}

	// Move contents up one level if needed
	if tinygoDir != targetDir {
		// Read contents of tinygo directory
		tinygoEntries, err := os.ReadDir(tinygoDir)
		if err != nil {
			return fmt.Errorf("failed to read tinygo directory: %w", err)
		}

		// Move each entry to parent directory
		for _, entry := range tinygoEntries {
			src := filepath.Join(tinygoDir, entry.Name())
			dst := filepath.Join(targetDir, entry.Name())

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
		}

		// Remove the now-empty tinygo directory
		if err := os.Remove(tinygoDir); err != nil {
			// Ignore error if directory not empty
		}
	}

	return nil
}

// Verify checks that TinyGo compiler works correctly
func (t *TinyGoDownloader) Verify(installPath string) error {
	tinygoBinary := filepath.Join(installPath, "bin", "tinygo")
	if runtime.GOOS == "windows" {
		tinygoBinary += ".exe"
	}

	// Check if binary exists
	if _, err := os.Stat(tinygoBinary); err != nil {
		return fmt.Errorf("tinygo binary not found at %s: %w", tinygoBinary, err)
	}

	// Make binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tinygoBinary, 0755); err != nil {
			return fmt.Errorf("failed to make tinygo executable: %w", err)
		}
	}

	// Run 'tinygo version' to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, tinygoBinary, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tinygo version check failed: %w (output: %s)", err, string(output))
	}

	// Check output contains "tinygo version"
	if !strings.Contains(string(output), "tinygo version") {
		return fmt.Errorf("unexpected tinygo version output: %s", string(output))
	}

	return nil
}

// GetMetadata returns metadata about TinyGo compiler download
func (t *TinyGoDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	// Get latest release
	release, err := t.getLatestRelease(req.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	// Map platform and arch to asset name
	assetName := t.getAssetName(req.Platform, req.Arch, release.TagName)

	// Find matching asset
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return &domain.CompilerMetadata{
				Language:    "tinygo",
				Version:     release.TagName,
				DownloadURL: asset.DownloadURL,
				Size:        asset.Size,
				Checksum:    "", // TinyGo doesn't provide checksums in releases
				ReleaseDate: release.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("no asset found for platform %s/%s", req.Platform, req.Arch)
}

// getLatestRelease fetches the latest TinyGo release from GitHub API
func (t *TinyGoDownloader) getLatestRelease(version string) (*tinygoRelease, error) {
	var apiURL string

	if version == "latest" || version == "" {
		// Get latest release
		apiURL = "https://api.github.com/repos/tinygo-org/tinygo/releases/latest"
	} else {
		// Get specific release by tag
		apiURL = fmt.Sprintf("https://api.github.com/repos/tinygo-org/tinygo/releases/tags/v%s", version)
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release tinygoRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// getAssetName returns the expected asset name for the given platform/arch
func (t *TinyGoDownloader) getAssetName(platform, arch, version string) string {
	// Remove 'v' prefix from version if present
	version = strings.TrimPrefix(version, "v")

	// Map platform
	osName := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "windows",
	}[platform]

	// Map arch
	archName := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
	}[arch]

	// Format: tinygo0.33.0.darwin-amd64.tar.gz
	ext := ".tar.gz"
	if platform == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("tinygo%s.%s-%s%s", version, osName, archName, ext)
}
