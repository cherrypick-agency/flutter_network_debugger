//go:build !darwin
// +build !darwin

package helper

import (
	"fmt"

	"network-debugger/internal/features/process/domain"
)

// stubHelperClient - stub for non-macOS platforms
type stubHelperClient struct{}

func (s *stubHelperClient) IsRunning() bool {
	return false
}

func (s *stubHelperClient) DetectProcess(port uint32) (*domain.ProcessInfo, error) {
	return nil, fmt.Errorf("helper tool not supported on this platform")
}

func (s *stubHelperClient) ExtractIcon(pid int32) (*domain.AppIcon, error) {
	return nil, fmt.Errorf("helper tool not supported on this platform")
}

func (s *stubHelperClient) Ping() error {
	return fmt.Errorf("helper tool not supported on this platform")
}

func (s *stubHelperClient) Close() error {
	return nil
}

// NewHelperClient - create stub client for non-macOS
func NewHelperClient() domain.IHelperClient {
	return &stubHelperClient{}
}
