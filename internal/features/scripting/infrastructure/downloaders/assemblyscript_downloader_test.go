package downloaders

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewAssemblyScriptDownloader(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

	if downloader == nil {
		t.Fatal("NewAssemblyScriptDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Fatal("BaseDownloader is nil")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

	releases := []nodeRelease{
		{
			Version: "v20.11.0",
			LTS:     "Hydrogen",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/dist/index.json") {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(releases)
		}
	}))
	defer server.Close()

	req := domain.DownloadRequest{
		Platform: runtime.GOOS,
		Arch:     runtime.GOARCH,
	}

	url, err := downloader.GetDownloadURL(req)
	if err != nil {
		t.Fatalf("GetDownloadURL() error = %v, want nil", err)
	}

	if url == "" {
		t.Error("GetDownloadURL() returned empty URL")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_GetMetadata(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

	testData := []byte("fake node binary")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/dist/index.json") {
			releases := []nodeRelease{
				{
					Version: "v20.11.0",
					LTS:     "Hydrogen",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(releases)
		} else if r.Method == http.MethodHead {
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

	if metadata.Language != "assemblyscript" {
		t.Errorf("Language = %q, want %q", metadata.Language, "assemblyscript")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_Extract_InvalidArchive(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "nonexistent.tar.gz")

	err := downloader.Extract(archivePath, targetDir)
	if err == nil {
		t.Error("Extract() should return error for nonexistent archive")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_Verify_NoBinary(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()
	tmpDir := t.TempDir()

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Error("Verify() should return error when asc binary doesn't exist")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_GetNpmPath(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()
	tmpDir := t.TempDir()

	path := downloader.getNpmPath(tmpDir)
	if path == "" {
		t.Error("getNpmPath() returned empty path")
	}

	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(path, "npm.cmd") {
			t.Errorf("getNpmPath() = %q, should end with npm.cmd on Windows", path)
		}
	} else {
		if !strings.Contains(path, "bin/npm") {
			t.Errorf("getNpmPath() = %q, should contain bin/npm on Unix", path)
		}
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_GetAscPath(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()
	tmpDir := t.TempDir()

	path := downloader.getAscPath(tmpDir)
	if path == "" {
		t.Error("getAscPath() returned empty path")
	}

	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(path, "asc.cmd") {
			t.Errorf("getAscPath() = %q, should end with asc.cmd on Windows", path)
		}
	} else {
		if !strings.Contains(path, "bin/asc") {
			t.Errorf("getAscPath() = %q, should contain bin/asc on Unix", path)
		}
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_Download_ErrorGettingNodeVersion(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

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
