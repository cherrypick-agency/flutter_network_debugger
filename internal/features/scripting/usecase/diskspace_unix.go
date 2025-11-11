//go:build unix

package usecase

import (
	"fmt"
	"syscall"
)

// getAvailableDiskSpace возвращает доступное дисковое пространство (байт) для указанного пути.
func getAvailableDiskSpace(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("failed to stat filesystem: %w", err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
