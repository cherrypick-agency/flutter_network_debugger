package detector

import (
	"network-debugger/internal/features/process/domain"
)

// NewDetector - создать детектор процессов для текущей платформы
func NewDetector(privileged bool) (domain.IProcessDetector, error) {
	// Пока privileged и unprivileged детекторы идентичны
	// В будущем privileged будет использовать helper tool
	return newDetectorForPlatform()
}

func newDetectorForPlatform() (domain.IProcessDetector, error) {
	// Используем gopsutil adapter как универсальное решение для всех платформ
	// Platform-specific детекторы (darwin/windows/linux) доступны но не используются пока
	return &gopsutilAdapter{}, nil
}
