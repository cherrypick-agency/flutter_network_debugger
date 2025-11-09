package downloaders

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// SwiftDownloader implements CompilerDownloader for Swift language
// Swift compiles to WASM via SwiftWasm SDK
// Requires system Swift 6.1+ to be installed
type SwiftDownloader struct {
	*BaseDownloader
	minSwiftVersion string // Minimum required Swift version
}

// NewSwiftDownloader creates a new Swift downloader
func NewSwiftDownloader() *SwiftDownloader {
	return &SwiftDownloader{
		BaseDownloader:  NewBaseDownloader(),
		minSwiftVersion: "6.1.0",
	}
}

// GetDownloadURL returns the download URL for SwiftWasm SDK
func (s *SwiftDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	platform := req.Platform
	arch := req.Arch

	// Map platform and arch to SwiftWasm naming
	targetKey, err := s.mapPlatformToSwiftTarget(platform, arch)
	if err != nil {
		return "", err
	}

	version := req.Version
	if version == "latest" || version == "" {
		version = "6.1.0" // Latest stable SwiftWasm release
	}

	// SwiftWasm SDK format:
	// https://github.com/swiftwasm/swift/releases/download/swift-wasm-6.1.0-RELEASE/swift-wasm-6.1.0-RELEASE-macos_arm64.pkg
	// For non-pkg formats:
	// https://github.com/swiftwasm/swift/releases/download/swift-wasm-6.1.0-RELEASE/swift-wasm-6.1.0-RELEASE-ubuntu22.04_x86_64.tar.gz

	releaseTag := fmt.Sprintf("swift-wasm-%s-RELEASE", version)

	// Determine file extension based on platform
	ext := ".tar.gz"
	if platform == "darwin" {
		ext = ".pkg"
	}

	url := fmt.Sprintf(
		"https://github.com/swiftwasm/swift/releases/download/%s/%s-%s%s",
		releaseTag, releaseTag, targetKey, ext,
	)

	return url, nil
}

// Download downloads and installs SwiftWasm SDK
func (s *SwiftDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Step 1: Check system Swift
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "checking",
			Message: "Checking for system Swift...",
		})
	}

	swiftVersion, err := s.getSystemSwiftVersion()
	if err != nil {
		return fmt.Errorf("system Swift not found. Please install Swift %s or later from https://swift.org/download/", s.minSwiftVersion)
	}

	if !s.isVersionSufficient(swiftVersion, s.minSwiftVersion) {
		return fmt.Errorf("system Swift version %s is too old. Please upgrade to Swift %s or later", swiftVersion, s.minSwiftVersion)
	}

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "checking",
			Message: fmt.Sprintf("Using system Swift %s", swiftVersion),
		})
	}

	// Step 2: Download SwiftWasm SDK
	url, err := s.GetDownloadURL(req)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	version := req.Version
	if version == "latest" || version == "" {
		version = "6.1.0"
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "swiftwasm-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Determine archive type
	archiveName := "swiftwasm-sdk.tar.gz"
	if req.Platform == "darwin" {
		archiveName = "swiftwasm-sdk.pkg"
	}

	archivePath := filepath.Join(tmpDir, archiveName)

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "downloading",
			Message: fmt.Sprintf("Downloading SwiftWasm SDK %s...", version),
		})
	}

	err = s.DownloadFile(ctx, url, archivePath, func(progress domain.DownloadProgress) {
		progress.Message = fmt.Sprintf("Downloading SwiftWasm: %s", progress.Message)
		if progressCb != nil {
			progressCb(progress)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download SwiftWasm SDK: %w", err)
	}

	// Extract/Install
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Installing SwiftWasm SDK...",
		})
	}

	err = s.Extract(archivePath, req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to install SwiftWasm SDK: %w", err)
	}

	// Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = s.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    fmt.Sprintf("SwiftWasm SDK %s installed successfully!", version),
		})
	}

	return nil
}

// Extract extracts SwiftWasm SDK archive to target directory
func (s *SwiftDownloader) Extract(archivePath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Check file extension
	if strings.HasSuffix(archivePath, ".pkg") {
		// macOS .pkg installer
		return s.extractPkg(archivePath, targetDir)
	} else if strings.HasSuffix(archivePath, ".tar.gz") {
		// Linux tar.gz archive
		return s.ExtractTarGz(archivePath, targetDir)
	}

	return fmt.Errorf("unsupported archive format: %s", archivePath)
}

