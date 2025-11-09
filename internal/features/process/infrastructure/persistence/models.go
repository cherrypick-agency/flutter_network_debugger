package persistence

import "time"

// ProcessDetectionConfigModel - GORM модель для конфигурации детекции процессов
type ProcessDetectionConfigModel struct {
	ID              int64     `gorm:"primaryKey"`
	Enabled         bool      `gorm:"not null;default:true"`
	UseHelperTool   bool      `gorm:"not null;default:false"`
	HelperInstalled bool      `gorm:"not null;default:false"`
	CacheEnabled    bool      `gorm:"not null;default:true"`
	CacheTTLSeconds int       `gorm:"not null;default:300"`
	FallbackEnabled bool      `gorm:"not null;default:true"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

// TableName - имя таблицы в базе данных
func (ProcessDetectionConfigModel) TableName() string {
	return "process_detection_config"
}

// IconCacheModel - GORM модель для кеша иконок
type IconCacheModel struct {
	CacheKey   string `gorm:"primaryKey"`
	IconFormat string `gorm:"not null"`
	IconData   []byte `gorm:"not null"`
	IconPath   *string
	ExpiresAt  time.Time `gorm:"not null;index:idx_icon_cache_expires"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// TableName - имя таблицы в базе данных
func (IconCacheModel) TableName() string {
	return "icon_cache"
}
