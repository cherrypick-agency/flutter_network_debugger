//go:build darwin
// +build darwin

package ipc

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

const (
	// SocketPath - Unix socket path for IPC on macOS
	SocketPath = "/var/run/network-debugger-helper.sock"
)

// CreateListener - create Unix socket listener
// Socket will be created with 0600 permissions (only owner can read/write)
func CreateListener() (net.Listener, error) {
	// Remove existing socket if present
	if err := os.Remove(SocketPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Save current umask and set 0177 to create socket with 0600
	// This atomically sets permissions without security window
	oldMask := syscall.Umask(0177) // 0777 - 0177 = 0600
	defer syscall.Umask(oldMask)

	// Create Unix socket listener (will be created with permissions 0600)
	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket listener: %w", err)
	}

	return listener, nil
}

// CleanupSocket - remove socket file
func CleanupSocket() {
	_ = os.Remove(SocketPath)
}
