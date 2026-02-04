//go:build windows

package usecase

import (
	"syscall"
	"unsafe"
)

// getAvailableDiskSpace returns available disk space (bytes) for the specified path.
// On Windows uses GetDiskFreeSpaceExW.
func getAvailableDiskSpace(path string) (int64, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetDiskFreeSpaceExW")

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	var freeBytesAvailable int64
	var totalBytes int64
	var totalFreeBytes int64

	ret, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		// Return error so caller can handle best-effort scenario
		if callErr != nil {
			return 0, callErr
		}
		return 0, syscall.EINVAL
	}

	return freeBytesAvailable, nil
}
