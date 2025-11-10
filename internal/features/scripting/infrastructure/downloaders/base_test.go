package downloaders

import (
	"archive/tar"
	"archive/zip"
	"bytes"
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

	// Создаем тестовый HTTP сервер
	testData := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	// Тестируем загрузку
	var progressCalls []domain.DownloadProgress
	progressCb := func(progress domain.DownloadProgress) {
		progressCalls = append(progressCalls, progress)
	}

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	// Проверяем что файл создан
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
	}

	// Проверяем содержимое
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if !bytes.Equal(content, testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}

	// Проверяем что progress callback вызывался
	if len(progressCalls) == 0 {
		t.Error("Progress callback was not called")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_Cancellation(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	// Создаем медленный сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		// Пишем медленно
		for i := 0; i < 100; i++ {
			w.Write([]byte("x"))
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Отменяем контекст через короткое время
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("DownloadFile() should return error on cancellation")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	}

	if dlErr.Type != "cancelled" {
		t.Errorf("Expected error type 'cancelled', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_RetryOnNetworkError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	attempts := 0
	testData := []byte("test content")

	// Сервер падает первые 2 раза, затем работает
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Симулируем сетевую ошибку - закрываем соединение
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
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
		t.Fatalf("DownloadFile() error = %v, want nil after retries", err)
	}

	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_Resume(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	// Создаем частично загруженный файл
	partialData := []byte("partial ")
	if err := os.WriteFile(partialPath, partialData, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	testData := []byte("test file content")
	var rangeHeaderSent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем Range header
		rangeHeaderSent = r.Header.Get("Range")
		if rangeHeaderSent == "" {
			// Первый запрос без Range - возвращаем полный файл
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
			return
		}

		// Запрос с Range - возвращаем только недостающую часть
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", len(partialData), len(testData)-1, len(testData)))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)-len(partialData)))
		w.WriteHeader(http.StatusPartialContent)
		w.Write(testData[len(partialData):])
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	// Проверяем что Range header был отправлен
	if rangeHeaderSent == "" {
		t.Error("Expected Range header to be sent for resume")
	}

	// Проверяем что файл существует
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
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
		t.Error("DownloadFile() should return error for 404 status")
	}

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Errorf("Expected DownloadError, got %T", err)
	}

	if dlErr.Type != "network" {
		t.Errorf("Expected error type 'network', got %q", dlErr.Type)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_ProgressCallback(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := make([]byte, 1024*10) // 10 KB
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	var progressCalls []domain.DownloadProgress
	progressCb := func(progress domain.DownloadProgress) {
		progressCalls = append(progressCalls, progress)
	}

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	// Проверяем что progress callback вызывался несколько раз
	if len(progressCalls) == 0 {
		t.Error("Progress callback was not called")
	}

	// Проверяем последний вызов - должен быть 100%
	lastProgress := progressCalls[len(progressCalls)-1]
	if lastProgress.Percentage < 99 {
		t.Errorf("Expected final percentage >= 99, got %.1f", lastProgress.Percentage)
	}

	if lastProgress.BytesDownloaded != int64(len(testData)) {
		t.Errorf("Expected final bytes downloaded = %d, got %d", len(testData), lastProgress.BytesDownloaded)
	}
}
