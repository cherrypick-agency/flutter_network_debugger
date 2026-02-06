package domain

import "errors"

// InterceptConfig holds intercept feature configuration
type InterceptConfig struct {
	Enabled      bool   `json:"enabled"`
	Requests     bool   `json:"requests"`
	Responses    bool   `json:"responses"`
	TimeoutMs    int    `json:"timeoutMs"`
	QueueMax     int    `json:"queueMax"`
	BodyMaxBytes int    `json:"bodyMaxBytes"`
	Reencode     bool   `json:"reencode"`
	Overflow     string `json:"overflow"`
}

// SetDefaults sets default values for unset fields
func (c *InterceptConfig) SetDefaults() {
	if c.TimeoutMs <= 0 {
		c.TimeoutMs = 60000
	}
	if c.QueueMax <= 0 {
		c.QueueMax = 200
	}
	if c.BodyMaxBytes <= 0 {
		c.BodyMaxBytes = 1 << 20 // 1MB
	}
	if c.Overflow == "" {
		c.Overflow = "auto-continue-oldest"
	}
}

// Validate performs domain-level validation
func (c *InterceptConfig) Validate() error {
	if c.TimeoutMs <= 0 {
		return errors.New("timeoutMs must be positive")
	}
	if c.TimeoutMs > 300000 {
		return errors.New("timeoutMs must not exceed 300000 (5 minutes)")
	}
	if c.QueueMax < 0 {
		return errors.New("queueMax must be non-negative")
	}
	if c.QueueMax > 10000 {
		return errors.New("queueMax must not exceed 10000")
	}
	if c.BodyMaxBytes < 0 {
		return errors.New("bodyMaxBytes must be non-negative")
	}
	if c.BodyMaxBytes > 100<<20 {
		return errors.New("bodyMaxBytes must not exceed 104857600 (100MB)")
	}
	if c.Overflow != "" && c.Overflow != "auto-continue-oldest" && c.Overflow != "drop-new" {
		return errors.New("overflow must be one of: auto-continue-oldest, drop-new")
	}
	return nil
}
