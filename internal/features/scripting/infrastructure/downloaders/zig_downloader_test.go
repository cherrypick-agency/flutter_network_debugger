package downloaders

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"

	"github.com/ulikunitz/xz"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewZigDownloader(t *testing.T) {
	downloader := NewZigDownloader()

	if downloader == nil {
		t.Fatal("NewZigDownloader returned nil")
	}

	if downloader.BaseDownloader == nil {
		t.Error("BaseDownloader is nil")
	}

	if downloader.indexURL == "" {
		t.Error("indexURL is empty")
	}
}

// Composer 1.
func TestZigDownloader_mapPlatformToZigTarget(t *testing.T) {
	downloader := NewZigDownloader()

	tests := []struct {
		name     string
		platform string
		arch     string
		expected string
	}{
		{
			name:     "darwin amd64",
			platform: "darwin",
			arch:     "amd64",
			expected: "x86_64-macos",
		},
		{
			name:     "darwin arm64",
			platform: "darwin",
			arch:     "arm64",
			expected: "aarch64-macos",
		},
		{
			name:     "linux amd64",
			platform: "linux",
			arch:     "amd64",
			expected: "x86_64-linux",
		},
		{
			name:     "linux arm64",
			platform: "linux",
			arch:     "arm64",
			expected: "aarch64-linux",
		},
		{
			name:     "windows amd64",
			platform: "windows",
			arch:     "amd64",
			expected: "x86_64-windows",
		},
		{
			name:     "linux 386",
			platform: "linux",
			arch:     "386",
			expected: "i386-linux",
		},
		{
			name:     "linux arm",
			platform: "linux",
			arch:     "arm",
			expected: "armv7a-linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := downloader.mapPlatformToZigTarget(tt.platform, tt.arch)
			if got != tt.expected {
				t.Errorf("mapPlatformToZigTarget(%q, %q) = %q, want %q", tt.platform, tt.arch, got, tt.expected)
			}
		})
	}
}

// Composer 1.
func TestZigDownloader_GetDownloadURL_Error(t *testing.T) {
	downloader := NewZigDownloader()

	req := domain.DownloadRequest{
		Platform: "unsupported",
		Arch:     "amd64",
		Version:  "latest",
	}

	_, err := downloader.GetDownloadURL(req)
	if err == nil {
		t.Error("GetDownloadURL() should return error for unsupported platform")
	}
}

// Composer 1.
func TestZigDownloader_Extract_Success(t *testing.T) {
	downloader := NewZigDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "zig.tar.xz")

	// Создаем tar.xz архив с zig-* директорией
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	// Добавляем директорию zig-0.14.0
	header := &tar.Header{
		Name:     "zig-darwin-arm64-0.14.0/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	// Добавляем файл в директорию
	content := []byte("zig binary")
	header = &tar.Header{
		Name:     "zig-darwin-arm64-0.14.0/zig",
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
	xzWriter.Close()
	file.Close()

	// Извлекаем
	err = downloader.Extract(archivePath, targetDir)
	if err != nil {
		t.Fatalf("Extract() error = %v, want nil", err)
	}

	// Проверяем что файл извлечен в корень targetDir
	zigFile := filepath.Join(targetDir, "zig")
	if _, err := os.Stat(zigFile); os.IsNotExist(err) {
		t.Error("zig binary was not extracted to target directory root")
	}
}

// Composer 1.
func TestZigDownloader_Extract_NoZigDir(t *testing.T) {
	downloader := NewZigDownloader()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "target")
	archivePath := filepath.Join(tmpDir, "zig.tar.xz")

	// Создаем tar.xz архив без zig-* директории
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	// Добавляем обычный файл без zig-* директории
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
	xzWriter.Close()
	file.Close()

	// Извлечение должно вернуть ошибку
	err = downloader.Extract(archivePath, targetDir)
	if err == nil {
		t.Error("Extract() should return error when zig directory not found")
	}
}
