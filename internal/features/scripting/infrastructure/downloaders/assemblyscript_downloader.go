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

// AssemblyScriptDownloader implements CompilerDownloader for AssemblyScript
// AssemblyScript is a TypeScript-like language that compiles to WebAssembly
// Strategy: Download Node.js + install assemblyscript via npm
type AssemblyScriptDownloader struct {
	*BaseDownloader
}

// nodeRelease represents a Node.js release
type nodeRelease struct {
	Version string      `json:"version"`
	Date    string      `json:"date"`
	Files   []string    `json:"files"`
	LTS     interface{} `json:"lts"` // false or string
}

// NewAssemblyScriptDownloader creates a new AssemblyScript downloader
func NewAssemblyScriptDownloader() *AssemblyScriptDownloader {
	return &AssemblyScriptDownloader{
		BaseDownloader: NewBaseDownloader(),
	}
}

// GetDownloadURL returns the download URL for Node.js
func (a *AssemblyScriptDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	// Get LTS version
	version, err := a.getNodeLTSVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get Node.js LTS version: %w", err)
	}

	// Build download URL
	baseURL := "https://nodejs.org/dist"
	assetName := a.getAssetName(req.Platform, req.Arch, version)

	return fmt.Sprintf("%s/%s/%s", baseURL, version, assetName), nil
}

// Download downloads and installs AssemblyScript compiler
func (a *AssemblyScriptDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Step 1: Download Node.js
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "downloading",
			Message: "Downloading Node.js...",
		})
	}

	// Get Node.js LTS version
	nodeVersion, err := a.getNodeLTSVersion()
	if err != nil {
		return fmt.Errorf("failed to get Node.js version: %w", err)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "assemblyscript-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download Node.js
	nodeURL, err := a.GetDownloadURL(req)
	if err != nil {
		return fmt.Errorf("failed to get Node.js URL: %w", err)
	}

	archivePath := filepath.Join(tmpDir, "node.tar.gz")
	if runtime.GOOS == "windows" {
		archivePath = filepath.Join(tmpDir, "node.zip")
	}

	err = a.DownloadFile(ctx, nodeURL, archivePath, progressCb)
	if err != nil {
		return fmt.Errorf("failed to download Node.js: %w", err)
	}

	// Step 2: Extract Node.js
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Extracting Node.js...",
		})
	}

	err = a.Extract(archivePath, req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to extract Node.js: %w", err)
	}

	// Step 3: Install AssemblyScript via npm
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "installing",
			Message: "Installing AssemblyScript compiler...",
		})
	}

	err = a.installAssemblyScript(req.TargetDir, nodeVersion)
	if err != nil {
		return fmt.Errorf("failed to install AssemblyScript: %w", err)
	}

	// Step 4: Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = a.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    "AssemblyScript compiler installed successfully!",
		})
	}

	return nil
}

// Extract extracts Node.js archive to target directory
func (a *AssemblyScriptDownloader) Extract(archivePath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract based on platform
	if runtime.GOOS == "windows" {
		if err := a.ExtractZip(archivePath, targetDir); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		if err := a.ExtractTarGz(archivePath, targetDir); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	}

	// Node.js extracts to a versioned directory (e.g., node-v20.11.0-darwin-x64)
	// Find and flatten it
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return fmt.Errorf("failed to read target directory: %w", err)
	}

	var nodeDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "node-") {
			nodeDir = filepath.Join(targetDir, entry.Name())
			break
		}
	}

	if nodeDir == "" {
		return fmt.Errorf("node directory not found in archive")
	}

	// Move contents up one level if needed
	if nodeDir != targetDir {
		nodeEntries, err := os.ReadDir(nodeDir)
		if err != nil {
			return fmt.Errorf("failed to read node directory: %w", err)
		}

		for _, entry := range nodeEntries {
			src := filepath.Join(nodeDir, entry.Name())
			dst := filepath.Join(targetDir, entry.Name())

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
		}

		// Remove the now-empty node directory
		os.Remove(nodeDir)
	}

	return nil
}

