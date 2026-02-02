package downloaders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// KotlinDownloader implements CompilerDownloader for Kotlin language
// Kotlin compiles to WASM via Kotlin/Wasm (experimental)
// Requires JRE to run the Kotlin compiler
type KotlinDownloader struct {
	*BaseDownloader
}

// NewKotlinDownloader creates a new Kotlin downloader
func NewKotlinDownloader() *KotlinDownloader {
	return &KotlinDownloader{
		BaseDownloader: NewBaseDownloader(),
	}
}

// GetDownloadURL returns the download URL for Kotlin compiler
func (k *KotlinDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	version := req.Version
	if version == "latest" || version == "" {
		version = "2.2.20" // Latest stable as of 2025
	}

	// Kotlin compiler is distributed as a zip file
	// Format: https://github.com/JetBrains/kotlin/releases/download/v{version}/kotlin-compiler-{version}.zip
	url := fmt.Sprintf("https://github.com/JetBrains/kotlin/releases/download/v%s/kotlin-compiler-%s.zip", version, version)

	return url, nil
}

// Download downloads and installs Kotlin compiler + JRE if needed
func (k *KotlinDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Step 1: Check JRE
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "checking",
			Message: "Checking for Java Runtime Environment...",
		})
	}

	jrePath, err := k.findOrInstallJRE(ctx, req.TargetDir, progressCb)
	if err != nil {
		return fmt.Errorf("JRE setup failed: %w", err)
	}

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "checking",
			Message: fmt.Sprintf("Using JRE at: %s", jrePath),
		})
	}

	// Step 2: Download Kotlin compiler
	url, err := k.GetDownloadURL(req)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Get version for messages
	version := req.Version
	if version == "latest" || version == "" {
		version = "2.2.20"
	}

	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "kotlin-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download archive
	archivePath := filepath.Join(tmpDir, "kotlin-compiler.zip")

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "downloading",
			Message: fmt.Sprintf("Downloading Kotlin compiler %s...", version),
		})
	}

	err = k.DownloadFile(ctx, url, archivePath, func(progress domain.DownloadProgress) {
		// Adjust message for Kotlin
		progress.Message = fmt.Sprintf("Downloading Kotlin %s: %s", version, progress.Message)
		if progressCb != nil {
			progressCb(progress)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download Kotlin: %w", err)
	}

	// Extract archive
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Extracting Kotlin compiler...",
		})
	}

	err = k.Extract(archivePath, req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to extract Kotlin: %w", err)
	}

	// Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = k.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    fmt.Sprintf("Kotlin %s installed successfully!", version),
		})
	}

	return nil
}

// Extract extracts Kotlin compiler archive to target directory
func (k *KotlinDownloader) Extract(archivePath, targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Extract zip
	if err := k.ExtractZip(archivePath, targetDir); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	// Kotlin extracts to a 'kotlinc' directory
	// Move its contents to target dir
	kotlincDir := filepath.Join(targetDir, "kotlinc")
	if _, err := os.Stat(kotlincDir); err == nil {
		// Read contents
		entries, err := os.ReadDir(kotlincDir)
		if err != nil {
			return fmt.Errorf("failed to read kotlinc directory: %w", err)
		}

		// Move each entry to parent directory
		for _, entry := range entries {
			src := filepath.Join(kotlincDir, entry.Name())
			dst := filepath.Join(targetDir, entry.Name())

			if err := os.Rename(src, dst); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
		}

		// Remove the now-empty kotlinc directory
		if err := os.Remove(kotlincDir); err != nil {
			// Ignore error if directory not empty
		}
	}

	return nil
}

// Verify checks that Kotlin compiler works correctly
func (k *KotlinDownloader) Verify(installPath string) error {
	// Find kotlinc binary
	kotlincBinary := filepath.Join(installPath, "bin", "kotlinc")
	if runtime.GOOS == "windows" {
		kotlincBinary = filepath.Join(installPath, "bin", "kotlinc.bat")
	}

	// Check if binary exists
	if _, err := os.Stat(kotlincBinary); err != nil {
		return fmt.Errorf("kotlinc binary not found at %s: %w", kotlincBinary, err)
	}

	// Make binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(kotlincBinary, 0755); err != nil {
			return fmt.Errorf("failed to make kotlinc executable: %w", err)
		}
	}

	// Run 'kotlinc -version' to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, kotlincBinary, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kotlinc version check failed: %w (output: %s)", err, string(output))
	}

	// Check output contains "Kotlin"
	if !strings.Contains(string(output), "Kotlin") {
		return fmt.Errorf("unexpected kotlinc version output: %s", string(output))
	}

	return nil
}

