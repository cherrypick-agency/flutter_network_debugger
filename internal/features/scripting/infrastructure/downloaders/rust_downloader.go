package downloaders

import (
	"context"
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

// RustDownloader implements CompilerDownloader for Rust
// Uses rustup-init for automated installation with wasm32-unknown-unknown target
type RustDownloader struct {
	*BaseDownloader
}

// NewRustDownloader creates a new Rust downloader
func NewRustDownloader() *RustDownloader {
	return &RustDownloader{
		BaseDownloader: NewBaseDownloader(),
	}
}

// GetDownloadURL returns the download URL for rustup-init
func (r *RustDownloader) GetDownloadURL(req domain.DownloadRequest) (string, error) {
	baseURL := "https://static.rust-lang.org/rustup/dist"

	// Map platform and arch to rustup naming
	platform := req.Platform
	arch := req.Arch

	// rustup naming convention
	var target string
	switch {
	case platform == "darwin" && arch == "amd64":
		target = "x86_64-apple-darwin"
	case platform == "darwin" && arch == "arm64":
		target = "aarch64-apple-darwin"
	case platform == "linux" && arch == "amd64":
		target = "x86_64-unknown-linux-gnu"
	case platform == "linux" && arch == "arm64":
		target = "aarch64-unknown-linux-gnu"
	case platform == "windows" && arch == "amd64":
		target = "x86_64-pc-windows-msvc"
	case platform == "windows" && arch == "386":
		target = "i686-pc-windows-msvc"
	default:
		return "", fmt.Errorf("unsupported platform/arch: %s/%s", platform, arch)
	}

	// Construct URL
	ext := ""
	if platform == "windows" {
		ext = ".exe"
	}

	return fmt.Sprintf("%s/%s/rustup-init%s", baseURL, target, ext), nil
}

// Download downloads and installs Rust compiler with rustup
func (r *RustDownloader) Download(
	ctx context.Context,
	req domain.DownloadRequest,
	progressCb func(domain.DownloadProgress),
) error {
	// Get download URL
	downloadURL, err := r.GetDownloadURL(req)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "downloading",
			Message: "Downloading rustup-init...",
		})
	}

	// Create temp directory for download
	tmpDir, err := os.MkdirTemp("", "rust-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download rustup-init
	rustupInit := filepath.Join(tmpDir, "rustup-init")
	if runtime.GOOS == "windows" {
		rustupInit += ".exe"
	}

	// Download file
	err = r.downloadRustupInit(ctx, downloadURL, rustupInit, progressCb)
	if err != nil {
		return fmt.Errorf("failed to download rustup-init: %w", err)
	}

	// Make executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(rustupInit, 0755); err != nil {
			return fmt.Errorf("failed to make rustup-init executable: %w", err)
		}
	}

	// Install Rust using rustup-init
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "installing",
			Message: "Installing Rust toolchain...",
		})
	}

	err = r.installRust(ctx, rustupInit, req.TargetDir, progressCb)
	if err != nil {
		return fmt.Errorf("failed to install Rust: %w", err)
	}

	// Install wasm32-unknown-unknown target
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "installing",
			Message: "Installing wasm32-unknown-unknown target...",
		})
	}

	err = r.installWasmTarget(req.TargetDir)
	if err != nil {
		return fmt.Errorf("failed to install wasm32 target: %w", err)
	}

	// Verify installation
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:   "verifying",
			Message: "Verifying installation...",
		})
	}

	err = r.Verify(req.TargetDir)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}

	// Complete
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:      "complete",
			Percentage: 100,
			Message:    "Rust toolchain installed successfully!",
		})
	}

	return nil
}

