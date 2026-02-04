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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ulikunitz/xz"

	"network-debugger/internal/features/scripting/domain"
)

// Custom error types for better error handling
var (
	ErrInsufficientDiskSpace = errors.New("insufficient disk space")
	ErrNetworkFailure        = errors.New("network failure")
	ErrInvalidChecksum       = errors.New("invalid checksum")
	ErrDownloadCancelled     = errors.New("download cancelled")
)

// DownloadError wraps download errors with additional context
type DownloadError struct {
	Type    string // "network", "disk_space", "checksum", "cancelled"
	Message string
	Err     error
}

func (e *DownloadError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *DownloadError) Unwrap() error {
	return e.Err
}

// BaseDownloader provides shared functionality for all compiler downloaders
// Follows DRY principle - Don't Repeat Yourself
type BaseDownloader struct {
	httpClient *http.Client
}

func readAllLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(r, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response body exceeds %s", formatBytes(maxBytes))
	}
	return data, nil
}

func sanitizeArchiveFileMode(mode int64) os.FileMode {
	// Tar/zip archives may contain strange bits (setuid/setgid/sticky, etc.).
	// We only need regular rwx permissions.
	m := os.FileMode(mode) & 0o777
	if m == 0 {
		return 0o644
	}
	return m
}

func safeExtractPath(targetDir, entryName string) (string, error) {
	if strings.TrimSpace(entryName) == "" {
		return "", fmt.Errorf("invalid file path in archive: empty name")
	}

	cleanName := filepath.Clean(entryName)
	if filepath.IsAbs(cleanName) {
		return "", fmt.Errorf("invalid file path in archive: %s", entryName)
	}
	if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid file path in archive: %s", entryName)
	}

	targetPath := filepath.Join(targetDir, cleanName)
	rel, err := filepath.Rel(targetDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("invalid file path in archive: %s", entryName)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid file path in archive: %s", entryName)
	}

	return targetPath, nil
}

func safeSymlinkTarget(targetDir, linkPath, linkName string) error {
	if strings.TrimSpace(linkName) == "" {
		return fmt.Errorf("invalid symlink target in archive: empty link")
	}
	if filepath.IsAbs(linkName) {
		return fmt.Errorf("invalid symlink target in archive: %s", linkName)
	}

	// For relative links, target is considered relative to the link directory.
	// Important: we verify the resolved path doesn't escape targetDir.
	resolved := filepath.Clean(filepath.Join(filepath.Dir(linkPath), linkName))
	rel, err := filepath.Rel(targetDir, resolved)
	if err != nil {
		return fmt.Errorf("invalid symlink target in archive: %s", linkName)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("invalid symlink target in archive: %s", linkName)
	}
	return nil
}

func mkdirAllNoSymlink(baseDir, dirPath string, perm os.FileMode) error {
	rel, err := filepath.Rel(baseDir, dirPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if rel == "." {
		return nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("path escapes base dir")
	}

	cur := baseDir
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		cur = filepath.Join(cur, part)

		fi, err := os.Lstat(cur)
		if err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to follow symlink in path: %s", cur)
			}
			if !fi.IsDir() {
				return fmt.Errorf("path component is not a directory: %s", cur)
			}
			continue
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", cur, err)
		}
		if err := os.Mkdir(cur, perm); err != nil {
			// If there's a race — double-check.
			if fi2, err2 := os.Lstat(cur); err2 == nil {
				if fi2.Mode()&os.ModeSymlink != 0 {
					return fmt.Errorf("refusing to follow symlink in path: %s", cur)
				}
				if !fi2.IsDir() {
					return fmt.Errorf("path component is not a directory: %s", cur)
				}
				continue
			}
			return fmt.Errorf("mkdir %s: %w", cur, err)
		}
	}

	return nil
}

func ensureNoSymlinkFileTarget(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write to symlink: %s", path)
	}
	return nil
}

// NewBaseDownloader creates a new BaseDownloader with configured HTTP client
func NewBaseDownloader() *BaseDownloader {
	return &BaseDownloader{
		httpClient: &http.Client{
			Timeout: 30 * time.Minute, // Allow large downloads
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		},
	}
}

