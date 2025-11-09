package domain

// CompilerStatus represents the installation status of a compiler
type CompilerStatus string

const (
	CompilerStatusNotInstalled CompilerStatus = "not_installed"
	CompilerStatusInstalling   CompilerStatus = "installing"
	CompilerStatusInstalled    CompilerStatus = "installed"
	CompilerStatusError        CompilerStatus = "error"
)

// CompilerInfo contains information about a compiler's status
type CompilerInfo struct {
	Language      string         // "zig", "kotlin", "swift", "rust", etc.
	Version       string         // Installed version or "latest"
	Status        CompilerStatus // Current status
	InstalledPath string         // Path to compiler binary (empty if not installed)
	Size          int64          // Installed size in bytes (0 if not installed)
	DownloadSize  int64          // Size to download in bytes (0 if already installed)
	Error         string         // Error message if Status is "error"
}

// CompilerManager is the PORT interface for managing compiler lifecycle
// Implementations are ADAPTERS
type CompilerManager interface {
	// GetStatus returns the current status of a compiler
	GetStatus(language string) (*CompilerInfo, error)

	// IsInstalled checks if a compiler is installed (system or cache)
	IsInstalled(language string) bool

	// GetInstalledPath returns the path to installed compiler binary
	// Returns error if not installed
	GetInstalledPath(language string) (string, error)

	// ListAll returns information about all supported compilers
	ListAll() ([]*CompilerInfo, error)
}
