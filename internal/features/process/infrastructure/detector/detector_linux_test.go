//go:build linux
// +build linux

package detector

import (
	"context"
	"os"
	"strings"
	"testing"

	"network-debugger/internal/features/process/domain"
)

// TestLinuxDetector_DetectByPID tests DetectByPID on Linux
func TestLinuxDetector_DetectByPID(t *testing.T) {
	detector := &linuxDetector{}
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

// TestLinuxDetector_DetectByPID_InvalidPID tests DetectByPID with invalid PID
func TestLinuxDetector_DetectByPID_InvalidPID(t *testing.T) {
	detector := &linuxDetector{}
	ctx := context.Background()

	// Use a very large PID that likely doesn't exist
	invalidPID := int32(999999999)

	_, err := detector.DetectByPID(ctx, invalidPID)
	if err == nil {
		t.Error("Expected error for invalid PID")
	}
}

// TestLinuxDetector_DetectByPort tests DetectByPort on Linux
func TestLinuxDetector_DetectByPort(t *testing.T) {
	detector := &linuxDetector{}
	ctx := context.Background()

	// Try to detect a process on a common port
	// This might fail if /proc/net/tcp is not accessible or no process is listening
	port := uint32(8080)

	info, err := detector.DetectByPort(ctx, port)
	if err != nil {
		// It's OK if no process is found or /proc/net/tcp is not accessible
		if !strings.Contains(err.Error(), "port") && !strings.Contains(err.Error(), "failed to read") {
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

// TestLinuxDetector_RequiresPrivileges tests RequiresPrivileges on Linux
func TestLinuxDetector_RequiresPrivileges(t *testing.T) {
	detector := &linuxDetector{}

	if detector.RequiresPrivileges() {
		t.Error("Linux detector should not require privileges for basic detection")
	}
}

// TestFindInodeByPort tests findInodeByPort helper function
func TestFindInodeByPort(t *testing.T) {
	tcpData := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345
   1: 0100007F:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12346`

	// Test with port 8080 (hex: 1F90)
	inode, err := findInodeByPort(tcpData, 8080)
	if err != nil {
		t.Fatalf("findInodeByPort failed: %v", err)
	}

	if inode != 12345 {
		t.Errorf("Expected inode 12345, got %d", inode)
	}

	// Test with port 80 (hex: 0050)
	inode, err = findInodeByPort(tcpData, 80)
	if err != nil {
		t.Fatalf("findInodeByPort failed: %v", err)
	}

	if inode != 12346 {
		t.Errorf("Expected inode 12346, got %d", inode)
	}

	// Test with non-existent port
	_, err = findInodeByPort(tcpData, 9999)
	if err == nil {
		t.Error("Expected error for non-existent port")
	}
}

// TestFindPIDByInode tests findPIDByInode helper function
func TestFindPIDByInode(t *testing.T) {
	// This test might fail if /proc is not accessible or the inode doesn't exist
	// We just test that the function doesn't panic
	inode := uint64(999999999)

	_, err := findPIDByInode(inode)
	if err == nil {
		t.Log("findPIDByInode found PID (unexpected, but OK)")
	} else {
		if !strings.Contains(err.Error(), "no PID found") {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}