// DownloadFile downloads a file from URL to destination with progress tracking,
// resume support, retry logic, and disk space checking.
// progressCb is called periodically with download progress
// Returns error if download fails or is cancelled
func (b *BaseDownloader) DownloadFile(
	ctx context.Context,
	url string,
	destPath string,
	progressCb func(domain.DownloadProgress),
) error {
	const (
		maxRetries      = 3
		initialBackoff  = 2 * time.Second
		maxBackoff      = 30 * time.Second
		diskSpaceBuffer = 100 * 1024 * 1024 // 100MB safety buffer
	)

	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check context cancellation before each attempt
		select {
		case <-ctx.Done():
			return &DownloadError{
				Type:    "cancelled",
				Message: "Download cancelled by user",
				Err:     ErrDownloadCancelled,
			}
		default:
		}

		// Attempt download (with resume support)
		err := b.attemptDownload(ctx, url, destPath, progressCb)
		if err == nil {
			return nil // Success!
		}

		lastErr = err

		// Don't retry on certain errors
		var dlErr *DownloadError
		if errors.As(err, &dlErr) {
			switch dlErr.Type {
			case "cancelled", "disk_space", "checksum":
				return err // Don't retry these
			}
		}

		// Check if context was cancelled
		if ctx.Err() != nil {
			return &DownloadError{
				Type:    "cancelled",
				Message: "Download cancelled",
				Err:     ctx.Err(),
			}
		}

		// Last attempt failed
		if attempt == maxRetries {
			break
		}

		// Wait before retry with exponential backoff
		if progressCb != nil {
			progressCb(domain.DownloadProgress{
				Stage:   "retrying",
				Message: fmt.Sprintf("Retrying in %v (attempt %d/%d)...", backoff, attempt+2, maxRetries+1),
			})
		}

		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-ctx.Done():
			return &DownloadError{
				Type:    "cancelled",
				Message: "Download cancelled during retry",
				Err:     ctx.Err(),
			}
		}
	}

	// All retries exhausted
	return &DownloadError{
		Type:    "network",
		Message: fmt.Sprintf("Download failed after %d attempts", maxRetries+1),
		Err:     lastErr,
	}
}

// attemptDownload performs a single download attempt with resume support
func (b *BaseDownloader) attemptDownload(
	ctx context.Context,
	url string,
	destPath string,
	progressCb func(domain.DownloadProgress),
) error {
	// On some servers Range may be ignored. In this case:
	// - don't append to .partial (otherwise the file silently gets corrupted)
	// - restart the download from scratch once
	const maxResumeFallbacks = 1

	// Check if partial file exists (for resume)
	var resumeFrom int64
	partialPath := destPath + ".partial"
	if stat, err := os.Stat(partialPath); err == nil {
		resumeFrom = stat.Size()
	}

	var resp *http.Response
	var err error
	resumeFallbacksUsed := 0

	for {
		// Create HTTP request with context for cancellation
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			return fmt.Errorf("failed to create request: %w", reqErr)
		}

		// Add Range header for resume
		if resumeFrom > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeFrom))
		}

		// Execute request
		resp, err = b.httpClient.Do(req)
		if err != nil {
			return &DownloadError{
				Type:    "network",
				Message: "Failed to connect to server",
				Err:     err,
			}
		}

		// Check status code
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
			resp.Body.Close()
			// If server doesn't support resume, start from beginning
			if resumeFrom > 0 && resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
				resumeFrom = 0
				_ = os.Remove(partialPath)
				if stat, statErr := os.Stat(partialPath); statErr == nil {
					_ = stat
				}
				continue
			}

			return &DownloadError{
				Type:    "network",
				Message: fmt.Sprintf("Server returned status code %d", resp.StatusCode),
				Err:     fmt.Errorf("unexpected status code: %d", resp.StatusCode),
			}
		}

		// Server ignored Range header: restart from scratch once.
		if resumeFrom > 0 && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			if resumeFallbacksUsed >= maxResumeFallbacks {
				return &DownloadError{
					Type:    "network",
					Message: "Server does not support resume (range ignored)",
					Err:     fmt.Errorf("server ignored Range header"),
				}
			}
			resumeFallbacksUsed++
			resumeFrom = 0
			_ = os.Remove(partialPath)
			continue
		}

		break
	}
	defer resp.Body.Close()

	// Get total size from Content-Length or Content-Range header
	totalBytes := resp.ContentLength
	if resp.StatusCode == http.StatusPartialContent {
		totalBytes += resumeFrom
	}

	// Check disk space
	if totalBytes > 0 {
		if err := checkDiskSpace(filepath.Dir(destPath), totalBytes); err != nil {
			return &DownloadError{
				Type:    "disk_space",
				Message: fmt.Sprintf("Not enough disk space (need %s)", formatBytes(totalBytes)),
				Err:     err,
			}
		}
	}

	// Open/create destination file
	var out *os.File
	if resumeFrom > 0 && resp.StatusCode == http.StatusPartialContent {
		// Resume: open for append
		out, err = os.OpenFile(partialPath, os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("failed to open partial file: %w", err)
		}
	} else {
		// New download: create file
		out, err = os.Create(partialPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}
	defer out.Close()

	// Track progress
	var downloaded int64 = resumeFrom
	startTime := time.Now()
	lastProgressTime := startTime
	progressUpdateInterval := 500 * time.Millisecond // Update every 500ms

	// Create progress reader
	progressReader := &progressReader{
		reader: resp.Body,
		total:  totalBytes,
		onProgress: func(bytesRead int64) {
			downloaded = resumeFrom + bytesRead
			now := time.Now()

			// Throttle progress updates
			if now.Sub(lastProgressTime) < progressUpdateInterval {
				return
			}
			lastProgressTime = now

			// Calculate speed and percentage
			elapsed := now.Sub(startTime).Seconds()
			speed := int64(0)
			if elapsed > 0 {
				speed = int64(float64(bytesRead) / elapsed)
			}

			percentage := float64(0)
			if totalBytes > 0 {
				percentage = float64(downloaded) / float64(totalBytes) * 100
			}

			message := fmt.Sprintf("Downloaded %s of %s (%.1f%%)", formatBytes(downloaded), formatBytes(totalBytes), percentage)
			if resumeFrom > 0 {
				message = fmt.Sprintf("Resuming: %s", message)
			}

			if progressCb != nil {
				progressCb(domain.DownloadProgress{
					Stage:           "downloading",
					BytesDownloaded: downloaded,
					TotalBytes:      totalBytes,
					Percentage:      percentage,
					Speed:           speed,
					Message:         message,
				})
			}
		},
	}

	// Copy with progress tracking
	_, err = io.Copy(out, progressReader)
	if err != nil {
		// Save partial file for resume
		return &DownloadError{
			Type:    "network",
			Message: "Download interrupted (will resume)",
			Err:     err,
		}
	}

	// Sync to disk
	if err := out.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	out.Close()

	// Rename partial file to final destination
	if err := os.Rename(partialPath, destPath); err != nil {
		return fmt.Errorf("failed to finalize file: %w", err)
	}

	// Final progress update
	if progressCb != nil {
		progressCb(domain.DownloadProgress{
			Stage:           "downloading",
			BytesDownloaded: totalBytes,
			TotalBytes:      totalBytes,
			Percentage:      100,
			Speed:           0,
			Message:         fmt.Sprintf("Download complete: %s", formatBytes(totalBytes)),
		})
	}

	return nil
}

