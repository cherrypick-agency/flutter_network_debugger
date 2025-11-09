package domain

import (
	"errors"
	"time"
)

// ProcessInfo - информация о процессе, создавшем сетевое соединение
type ProcessInfo struct {
	PID            int32
	Name           string
	ExecutablePath string
	BundleID       *string // macOS only (e.g., "com.apple.Safari")
	Icon           *AppIcon
	DetectedAt     time.Time
}

// AppIcon - иконка приложения
type AppIcon struct {
	Format string  // "png", "icns", "ico"
	Data   []byte  // binary data
	Path   *string // cached file path (optional)
}

// DetectionConfig - конфигурация детекции процессов
type DetectionConfig struct {
	ID              int64
	Enabled         bool // включена ли детекция
	UseHelperTool   bool // использовать привилегированный helper tool
	HelperInstalled bool // установлен ли helper
	CacheEnabled    bool // кешировать иконки
	CacheTTLSeconds int  // TTL кеша в секундах
	FallbackEnabled bool // показывать "Unknown" при ошибке детекции
	UpdatedAt       time.Time
}

// Validate - валидация конфигурации
func (c *DetectionConfig) Validate() error {
	if c.CacheTTLSeconds < 0 {
		return errors.New("cache TTL must be non-negative")
	}
	if c.CacheTTLSeconds > 86400 {
		return errors.New("cache TTL must be less than 24 hours")
	}
	return nil
}
