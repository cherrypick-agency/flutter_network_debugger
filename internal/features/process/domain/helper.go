package domain

// IHelperClient - интерфейс для коммуникации с привилегированным helper tool
type IHelperClient interface {
	// IsRunning - проверить, запущен ли helper tool
	IsRunning() bool

	// DetectProcess - детекция процесса через helper (с привилегиями)
	DetectProcess(port uint32) (*ProcessInfo, error)

	// ExtractIcon - извлечение иконки через helper (с привилегиями)
	ExtractIcon(pid int32) (*AppIcon, error)

	// Ping - проверка доступности helper tool
	Ping() error

	// Close - закрыть соединение с helper tool
	Close() error
}
