package httpapi

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupSpoolDir_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	// Should not panic with empty directory
	cleanupSpoolDir(tempDir, 24*time.Hour)

	// Directory should still exist
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("directory should still exist")
	}
}

func TestCleanupSpoolDir_NoMatchingFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create files that don't match the pattern
	files := []string{"test.txt", "data.json", "other-file.bin"}
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	cleanupSpoolDir(tempDir, 1*time.Second)

	// All files should still exist (not matching pattern)
	for _, f := range files {
		path := filepath.Join(tempDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file %s should not be deleted (doesn't match pattern)", f)
		}
	}
}

func TestCleanupSpoolDir_OldFilesRemoved(t *testing.T) {
	tempDir := t.TempDir()

	// Create old files matching the pattern
	oldFiles := []string{"gpx-req-old.bin", "gpx-resp-old.bin", "gpx-ws-old.bin"}
	for _, f := range oldFiles {
		path := filepath.Join(tempDir, f)
		if err := os.WriteFile(path, []byte("old data"), 0o600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		// Set modification time to 2 hours ago
		oldTime := time.Now().Add(-2 * time.Hour)
		if err := os.Chtimes(path, oldTime, oldTime); err != nil {
			t.Fatalf("failed to set file time: %v", err)
		}
	}

	// Clean up files older than 1 hour
	cleanupSpoolDir(tempDir, 1*time.Hour)

	// All old files should be deleted
	for _, f := range oldFiles {
		path := filepath.Join(tempDir, f)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file %s should be deleted (too old)", f)
		}
	}
}

func TestCleanupSpoolDir_RecentFilesKept(t *testing.T) {
	tempDir := t.TempDir()

	// Create recent files matching the pattern
	recentFiles := []string{"gpx-req-recent.bin", "gpx-resp-recent.bin", "gpx-ws-recent.bin"}
	for _, f := range recentFiles {
		path := filepath.Join(tempDir, f)
		if err := os.WriteFile(path, []byte("recent data"), 0o600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		// File is just created, so it's recent
	}

	// Clean up files older than 1 hour (these files are recent)
	cleanupSpoolDir(tempDir, 1*time.Hour)

	// All recent files should still exist
	for _, f := range recentFiles {
		path := filepath.Join(tempDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file %s should not be deleted (still recent)", f)
		}
	}
}

func TestCleanupSpoolDir_MixedFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create mix of old and recent files
	oldFile := filepath.Join(tempDir, "gpx-req-old.bin")
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set old file time: %v", err)
	}

	recentFile := filepath.Join(tempDir, "gpx-req-recent.bin")
	if err := os.WriteFile(recentFile, []byte("recent"), 0o600); err != nil {
		t.Fatalf("failed to create recent file: %v", err)
	}

	// Clean up files older than 1 hour
	cleanupSpoolDir(tempDir, 1*time.Hour)

	// Old file should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old file should be deleted")
	}

	// Recent file should still exist
	if _, err := os.Stat(recentFile); os.IsNotExist(err) {
		t.Error("recent file should not be deleted")
	}
}

func TestCleanupSpoolDir_DefaultDir(t *testing.T) {
	// Should not panic with empty string (uses os.TempDir())
	cleanupSpoolDir("", 24*time.Hour)
}

func TestCleanupSpoolDir_NonexistentDir(t *testing.T) {
	// Should not panic with nonexistent directory
	cleanupSpoolDir("/nonexistent/directory/path", 24*time.Hour)
}

func TestCleanupSpoolDir_WithSubdirectories(t *testing.T) {
	tempDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create old file in subdirectory
	oldFile := filepath.Join(subDir, "gpx-req-old.bin")
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatalf("failed to create file in subdir: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	// Clean up files older than 1 hour
	cleanupSpoolDir(tempDir, 1*time.Hour)

	// Old file in subdirectory should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("old file in subdirectory should be deleted")
	}

	// Subdirectory should still exist
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("subdirectory should still exist")
	}
}

func TestCleanupSpoolDir_AllPatterns(t *testing.T) {
	tempDir := t.TempDir()

	// Test all three pattern prefixes
	patterns := []string{"gpx-req-", "gpx-resp-", "gpx-ws-"}
	for _, prefix := range patterns {
		oldFile := filepath.Join(tempDir, prefix+"old.bin")
		if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		oldTime := time.Now().Add(-2 * time.Hour)
		if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
			t.Fatalf("failed to set file time: %v", err)
		}
	}

	// Clean up files older than 1 hour
	cleanupSpoolDir(tempDir, 1*time.Hour)

	// All pattern files should be deleted
	for _, prefix := range patterns {
		oldFile := filepath.Join(tempDir, prefix+"old.bin")
		if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
			t.Errorf("file with prefix %s should be deleted", prefix)
		}
	}
}

func TestCleanupSpoolDir_ZeroDuration(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file
	file := filepath.Join(tempDir, "gpx-req-test.bin")
	if err := os.WriteFile(file, []byte("test"), 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Clean up with zero duration (everything is old)
	cleanupSpoolDir(tempDir, 0)

	// File should be deleted
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Error("file should be deleted with zero duration")
	}
}

func TestCleanupSpoolDir_VeryLongDuration(t *testing.T) {
	tempDir := t.TempDir()

	// Create an old file
	oldFile := filepath.Join(tempDir, "gpx-req-old.bin")
	if err := os.WriteFile(oldFile, []byte("old"), 0o600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set file time: %v", err)
	}

	// Clean up with very long duration (nothing is old enough)
	cleanupSpoolDir(tempDir, 1000*time.Hour)

	// File should still exist
	if _, err := os.Stat(oldFile); os.IsNotExist(err) {
		t.Error("file should not be deleted with very long duration")
	}
}
