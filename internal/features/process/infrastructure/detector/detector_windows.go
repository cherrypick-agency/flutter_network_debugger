//go:build windows
// +build windows

package detector

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"network-debugger/internal/features/process/domain"
)

type windowsDetector struct{}

// DetectByPort - find process by local port using gopsutil
func (w *windowsDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	// gopsutil works well on Windows without admin
	conns, err := net.ConnectionsWithContext(ctx, "tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Find connection by port
	for _, conn := range conns {
		if conn.Laddr.Port == port {
			return w.DetectByPID(ctx, conn.Pid)
		}
	}

	return nil, fmt.Errorf("no process found for port %d", port)
}

// DetectByPID - get process information by PID using gopsutil
func (w *windowsDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
	proc, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("process not found: %w", err)
	}

	name, _ := proc.Name()
	exe, _ := proc.Exe()

	return &domain.ProcessInfo{
		PID:            pid,
		Name:           name,
		ExecutablePath: exe,
		DetectedAt:     time.Now(),
	}, nil
}

// RequiresPrivileges - Windows is more permissive, doesn't require admin for basic information
func (w *windowsDetector) RequiresPrivileges() bool {
	return false
}
