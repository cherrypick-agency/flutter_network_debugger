package domain

import "context"

// IIconExtractor - абстракция извлечения иконок приложений
type IIconExtractor interface {
	// ExtractByPID - извлечь иконку приложения по PID процесса
	ExtractByPID(ctx context.Context, pid int32) (*AppIcon, error)

	// ExtractByPath - извлечь иконку по пути к приложению/executable
	ExtractByPath(ctx context.Context, path string) (*AppIcon, error)
}