// GetMetadata returns metadata about Kotlin compiler download
func (k *KotlinDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	version := req.Version
	if version == "latest" || version == "" {
		version = "2.2.20"
	}

	url, err := k.GetDownloadURL(req)
	if err != nil {
		return nil, err
	}

	// Approximate size: Kotlin compiler is ~75 MB
	// JRE adds 100-300 MB if not present
	size := int64(75 * 1024 * 1024) // 75 MB

	// Check if JRE is available
	if !k.isJREAvailable() {
		size += int64(200 * 1024 * 1024) // Add ~200 MB for JRE
	}

	return &domain.CompilerMetadata{
		Language:    "kotlin",
		Version:     version,
		DownloadURL: url,
		Size:        size,
		Checksum:    "", // Kotlin doesn't provide checksums in releases
		ReleaseDate: time.Time{},
	}, nil
}

// findOrInstallJRE finds system JRE or installs minimal JRE
// Returns path to JRE home directory
func (k *KotlinDownloader) findOrInstallJRE(
	ctx context.Context,
	installDir string,
	progressCb func(domain.DownloadProgress),
) (string, error) {
	// Try to find system JRE first
	jrePath := k.findSystemJRE()
	if jrePath != "" {
		return jrePath, nil
	}

	// System JRE not found, install minimal JRE
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "downloading",
			Message: "System JRE not found. Downloading minimal JRE...",
		})
	}

	jreDir := filepath.Join(installDir, "jre")
	err := k.downloadMinimalJRE(ctx, jreDir, progressCb)
	if err != nil {
		return "", fmt.Errorf("failed to download JRE: %w", err)
	}

	return jreDir, nil
}

// findSystemJRE tries to find system-installed JRE
// Returns empty string if not found
func (k *KotlinDownloader) findSystemJRE() string {
	// Try 'java -version' command
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "java", "-version")
	err := cmd.Run()
	if err == nil {
		// Java is available in PATH
		// Get JAVA_HOME
		javaHome := os.Getenv("JAVA_HOME")
		if javaHome != "" {
			return javaHome
		}

		// Try to find java location
		ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel2()

		cmd := exec.CommandContext(ctx2, "which", "java")
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx2, "where", "java")
		}

		output, err := cmd.Output()
		if err == nil {
			javaPath := strings.TrimSpace(string(output))
			// Java is typically at $JAVA_HOME/bin/java
			// So JAVA_HOME is parent of bin
			binDir := filepath.Dir(javaPath)
			javaHome := filepath.Dir(binDir)
			return javaHome
		}
	}

	return ""
}

// isJREAvailable checks if JRE is available on the system
func (k *KotlinDownloader) isJREAvailable() bool {
	return k.findSystemJRE() != ""
}

// downloadMinimalJRE downloads a minimal JRE for running Kotlin compiler
// This is a fallback if system JRE is not found
func (k *KotlinDownloader) downloadMinimalJRE(
	ctx context.Context,
	jreDir string,
	progressCb func(domain.DownloadProgress),
) error {
	// Determine JRE download URL based on platform
	jreURL, err := k.getJREDownloadURL()
	if err != nil {
		return err
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "jre-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download JRE
	archivePath := filepath.Join(tmpDir, "jre.tar.gz")

	err = k.DownloadFile(ctx, jreURL, archivePath, func(progress domain.DownloadProgress) {
		progress.Message = fmt.Sprintf("Downloading JRE: %s", progress.Message)
		if progressCb != nil {
			progressCb(progress)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to download JRE: %w", err)
	}

	// Extract JRE
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "extracting",
			Message: "Extracting JRE...",
		})
	}

	if err := os.MkdirAll(jreDir, 0755); err != nil {
		return fmt.Errorf("failed to create JRE directory: %w", err)
	}

	if err := k.ExtractTarGz(archivePath, jreDir); err != nil {
		return fmt.Errorf("failed to extract JRE: %w", err)
	}

	return nil
}

// getJREDownloadURL returns download URL for minimal JRE
// Uses Adoptium (Eclipse Temurin) builds
func (k *KotlinDownloader) getJREDownloadURL() (string, error) {
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map arch
	var adoptiumArch string
	switch arch {
	case "amd64":
		adoptiumArch = "x64"
	case "arm64":
		adoptiumArch = "aarch64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	// Map platform
	var adoptiumOS string
	switch platform {
	case "darwin":
		adoptiumOS = "mac"
	case "linux":
		adoptiumOS = "linux"
	case "windows":
		adoptiumOS = "windows"
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}

	// Use JRE 21 (LTS) from Adoptium
	// Format: https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.5%2B11/OpenJDK21U-jre_{arch}_{os}_hotspot_21.0.5_11.tar.gz
	version := "21.0.5"
	buildNumber := "11"

	ext := "tar.gz"
	if platform == "windows" {
		ext = "zip"
	}

	url := fmt.Sprintf(
		"https://github.com/adoptium/temurin21-binaries/releases/download/jdk-%s%%2B%s/OpenJDK21U-jre_%s_%s_hotspot_%s_%s.%s",
		version, buildNumber, adoptiumArch, adoptiumOS, strings.ReplaceAll(version, ".", "_"), buildNumber, ext,
	)

	return url, nil
}
