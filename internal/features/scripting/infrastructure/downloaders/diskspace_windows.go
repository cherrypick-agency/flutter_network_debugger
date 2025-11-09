//go:build windows

package downloaders

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpace = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// checkDiskSpace checks if there's enough disk space to download a file
func checkDiskSpace(dir string, requiredBytes int64) error {
	const diskSpaceBuffer = 100 * 1024 * 1024 // 100MB safety buffer

	var freeBytesAvailable int64
	var totalBytes int64
	var totalFreeBytes int64

	pathPtr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		// If we can't check disk space, proceed anyway
		return nil
	}

	ret, _, _ := getDiskFreeSpace.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)

	if ret == 0 {
		// If we can't check disk space, proceed anyway
		return nil
	}

	requiredWithBuffer := requiredBytes + diskSpaceBuffer

	if freeBytesAvailable < requiredWithBuffer {
		return fmt.Errorf("%w: have %s, need %s (including 100MB buffer)",
			ErrInsufficientDiskSpace,
			formatBytes(freeBytesAvailable),
			formatBytes(requiredWithBuffer))
	}

	return nil
}