// installAssemblyScript installs AssemblyScript using npm
func (a *AssemblyScriptDownloader) installAssemblyScript(installPath, nodeVersion string) error {
	npmBinary := a.getNpmPath(installPath)

	// Create node_modules directory
	nodeModulesDir := filepath.Join(installPath, "lib", "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create node_modules directory: %w", err)
	}

	// Install assemblyscript globally
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, npmBinary, "install", "-g", "--prefix", installPath, "assemblyscript")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Verify checks that AssemblyScript compiler works correctly
func (a *AssemblyScriptDownloader) Verify(installPath string) error {
	ascBinary := a.getAscPath(installPath)

	// Check if asc binary exists
	if _, err := os.Stat(ascBinary); err != nil {
		return fmt.Errorf("asc binary not found at %s: %w", ascBinary, err)
	}

	// Make binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(ascBinary, 0755); err != nil {
			return fmt.Errorf("failed to make asc executable: %w", err)
		}
	}

	// Run 'asc --version' to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ascBinary, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("asc version check failed: %w (output: %s)", err, string(output))
	}

	// Check output contains version number
	if len(output) == 0 {
		return fmt.Errorf("unexpected asc version output: empty")
	}

	return nil
}

// GetMetadata returns metadata about AssemblyScript compiler download
func (a *AssemblyScriptDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	// Get Node.js LTS version
	version, err := a.getNodeLTSVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get Node.js version: %w", err)
	}

	// Get download URL
	downloadURL, err := a.GetDownloadURL(req)
	if err != nil {
		return nil, err
	}

	// Get size via HEAD request
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	headReq, err := http.NewRequestWithContext(ctx, http.MethodHead, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(headReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Node.js is ~35-50 MB, AssemblyScript adds ~10 MB
	size := resp.ContentLength
	if size <= 0 {
		size = 50 * 1024 * 1024 // 50 MB estimate
	}

	return &domain.CompilerMetadata{
		Language:    "assemblyscript",
		Version:     version,
		DownloadURL: downloadURL,
		Size:        size,
		Checksum:    "", // Node.js provides checksums separately
		ReleaseDate: time.Now(),
	}, nil
}

// getNodeLTSVersion fetches the current LTS version of Node.js
func (a *AssemblyScriptDownloader) getNodeLTSVersion() (string, error) {
	// Fetch from Node.js API
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://nodejs.org/dist/index.json", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Node.js releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Node.js API returned status: %d", resp.StatusCode)
	}

	body, err := readAllLimited(resp.Body, 10*1024*1024)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var releases []nodeRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", fmt.Errorf("failed to parse releases: %w", err)
	}

	// Find latest LTS version
	for _, release := range releases {
		if release.LTS != nil && release.LTS != false {
			return release.Version, nil
		}
	}

	return "", fmt.Errorf("no LTS version found")
}

// getAssetName returns the expected asset name for the given platform/arch
func (a *AssemblyScriptDownloader) getAssetName(platform, arch, version string) string {
	// Map platform
	osName := map[string]string{
		"darwin":  "darwin",
		"linux":   "linux",
		"windows": "win",
	}[platform]

	// Map arch
	archName := map[string]string{
		"amd64": "x64",
		"arm64": "arm64",
		"386":   "x86",
	}[arch]

	// Format: node-v20.11.0-darwin-x64.tar.gz
	ext := ".tar.gz"
	if platform == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("node-%s-%s-%s%s", version, osName, archName, ext)
}

// getNpmPath returns the path to npm binary
func (a *AssemblyScriptDownloader) getNpmPath(installPath string) string {
	npmBinary := filepath.Join(installPath, "bin", "npm")
	if runtime.GOOS == "windows" {
		npmBinary = filepath.Join(installPath, "npm.cmd")
	}
	return npmBinary
}

// getAscPath returns the path to asc (AssemblyScript compiler) binary
func (a *AssemblyScriptDownloader) getAscPath(installPath string) string {
	ascBinary := filepath.Join(installPath, "bin", "asc")
	if runtime.GOOS == "windows" {
		ascBinary = filepath.Join(installPath, "asc.cmd")
	}
	return ascBinary
}
