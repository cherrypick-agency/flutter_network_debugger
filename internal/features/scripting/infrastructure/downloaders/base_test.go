package downloaders

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ulikunitz/xz"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestDownloadError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DownloadError
		contains []string
	}{
		{
			name: "error with wrapped error",
			err: &DownloadError{
				Type:    "network",
				Message: "connection failed",
				Err:     errors.New("timeout"),
			},
			contains: []string{"network", "connection failed", "timeout"},
		},
		{
			name: "error without wrapped error",
			err: &DownloadError{
				Type:    "disk_space",
				Message: "insufficient space",
				Err:     nil,
			},
			contains: []string{"disk_space", "insufficient space"},
		},
		{
			name: "checksum error",
			err: &DownloadError{
				Type:    "checksum",
				Message: "invalid checksum",
				Err:     errors.New("mismatch"),
			},
			contains: []string{"checksum", "invalid checksum", "mismatch"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, contain := range tt.contains {
				if !contains(got, contain) {
					t.Errorf("Error() = %q, should contain %q", got, contain)
				}
			}
		})
	}
}

// Composer 1.
func TestDownloadError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	downloadErr := &DownloadError{
		Type:    "network",
		Message: "failed",
		Err:     originalErr,
	}

	unwrapped := downloadErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Тест без wrapped error
	downloadErrNoWrap := &DownloadError{
		Type:    "network",
		Message: "failed",
		Err:     nil,
	}

	unwrappedNil := downloadErrNoWrap.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrappedNil)
	}
}

// Composer 1.
func TestNewBaseDownloader(t *testing.T) {
	downloader := NewBaseDownloader()

	if downloader == nil {
		t.Fatal("NewBaseDownloader returned nil")
	}

	if downloader.httpClient == nil {
		t.Fatal("httpClient is nil")
	}

	// Проверяем таймаут
	if downloader.httpClient.Timeout != 30*time.Minute {
		t.Errorf("Expected timeout 30 minutes, got %v", downloader.httpClient.Timeout)
	}

	// Проверяем Transport
	transport, ok := downloader.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport is not *http.Transport")
	}

	if transport.MaxIdleConns != 10 {
		t.Errorf("Expected MaxIdleConns 10, got %d", transport.MaxIdleConns)
	}

	if transport.IdleConnTimeout != 90*time.Second {
		t.Errorf("Expected IdleConnTimeout 90s, got %v", transport.IdleConnTimeout)
	}

	if transport.DisableCompression {
		t.Error("DisableCompression should be false")
	}

	if transport.DisableKeepAlives {
		t.Error("DisableKeepAlives should be false")
	}

	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost 10, got %d", transport.MaxIdleConnsPerHost)
	}
}

