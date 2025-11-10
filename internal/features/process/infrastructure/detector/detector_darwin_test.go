//go:build darwin
// +build darwin

package detector

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestDarwinDetector_DetectByPID tests DetectByPID on Darwin
func TestDarwinDetector_DetectByPID(t *testing.T) {
	detector := &darwinDetector{}
	ctx := context.Background()

	// Use current process PID
	pid := int32(os.Getpid())

	info, err := detector.DetectByPID(ctx, pid)
	if err != nil {
		t.Fatalf("DetectByPID failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected non-nil ProcessInfo")
	}

	if info.PID != pid {
		t.Errorf("Expected PID %d, got %d", pid, info.PID)
	}

	if info.Name == "" {
		t.Error("Expected non-empty process name")
	}
}

// TestDarwinDetector_DetectByPID_InvalidPID tests DetectByPID with invalid PID
func TestDarwinDetector_DetectByPID_InvalidPID(t *testing.T) {
	detector := &darwinDetector{}
	ctx := context.Background()

	// Use a very large PID that likely doesn't exist
	invalidPID := int32(999999999)

	_, err := detector.DetectByPID(ctx, invalidPID)
	if err == nil {
		t.Error("Expected error for invalid PID")
	}

	if err != nil && !strings.Contains(err.Error(), "process not found") && !strings.Contains(err.Error(), "ps failed") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestDarwinDetector_DetectByPort tests DetectByPort on Darwin
func TestDarwinDetector_DetectByPort(t *testing.T) {
	detector := &darwinDetector{}
	ctx := context.Background()

	// Try to detect a process on a common port
	// This might fail if lsof is not available or no process is listening
	port := uint32(8080)

	info, err := detector.DetectByPort(ctx, port)
	if err != nil {
		// It's OK if no process is found or lsof fails
		if !strings.Contains(err.Error(), "no process found") && !strings.Contains(err.Error(), "lsof failed") {
			t.Errorf("Unexpected error: %v", err)
		}
		return
	}

	// If a process was found, verify the result
	if info == nil {
		t.Fatal("Expected non-nil ProcessInfo when no error")
	}

	if info.PID == 0 {
		t.Error("Expected non-zero PID")
	}
}

// TestDarwinDetector_RequiresPrivileges tests RequiresPrivileges on Darwin
func TestDarwinDetector_RequiresPrivileges(t *testing.T) {
	detector := &darwinDetector{}

	if detector.RequiresPrivileges() {
		t.Error("Darwin detector should not require privileges for basic detection")
	}
}

// TestParseLsofOutput tests parseLsofOutput helper function
func TestParseLsofOutput(t *testing.T) {
	output := `p12345
cmyapp
n127.0.0.1:8080`

	info, err := parseLsofOutput(output)
	if err != nil {
		t.Fatalf("parseLsofOutput failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected non-nil ProcessInfo")
	}

	if info.PID != 12345 {
		t.Errorf("Expected PID 12345, got %d", info.PID)
	}

	if info.Name != "myapp" {
		t.Errorf("Expected name 'myapp', got '%s'", info.Name)
	}

	if info.ExecutablePath != "127.0.0.1:8080" {
		t.Errorf("Expected executable path '127.0.0.1:8080', got '%s'", info.ExecutablePath)
	}
}

// TestParseLsofOutput_Invalid tests parseLsofOutput with invalid input
func TestParseLsofOutput_Invalid(t *testing.T) {
	// Empty output
	_, err := parseLsofOutput("")
	if err == nil {
		t.Error("Expected error for empty output")
	}

	// Output without PID
	_, err = parseLsofOutput("cmyapp\nn127.0.0.1:8080")
	if err == nil {
		t.Error("Expected error for output without PID")
	}
}