// downloadRustupInit downloads rustup-init binary
func (r *RustDownloader) downloadRustupInit(ctx context.Context, url, dest string, progressCb func(domain.DownloadProgress)) error {
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// Execute request
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create destination file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy with progress
	totalBytes := resp.ContentLength
	var downloadedBytes int64

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloadedBytes += int64(n)

			// Report progress
			if progressCb != nil && totalBytes > 0 {
				percentage := float64(downloadedBytes) / float64(totalBytes) * 100
				progressCb(domain.DownloadProgress{
					Stage:           "downloading",
					Percentage:      percentage,
					BytesDownloaded: downloadedBytes,
					TotalBytes:      totalBytes,
					Message:         fmt.Sprintf("Downloading rustup-init... %d%%", int(percentage)),
				})
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// installRust runs rustup-init to install Rust
func (r *RustDownloader) installRust(ctx context.Context, rustupInit, targetDir string, progressCb func(domain.DownloadProgress)) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Set RUSTUP_HOME and CARGO_HOME to target directory
	rustupHome := filepath.Join(targetDir, "rustup")
	cargoHome := filepath.Join(targetDir, "cargo")

	// Run rustup-init with minimal profile
	// -y: assume yes to prompts
	// --no-modify-path: don't modify PATH
	// --default-toolchain stable: install stable toolchain
	// --profile minimal: minimal installation (rustc, cargo, rust-std)
	cmd := exec.CommandContext(ctx, rustupInit,
		"-y",
		"--no-modify-path",
		"--default-toolchain", "stable",
		"--profile", "minimal",
	)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("RUSTUP_HOME=%s", rustupHome),
		fmt.Sprintf("CARGO_HOME=%s", cargoHome),
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rustup-init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// installWasmTarget installs wasm32-unknown-unknown target
func (r *RustDownloader) installWasmTarget(installPath string) error {
	rustupBinary := r.getRustupPath(installPath)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Run: rustup target add wasm32-unknown-unknown
	cmd := exec.CommandContext(ctx, rustupBinary, "target", "add", "wasm32-unknown-unknown")

	// Set environment variables
	rustupHome := filepath.Join(installPath, "rustup")
	cargoHome := filepath.Join(installPath, "cargo")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("RUSTUP_HOME=%s", rustupHome),
		fmt.Sprintf("CARGO_HOME=%s", cargoHome),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add wasm32 target: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Verify checks that Rust compiler works correctly
func (r *RustDownloader) Verify(installPath string) error {
	cargoBinary := r.getCargoBinPath(installPath)

	// Check if cargo binary exists
	if _, err := os.Stat(cargoBinary); err != nil {
		return fmt.Errorf("cargo binary not found at %s: %w", cargoBinary, err)
	}

	// Make binary executable (Unix-like systems)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(cargoBinary, 0755); err != nil {
			return fmt.Errorf("failed to make cargo executable: %w", err)
		}
	}

	// Set environment variables
	rustupHome := filepath.Join(installPath, "rustup")
	cargoHome := filepath.Join(installPath, "cargo")

	// Run 'cargo --version' to verify it works
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, cargoBinary, "--version")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("RUSTUP_HOME=%s", rustupHome),
		fmt.Sprintf("CARGO_HOME=%s", cargoHome),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo version check failed: %w (output: %s)", err, string(output))
	}

	// Check output contains "cargo"
	if !strings.Contains(string(output), "cargo") {
		return fmt.Errorf("unexpected cargo version output: %s", string(output))
	}

	// Verify wasm32 target is installed
	rustupBinary := r.getRustupPath(installPath)
	cmd = exec.CommandContext(ctx, rustupBinary, "target", "list", "--installed")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("RUSTUP_HOME=%s", rustupHome),
		fmt.Sprintf("CARGO_HOME=%s", cargoHome),
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rustup target list failed: %w", err)
	}

	if !strings.Contains(string(output), "wasm32-unknown-unknown") {
		return fmt.Errorf("wasm32-unknown-unknown target not installed")
	}

	return nil
}

// GetMetadata returns metadata about Rust compiler download
func (r *RustDownloader) GetMetadata(req domain.DownloadRequest) (*domain.CompilerMetadata, error) {
	// Get download URL
	downloadURL, err := r.GetDownloadURL(req)
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

	resp, err := r.httpClient.Do(headReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// rustup-init is small (~10 MB), but total installation will be ~600 MB
	// We return estimated total size
	estimatedSize := int64(600 * 1024 * 1024) // 600 MB

	return &domain.CompilerMetadata{
		Language:    "rust",
		Version:     "stable",
		DownloadURL: downloadURL,
		Size:        estimatedSize,
		Checksum:    "", // Rustup verifies checksums internally
		ReleaseDate: time.Now(),
	}, nil
}

// Extract is not used for Rust (rustup-init handles installation)
func (r *RustDownloader) Extract(archivePath, targetDir string) error {
	return fmt.Errorf("extract not implemented for Rust (uses rustup-init)")
}

// getRustupPath returns the path to rustup binary
func (r *RustDownloader) getRustupPath(installPath string) string {
	rustupBinary := filepath.Join(installPath, "cargo", "bin", "rustup")
	if runtime.GOOS == "windows" {
		rustupBinary += ".exe"
	}
	return rustupBinary
}

// getCargoBinPath returns the path to cargo binary
func (r *RustDownloader) getCargoBinPath(installPath string) string {
	cargoBinary := filepath.Join(installPath, "cargo", "bin", "cargo")
	if runtime.GOOS == "windows" {
		cargoBinary += ".exe"
	}
	return cargoBinary
}
