package detector

import (
	"network-debugger/internal/features/process/domain"
)

// NewDetector - create process detector for the current platform
func NewDetector(privileged bool) (domain.IProcessDetector, error) {
	// For now privileged and unprivileged detectors are identical
	// In the future privileged will use helper tool
	return newDetectorForPlatform()
}

func newDetectorForPlatform() (domain.IProcessDetector, error) {
	// Using gopsutil adapter as universal solution for all platforms
	// Platform-specific detectors (darwin/windows/linux) are available but not used for now
	return &gopsutilAdapter{}, nil
}