// extractPkg extracts a macOS .pkg file
// Note: .pkg files need special handling on macOS
func (s *SwiftDownloader) extractPkg(pkgPath, targetDir string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf(".pkg files can only be extracted on macOS")
	}

	// Create temp directory for extraction
	tmpDir, err := os.MkdirTemp("", "pkg-extract-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract pkg using pkgutil
	// pkgutil --expand-full <pkg> <output-dir>
	cmd := exec.Command("pkgutil", "--expand-full", pkgPath, tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pkgutil failed: %w (output: %s)", err, string(output))
	}

	// Find the extracted SDK directory
	// SwiftWasm SDK typically extracts to a specific structure
	// We need to find the actual SDK root and copy it to targetDir

	// For now, use a simpler approach: just copy all contents
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, entry := range entries {
		src := filepath.Join(tmpDir, entry.Name())
		dst := filepath.Join(targetDir, entry.Name())

		if entry.IsDir() {
			// Copy directory recursively
			if err := s.copyDir(src, dst); err != nil {
				return fmt.Errorf("failed to copy directory: %w", err)
			}
		} else {
			// Copy file
			if err := s.copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy file: %w", err)
			}
		}
	}

	return nil
}

// Verify checks that SwiftWasm SDK is installed correctly
func (s *SwiftDownloader) Verify(installPath string) error {
	// Check if system Swift is still available
	_, err := s.getSystemSwiftVersion()
	if err != nil {
		return fmt.Errorf("system Swift not found: %w", err)
	}

	// Check if SwiftWasm SDK files exist
	// Look for typical SwiftWasm SDK structure
	// The exact structure depends on how it was extracted

	// For now, just check if install path exists and is not empty
	entries, err := os.ReadDir(installPath)
	if err != nil {
		return fmt.Errorf("failed to read install directory: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("install directory is empty")
	}

	return nil
}

// GetMetadata returns metadata about SwiftWasm SDK download
func (s *SwiftDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	version := req.Version
	if version == "latest" || version == "" {
		version = "6.1.0"
	}

	url, err := s.GetDownloadURL(req)
	if err != nil {
		return nil, err
	}

	// Approximate size: SwiftWasm SDK is ~50-100 MB
	size := int64(75 * 1024 * 1024) // 75 MB average

	return &domain.CompilerMetadata{
		Language:    "swift",
		Version:     version,
		DownloadURL: url,
		Size:        size,
		Checksum:    "", // SwiftWasm doesn't provide checksums
		ReleaseDate: time.Time{},
	}, nil
}

// getSystemSwiftVersion returns the version of system-installed Swift
func (s *SwiftDownloader) getSystemSwiftVersion() (string, error) {
	cmd := exec.Command("swift", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("swift command not found: %w", err)
	}

	// Parse version from output
	// Example: "Swift version 6.1 (swift-6.1-RELEASE)"
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Swift version") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "version" && i+1 < len(parts) {
					return parts[i+1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not parse Swift version from output: %s", string(output))
}

// isVersionSufficient checks if actual version meets minimum requirement
func (s *SwiftDownloader) isVersionSufficient(actual, minimum string) bool {
	actualParts := s.parseVersion(actual)
	minimumParts := s.parseVersion(minimum)

	for i := 0; i < len(minimumParts) && i < len(actualParts); i++ {
		if actualParts[i] > minimumParts[i] {
			return true
		}
		if actualParts[i] < minimumParts[i] {
			return false
		}
	}

	return true // Equal or actual has more parts
}

// parseVersion parses version string into numeric parts
func (s *SwiftDownloader) parseVersion(version string) []int {
	// Remove any non-numeric suffix
	version = strings.Split(version, "-")[0]

	parts := strings.Split(version, ".")
	result := make([]int, len(parts))

	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err == nil {
			result[i] = num
		}
	}

	return result
}

// mapPlatformToSwiftTarget maps Go's GOOS/GOARCH to SwiftWasm naming
func (s *SwiftDownloader) mapPlatformToSwiftTarget(platform, arch string) (string, error) {
	switch platform {
	case "darwin":
		switch arch {
		case "amd64":
			return "macos_x86_64", nil
		case "arm64":
			return "macos_arm64", nil
		default:
			return "", fmt.Errorf("unsupported macOS architecture: %s", arch)
		}

	case "linux":
		// SwiftWasm supports Ubuntu 22.04 and 20.04
		switch arch {
		case "amd64":
			return "ubuntu22.04_x86_64", nil
		case "arm64":
			return "ubuntu22.04_aarch64", nil
		default:
			return "", fmt.Errorf("unsupported Linux architecture: %s", arch)
		}

	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}
}

// copyFile copies a file from src to dst
func (s *SwiftDownloader) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// copyDir recursively copies a directory
func (s *SwiftDownloader) copyDir(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := s.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := s.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
