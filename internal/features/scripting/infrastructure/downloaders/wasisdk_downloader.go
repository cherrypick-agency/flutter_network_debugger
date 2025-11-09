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

// WASISDKDownloader implements CompilerDownloader for WASI-SDK
// WASI-SDK is a clang/LLVM-based toolchain for compiling C/C++ to WebAssembly
type WASISDKDownloader struct {
	*BaseDownloader
}

// wasiRelease represents a GitHub release
type wasiRelease struct {
	TagName    string      `json:"tag_name"`
	Name       string      `json:"name"`
	Draft      bool        `json:"draft"`
	Prerelease bool        `json:"prerelease"`
	CreatedAt  time.Time   `json:"created_at"`
	Assets     []wasiAsset `json:"assets"`
}

type wasiAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

// NewWASISDKDownloader creates a new WASI-SDK downloader
func NewWASISDKDownloader() *WASISDKDownloader {
	return &WASISDKDownloader{
		BaseDownloader: NewBaseDownloader(),
	}
}

// GetDownloadURL returns the download URL for WASI-SDK
func (w *WASISDKDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	// Get latest release from GitHub API
	release, err := w.getLatestRelease(req.Version)
	if err != nil {
		return "", fmt.Errorf("failed to get latest release: %w", err)
	}

	// Map platform and arch to WASI-SDK asset naming
	assetName := w.getAssetName(req.Platform, req.Arch, release.TagName)

	// Find matching asset
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset.DownloadURL, nil
		}
	}

	return "", fmt.Errorf("no asset found for platform %s/%s", req.Platform, req.Arch)
}

// Download downloads and installs WASI-SDK
func (w *WASISDKDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Get metadata
	metadata, err := w.GetMetadata(req)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "wasisdk-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download archive
	archivePath := filepath.Join(tmpDir, "wasi-sdk.tar.gz")

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "downloading",
			Message:    fmt.Sprintf("Downloading WASI-SDK %s...", metadata.Version),
			TotalBytes: metadata.Size,
		})
	}

	err = w.DownloadFile(ctx, metadata.DownloadURL, archivePath, progressCb)
	if err != nil {
		return fmt.Errorf("failed to download WASI-SDK: %w", err)
	}

	// Extract archive
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Extracting WASI-SDK...",
		})
	}

	err = w.Extract(archivePath, req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to extract WASI-SDK: %w", err)
	}

	// Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = w.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    fmt.Sprintf("WASI-SDK %s installed successfully!", metadata.Version),
		})
	}

	return nil
}

// Extract extracts WASI-SDK archive to target directory
func (w *WASISDKDownloader) Extract(archivePath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract tar.gz
	if err := w.ExtractTarGz(archivePath, targetDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// WASI-SDK extracts to a versioned directory (e.g., wasi-sdk-20.0)
	// Find the extracted directory and move its contents to target dir
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	// Find the wasi-sdk* directory
	var wasiDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "wasi-sdk") {
			wasiDir = filepath.Join(targetDir, entry.Name())
			break
		}
	}

	if wasiDir == "" {
		return fmt.Errorf("wasi-sdk directory not found in archive")
	}

	// Move contents up one level if needed
	if wasiDir != targetDir {
		// Read contents of wasi-sdk directory
		wasiEntries, err := os.ReadDir(wasiDir)
		if err != nil {
			return fmt.Errorf("failed to read wasi-sdk directory: %w", err)
		}

		// Move each entry to parent directory
		for _, entry := range wasiEntries {
			src := filepath.Join(wasiDir, entry.Name())
			dst := filepath.Join(targetDir, entry.Name())

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
		}

		// Remove the now-empty wasi-sdk directory
		os.Remove(wasiDir)
	}

	return nil
}

// Verify checks that WASI-SDK works correctly
func (w *WASISDKDownloader) Verify(installPath string) error {
	clangBinary := filepath.Join(installPath, "bin", "clang")
	if runtime.GOOS == "windows" {
		clangBinary += ".exe"
	}

	// Check if clang binary exists
	if _, err := os.Stat(clangBinary); err != nil {
		return fmt.Errorf("clang binary not found at %s: %w", clangBinary, err)
	}

	// Make binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(clangBinary, 0755); err != nil {
			return fmt.Errorf("failed to make clang executable: %w", err)
		}
	}

	// Run 'clang --version' to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, clangBinary, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clang version check failed: %w (output: %s)", err, string(output))
	}

	// Check output contains "clang version"
	if !strings.Contains(string(output), "clang version") {
		return fmt.Errorf("unexpected clang version output: %s", string(output))
	}

	return nil
}

// GetMetadata returns metadata about WASI-SDK download
func (w *WASISDKDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	// Get latest release
	release, err := w.getLatestRelease(req.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}

	// Map platform and arch to asset name
	assetName := w.getAssetName(req.Platform, req.Arch, release.TagName)

	// Find matching asset
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return &domain.CompilerMetadata{
				Language:    "wasi-sdk",
				Version:     release.TagName,
				DownloadURL: asset.DownloadURL,
				Size:        asset.Size,
				Checksum:    "", // WASI-SDK provides checksums separately
				ReleaseDate: release.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("no asset found for platform %s/%s", req.Platform, req.Arch)
}

// getLatestRelease fetches the latest WASI-SDK release from GitHub API
func (w *WASISDKDownloader) getLatestRelease(version string) (*wasiRelease, error) {
	var apiURL string

	if version == "latest" || version == "" {
		// Get latest release
		apiURL = "https://api.github.com/repos/WebAssembly/wasi-sdk/releases/latest"
	} else {
		// Get specific release by tag
		apiURL = fmt.Sprintf("https://api.github.com/repos/WebAssembly/wasi-sdk/releases/tags/wasi-sdk-%s", version)
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

	var release wasiRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	return &release, nil
}

// getAssetName returns the expected asset name for the given platform/arch
func (w *WASISDKDownloader) getAssetName(platform, arch, version string) string {
	// Remove 'wasi-sdk-' prefix from version if present
	version = strings.TrimPrefix(version, "wasi-sdk-")

	// Map platform
	osName := map[string]string{
		"darwin":  "macos",
		"linux":   "linux",
		"windows": "mingw", // WASI-SDK uses mingw for Windows
	}[platform]

	// WASI-SDK currently only provides x86_64 builds
	// arm64 Macs can use x86_64 via Rosetta
	// Format: wasi-sdk-20.0-macos.tar.gz
	return fmt.Sprintf("wasi-sdk-%s-%s.tar.gz", version, osName)
}
