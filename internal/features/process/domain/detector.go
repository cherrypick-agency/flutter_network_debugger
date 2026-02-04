package domain

import "context"

// IProcessDetector - abstraction for process detection
// Implementations for different OS in infrastructure layer
type IProcessDetector interface {
	// DetectByPort - find process by local TCP connection port
	DetectByPort(ctx context.Context, port uint32) (*ProcessInfo, error)

	// DetectByPID - get process information by PID
	DetectByPID(ctx context.Context, pid int32) (*ProcessInfo, error)

	// RequiresPrivileges - are root/admin rights required for full detection
	RequiresPrivileges() bool
}
