//go:build windows
// +build windows

package detector

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestWindowsDetector_DetectByPID tests DetectByPID on Windows
func TestWindowsDetector_DetectByPID(t *testing.T) {
	detector := &windowsDetector{}
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

// TestWindowsDetector_DetectByPID_InvalidPID tests DetectByPID with invalid PID
func TestWindowsDetector_DetectByPID_InvalidPID(t *testing.T) {
	detector := &windowsDetector{}
	ctx := context.Background()

	// Use a very large PID that likely doesn't exist
	invalidPID := int32(999999999)

	_, err := detector.DetectByPID(ctx, invalidPID)
	if err == nil {
		t.Error("Expected error for invalid PID")
	}

	if err != nil && !strings.Contains(err.Error(), "process not found") {
		t.Errorf("Expected 'process not found' error, got: %v", err)
	}
}

// TestWindowsDetector_DetectByPort tests DetectByPort on Windows
func TestWindowsDetector_DetectByPort(t *testing.T) {
	detector := &windowsDetector{}
	ctx := context.Background()

	// Try to detect a process on a common port
	// This might fail if no process is listening
	port := uint32(8080)

	info, err := detector.DetectByPort(ctx, port)
	if err != nil {
		// It's OK if no process is found on this port
		if !strings.Contains(err.Error(), "no process found") && !strings.Contains(err.Error(), "failed to get connections") {
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

// TestWindowsDetector_RequiresPrivileges tests RequiresPrivileges on Windows
func TestWindowsDetector_RequiresPrivileges(t *testing.T) {
	detector := &windowsDetector{}

	if detector.RequiresPrivileges() {
		t.Error("Windows detector should not require privileges for basic detection")
	}
}
