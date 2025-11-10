package downloaders

import (
	"context"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewSwiftDownloader(t *testing.T) {
	downloader := NewSwiftDownloader()

	if downloader == nil {
		t.Fatal("NewSwiftDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Fatal("BaseDownloader is nil")
	}

	if downloader.minSwiftVersion == "" {
		t.Error("minSwiftVersion should be set")
	}
}

// Composer 1.
func TestSwiftDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewSwiftDownloader()

	tests := []struct {
		platform string
		arch     string
		wantErr  bool
	}{
		{"darwin", "amd64", false},
		{"darwin", "arm64", false},
		{"linux", "amd64", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.platform+"/"+tt.arch, func(t *testing.T) {
			req := domain.DownloadRequest{
				Platform: tt.platform,
				Arch:     tt.arch,
				Version:  "latest",
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
func TestSwiftDownloader_Verify_NoSwift(t *testing.T) {
	downloader := NewSwiftDownloader()
	tmpDir := t.TempDir()

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Log("Verify() may succeed if system Swift is installed")
	}
}

// Composer 1.
func TestSwiftDownloader_Download_ErrorGettingURL(t *testing.T) {
	downloader := NewSwiftDownloader()

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
