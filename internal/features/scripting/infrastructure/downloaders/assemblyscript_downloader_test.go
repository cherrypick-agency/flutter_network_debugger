package downloaders

import (
	"archive/tar"
	"compress/gzip"
	"os"
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
		t.Error("BaseDownloader is nil")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_getAssetName(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

	tests := []struct {
		name     string
		platform string
		arch     string
		version  string
		contains string
	}{
		{
			name:     "darwin amd64",
			platform: "darwin",
			arch:     "amd64",
			version:  "v20.11.0",
			contains: "node-v20.11.0-darwin-x64.tar.gz",
		},
		{
			name:     "darwin arm64",
			platform: "darwin",
			arch:     "arm64",
			version:  "v20.11.0",
			contains: "node-v20.11.0-darwin-arm64.tar.gz",
		},
		{
			name:     "linux amd64",
			platform: "linux",
			arch:     "amd64",
			version:  "v20.11.0",
			contains: "node-v20.11.0-linux-x64.tar.gz",
		},
		{
			name:     "windows amd64",
			platform: "windows",
			arch:     "amd64",
			version:  "v20.11.0",
			contains: "node-v20.11.0-win-x64.zip",
		},
		{
			name:     "windows 386",
			platform: "windows",
			arch:     "386",
			version:  "v20.11.0",
			contains: "node-v20.11.0-win-x86.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := downloader.getAssetName(tt.platform, tt.arch, tt.version)
			if !contains(got, tt.contains) {
				t.Errorf("getAssetName() = %q, should contain %q", got, tt.contains)
			}

			if tt.platform == "windows" && !contains(got, ".zip") {
				t.Error("Windows asset should be .zip")
			}

			if tt.platform != "windows" && !contains(got, ".tar.gz") {
				t.Error("Non-Windows asset should be .tar.gz")
			}
		})
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

	req := domain.DownloadRequest{
		Platform: "darwin",
		Arch:     "amd64",
	}

	url, err := downloader.GetDownloadURL(req)
	if err != nil {
		t.Fatalf("GetDownloadURL() error = %v, want nil", err)
	}

	if url == "" {
		t.Error("GetDownloadURL() returned empty URL")
	}

	if !contains(url, "nodejs.org/dist") {
		t.Error("URL should point to nodejs.org/dist")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_GetMetadata(t *testing.T) {
	downloader := NewAssemblyScriptDownloader()

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

	if metadata.Language != "assemblyscript" {
		t.Errorf("Expected language 'assemblyscript', got %q", metadata.Language)
	}

	if metadata.Version == "" {
		t.Error("Version is empty")
	}

	if metadata.DownloadURL == "" {
		t.Error("DownloadURL is empty")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_Extract_TarGz(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping tar.gz test on Windows")
	}

	downloader := NewAssemblyScriptDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "node.tar.gz")

	// Create tar.gz archive with node-* directory
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add directory node-v20.11.0
	header := &tar.Header{
		Name:     "node-v20.11.0-darwin-x64/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	// Add file to directory
	content := []byte("node binary")
	header = &tar.Header{
		Name:     "node-v20.11.0-darwin-x64/bin/node",
		Size:     int64(len(content)),
		Mode:     0755,
		Typeflag: tar.TypeReg,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Extract
	err = downloader.Extract(archivePath, targetDir)
	if err != nil {
		t.Fatalf("Extract() error = %v, want nil", err)
	}

	// Check that file is extracted to targetDir root
	nodeFile := filepath.Join(targetDir, "bin", "node")
	if _, err := os.Stat(nodeFile); os.IsNotExist(err) {
		t.Error("node binary was not extracted to target directory root")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_Extract_NoNodeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping tar.gz test on Windows")
	}

	downloader := NewAssemblyScriptDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "node.tar.gz")

	// Create tar.gz archive without node-* directory
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add regular file
	content := []byte("test")
	header := &tar.Header{
		Name:     "test.txt",
		Size:     int64(len(content)),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Extraction should return error
	err = downloader.Extract(archivePath, targetDir)
	if err == nil {
		t.Error("Extract() should return error when node directory not found")
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_getNpmPath(t *testing.T) {
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
			t.Errorf("getNpmPath() = %q, should contain bin/npm", path)
		}
	}
}

// Composer 1.
func TestAssemblyScriptDownloader_getAscPath(t *testing.T) {
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
			t.Errorf("getAscPath() = %q, should contain bin/asc", path)
		}
	}
}
