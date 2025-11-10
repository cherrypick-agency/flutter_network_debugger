package downloaders

import (
	"runtime"
	"strings"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewRustDownloader(t *testing.T) {
	downloader := NewRustDownloader()

	if downloader == nil {
		t.Fatal("NewRustDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Error("BaseDownloader is nil")
	}
}

// Composer 1.
func TestRustDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewRustDownloader()

	tests := []struct {
		name     string
		req      domain.DownloadRequest
		wantErr  bool
		contains string
	}{
		{
			name: "darwin amd64",
			req: domain.DownloadRequest{
				Platform: "darwin",
				Arch:     "amd64",
			},
			wantErr:  false,
			contains: "x86_64-apple-darwin",
		},
		{
			name: "darwin arm64",
			req: domain.DownloadRequest{
				Platform: "darwin",
				Arch:     "arm64",
			},
			wantErr:  false,
			contains: "aarch64-apple-darwin",
		},
		{
			name: "linux amd64",
			req: domain.DownloadRequest{
				Platform: "linux",
				Arch:     "amd64",
			},
			wantErr:  false,
			contains: "x86_64-unknown-linux-gnu",
		},
		{
			name: "linux arm64",
			req: domain.DownloadRequest{
				Platform: "linux",
				Arch:     "arm64",
			},
			wantErr:  false,
			contains: "aarch64-unknown-linux-gnu",
		},
		{
			name: "windows amd64",
			req: domain.DownloadRequest{
				Platform: "windows",
				Arch:     "amd64",
			},
			wantErr:  false,
			contains: "x86_64-pc-windows-msvc",
		},
		{
			name: "windows 386",
			req: domain.DownloadRequest{
				Platform: "windows",
				Arch:     "386",
			},
			wantErr:  false,
			contains: "i686-pc-windows-msvc",
		},
		{
			name: "unsupported platform",
			req: domain.DownloadRequest{
				Platform: "unsupported",
				Arch:     "amd64",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := downloader.GetDownloadURL(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDownloadURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if url == "" {
					t.Error("GetDownloadURL() returned empty URL")
				}
				if tt.contains != "" && !contains(url, tt.contains) {
					t.Errorf("GetDownloadURL() = %q, should contain %q", url, tt.contains)
				}
				if tt.req.Platform == "windows" && !contains(url, ".exe") {
					t.Error("Windows URL should contain .exe")
				}
			}
		})
	}
}

// Composer 1.
func TestRustDownloader_GetMetadata(t *testing.T) {
	downloader := NewRustDownloader()

	req := domain.DownloadRequest{
		Platform: "darwin",
		Arch:     "amd64",
	}

	metadata, err := downloader.GetMetadata(req)
	if err != nil {
		t.Fatalf("GetMetadata() error = %v, want nil", err)
	}

	if metadata == nil {
		t.Fatal("GetMetadata() returned nil")
	}

	if metadata.Language != "rust" {
		t.Errorf("Expected language 'rust', got %q", metadata.Language)
	}

	if metadata.Version != "stable" {
		t.Errorf("Expected version 'stable', got %q", metadata.Version)
	}

	if metadata.DownloadURL == "" {
		t.Error("DownloadURL is empty")
	}

	if metadata.Size <= 0 {
		t.Error("Size should be positive")
	}
}

// Composer 1.
func TestRustDownloader_getRustupPath(t *testing.T) {
	downloader := NewRustDownloader()
	tmpDir := t.TempDir()

	path := downloader.getRustupPath(tmpDir)
	if path == "" {
		t.Error("getRustupPath() returned empty path")
	}

	expectedSuffix := "rustup"
	if runtime.GOOS == "windows" {
		expectedSuffix = "rustup.exe"
	}

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("getRustupPath() = %q, should end with %q", path, expectedSuffix)
	}
}

// Composer 1.
func TestRustDownloader_getCargoBinPath(t *testing.T) {
	downloader := NewRustDownloader()
	tmpDir := t.TempDir()

	path := downloader.getCargoBinPath(tmpDir)
	if path == "" {
		t.Error("getCargoBinPath() returned empty path")
	}

	expectedSuffix := "cargo"
	if runtime.GOOS == "windows" {
		expectedSuffix = "cargo.exe"
	}

	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("getCargoBinPath() = %q, should end with %q", path, expectedSuffix)
	}
}

// Composer 1.
func TestRustDownloader_Extract(t *testing.T) {
	downloader := NewRustDownloader()

	err := downloader.Extract("archive", "target")
	if err == nil {
		t.Error("Extract() should return error (not implemented for Rust)")
	}

	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("Expected error about not implemented, got %q", err.Error())
	}
}
