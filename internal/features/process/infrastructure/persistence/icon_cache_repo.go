package persistence

import (
	"time"

	"gorm.io/gorm"

	"network-debugger/internal/features/process/domain"
)

// IconCacheRepo - implementation of IIconCacheRepository for GORM/SQLite
type IconCacheRepo struct {
	db *gorm.DB
}

// NewIconCacheRepo - create new icon cache repository
func NewIconCacheRepo(db *gorm.DB) *IconCacheRepo {
	return &IconCacheRepo{db: db}
}

// Get - get icon from cache by key
func (r *IconCacheRepo) Get(key string) (*domain.AppIcon, error) {
	var model IconCacheModel

	err := r.db.Where("cache_key = ? AND expires_at > ?", key, time.Now()).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return &domain.AppIcon{
		Format: model.IconFormat,
		Data:   model.IconData,
		Path:   model.IconPath,
	}, nil
}

// Set - save icon to cache with specified TTL
func (r *IconCacheRepo) Set(key string, icon *domain.AppIcon, ttl time.Duration) error {
	model := IconCacheModel{
		CacheKey:   key,
		IconFormat: icon.Format,
		IconData:   icon.Data,
		IconPath:   icon.Path,
		ExpiresAt:  time.Now().Add(ttl),
	}

	return r.db.Save(&model).Error
}

// Delete - delete icon from cache
func (r *IconCacheRepo) Delete(key string) error {
	return r.db.Delete(&IconCacheModel{}, "cache_key = ?", key).Error
}

// Clear - clear entire cache
func (r *IconCacheRepo) Clear() error {
	return r.db.Exec("DELETE FROM icon_cache").Error
}

// CleanupExpired - remove expired entries from cache
func (r *IconCacheRepo) CleanupExpired() error {
	return r.db.Delete(&IconCacheModel{}, "expires_at < ?", time.Now()).Error
}
