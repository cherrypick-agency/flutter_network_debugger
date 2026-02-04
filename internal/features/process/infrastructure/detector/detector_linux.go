//go:build linux
// +build linux

package detector

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/features/process/domain"
)

type linuxDetector struct{}

// DetectByPort - find process by local port via /proc/net/tcp
func (l *linuxDetector) DetectByPort(ctx context.Context, port uint32) (*domain.ProcessInfo, error) {
	// 1. Read /proc/net/tcp
	tcpData, err := os.ReadFile("/proc/net/tcp")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/net/tcp: %w", err)
	}

	// 2. Find inode by port
	inode, err := findInodeByPort(string(tcpData), port)
	if err != nil {
		return nil, err
	}

	// 3. Find PID by inode
	pid, err := findPIDByInode(inode)
	if err != nil {
		return nil, err
	}

	return l.DetectByPID(ctx, int32(pid))
}

// DetectByPID - get process information by PID via /proc/[pid]
func (l *linuxDetector) DetectByPID(ctx context.Context, pid int32) (*domain.ProcessInfo, error) {
	// Read /proc/[pid]/comm for name
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := os.ReadFile(commPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read comm: %w", err)
	}

	// Read /proc/[pid]/exe for path
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	exe, _ := os.Readlink(exePath)

	return &domain.ProcessInfo{
		PID:            pid,
		Name:           strings.TrimSpace(string(commData)),
		ExecutablePath: exe,
		DetectedAt:     time.Now(),
	}, nil
}

// RequiresPrivileges - /proc/net/tcp is accessible, but /proc/[pid]/fd may require permissions
func (l *linuxDetector) RequiresPrivileges() bool {
	return false
}

// findInodeByPort - find socket inode by port in /proc/net/tcp
func findInodeByPort(tcpData string, port uint32) (uint64, error) {
	lines := strings.Split(tcpData, "\n")

	for _, line := range lines[1:] { // skip header
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// local_address in hex: "0100007F:1F90" = 127.0.0.1:8080
		localAddr := fields[1]
		parts := strings.Split(localAddr, ":")
		if len(parts) != 2 {
			continue
		}

		// Parse hex port
		portHex := parts[1]
		p, err := strconv.ParseUint(portHex, 16, 32)
		if err != nil {
			continue
		}

		if uint32(p) == port {
			// inode in 9th field
			inode, _ := strconv.ParseUint(fields[9], 10, 64)
			return inode, nil
		}
	}

	return 0, fmt.Errorf("port %d not found in /proc/net/tcp", port)
}

// findPIDByInode - find process PID by socket inode
func findPIDByInode(inode uint64) (int, error) {
	// Iterate over all /proc/[pid]/fd/* looking for socket:[inode]
	procDir := "/proc"
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return 0, err
	}

	target := fmt.Sprintf("socket:[%d]", inode)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		fdDir := filepath.Join(procDir, entry.Name(), "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue // permission denied - skip
		}

		for _, fd := range fds {
			linkPath := filepath.Join(fdDir, fd.Name())
			link, err := os.Readlink(linkPath)
			if err != nil {
				continue
			}

			if link == target {
				return pid, nil
			}
		}
	}

	return 0, fmt.Errorf("no PID found for inode %d", inode)
}
