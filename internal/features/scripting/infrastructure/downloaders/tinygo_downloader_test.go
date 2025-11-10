package downloaders

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewTinyGoDownloader(t *testing.T) {
	downloader := NewTinyGoDownloader()

	if downloader == nil {
		t.Fatal("NewTinyGoDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Fatal("BaseDownloader is nil")
	}
}

// Composer 1.
func TestTinyGoDownloader_GetDownloadURL_Error(t *testing.T) {
	downloader := NewTinyGoDownloader()

	req := domain.DownloadRequest{
		Platform: "invalid",
		Arch:     "invalid",
		Version:  "latest",
	}

	_, err := downloader.GetDownloadURL(req)
	if err == nil {
		t.Error("GetDownloadURL() should return error for invalid platform/arch")
	}
}

// Composer 1.
func TestTinyGoDownloader_GetMetadata_Error(t *testing.T) {
	downloader := NewTinyGoDownloader()

	req := domain.DownloadRequest{
		Platform: "invalid",
		Arch:     "invalid",
		Version:  "latest",
	}

	_, err := downloader.GetMetadata(req)
	if err == nil {
		t.Error("GetMetadata() should return error for invalid platform/arch")
	}
}

// Composer 1.
func TestTinyGoDownloader_Extract_InvalidArchive(t *testing.T) {
	downloader := NewTinyGoDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "nonexistent.tar.gz")

	err := downloader.Extract(archivePath, targetDir)
	if err == nil {
		t.Error("Extract() should return error for nonexistent archive")
	}
}

// Composer 1.
func TestTinyGoDownloader_Verify(t *testing.T) {
	downloader := NewTinyGoDownloader()
	tmpDir := t.TempDir()

	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin directory: %v", err)
	}

	binaryName := "tinygo"
	if runtime.GOOS == "windows" {
		binaryName = "tinygo.exe"
	}
	binaryPath := filepath.Join(binDir, binaryName)

	err := downloader.Verify(tmpDir)
	if err == nil {
		t.Error("Verify() should return error when binary doesn't exist")
	}

	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\necho 'tinygo version 0.33.0'\n"), 0755); err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0755); err != nil {
			t.Fatalf("Failed to make binary executable: %v", err)
		}
	}

	err = downloader.Verify(tmpDir)
	if err != nil {
		t.Logf("Verify() error = %v (expected for fake binary)", err)
	}
}

// Composer 1.
func TestTinyGoDownloader_Download_ErrorGettingMetadata(t *testing.T) {
	downloader := NewTinyGoDownloader()

	req := domain.DownloadRequest{
		Platform:  "invalid",
		Arch:      "invalid",
		Version:   "latest",
		TargetDir: t.TempDir(),
	}

	ctx := context.Background()
	err := downloader.Download(ctx, req, nil)
	if err == nil {
		t.Error("Download() should return error when GetMetadata fails")
	}
}
