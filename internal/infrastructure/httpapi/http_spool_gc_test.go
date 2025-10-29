package httpapi

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanupSpoolDir_RemovesOldFiles(t *testing.T) {
	dir := t.TempDir()
	// create files matching patterns and others
	old := filepath.Join(dir, "gpx-resp-old.bin")
	_ = os.WriteFile(old, []byte("x"), 0o644)
	// set mtime in the past
	past := time.Now().Add(-2 * time.Hour)
	_ = os.Chtimes(old, past, past)
	recent := filepath.Join(dir, "gpx-req-recent.bin")
	_ = os.WriteFile(recent, []byte("y"), 0o644)
	other := filepath.Join(dir, "not-ours.bin")
	_ = os.WriteFile(other, []byte("z"), 0o644)

	cleanupSpoolDir(dir, time.Hour)
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("old file must be removed")
	}
	if _, err := os.Stat(recent); err != nil {
		t.Fatalf("recent file must remain: %v", err)
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatalf("foreign file must remain: %v", err)
	}
}
