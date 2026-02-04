//go:build darwin
// +build darwin

package helper

import "network-debugger/internal/features/process/domain"

const (
	// HelperSocketPath - path to Unix socket of helper daemon on macOS
	HelperSocketPath = "/var/run/network-debugger-helper.sock"
)

// NewHelperClient - create IPC client for macOS
func NewHelperClient() domain.IHelperClient {
	return NewClient(HelperSocketPath)
}
