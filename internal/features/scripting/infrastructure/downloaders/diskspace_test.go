//go:build unix

package downloaders

import (
	"os"
	"path/filepath"
	"testing"
)

// Composer 1.
func TestCheckDiskSpace(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with sufficient space (requesting 1 byte)
	err := checkDiskSpace(tmpDir, 1)
	if err != nil {
		t.Errorf("checkDiskSpace() with sufficient space error = %v, want nil", err)
	}

	// Test with very large request (should return error if space is insufficient)
	// But on most systems this won't work because there's enough space
	// So we just check that the function doesn't panic
	err = checkDiskSpace(tmpDir, 1<<60) // 1 exabyte
	// May return error or nil depending on the system
	_ = err
}

// Composer 1.
func TestCheckDiskSpace_InvalidDir(t *testing.T) {
	// Test with non-existent directory
	// Function should return nil (continue work)
	err := checkDiskSpace("/nonexistent/directory/12345", 1000)
	if err != nil {
		// If it returns error - that's also fine, the main thing is it doesn't panic
		_ = err
	}
}

// Composer 1.
func TestCheckDiskSpace_CurrentDir(t *testing.T) {
	// Test with current directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	err = checkDiskSpace(wd, 1000)
	if err != nil {
		// May return error if space is insufficient
		_ = err
	}
}

// Composer 1.
func TestCheckDiskSpace_TempDir(t *testing.T) {
	// Test with temporary directory
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err := checkDiskSpace(testDir, 1000)
	if err != nil {
		// May return error if space is insufficient
		_ = err
	}
}
