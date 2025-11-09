//go:build darwin
// +build darwin

package helper

import "network-debugger/internal/features/process/domain"

const (
	// HelperSocketPath - путь к Unix socket helper daemon на macOS
	HelperSocketPath = "/var/run/network-debugger-helper.sock"
)

// NewHelperClient - создать IPC client для macOS
func NewHelperClient() domain.IHelperClient {
	return NewClient(HelperSocketPath)
}
