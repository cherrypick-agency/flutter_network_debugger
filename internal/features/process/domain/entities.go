package domain

import (
	"errors"
	"time"
)

// ProcessInfo - information about the process that created the network connection
type ProcessInfo struct {
	PID            int32
	Name           string
	ExecutablePath string
	BundleID       *string // macOS only (e.g., "com.apple.Safari")
	Icon           *AppIcon
	DetectedAt     time.Time
}

// AppIcon - application icon
type AppIcon struct {
	Format string  // "png", "icns", "ico"
	Data   []byte  // binary data
	Path   *string // cached file path (optional)
}

// DetectionConfig - process detection configuration
type DetectionConfig struct {
	ID              int64
	Enabled         bool // is detection enabled
	UseHelperTool   bool // use privileged helper tool
	HelperInstalled bool // is helper installed
	CacheEnabled    bool // cache icons
	CacheTTLSeconds int  // cache TTL in seconds
	FallbackEnabled bool // show "Unknown" on detection error
	UpdatedAt       time.Time
}

// Validate - validate configuration
func (c *DetectionConfig) Validate() error {
	if c.CacheTTLSeconds < 0 {
		return errors.New("cache TTL must be non-negative")
	}
	if c.CacheTTLSeconds > 86400 {
		return errors.New("cache TTL must be less than 24 hours")
	}
	return nil
}
