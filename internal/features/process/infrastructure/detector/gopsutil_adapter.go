package detector

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"network-debugger/internal/features/process/domain"
)

// gopsutilAdapter - universal detector using gopsutil (fallback for all platforms)
type gopsutilAdapter struct{}

// DetectByPort - find process by local port
func (g *gopsutilAdapter) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	conns, err := net.ConnectionsWithContext(ctx, "tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Find connection by port
	for _, conn := range conns {
		if conn.Laddr.Port == port {
			return g.DetectByPID(ctx, conn.Pid)
		}
	}

	return nil, fmt.Errorf("no process found for port %d", port)
}

// DetectByPID - get process information by PID
func (g *gopsutilAdapter) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
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

// RequiresPrivileges - depends on platform
func (g *gopsutilAdapter) RequiresPrivileges() bool {
	return false // gopsutil gracefully handles permission errors
}
