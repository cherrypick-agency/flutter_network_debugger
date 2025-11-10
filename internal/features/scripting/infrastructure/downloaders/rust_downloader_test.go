package downloaders

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		t.Fatal("BaseDownloader is nil")
	}
}

// Composer 1.
func TestRustDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewRustDownloader()

	tests := []struct {
		platform string
		arch     string
		wantErr  bool
	}{
		{"darwin", "amd64", false},
		{"darwin", "arm64", false},
		{"linux", "amd64", false},
		{"linux", "arm64", false},
		{"windows", "amd64", false},
		{"windows", "386", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.platform+"/"+tt.arch, func(t *testing.T) {
			req := domain.DownloadRequest{
				Platform: tt.platform,
				Arch:     tt.arch,
			}

			url, err := downloader.GetDownloadURL(req)
			if tt.wantErr {
				if err == nil {
					t.Error("GetDownloadURL() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetDownloadURL() error = %v, want nil", err)
			}

			if url == "" {
				t.Error("GetDownloadURL() returned empty URL")
			}
		})
	}
}

// Composer 1.
func TestRustDownloader_GetMetadata(t *testing.T) {
	downloader := NewRustDownloader()

	testData := []byte("fake rustup-init")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	req := domain.DownloadRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
	}

	metadata, err := downloader.GetMetadata(req)
	if err != nil {
		t.Fatalf("GetMetadata() error = %v, want nil", err)
	}

	if metadata == nil {
		t.Fatal("GetMetadata() returned nil")
	}

	if metadata.Language != "rust" {
		t.Errorf("Language = %q, want %q", metadata.Language, "rust")
	}

	if metadata.Size <= 0 {
		t.Errorf("Size = %d, want > 0", metadata.Size)
	}
}

// Composer 1.
func TestRustDownloader_Extract(t *testing.T) {
	downloader := NewRustDownloader()

	err := downloader.Extract("archive", "target")
	if err == nil {
		t.Error("Extract() should return error (not implemented for Rust)")
	}
}

// Composer 1.
func TestRustDownloader_Verify_NoBinary(t *testing.T) {
	downloader := NewRustDownloader()
	tmpDir := t.TempDir()

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Error("Verify() should return error when cargo binary doesn't exist")
	}
}

// Composer 1.
func TestRustDownloader_GetRustupPath(t *testing.T) {
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
func TestRustDownloader_GetCargoBinPath(t *testing.T) {
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
func TestRustDownloader_Download_ErrorGettingURL(t *testing.T) {
	downloader := NewRustDownloader()

	req := domain.DownloadRequest{
		Platform:  "invalid",
		Arch:      "invalid",
		TargetDir: t.TempDir(),
	}

	ctx := context.Background()
	err := downloader.Download(ctx, req, nil)
	if err == nil {
		t.Error("Download() should return error when GetDownloadURL fails")
	}
}