// ExtractTarXz extracts a .tar.xz archive to target directory
func (b *BaseDownloader) ExtractTarXz(archivePath, targetDir string) error {
	// Base directory must exist, otherwise nested paths cannot be created safely.
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	if fi, err := os.Lstat(targetDir); err != nil {
		return fmt.Errorf("failed to stat target directory: %w", err)
	} else if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("target directory is not a real directory: %s", targetDir)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create XZ reader
	xzReader, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	// Create tar reader
	tarReader := tar.NewReader(xzReader)

	// Extract all files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Construct target path
		targetPath, err := safeExtractPath(targetDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory (don't follow symlink chains)
			if err := mkdirAllNoSymlink(targetDir, targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			// Create parent directory if needed
			if err := mkdirAllNoSymlink(targetDir, filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			if err := ensureNoSymlinkFileTarget(targetPath); err != nil {
				return err
			}

			// Create file
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sanitizeArchiveFileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// Copy file contents
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			if err := mkdirAllNoSymlink(targetDir, filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			if err := safeSymlinkTarget(targetDir, targetPath, header.Linkname); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				// Ignore symlink errors (may fail on Windows)
				continue
			}
		case tar.TypeLink:
			// Hard links in archives are not needed and often create security surprises.
			continue
		}
	}

	return nil
}

// ExtractTarGz extracts a .tar.gz archive to target directory
func (b *BaseDownloader) ExtractTarGz(archivePath, targetDir string) error {
	// Base directory must exist, otherwise nested paths cannot be created safely.
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	if fi, err := os.Lstat(targetDir); err != nil {
		return fmt.Errorf("failed to stat target directory: %w", err)
	} else if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("target directory is not a real directory: %s", targetDir)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract all files (same logic as ExtractTarXz)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		targetPath, err := safeExtractPath(targetDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := mkdirAllNoSymlink(targetDir, targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case tar.TypeReg:
			if err := mkdirAllNoSymlink(targetDir, filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			if err := ensureNoSymlinkFileTarget(targetPath); err != nil {
				return err
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sanitizeArchiveFileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()

		case tar.TypeSymlink:
			if err := mkdirAllNoSymlink(targetDir, filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			if err := safeSymlinkTarget(targetDir, targetPath, header.Linkname); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				continue
			}
		case tar.TypeLink:
			continue
		}
	}

	return nil
}

// ExtractZip extracts a .zip archive to target directory
func (b *BaseDownloader) ExtractZip(archivePath, targetDir string) error {
	// Base directory must exist, otherwise nested paths cannot be created safely.
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	if fi, err := os.Lstat(targetDir); err != nil {
		return fmt.Errorf("failed to stat target directory: %w", err)
	} else if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return fmt.Errorf("target directory is not a real directory: %s", targetDir)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath, err := safeExtractPath(targetDir, file.Name)
		if err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			// Create directory
			if err := mkdirAllNoSymlink(targetDir, targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create parent directory
		if err := mkdirAllNoSymlink(targetDir, filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		if err := ensureNoSymlinkFileTarget(targetPath); err != nil {
			return err
		}

		// Open source file
		srcFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		// Create destination file
		dstFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sanitizeArchiveFileMode(int64(file.Mode())))
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Copy contents
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			dstFile.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		srcFile.Close()
		dstFile.Close()
	}

	return nil
}

// VerifyChecksum verifies SHA256 checksum of a file
// Returns error if checksum doesn't match
func (b *BaseDownloader) VerifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// progressReader wraps an io.Reader and calls onProgress callback
type progressReader struct {
	reader     io.Reader
	total      int64
	read       int64
	onProgress func(bytesRead int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)

	if pr.onProgress != nil {
		pr.onProgress(pr.read)
	}

	return n, err
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
