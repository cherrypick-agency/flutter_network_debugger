package observability

import (
	"testing"
)

func TestVersion_DefaultValues(t *testing.T) {
	// Test that version variables have default values
	if Version == "" {
		t.Error("Version should have a default value")
	}
	if Commit == "" {
		t.Error("Commit should have a default value")
	}

	// These should match the default values from version.go
	expectedVersion := "dev"
	expectedCommit := "none"

	if Version != expectedVersion {
		t.Errorf("Version = %q, want %q", Version, expectedVersion)
	}
	if Commit != expectedCommit {
		t.Errorf("Commit = %q, want %q", Commit, expectedCommit)
	}
}

func TestVersion_CanBeModified(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := Commit
	origDate := Date

	// Verify we can modify them (simulates -ldflags behavior)
	Version = "1.0.0"
	Commit = "abc123"
	Date = "2024-01-01"

	if Version != "1.0.0" {
		t.Error("Version should be modifiable")
	}
	if Commit != "abc123" {
		t.Error("Commit should be modifiable")
	}
	if Date != "2024-01-01" {
		t.Error("Date should be modifiable")
	}

	// Restore original values
	Version = origVersion
	Commit = origCommit
	Date = origDate
}
