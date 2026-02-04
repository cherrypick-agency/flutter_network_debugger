package downloaders

import (
	"archive/zip"
	"os"
	"path/filepath"
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
		t.Error("BaseDownloader is nil")
	}
}

// Composer 1.
func TestKotlinDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewKotlinDownloader()

	tests := []struct {
		name     string
		req      domain.DownloadRequest
		contains string
	}{
		{
			name: "latest version",
			req: domain.DownloadRequest{
				Version: "latest",
			},
			contains: "kotlin-compiler-2.2.20.zip",
		},
		{
			name: "specific version",
			req: domain.DownloadRequest{
				Version: "2.1.0",
			},
			contains: "kotlin-compiler-2.1.0.zip",
		},
		{
			name: "empty version",
			req: domain.DownloadRequest{
				Version: "",
			},
			contains: "kotlin-compiler-2.2.20.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := downloader.GetDownloadURL(tt.req)
			if err != nil {
				t.Fatalf("GetDownloadURL() error = %v, want nil", err)
			}

			if url == "" {
				t.Error("GetDownloadURL() returned empty URL")
			}

			if !contains(url, tt.contains) {
				t.Errorf("GetDownloadURL() = %q, should contain %q", url, tt.contains)
			}

			if !contains(url, "github.com/JetBrains/kotlin/releases") {
				t.Error("URL should point to GitHub releases")
			}
		})
	}
}

// Composer 1.
func TestKotlinDownloader_GetMetadata(t *testing.T) {
	downloader := NewKotlinDownloader()

	req := domain.DownloadRequest{
		Version: "latest",
	}

	metadata, err := downloader.GetMetadata(req)
	if err != nil {
		t.Fatalf("GetMetadata() error = %v, want nil", err)
	}

	if metadata == nil {
		t.Fatal("GetMetadata() returned nil")
	}

	if metadata.Language != "kotlin" {
		t.Errorf("Expected language 'kotlin', got %q", metadata.Language)
	}

	if metadata.Version == "" {
		t.Error("Version is empty")
	}

	if metadata.DownloadURL == "" {
		t.Error("DownloadURL is empty")
	}

	if metadata.Size <= 0 {
		t.Error("Size should be positive")
	}
}

// Composer 1.
func TestKotlinDownloader_Extract_Success(t *testing.T) {
	downloader := NewKotlinDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "kotlin.zip")

	// Create zip archive with kotlinc directory
	zipFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create zip: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add kotlinc directory
	_, err = zipWriter.Create("kotlinc/")
	if err != nil {
		t.Fatalf("Failed to create directory in zip: %v", err)
	}

	// Add file to kotlinc
	f, err := zipWriter.Create("kotlinc/bin/kotlinc")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("kotlinc binary"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zipWriter.Close()
	zipFile.Close()

	// Extract
	err = downloader.Extract(archivePath, targetDir)
	if err != nil {
		t.Fatalf("Extract() error = %v, want nil", err)
	}

	// Check that file is extracted to targetDir root
	kotlincFile := filepath.Join(targetDir, "bin", "kotlinc")
	if _, err := os.Stat(kotlincFile); os.IsNotExist(err) {
		t.Error("kotlinc binary was not extracted to target directory root")
	}
}

// Composer 1.
func TestKotlinDownloader_Extract_NoKotlincDir(t *testing.T) {
	downloader := NewKotlinDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "kotlin.zip")

	// Create zip archive without kotlinc directory
	zipFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create zip: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add regular file
	f, err := zipWriter.Create("test.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zipWriter.Close()
	zipFile.Close()

	// Extraction should succeed (kotlinc directory is optional)
	err = downloader.Extract(archivePath, targetDir)
	if err != nil {
		t.Fatalf("Extract() error = %v, want nil (kotlinc dir is optional)", err)
	}
}
