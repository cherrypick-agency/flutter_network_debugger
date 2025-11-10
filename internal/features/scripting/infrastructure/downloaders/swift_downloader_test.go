package downloaders

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
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
		t.Error("BaseDownloader is nil")
	}

	if downloader.minSwiftVersion == "" {
		t.Error("minSwiftVersion is empty")
	}
}

// Composer 1.
func TestSwiftDownloader_mapPlatformToSwiftTarget(t *testing.T) {
	downloader := NewSwiftDownloader()

	tests := []struct {
		name     string
		platform string
		arch     string
		wantErr  bool
		expected string
	}{
		{
			name:     "darwin amd64",
			platform: "darwin",
			arch:     "amd64",
			wantErr:  false,
			expected: "macos_x86_64",
		},
		{
			name:     "darwin arm64",
			platform: "darwin",
			arch:     "arm64",
			wantErr:  false,
			expected: "macos_arm64",
		},
		{
			name:     "linux amd64",
			platform: "linux",
			arch:     "amd64",
			wantErr:  false,
			expected: "ubuntu22.04_x86_64",
		},
		{
			name:     "linux arm64",
			platform: "linux",
			arch:     "arm64",
			wantErr:  false,
			expected: "ubuntu22.04_aarch64",
		},
		{
			name:     "unsupported platform",
			platform: "windows",
			arch:     "amd64",
			wantErr:  true,
		},
		{
			name:     "unsupported arch",
			platform: "darwin",
			arch:     "386",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := downloader.mapPlatformToSwiftTarget(tt.platform, tt.arch)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapPlatformToSwiftTarget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.expected {
				t.Errorf("mapPlatformToSwiftTarget() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// Composer 1.
func TestSwiftDownloader_GetDownloadURL(t *testing.T) {
	downloader := NewSwiftDownloader()

	tests := []struct {
		name     string
		req      domain.DownloadRequest
		contains string
	}{
		{
			name: "darwin arm64 latest",
			req: domain.DownloadRequest{
				Platform: "darwin",
				Arch:     "arm64",
				Version:  "latest",
			},
			contains: "swift-wasm-6.1.0-RELEASE",
		},
		{
			name: "linux amd64 specific version",
			req: domain.DownloadRequest{
				Platform: "linux",
				Arch:     "amd64",
				Version:  "6.0.0",
			},
			contains: "swift-wasm-6.0.0-RELEASE",
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

			if tt.req.Platform == "darwin" && !contains(url, ".pkg") {
				t.Error("macOS URL should contain .pkg")
			}

			if tt.req.Platform == "linux" && !contains(url, ".tar.gz") {
				t.Error("Linux URL should contain .tar.gz")
			}
		})
	}
}

// Composer 1.
func TestSwiftDownloader_GetMetadata(t *testing.T) {
	downloader := NewSwiftDownloader()

	req := domain.DownloadRequest{
		Platform: "darwin",
		Arch:     "arm64",
		Version:  "latest",
	}

	metadata, err := downloader.GetMetadata(req)
	if err != nil {
		t.Fatalf("GetMetadata() error = %v, want nil", err)
	}

	if metadata == nil {
		t.Fatal("GetMetadata() returned nil")
	}

	if metadata.Language != "swift" {
		t.Errorf("Expected language 'swift', got %q", metadata.Language)
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
func TestSwiftDownloader_Extract_TarGz(t *testing.T) {
	downloader := NewSwiftDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "swift.tar.gz")

	// Создаем tar.gz архив
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Добавляем файл
	content := []byte("swift wasm sdk")
	header := &tar.Header{
		Name:     "swift.txt",
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

	// Извлекаем
	err = downloader.Extract(archivePath, targetDir)
	if err != nil {
		t.Fatalf("Extract() error = %v, want nil", err)
	}

	// Проверяем что файл извлечен
	swiftFile := filepath.Join(targetDir, "swift.txt")
	if _, err := os.Stat(swiftFile); os.IsNotExist(err) {
		t.Error("swift file was not extracted")
	}
}

// Composer 1.
func TestSwiftDownloader_Extract_UnsupportedFormat(t *testing.T) {
	downloader := NewSwiftDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "swift.unknown")

	err := downloader.Extract(archivePath, targetDir)
	if err == nil {
		t.Error("Extract() should return error for unsupported format")
	}
}
