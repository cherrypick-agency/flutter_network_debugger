//go:build darwin
// +build darwin

package ipc

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

const (
	// SocketPath - Unix socket path for IPC on macOS
	SocketPath = "/var/run/network-debugger-helper.sock"
)

// CreateListener - создать Unix socket listener
// Socket будет создан с правами 0600 (только owner может читать/писать)
func CreateListener() (net.Listener, error) {
	// Удалить существующий socket если есть
	if err := os.Remove(SocketPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Сохранить текущий umask и установить 0177 для создания socket с 0600
	// Это атомарно устанавливает permissions без security window
	oldMask := syscall.Umask(0177) // 0777 - 0177 = 0600
	defer syscall.Umask(oldMask)

	// Создать Unix socket listener (будет создан с permissions 0600)
	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket listener: %w", err)
	}

	return listener, nil
}

// CleanupSocket - удалить socket файл
func CleanupSocket() {
	_ = os.Remove(SocketPath)
}
