//go:build linux
// +build linux

package icon

import (
	"context"
	"fmt"

	"network-debugger/internal/features/process/domain"
)

type linuxExtractor struct{}

// ExtractByPID - извлечь иконку по PID (TODO: реализовать через .desktop files)
func (e *linuxExtractor) ExtractByPID(ctx context.Context, pid int32) (*domain.AppIcon, error) {
	// TODO: Реализовать через:
	// 1. Получить executable path процесса
	// 2. Найти соответствующий .desktop файл в /usr/share/applications/
	// 3. Извлечь Icon= поле
	// 4. Найти иконку в /usr/share/icons/ или ~/.icons/
	// 5. Вернуть PNG/SVG данные
	return nil, fmt.Errorf("Linux icon extraction not implemented yet")
}

// ExtractByPath - извлечь иконку по пути (TODO: реализовать через freedesktop spec)
func (e *linuxExtractor) ExtractByPath(ctx context.Context, path string) (*domain.AppIcon, error) {
	// TODO: Реализовать через freedesktop.org icon theme specification
	return nil, fmt.Errorf("Linux icon extraction not implemented yet")
}
