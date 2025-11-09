package domain

import (
	"context"
	"time"
)

// DownloadProgress represents the progress of a compiler download
type DownloadProgress struct {
	Stage           string  // "downloading", "extracting", "verifying", "complete"
	BytesDownloaded int64   // Bytes downloaded so far
	TotalBytes      int64   // Total bytes to download
	Percentage      float64 // Progress percentage (0-100)
	Speed           int64   // Download speed in bytes/sec
	Message         string  // Human-readable message
}

// DownloadRequest contains parameters for downloading a compiler
type DownloadRequest struct {
	Language  string // "zig", "kotlin", "swift"
	Version   string // "latest" or specific version like "0.14.0"
	Platform  string // runtime.GOOS ("darwin", "linux", "windows")
	Arch      string // runtime.GOARCH ("amd64", "arm64")
	TargetDir string // Where to install compiler
}

// CompilerMetadata contains information about a compiler download
type CompilerMetadata struct {
	Language    string    // "zig", "kotlin", "swift"
	Version     string    // "0.14.0", "2.2.20", etc.
	DownloadURL string    // URL to download from
	Size        int64     // Download size in bytes
	Checksum    string    // SHA256 checksum for verification
	ReleaseDate time.Time // Release date
}

// CompilerDownloader is the PORT interface for downloading and installing compilers
// Implementations are ADAPTERS (e.g., ZigDownloader, KotlinDownloader)
type CompilerDownloader interface {
	// GetDownloadURL returns the download URL for the given request
	GetDownloadURL(req DownloadRequest) (string, error)

	// Download downloads and installs the compiler
	// progressCb is called periodically with progress updates
	// Context can be used to cancel the download
	Download(ctx context.Context, req DownloadRequest, progressCb func(DownloadProgress)) error

	// Extract extracts the downloaded archive to target directory
	// archivePath is the path to downloaded file
	// targetDir is where to extract
	Extract(archivePath, targetDir string) error

	// Verify checks that the installed compiler works correctly
	// installPath is the directory containing the compiler
	Verify(installPath string) error

	// GetMetadata returns metadata about the compiler (size, version, etc.)
	// Useful for showing download size before downloading
	GetMetadata(req DownloadRequest) (*CompilerMetadata, error)
}
