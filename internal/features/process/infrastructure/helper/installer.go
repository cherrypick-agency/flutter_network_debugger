package helper

// Installer - интерфейс для установки/удаления helper tool
type Installer interface {
	// IsInstalled - проверить установлен ли helper
	IsInstalled() bool

	// Install - установить helper tool
	// Может потребовать пароль администратора
	Install(helperBinaryPath string) error

	// Uninstall - удалить helper tool
	Uninstall() error

	// GetVersion - получить версию установленного helper
	GetVersion() string
}
