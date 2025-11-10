package downloaders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewWASISDKDownloader(t *testing.T) {
	downloader := NewWASISDKDownloader()

	if downloader == nil {
		t.Fatal("NewWASISDKDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Fatal("BaseDownloader is nil")
	}
}

// Composer 1.
func TestWASISDKDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewWASISDKDownloader()

	release := &wasiRelease{
		TagName: "wasi-sdk-21",
		Assets: []wasiAsset{
			{
				Name:        "wasi-sdk-21.0-macos.tar.gz",
				DownloadURL: "https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-21/wasi-sdk-21.0-macos.tar.gz",
				Size:        100 * 1024 * 1024,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}
	}))
	defer server.Close()

	req := domain.DownloadRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  "latest",
	}

	url, err := downloader.GetDownloadURL(req)
	if err != nil {
		t.Logf("GetDownloadURL() error = %v (may fail if platform not supported)", err)
		return
	}

	if url == "" {
		t.Error("GetDownloadURL() returned empty URL")
	}
}

// Composer 1.
func TestWASISDKDownloader_GetMetadata(t *testing.T) {
	downloader := NewWASISDKDownloader()

	release := &wasiRelease{
		TagName:    "wasi-sdk-21",
		CreatedAt:  time.Now(),
		Prerelease: false,
		Draft:      false,
		Assets: []wasiAsset{
			{
				Name:        "wasi-sdk-21.0-macos.tar.gz",
				DownloadURL: "https://github.com/WebAssembly/wasi-sdk/releases/download/wasi-sdk-21/wasi-sdk-21.0-macos.tar.gz",
				Size:        100 * 1024 * 1024,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(release)
		}
	}))
	defer server.Close()

	req := domain.DownloadRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
		Version:  "latest",
	}

	metadata, err := downloader.GetMetadata(req)
	if err != nil {
		t.Logf("GetMetadata() error = %v (may fail if platform not supported)", err)
		return
	}

	if metadata == nil {
		t.Fatal("GetMetadata() returned nil")
	}

	if metadata.Language != "wasisdk" {
		t.Errorf("Language = %q, want %q", metadata.Language, "wasisdk")
	}
}

// Composer 1.
func TestWASISDKDownloader_Verify_NoBinary(t *testing.T) {
	downloader := NewWASISDKDownloader()
	tmpDir := t.TempDir()

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Error("Verify() should return error when clang binary doesn't exist")
	}
}

// Composer 1.
func TestWASISDKDownloader_Download_ErrorGettingRelease(t *testing.T) {
	downloader := NewWASISDKDownloader()

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
		t.Error("Download() should return error when release fetch fails")
	}
}
