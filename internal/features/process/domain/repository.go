package domain

import (
	"context"
	"time"
)

// IConfigRepository - интерфейс для персистентности конфигурации детекции
type IConfigRepository interface {
	// Load - загрузить конфигурацию из базы данных
	Load(ctx context.Context) (*DetectionConfig, error)

	// Save - сохранить конфигурацию в базу данных
	Save(ctx context.Context, cfg *DetectionConfig) error
}

// IIconCacheRepository - интерфейс для кеша иконок
type IIconCacheRepository interface {
	// Get - получить иконку из кеша по ключу
	Get(key string) (*AppIcon, error)

	// Set - сохранить иконку в кеш с указанным TTL
	Set(key string, icon *AppIcon, ttl time.Duration) error

	// Delete - удалить иконку из кеша
	Delete(key string) error

	// Clear - очистить весь кеш
	Clear() error

	// CleanupExpired - удалить истекшие записи из кеша
	CleanupExpired() error
}
