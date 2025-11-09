//go:build darwin
// +build darwin

package detector

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/features/process/domain"
)

type darwinDetector struct{}

// DetectByPort - найти процесс по локальному порту используя lsof
func (d *darwinDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	// lsof -i :PORT -P -n -F (field output mode)
	cmd := exec.CommandContext(
		ctx,
		"lsof",
		"-i", fmt.Sprintf(":%d", port),
		"-P",        // numeric ports
		"-n",        // numeric IPs
		"-F", "pcn", // fields: PID, command, name
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lsof failed: %w", err)
	}

	// Парсинг вывода lsof
	info, err := parseLsofOutput(string(output))
	if err != nil {
		return nil, err
	}

	return info, nil
}

// DetectByPID - получить информацию о процессе по PID используя ps
func (d *darwinDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
	// ps -p PID -o comm=,args=
	cmd := exec.CommandContext(ctx, "ps", "-p", fmt.Sprint(pid), "-o", "comm=,args=")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ps failed: %w", err)
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return nil, fmt.Errorf("process not found")
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid ps output")
	}

	return &domain.ProcessInfo{
		PID:            pid,
		Name:           parts[0],
		ExecutablePath: parts[0],
		DetectedAt:     time.Now(),
	}, nil
}

// RequiresPrivileges - lsof работает для своих процессов без sudo
func (d *darwinDetector) RequiresPrivileges() bool {
	return false
}

// parseLsofOutput - парсинг вывода lsof в field mode
// Format:
// p<PID>
// c<command>
// n<name>
func parseLsofOutput(output string) (*domain.ProcessInfo, error) {
	lines := strings.Split(output, "\n")

	var pid int32
	var name, cmd string

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		prefix := line[0]
		value := line[1:]

		switch prefix {
		case 'p':
			p, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				continue
			}
			pid = int32(p)
		case 'c':
			cmd = value
		case 'n':
			name = value
		}
	}

	if pid == 0 {
		return nil, fmt.Errorf("no process found in lsof output")
	}

	return &domain.ProcessInfo{
		PID:            pid,
		Name:           cmd,
		ExecutablePath: name,
		DetectedAt:     time.Now(),
	}, nil
}
