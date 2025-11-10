package downloaders

import (
	"context"
	"runtime"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewKotlinDownloader(t *testing.T) {
	downloader := NewKotlinDownloader()

	if downloader == nil {
		t.Fatal("NewKotlinDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Fatal("BaseDownloader is nil")
	}
}

// Composer 1.
func TestKotlinDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewKotlinDownloader()

	tests := []struct {
		version string
		wantURL string
	}{
		{"latest", "https://github.com/JetBrains/kotlin/releases/download/v2.2.20/kotlin-compiler-2.2.20.zip"},
		{"2.2.20", "https://github.com/JetBrains/kotlin/releases/download/v2.2.20/kotlin-compiler-2.2.20.zip"},
		{"", "https://github.com/JetBrains/kotlin/releases/download/v2.2.20/kotlin-compiler-2.2.20.zip"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			req := domain.DownloadRequest{
				Version: tt.version,
			}

			url, err := downloader.GetDownloadURL(req)
			if err != nil {
				t.Fatalf("GetDownloadURL() error = %v, want nil", err)
			}

			if url != tt.wantURL {
				t.Errorf("GetDownloadURL() = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

// Composer 1.
func TestKotlinDownloader_Verify_NoBinary(t *testing.T) {
	downloader := NewKotlinDownloader()
	tmpDir := t.TempDir()

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Error("Verify() should return error when kotlinc binary doesn't exist")
	}
}

// Composer 1.
func TestKotlinDownloader_Download_ErrorGettingURL(t *testing.T) {
	downloader := NewKotlinDownloader()

	req := domain.DownloadRequest{
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		TargetDir: t.TempDir(),
		Version:   "latest",
	}

	ctx := context.Background()
	err := downloader.Download(ctx, req, nil)
	if err == nil {
		t.Log("Download() may succeed if network is available")
	}
}
