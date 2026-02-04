package domain

// IHelperClient - interface for communication with privileged helper tool
type IHelperClient interface {
	// IsRunning - check if helper tool is running
	IsRunning() bool

	// DetectProcess - detect process via helper (with privileges)
	DetectProcess(port uint32) (*ProcessInfo, error)

	// ExtractIcon - extract icon via helper (with privileges)
	ExtractIcon(pid int32) (*AppIcon, error)

	// Ping - check helper tool availability
	Ping() error

	// Close - close connection to helper tool
	Close() error
}
