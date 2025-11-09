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

// DetectByPort - найти процесс по локальному порту используя gopsutil
func (w *windowsDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	// gopsutil работает хорошо на Windows без admin
	conns, err := net.ConnectionsWithContext(ctx, "tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Найти соединение по порту
	for _, conn := range conns {
		if conn.Laddr.Port == port {
			return w.DetectByPID(ctx, conn.Pid)
		}
	}

	return nil, fmt.Errorf("no process found for port %d", port)
}

// DetectByPID - получить информацию о процессе по PID используя gopsutil
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

// RequiresPrivileges - Windows более permissive, не требует admin для базовой информации
func (w *windowsDetector) RequiresPrivileges() bool {
	return false
}