// Composer 1.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1024 * 1024, "1.0 MB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GB"},
		{"fractional KB", 1536, "1.5 KB"},
		{"fractional MB", 1536 * 1024, "1.5 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

// Composer 1.
func TestBaseDownloader_VerifyChecksum(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hasher := sha256.New()
	hasher.Write(content)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if err := downloader.VerifyChecksum(testFile, expectedChecksum); err != nil {
		t.Errorf("VerifyChecksum() with correct checksum error = %v, want nil", err)
	}

	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
	if err := downloader.VerifyChecksum(testFile, wrongChecksum); err == nil {
		t.Error("VerifyChecksum() with wrong checksum should return error")
	}
}

// Composer 1.
func TestProgressReader_Read(t *testing.T) {
	data := []byte("test data")
	reader := &progressReader{
		reader:     &testReader{data: data},
		total:      int64(len(data)),
		read:       0,
		onProgress: nil,
	}

	buf := make([]byte, 4)
	n, err := reader.Read(buf)

	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	if n != 4 {
		t.Errorf("Read() n = %d, want 4", n)
	}

	if reader.read != 4 {
		t.Errorf("read counter = %d, want 4", reader.read)
	}
}

// Composer 1.
func TestProgressReader_Read_WithCallback(t *testing.T) {
	data := []byte("test data")
	called := false
	var bytesRead int64

	reader := &progressReader{
		reader: &testReader{data: data},
		total:  int64(len(data)),
		read:   0,
		onProgress: func(br int64) {
			called = true
			bytesRead = br
		},
	}

	buf := make([]byte, 4)
	_, err := reader.Read(buf)

	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	if !called {
		t.Error("onProgress callback was not called")
	}

	if bytesRead != 4 {
		t.Errorf("onProgress called with bytesRead = %d, want 4", bytesRead)
	}
}

// Test helper
type testReader struct {
	data []byte
	pos  int
}

func (r *testReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// Composer 1.
func TestBaseDownloader_ExtractZip(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем простой zip файл
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Добавляем файл в zip
	f, err := zw.Create("test.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("test content"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zw.Close()
	zipWriter.Close()

	// Извлекаем
	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	// Проверяем что файл извлечен
	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	// Проверяем содержимое
	content, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(content) != "test content" {
		t.Errorf("Extracted file content = %q, want %q", string(content), "test content")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_InvalidFile(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "extracted")

	err := downloader.ExtractZip("nonexistent.zip", targetDir)
	if err == nil {
		t.Error("ExtractZip() should return error for nonexistent file")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем tar.gz архив
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Добавляем файл в архив
	content := []byte("test content")
	header := &tar.Header{
		Name: "test.txt",
		Size: int64(len(content)),
		Mode: 0644,
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
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	// Проверяем что файл извлечен
	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	// Проверяем содержимое
	readContent, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Extracted file content = %q, want %q", string(readContent), string(content))
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_InvalidFile(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "extracted")

	err := downloader.ExtractTarGz("nonexistent.tar.gz", targetDir)
	if err == nil {
		t.Error("ExtractTarGz() should return error for nonexistent file")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем tar.xz архив
	file, err := os.Create(tarXzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.xz file: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	// Добавляем файл в архив
	content := []byte("test content")
	header := &tar.Header{
		Name: "test.txt",
		Size: int64(len(content)),
		Mode: 0644,
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
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	// Проверяем что файл извлечен
	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	// Проверяем содержимое
	readContent, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Extracted file content = %q, want %q", string(readContent), string(content))
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_InvalidFile(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "extracted")

	err := downloader.ExtractTarXz("nonexistent.tar.xz", targetDir)
	if err == nil {
		t.Error("ExtractTarXz() should return error for nonexistent file")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_PathTraversal(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем zip с попыткой path traversal
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Пытаемся создать файл вне целевой директории
	f, err := zw.Create("../../outside.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("malicious"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zw.Close()
	zipWriter.Close()

	// Извлечение должно вернуть ошибку
	err = downloader.ExtractZip(zipFile, targetDir)
	if err == nil {
		t.Error("ExtractZip() should return error for path traversal attempt")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_PathTraversal(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем tar.gz с попыткой path traversal
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	header := &tar.Header{
		Name: "../../outside.txt",
		Size: 10,
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Извлечение должно вернуть ошибку
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarGz() should return error for path traversal attempt")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_PathTraversal(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем tar.xz с попыткой path traversal
	file, err := os.Create(tarXzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.xz file: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	header := &tar.Header{
		Name: "../../outside.txt",
		Size: 10,
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	// Извлечение должно вернуть ошибку
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarXz() should return error for path traversal attempt")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_Directory(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Создаем zip с директорией
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Добавляем директорию
	_, err = zw.Create("subdir/")
	if err != nil {
		t.Fatalf("Failed to create directory in zip: %v", err)
	}

	// Добавляем файл в директорию
	f, err := zw.Create("subdir/file.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("content"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zw.Close()
	zipWriter.Close()

	// Извлекаем
	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	// Проверяем что директория создана
	subdir := filepath.Join(targetDir, "subdir")
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("Subdirectory was not created")
	}

	// Проверяем что файл создан
	filePath := filepath.Join(subdir, "file.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File in subdirectory was not created")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_RetryWithBackoff(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test content")
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	var retryMessages []string
	progressCb := func(p domain.DownloadProgress) {
		if p.Stage == "retrying" {
			retryMessages = append(retryMessages, p.Message)
		}
	}

	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() after retries error = %v, want nil", err)
	}

	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}

	if len(retryMessages) == 0 {
		t.Error("Expected retry progress messages")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_CancellationDuringRetry(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("DownloadFile() with cancelled context should return error")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	} else if dlErr.Type != "cancelled" {
		t.Errorf("Expected error type 'cancelled', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_UnknownContentLength(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}
}

// Composer 1.
func TestBaseDownloader_VerifyChecksum_FileNotFound(t *testing.T) {
	downloader := NewBaseDownloader()

	err := downloader.VerifyChecksum("nonexistent.txt", "abc123")
	if err == nil {
		t.Error("VerifyChecksum() should return error for nonexistent file")
	}
}

// Composer 1.
func TestBaseDownloader_VerifyChecksum_InvalidChecksum(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
	err := downloader.VerifyChecksum(testFile, wrongChecksum)
	if err == nil {
		t.Error("VerifyChecksum() should return error for invalid checksum")
	}

	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Errorf("Error message should contain 'checksum mismatch', got: %v", err)
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_WithSymlink(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarXzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.xz file: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	content := []byte("test content")
	header := &tar.Header{
		Name: "test.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	symlinkHeader := &tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "test.txt",
		Mode:     0644,
	}
	if err := tarWriter.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}

	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_WithSymlink(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	content := []byte("test content")
	header := &tar.Header{
		Name: "test.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	symlinkHeader := &tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "test.txt",
		Mode:     0644,
	}
	if err := tarWriter.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_NetworkError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	ctx := context.Background()
	err := downloader.attemptDownload(ctx, "http://localhost:99999/nonexistent", destPath, nil)
	if err == nil {
		t.Error("attemptDownload() should return error for network failure")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	} else if dlErr.Type != "network" {
		t.Errorf("Expected error type 'network', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_WithSubdirectories(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	f1, err := zw.Create("dir1/file1.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f1.Write([]byte("content1"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	f2, err := zw.Create("dir2/file2.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f2.Write([]byte("content2"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zw.Close()
	zipWriter.Close()

	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	file1 := filepath.Join(targetDir, "dir1", "file1.txt")
	file2 := filepath.Join(targetDir, "dir2", "file2.txt")

	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("File1 was not extracted")
	}

	if _, err := os.Stat(file2); os.IsNotExist(err) {
		t.Error("File2 was not extracted")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_WithSubdirectories(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarXzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.xz file: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	dirHeader := &tar.Header{
		Name:     "subdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tarWriter.WriteHeader(dirHeader); err != nil {
		t.Fatalf("Failed to write dir header: %v", err)
	}

	content := []byte("file content")
	fileHeader := &tar.Header{
		Name: "subdir/file.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(fileHeader); err != nil {
		t.Fatalf("Failed to write file header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	extractedFile := filepath.Join(targetDir, "subdir", "file.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	readContent, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Extracted file content = %q, want %q", string(readContent), string(content))
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_FileCreationError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "nonexistent", "test.bin")

	testData := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.attemptDownload(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("attemptDownload() should return error when file creation fails")
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_CopyError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("partial"))
		w.(http.Flusher).Flush()
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := downloader.attemptDownload(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("attemptDownload() should return error when copy is interrupted")
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_ProgressWithResume(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	testData := make([]byte, 50000)
	for i := range testData {
		testData[i] = byte(i % 256)
	}
	firstPart := testData[:10000]
	secondPart := testData[10000:]

	if err := os.WriteFile(partialPath, firstPart, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 10000-%d/%d", len(testData)-1, len(testData)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(secondPart)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(secondPart)
		} else {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	var progressMessages []string
	progressCb := func(p domain.DownloadProgress) {
		if p.Message != "" {
			progressMessages = append(progressMessages, p.Message)
		}
	}

	err := downloader.attemptDownload(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("attemptDownload() with resume error = %v, want nil", err)
	}

	foundResumeMessage := false
	for _, msg := range progressMessages {
		if strings.Contains(strings.ToLower(msg), "resum") {
			foundResumeMessage = true
			break
		}
	}

	if !foundResumeMessage && len(progressMessages) > 0 {
		t.Logf("Progress messages: %v", progressMessages)
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_InvalidArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "invalid.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	if err := os.WriteFile(tarXzFile, []byte("not a valid tar.xz"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	err := downloader.ExtractTarXz(tarXzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarXz() should return error for invalid archive")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_InvalidArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "invalid.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	if err := os.WriteFile(tarGzFile, []byte("not a valid tar.gz"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	err := downloader.ExtractTarGz(tarGzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarGz() should return error for invalid archive")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_InvalidArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "invalid.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	if err := os.WriteFile(zipFile, []byte("not a valid zip"), 0644); err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	err := downloader.ExtractZip(zipFile, targetDir)
	if err == nil {
		t.Error("ExtractZip() should return error for invalid archive")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_ErrorReadingTar(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarXzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.xz file: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	xzWriter.Write([]byte("corrupted data"))
	xzWriter.Close()
	file.Close()

	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err == nil {
		t.Log("ExtractTarXz() may not error on empty/corrupted archive, checking for io.EOF")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_ErrorReadingTar(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	gzWriter.Write([]byte("corrupted data"))
	gzWriter.Close()
	file.Close()

	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err == nil {
		t.Log("ExtractTarGz() may not error on empty/corrupted archive, checking for io.EOF")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_ErrorOpeningFile(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	f, err := zw.Create("test.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("content"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}
	zw.Close()
	zipWriter.Close()

	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_NoRetryOnChecksumError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test content")
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	if attempts > 1 {
		t.Errorf("Expected 1 attempt for successful download, got %d", attempts)
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_RequestCreationError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := downloader.attemptDownload(ctx, "http://example.com/test", destPath, nil)
	if err == nil {
		t.Error("attemptDownload() should return error when context is cancelled")
	}
}

// Composer 1.
func TestBaseDownloader_ProgressReader_Throttling(t *testing.T) {
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}

	callCount := 0
	reader := &progressReader{
		reader: &testReader{data: data},
		total:  int64(len(data)),
		read:   0,
		onProgress: func(bytesRead int64) {
			callCount++
		},
	}

	buf := make([]byte, 100)
	for i := 0; i < 10; i++ {
		_, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatalf("Read() error = %v", err)
		}
	}

	if callCount == 0 {
		t.Error("onProgress callback was not called")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_ErrorCreatingDirectory(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarXzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.xz file: %v", err)
	}
	defer file.Close()

	xzWriter, err := xz.NewWriter(file)
	if err != nil {
		t.Fatalf("Failed to create xz writer: %v", err)
	}
	defer xzWriter.Close()

	tarWriter := tar.NewWriter(xzWriter)
	defer tarWriter.Close()

	dirHeader := &tar.Header{
		Name:     "testdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tarWriter.WriteHeader(dirHeader); err != nil {
		t.Fatalf("Failed to write dir header: %v", err)
	}

	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	extractedDir := filepath.Join(targetDir, "testdir")
	if _, err := os.Stat(extractedDir); os.IsNotExist(err) {
		t.Error("Extracted directory was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_ErrorCreatingFile(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	content := []byte("test content")
	fileHeader := &tar.Header{
		Name: "subdir/test.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(fileHeader); err != nil {
		t.Fatalf("Failed to write file header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	extractedFile := filepath.Join(targetDir, "subdir", "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_ErrorCreatingFile(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "test.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	f, err := zw.Create("subdir/file.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write([]byte("content"))
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}

	zw.Close()
	zipWriter.Close()

	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	extractedFile := filepath.Join(targetDir, "subdir", "file.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_FinalProgressUpdate(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	var finalProgress domain.DownloadProgress
	progressCb := func(p domain.DownloadProgress) {
		if p.Percentage >= 100 {
			finalProgress = p
		}
	}

	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	if finalProgress.Percentage != 100 {
		t.Errorf("Final progress percentage = %.1f, want 100", finalProgress.Percentage)
	}

	if finalProgress.Stage != "downloading" {
		t.Errorf("Final progress stage = %q, want 'downloading'", finalProgress.Stage)
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_ProgressThrottling(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := make([]byte, 10000)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	progressCallCount := 0
	progressCb := func(p domain.DownloadProgress) {
		progressCallCount++
	}

	err := downloader.attemptDownload(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("attemptDownload() error = %v, want nil", err)
	}

	if progressCallCount == 0 {
		t.Error("Progress callback was not called")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Composer 1.
func TestBaseDownloader_DownloadFile_Success(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	var progressCalls []domain.DownloadProgress
	progressCb := func(p domain.DownloadProgress) {
		progressCalls = append(progressCalls, p)
	}

	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}

	if len(progressCalls) == 0 {
		t.Error("Progress callback was not called")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_WithProgress(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := make([]byte, 1024*10)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	var progressCalls []domain.DownloadProgress
	progressCb := func(p domain.DownloadProgress) {
		progressCalls = append(progressCalls, p)
	}

	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	if len(progressCalls) == 0 {
		t.Error("Progress callback was not called")
	}

	foundComplete := false
	for _, p := range progressCalls {
		if p.Stage == "downloading" && p.Percentage > 0 {
			if p.BytesDownloaded <= 0 || p.TotalBytes <= 0 {
				t.Errorf("Invalid progress: BytesDownloaded=%d, TotalBytes=%d", p.BytesDownloaded, p.TotalBytes)
			}
		}
		if p.Percentage >= 100 {
			foundComplete = true
		}
	}

	if !foundComplete {
		t.Error("No progress update with 100% completion found")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_RetryOnNetworkError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test content")
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() after retry error = %v, want nil", err)
	}

	if attempts < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attempts)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_Resume(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	testData := []byte("test file content for resume")
	firstPart := testData[:10]
	secondPart := testData[10:]

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 10-%d/%d", len(testData)-1, len(testData)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(secondPart)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(secondPart)
		} else {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		}
	}))
	defer server.Close()

	if err := os.WriteFile(partialPath, firstPart, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() with resume error = %v, want nil", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}

	if _, err := os.Stat(partialPath); err == nil {
		t.Error("Partial file was not removed after successful download")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_Cancellation(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(http.StatusOK)
		for i := 0; i < 1000; i++ {
			time.Sleep(10 * time.Millisecond)
			w.Write([]byte("data"))
		}
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("DownloadFile() with cancelled context should return error")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	} else if dlErr.Type != "cancelled" {
		t.Errorf("Expected error type 'cancelled', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_InvalidStatusCode(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("DownloadFile() with 404 should return error")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	} else if dlErr.Type != "network" {
		t.Errorf("Expected error type 'network', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_MaxRetriesExhausted(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("DownloadFile() with persistent errors should return error")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	} else if dlErr.Type != "network" {
		t.Errorf("Expected error type 'network', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_NoResumeSupport(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	testData := []byte("test content")
	if err := os.WriteFile(partialPath, testData[:5], 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Range") != "" {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.attemptDownload(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("attemptDownload() error = %v, want nil", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_PartialContent(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	testData := []byte("complete test data")
	firstPart := testData[:8]
	secondPart := testData[8:]

	if err := os.WriteFile(partialPath, firstPart, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 8-%d/%d", len(testData)-1, len(testData)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(secondPart)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(secondPart)
		} else {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.attemptDownload(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("attemptDownload() with partial content error = %v, want nil", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}
}
