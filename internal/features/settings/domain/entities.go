package domain

import "time"

// RuntimeSettings — single record with active runtime settings.
type RuntimeSettings struct {
	ID int64

	ResponseDelayMs    int
	ResponseDelayMinMs int
	ResponseDelayMaxMs int

	ThrottleEnabled       bool
	ThrottleDownKbps      int
	ThrottleUpKbps        int
	ThrottlePacketLoss    int
	ThrottleLatencyMs     int // Base latency (RTT/ping simulation)
	ThrottleLatencyJitter int // Random jitter ± ms
	ThrottleOffline       bool

	FontScale float64

	HighlightTheme      string
	HighlightThemeLight string
	HighlightThemeDark  string

	// Mapping (Map Local/Remote)
	MappingEnabled     bool
	MappingUploadMaxMB int

	UpdatedAt time.Time
}

// ThrottleProfile — user preset for speed/reliability.
type ThrottleProfile struct {
	ID            string
	Name          string
	DownKbps      int
	UpKbps        int
	PacketLossPct int
	LatencyMs     int // Base latency (RTT/ping simulation)
	LatencyJitter int // Random jitter ± ms
	Offline       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
