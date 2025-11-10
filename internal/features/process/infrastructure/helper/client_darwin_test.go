//go:build darwin
// +build darwin

package helper

import (
	"testing"

	"network-debugger/internal/features/process/domain"
)

// Composer 1.
func TestNewHelperClient_Darwin(t *testing.T) {
	client := NewHelperClient()

	if client == nil {
		t.Fatal("NewHelperClient returned nil")
	}

	realClient, ok := client.(*Client)
	if !ok {
		t.Fatal("NewHelperClient should return *Client on darwin")
	}

	if realClient.socketPath != HelperSocketPath {
		t.Errorf("socketPath = %q, want %q", realClient.socketPath, HelperSocketPath)
	}
}

// Composer 1.
func TestNewHelperClient_ImplementsInterface(t *testing.T) {
	client := NewHelperClient()

	if _, ok := client.(domain.IHelperClient); !ok {
		t.Error("NewHelperClient should implement domain.IHelperClient")
	}
}
