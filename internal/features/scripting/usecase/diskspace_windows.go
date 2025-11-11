//go:build windows

package usecase

import (
	"syscall"
	"unsafe"
)

// getAvailableDiskSpace возвращает доступное дисковое пространство (байт) для указанного пути.
// На Windows использует GetDiskFreeSpaceExW.
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
		// Возвращаем ошибку, чтобы вызывающий мог обработать best-effort сценарий
		if callErr != nil {
			return 0, callErr
		}
		return 0, syscall.EINVAL
	}

	return freeBytesAvailable, nil
}
