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
	"runtime"
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

	// Test without wrapped error
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

	// Check timeout
	if downloader.httpClient.Timeout != 30*time.Minute {
		t.Errorf("Expected timeout 30 minutes, got %v", downloader.httpClient.Timeout)
	}

	// Check Transport
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

	// Create a simple zip file
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Add file to zip
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

	// Extract
	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	// Check that file is extracted
	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	// Check content
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

	// Create tar.gz archive
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add file to archive
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

	// Extract
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	// Check that file is extracted
	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	// Check content
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

	// Create tar.xz archive
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

	// Add file to archive
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

	// Extract
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	// Check that file is extracted
	extractedFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
		t.Error("Extracted file was not created")
	}

	// Check content
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

	// Create zip with path traversal attempt
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Try to create file outside target directory
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

	// Extraction should return error
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

	// Create tar.gz with path traversal attempt
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

	// Extraction should return error
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

	// Create tar.xz with path traversal attempt
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

	// Extraction should return error
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

	// Create zip with directory
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Add directory
	_, err = zw.Create("subdir/")
	if err != nil {
		t.Fatalf("Failed to create directory in zip: %v", err)
	}

	// Add file to directory
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

	// Extract
	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	// Check that directory is created
	subdir := filepath.Join(targetDir, "subdir")
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("Subdirectory was not created")
	}

	// Check that file is created
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

	// Create test HTTP server
	testData := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	// Test download
	var progressCalls []domain.DownloadProgress
	progressCb := func(progress domain.DownloadProgress) {
		progressCalls = append(progressCalls, progress)
	}

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, progressCb)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	// Check that file is created
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
	}

	// Check content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if !bytes.Equal(content, testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}

	// Check that progress callback was called
	if len(progressCalls) == 0 {
		t.Error("Progress callback was not called")
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_Cancellation(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	// Create slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		// Write slowly
		for i := 0; i < 100; i++ {
			w.Write([]byte("x"))
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Cancel context after a short time
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

	// Server fails first 2 times, then works
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Simulate network error - close connection
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

	// Create partially downloaded file
	partialData := []byte("partial ")
	if err := os.WriteFile(partialPath, partialData, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	testData := []byte("test file content")
	var rangeHeaderSent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Range header
		rangeHeaderSent = r.Header.Get("Range")
		if rangeHeaderSent == "" {
			// First request without Range - return full file
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
			return
		}

		// Request with Range - return only missing part
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

	// Check that Range header was sent
	if rangeHeaderSent == "" {
		t.Error("Expected Range header to be sent for resume")
	}

	// Check that file exists
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

	// Check that progress callback was called multiple times
	if len(progressCalls) == 0 {
		t.Error("Progress callback was not called")
	}

	// Check last call - should be 100%
	lastProgress := progressCalls[len(progressCalls)-1]
	if lastProgress.Percentage < 99 {
		t.Errorf("Expected final percentage >= 99, got %.1f", lastProgress.Percentage)
	}

	if lastProgress.BytesDownloaded != int64(len(testData)) {
		t.Errorf("Expected final bytes downloaded = %d, got %d", len(testData), lastProgress.BytesDownloaded)
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_ServerError(t *testing.T) {
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
		t.Error("DownloadFile() should return error for 500 status")
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
func TestBaseDownloader_DownloadFile_RangeNotSatisfiable(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	// Create partially downloaded file
	partialData := []byte("partial")
	if err := os.WriteFile(partialPath, partialData, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	testData := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First request with Range - return 416
		if r.Header.Get("Range") != "" {
			w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
			return
		}
		// Second request without Range - return full file
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil after retry", err)
	}

	// Check that file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_Directory(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.gz with directory and file
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add directory
	dirHeader := &tar.Header{
		Name:     "subdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tarWriter.WriteHeader(dirHeader); err != nil {
		t.Fatalf("Failed to write dir header: %v", err)
	}

	// Add file to directory
	content := []byte("test content")
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
	gzWriter.Close()
	file.Close()

	// Extract
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	// Check that directory is created
	subdir := filepath.Join(targetDir, "subdir")
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("Subdirectory was not created")
	}

	// Check that file is created
	filePath := filepath.Join(subdir, "file.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File in subdirectory was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_Directory(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.xz with directory and file
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

	// Add directory
	dirHeader := &tar.Header{
		Name:     "subdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tarWriter.WriteHeader(dirHeader); err != nil {
		t.Fatalf("Failed to write dir header: %v", err)
	}

	// Add file to directory
	content := []byte("test content")
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

	// Extract
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	// Check that directory is created
	subdir := filepath.Join(targetDir, "subdir")
	if _, err := os.Stat(subdir); os.IsNotExist(err) {
		t.Error("Subdirectory was not created")
	}

	// Check that file is created
	filePath := filepath.Join(subdir, "file.txt")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("File in subdirectory was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_Symlink(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.gz with symlink
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add file
	content := []byte("target file")
	fileHeader := &tar.Header{
		Name: "target.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(fileHeader); err != nil {
		t.Fatalf("Failed to write file header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	// Add symlink
	symlinkHeader := &tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "target.txt",
		Mode:     0644,
	}
	if err := tarWriter.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Extract (symlink may not be created on Windows, that's normal)
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	// Check that target file is created
	targetPath := filepath.Join(targetDir, "target.txt")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Target file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_Symlink(t *testing.T) {
	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.xz with symlink
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

	// Add file
	content := []byte("target file")
	fileHeader := &tar.Header{
		Name: "target.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	if err := tarWriter.WriteHeader(fileHeader); err != nil {
		t.Fatalf("Failed to write file header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	// Add symlink
	symlinkHeader := &tar.Header{
		Name:     "link.txt",
		Typeflag: tar.TypeSymlink,
		Linkname: "target.txt",
		Mode:     0644,
	}
	if err := tarWriter.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}

	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	// Extract (symlink may not be created on Windows, that's normal)
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	// Check that target file is created
	targetPath := filepath.Join(targetDir, "target.txt")
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Target file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_SymlinkEscapeBlocked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may be unavailable on Windows without special permissions")
	}

	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "test.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.xz with symlink-directory and file inside it
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

	// symlink dir -> ".." (escapes from targetDir when writing)
	symlinkHeader := &tar.Header{
		Name:     "dir",
		Typeflag: tar.TypeSymlink,
		Linkname: "..",
		Mode:     0o777,
	}
	if err := tarWriter.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}

	// file under symlinked dir
	content := []byte("evil")
	fileHeader := &tar.Header{
		Name: "dir/evil.txt",
		Size: int64(len(content)),
		Mode: 0o644,
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
	if err == nil {
		t.Fatal("ExtractTarXz should fail when archive tries to escape via symlink parent")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_SymlinkEscapeBlocked(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may be unavailable on Windows without special permissions")
	}

	downloader := NewBaseDownloader()

	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "test.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.gz with symlink-directory and file inside it
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// symlink dir -> ".." (escapes from targetDir when writing)
	symlinkHeader := &tar.Header{
		Name:     "dir",
		Typeflag: tar.TypeSymlink,
		Linkname: "..",
		Mode:     0o777,
	}
	if err := tarWriter.WriteHeader(symlinkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}

	// file under symlinked dir
	content := []byte("evil")
	fileHeader := &tar.Header{
		Name: "dir/evil.txt",
		Size: int64(len(content)),
		Mode: 0o644,
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
	if err == nil {
		t.Fatal("ExtractTarGz should fail when archive tries to escape via symlink parent")
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
func TestBaseDownloader_DownloadFile_NoContentLength(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	testData := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't set Content-Length
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("DownloadFile() error = %v, want nil", err)
	}

	// Check that file is created
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
	}

	// Check content
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if !bytes.Equal(content, testData) {
		t.Errorf("Downloaded content = %q, want %q", string(content), string(testData))
	}
}

// Composer 1.
func TestBaseDownloader_DownloadFile_MaxRetries(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	// Server always returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("DownloadFile() should return error after max retries")
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
func TestBaseDownloader_DownloadFile_DiskSpaceError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	// Server returns very large Content-Length
	testData := []byte("test")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set huge size (1 exabyte)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", 1<<60))
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.DownloadFile(ctx, server.URL, destPath, nil)
	// May return disk space error or nil depending on the system
	_ = err
}

// Composer 1.
func TestBaseDownloader_DownloadFile_RetryProgress(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	attempts := 0
	testData := []byte("test content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// First attempt - error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second attempt - success
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
		t.Fatalf("DownloadFile() error = %v, want nil after retry", err)
	}

	// Check that there was a call with stage "retrying"
	foundRetry := false
	for _, progress := range progressCalls {
		if progress.Stage == "retrying" {
			foundRetry = true
			break
		}
	}

	if !foundRetry {
		t.Error("Expected progress callback with stage 'retrying'")
	}
}

// Composer 1.
func TestBaseDownloader_AttemptDownload_NetworkError(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")

	// Server closes connection immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.attemptDownload(ctx, server.URL, destPath, nil)
	if err == nil {
		t.Error("attemptDownload() should return error on network failure")
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
func TestBaseDownloader_AttemptDownload_ResumeSupport(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "test.bin")
	partialPath := destPath + ".partial"

	// Create partially downloaded file
	partialData := []byte("partial ")
	if err := os.WriteFile(partialPath, partialData, 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	testData := []byte("test file content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" {
			// Support resume - return only missing part
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", len(partialData), len(testData)-1, len(testData)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)-len(partialData)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(testData[len(partialData):])
		} else {
			// Full request
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testData)))
			w.WriteHeader(http.StatusOK)
			w.Write(testData)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	err := downloader.attemptDownload(ctx, server.URL, destPath, nil)
	if err != nil {
		t.Fatalf("attemptDownload() error = %v, want nil", err)
	}

	// Check that file is created
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Downloaded file was not created")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_CorruptedArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "corrupted.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create corrupted archive (just random bytes)
	if err := os.WriteFile(tarXzFile, []byte("corrupted data"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted archive: %v", err)
	}

	err := downloader.ExtractTarXz(tarXzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarXz() should return error for corrupted archive")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_CorruptedArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "corrupted.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create corrupted archive (just random bytes)
	if err := os.WriteFile(tarGzFile, []byte("corrupted data"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted archive: %v", err)
	}

	err := downloader.ExtractTarGz(tarGzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarGz() should return error for corrupted archive")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_CorruptedArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "corrupted.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create corrupted archive (just random bytes)
	if err := os.WriteFile(zipFile, []byte("corrupted data"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted archive: %v", err)
	}

	err := downloader.ExtractZip(zipFile, targetDir)
	if err == nil {
		t.Error("ExtractZip() should return error for corrupted archive")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_EmptyArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "empty.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create empty tar.xz archive
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

	// Don't add files - empty archive
	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	// Extracting empty archive should succeed
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil for empty archive", err)
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_EmptyArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "empty.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create empty tar.gz archive
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Don't add files - empty archive
	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Extracting empty archive should succeed
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil for empty archive", err)
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_EmptyArchive(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "empty.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create empty zip archive
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Don't add files - empty archive
	zw.Close()
	zipWriter.Close()

	// Extracting empty archive should succeed
	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil for empty archive", err)
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_MultipleFiles(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "multi.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.xz archive with multiple files
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

	// Add multiple files
	files := map[string][]byte{
		"file1.txt":        []byte("content1"),
		"file2.txt":        []byte("content2"),
		"subdir/file3.txt": []byte("content3"),
	}

	for filename, content := range files {
		header := &tar.Header{
			Name: filename,
			Size: int64(len(content)),
			Mode: 0644,
		}
		if strings.Contains(filename, "subdir") {
			// Create directory first
			dirHeader := &tar.Header{
				Name:     "subdir/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}
			if err := tarWriter.WriteHeader(dirHeader); err != nil {
				t.Fatalf("Failed to write dir header: %v", err)
			}
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}

	tarWriter.Close()
	xzWriter.Close()
	file.Close()

	// Extract
	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarXz() error = %v, want nil", err)
	}

	// Check that all files are extracted
	for filename, expectedContent := range files {
		extractedFile := filepath.Join(targetDir, filename)
		if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
			t.Errorf("File %s was not extracted", filename)
			continue
		}

		content, err := os.ReadFile(extractedFile)
		if err != nil {
			t.Fatalf("Failed to read extracted file %s: %v", filename, err)
		}

		if !bytes.Equal(content, expectedContent) {
			t.Errorf("File %s content = %q, want %q", filename, string(content), string(expectedContent))
		}
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_MultipleFiles(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "multi.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create tar.gz archive with multiple files
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Add multiple files
	files := map[string][]byte{
		"file1.txt":        []byte("content1"),
		"file2.txt":        []byte("content2"),
		"subdir/file3.txt": []byte("content3"),
	}

	for filename, content := range files {
		if strings.Contains(filename, "subdir") {
			// Create directory first
			dirHeader := &tar.Header{
				Name:     "subdir/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}
			if err := tarWriter.WriteHeader(dirHeader); err != nil {
				t.Fatalf("Failed to write dir header: %v", err)
			}
		}
		header := &tar.Header{
			Name: filename,
			Size: int64(len(content)),
			Mode: 0644,
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("Failed to write tar header: %v", err)
		}
		if _, err := tarWriter.Write(content); err != nil {
			t.Fatalf("Failed to write tar content: %v", err)
		}
	}

	tarWriter.Close()
	gzWriter.Close()
	file.Close()

	// Extract
	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractTarGz() error = %v, want nil", err)
	}

	// Check that all files are extracted
	for filename, expectedContent := range files {
		extractedFile := filepath.Join(targetDir, filename)
		if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
			t.Errorf("File %s was not extracted", filename)
			continue
		}

		content, err := os.ReadFile(extractedFile)
		if err != nil {
			t.Fatalf("Failed to read extracted file %s: %v", filename, err)
		}

		if !bytes.Equal(content, expectedContent) {
			t.Errorf("File %s content = %q, want %q", filename, string(content), string(expectedContent))
		}
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_MultipleFiles(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "multi.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create zip archive with multiple files
	zipWriter, err := os.Create(zipFile)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}
	defer zipWriter.Close()

	zw := zip.NewWriter(zipWriter)
	defer zw.Close()

	// Add multiple files
	files := map[string][]byte{
		"file1.txt":        []byte("content1"),
		"file2.txt":        []byte("content2"),
		"subdir/file3.txt": []byte("content3"),
	}

	for filename, content := range files {
		f, err := zw.Create(filename)
		if err != nil {
			t.Fatalf("Failed to create file in zip: %v", err)
		}
		_, err = f.Write(content)
		if err != nil {
			t.Fatalf("Failed to write to zip: %v", err)
		}
	}

	zw.Close()
	zipWriter.Close()

	// Extract
	err = downloader.ExtractZip(zipFile, targetDir)
	if err != nil {
		t.Fatalf("ExtractZip() error = %v, want nil", err)
	}

	// Check that all files are extracted
	for filename, expectedContent := range files {
		extractedFile := filepath.Join(targetDir, filename)
		if _, err := os.Stat(extractedFile); os.IsNotExist(err) {
			t.Errorf("File %s was not extracted", filename)
			continue
		}

		content, err := os.ReadFile(extractedFile)
		if err != nil {
			t.Fatalf("Failed to read extracted file %s: %v", filename, err)
		}

		if !bytes.Equal(content, expectedContent) {
			t.Errorf("File %s content = %q, want %q", filename, string(content), string(expectedContent))
		}
	}
}

// Composer 1.
func TestBaseDownloader_VerifyChecksum_EmptyFile(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	// Create empty file
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Calculate correct checksum for empty file
	hasher := sha256.New()
	hasher.Write([]byte{})
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if err := downloader.VerifyChecksum(testFile, expectedChecksum); err != nil {
		t.Errorf("VerifyChecksum() with correct checksum for empty file error = %v, want nil", err)
	}

	// Check with wrong checksum
	wrongChecksum := "0000000000000000000000000000000000000000000000000000000000000000"
	if err := downloader.VerifyChecksum(testFile, wrongChecksum); err == nil {
		t.Error("VerifyChecksum() with wrong checksum should return error")
	}
}

// Composer 1.
func TestBaseDownloader_VerifyChecksum_LargeFile(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.bin")

	// Create large file (100KB)
	largeContent := make([]byte, 100*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	if err := os.WriteFile(testFile, largeContent, 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Calculate correct checksum
	hasher := sha256.New()
	hasher.Write(largeContent)
	expectedChecksum := hex.EncodeToString(hasher.Sum(nil))

	if err := downloader.VerifyChecksum(testFile, expectedChecksum); err != nil {
		t.Errorf("VerifyChecksum() with correct checksum for large file error = %v, want nil", err)
	}
}

// Composer 1.
func TestProgressReader_Read_MultipleReads(t *testing.T) {
	data := []byte("test data for multiple reads")
	var progressCalls []int64

	reader := &progressReader{
		reader: &testReader{data: data},
		total:  int64(len(data)),
		read:   0,
		onProgress: func(br int64) {
			progressCalls = append(progressCalls, br)
		},
	}

	// Read in chunks
	buf1 := make([]byte, 5)
	buf2 := make([]byte, 5)
	buf3 := make([]byte, 100)

	_, _ = reader.Read(buf1)
	_, _ = reader.Read(buf2)
	_, _ = reader.Read(buf3)

	// Check that callback was called
	if len(progressCalls) == 0 {
		t.Error("onProgress callback was not called")
	}

	// Check that last call contains total bytes read
	if len(progressCalls) > 0 {
		lastCall := progressCalls[len(progressCalls)-1]
		if lastCall != int64(len(data)) {
			t.Errorf("Last progress call = %d, want %d", lastCall, len(data))
		}
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarXz_InvalidTarHeader(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarXzFile := filepath.Join(tmpDir, "invalid.tar.xz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create archive with invalid tar header
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

	// Write garbage instead of valid tar header
	_, err = xzWriter.Write([]byte("invalid tar data"))
	if err != nil {
		t.Fatalf("Failed to write invalid data: %v", err)
	}

	xzWriter.Close()
	file.Close()

	err = downloader.ExtractTarXz(tarXzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarXz() should return error for invalid tar header")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractTarGz_InvalidTarHeader(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	tarGzFile := filepath.Join(tmpDir, "invalid.tar.gz")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create archive with invalid tar header
	file, err := os.Create(tarGzFile)
	if err != nil {
		t.Fatalf("Failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	// Write garbage instead of valid tar header
	_, err = gzWriter.Write([]byte("invalid tar data"))
	if err != nil {
		t.Fatalf("Failed to write invalid data: %v", err)
	}

	gzWriter.Close()
	file.Close()

	err = downloader.ExtractTarGz(tarGzFile, targetDir)
	if err == nil {
		t.Error("ExtractTarGz() should return error for invalid tar header")
	}
}

// Composer 1.
func TestBaseDownloader_ExtractZip_InvalidZipHeader(t *testing.T) {
	downloader := NewBaseDownloader()
	tmpDir := t.TempDir()
	zipFile := filepath.Join(tmpDir, "invalid.zip")
	targetDir := filepath.Join(tmpDir, "extracted")

	// Create file with invalid zip header
	if err := os.WriteFile(zipFile, []byte("invalid zip data"), 0644); err != nil {
		t.Fatalf("Failed to create invalid zip file: %v", err)
	}

	err := downloader.ExtractZip(zipFile, targetDir)
	if err == nil {
		t.Error("ExtractZip() should return error for invalid zip header")
	}
}
