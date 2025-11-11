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

// TestDarwinDetector_DetectByPort_InvalidPort tests DetectByPort with invalid port
func TestDarwinDetector_DetectByPort_InvalidPort(t *testing.T) {
	detector := &darwinDetector{}
	ctx := context.Background()

	// Use a very large port number
	invalidPort := uint32(99999)

	_, err := detector.DetectByPort(ctx, invalidPort)
	if err == nil {
		t.Error("Expected error for invalid port")
	}

	if err != nil && !strings.Contains(err.Error(), "lsof failed") && !strings.Contains(err.Error(), "no process found") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestDarwinDetector_DetectByPort_ContextCancellation tests DetectByPort with cancelled context
func TestDarwinDetector_DetectByPort_ContextCancellation(t *testing.T) {
	detector := &darwinDetector{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	port := uint32(8080)

	_, err := detector.DetectByPort(ctx, port)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}

	if err != nil && !strings.Contains(err.Error(), "context canceled") && !strings.Contains(err.Error(), "lsof failed") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestParseLsofOutput_MultipleEntries tests parseLsofOutput with multiple entries
func TestParseLsofOutput_MultipleEntries(t *testing.T) {
	output := `p12345
cmyapp
n127.0.0.1:8080
p67890
canotherapp
n127.0.0.1:8081`

	info, err := parseLsofOutput(output)
	if err != nil {
		t.Fatalf("parseLsofOutput failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected non-nil ProcessInfo")
	}

	// Функция перезаписывает pid при каждом найденном 'p', поэтому возвращается последний
	if info.PID != 67890 {
		t.Errorf("Expected PID 67890 (last PID found), got %d", info.PID)
	}

	// Проверяем что команда и имя тоже соответствуют последней записи
	if info.Name != "anotherapp" {
		t.Errorf("Expected name 'anotherapp', got '%s'", info.Name)
	}
}

// TestParseLsofOutput_InvalidPID tests parseLsofOutput with invalid PID
func TestParseLsofOutput_InvalidPID(t *testing.T) {
	output := `pnotanumber
cmyapp
n127.0.0.1:8080`

	_, err := parseLsofOutput(output)
	if err == nil {
		t.Error("Expected error for invalid PID")
	}
}

// TestParseLsofOutput_EmptyLines tests parseLsofOutput with empty lines
func TestParseLsofOutput_EmptyLines(t *testing.T) {
	output := `p12345

cmyapp

n127.0.0.1:8080
`

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
}
