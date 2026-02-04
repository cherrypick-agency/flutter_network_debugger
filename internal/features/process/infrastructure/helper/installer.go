package helper

// Installer - interface for installing/uninstalling helper tool
type Installer interface {
	// IsInstalled - check if helper is installed
	IsInstalled() bool

	// Install - install helper tool
	// May require administrator password
	Install(helperBinaryPath string) error

	// Uninstall - uninstall helper tool
	Uninstall() error

	// GetVersion - get installed helper version
	GetVersion() string
}
