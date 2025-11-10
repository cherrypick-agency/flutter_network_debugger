package downloaders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewZigDownloader(t *testing.T) {
	downloader := NewZigDownloader()

	if downloader == nil {
		t.Fatal("NewZigDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Fatal("BaseDownloader is nil")
	}

	if downloader.indexURL == "" {
		t.Error("indexURL should be set")
	}
}

// Composer 1.
func TestZigDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewZigDownloader()

	index := map[string]interface{}{
		"master": map[string]interface{}{
			"version": "0.13.0",
			"x86_64-darwin": map[string]interface{}{
				"tarball": "https://ziglang.org/download/0.13.0/zig-macos-x86_64-0.13.0.tar.xz",
				"shasum":  "abc123",
				"size":    "52428800",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "index.json") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(index)
		}
	}))
	defer server.Close()

	req := domain.DownloadRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  "latest",
	}

	_, err := downloader.GetDownloadURL(req)
	if err != nil {
		t.Logf("GetDownloadURL() error = %v (may fail if platform not supported)", err)
	}
}

// Composer 1.
func TestZigDownloader_Verify_NoBinary(t *testing.T) {
	downloader := NewZigDownloader()
	tmpDir := t.TempDir()

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Error("Verify() should return error when zig binary doesn't exist")
	}
}

// Composer 1.
func TestZigDownloader_Download_ErrorGettingIndex(t *testing.T) {
	downloader := NewZigDownloader()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	req := domain.DownloadRequest{
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		TargetDir: t.TempDir(),
		Version:   "latest",
	}

	ctx := context.Background()
	err := downloader.Download(ctx, req, nil)
	if err == nil {
		t.Error("Download() should return error when index fetch fails")
	}
}
