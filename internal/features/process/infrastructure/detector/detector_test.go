package detector

import (
	"context"
	"os"
	"strings"
	"testing"

	"network-debugger/internal/features/process/domain"
)

// Composer 1.
func TestNewDetector(t *testing.T) {
	// Test without privileges
	detector, err := NewDetector(false)
	if err != nil {
		t.Fatalf("NewDetector(false) error = %v", err)
	}

	if detector == nil {
		t.Fatal("NewDetector(false) returned nil")
	}

	// Check that it implements the interface
	_, ok := detector.(domain.IProcessDetector)
	if !ok {
		t.Error("Detector does not implement IProcessDetector interface")
	}

	// Test with privileges
	detector2, err := NewDetector(true)
	if err != nil {
		t.Fatalf("NewDetector(true) error = %v", err)
	}

	if detector2 == nil {
		t.Fatal("NewDetector(true) returned nil")
	}
}

// Composer 1.
func TestNewDetectorForPlatform(t *testing.T) {
	detector, err := newDetectorForPlatform()
	if err != nil {
		t.Fatalf("newDetectorForPlatform() error = %v", err)
	}

	if detector == nil {
		t.Fatal("newDetectorForPlatform() returned nil")
	}

	// Check that gopsutilAdapter is returned
	_, ok := detector.(*gopsutilAdapter)
	if !ok {
		t.Error("newDetectorForPlatform() did not return gopsutilAdapter")
	}
}

// Composer 1.
func TestGopsutilAdapter_RequiresPrivileges(t *testing.T) {
	adapter := &gopsutilAdapter{}

	if adapter.RequiresPrivileges() {
		t.Error("RequiresPrivileges() should return false for gopsutilAdapter")
	}
}

// TestGopsutilAdapter_DetectByPID tests DetectByPID with valid PID
func TestGopsutilAdapter_DetectByPID(t *testing.T) {
	adapter := &gopsutilAdapter{}
	ctx := context.Background()

	// Use current process PID
	pid := int32(os.Getpid())

	info, err := adapter.DetectByPID(ctx, pid)
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

// TestGopsutilAdapter_DetectByPID_InvalidPID tests DetectByPID with invalid PID
func TestGopsutilAdapter_DetectByPID_InvalidPID(t *testing.T) {
	adapter := &gopsutilAdapter{}
	ctx := context.Background()

	// Use a very large PID that likely doesn't exist
	invalidPID := int32(999999999)

	_, err := adapter.DetectByPID(ctx, invalidPID)
	if err == nil {
		t.Error("Expected error for invalid PID")
	}

	if err != nil && !strings.Contains(err.Error(), "process not found") {
		t.Errorf("Expected 'process not found' error, got: %v", err)
	}
}

// TestGopsutilAdapter_DetectByPort tests DetectByPort
func TestGopsutilAdapter_DetectByPort(t *testing.T) {
	adapter := &gopsutilAdapter{}
	ctx := context.Background()

	// Try to detect a process on a common port (this might fail if no process is listening)
	// We test that the method doesn't panic and handles errors gracefully
	port := uint32(8080)

	info, err := adapter.DetectByPort(ctx, port)
	if err != nil {
		// It's OK if no process is found on this port
		if !strings.Contains(err.Error(), "no process found") {
			t.Errorf("Expected 'no process found' error, got: %v", err)
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

// TestGopsutilAdapter_DetectByPort_InvalidPort tests DetectByPort with invalid port
func TestGopsutilAdapter_DetectByPort_InvalidPort(t *testing.T) {
	adapter := &gopsutilAdapter{}
	ctx := context.Background()

	// Port 0 is invalid
	port := uint32(0)

	_, err := adapter.DetectByPort(ctx, port)
	if err == nil {
		t.Log("DetectByPort with port 0 might succeed or fail depending on system")
	}
}
