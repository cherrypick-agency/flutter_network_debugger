package detector

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	"network-debugger/internal/features/process/domain"
)

// gopsutilAdapter - универсальный детектор используя gopsutil (fallback для всех платформ)
type gopsutilAdapter struct{}

// DetectByPort - найти процесс по локальному порту
func (g *gopsutilAdapter) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	conns, err := net.ConnectionsWithContext(ctx, "tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Найти соединение по порту
	for _, conn := range conns {
		if conn.Laddr.Port == port {
			return g.DetectByPID(ctx, conn.Pid)
		}
	}

	return nil, fmt.Errorf("no process found for port %d", port)
}

// DetectByPID - получить информацию о процессе по PID
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

// RequiresPrivileges - зависит от платформы
func (g *gopsutilAdapter) RequiresPrivileges() bool {
	return false // gopsutil gracefully handles permission errors
}
