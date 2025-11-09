package domain

import "context"

// IProcessDetector - абстракция детекции процессов
// Реализации для разных OS в infrastructure layer
type IProcessDetector interface {
	// DetectByPort - найти процесс по локальному порту TCP соединения
	DetectByPort(ctx context.Context, port uint32) (*ProcessInfo, error)

	// DetectByPID - получить информацию о процессе по PID
	DetectByPID(ctx context.Context, pid int32) (*ProcessInfo, error)

	// RequiresPrivileges - требуются ли root/admin права для полной детекции
	RequiresPrivileges() bool
}
