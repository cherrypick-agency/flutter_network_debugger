//go:build unix

package downloaders

import (
	"fmt"
	"syscall"
)

// checkDiskSpace checks if there's enough disk space to download a file
func checkDiskSpace(dir string, requiredBytes int64) error {
	const diskSpaceBuffer = 100 * 1024 * 1024 // 100MB safety buffer

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		// If we can't check disk space, proceed anyway
		return nil
	}

	// Calculate available space
	availableBytes := int64(stat.Bavail) * int64(stat.Bsize)
	requiredWithBuffer := requiredBytes + diskSpaceBuffer

	if availableBytes < requiredWithBuffer {
		return fmt.Errorf("%w: have %s, need %s (including 100MB buffer)",
			ErrInsufficientDiskSpace,
			formatBytes(availableBytes),
			formatBytes(requiredWithBuffer))
	}

	return nil
}
