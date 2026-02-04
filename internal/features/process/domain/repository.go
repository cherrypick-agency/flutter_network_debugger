package domain

import (
	"context"
	"time"
)

// IConfigRepository - interface for detection configuration persistence
type IConfigRepository interface {
	// Load - load configuration from database
	Load(ctx context.Context) (*DetectionConfig, error)

	// Save - save configuration to database
	Save(ctx context.Context, cfg *DetectionConfig) error
}

// IIconCacheRepository - interface for icon cache
type IIconCacheRepository interface {
	// Get - get icon from cache by key
	Get(key string) (*AppIcon, error)

	// Set - save icon to cache with specified TTL
	Set(key string, icon *AppIcon, ttl time.Duration) error

	// Delete - delete icon from cache
	Delete(key string) error

	// Clear - clear entire cache
	Clear() error

	// CleanupExpired - remove expired entries from cache
	CleanupExpired() error
}
