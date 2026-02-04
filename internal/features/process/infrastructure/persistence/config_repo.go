package persistence

import (
	"context"

	"gorm.io/gorm"

	"network-debugger/internal/features/process/domain"
)

// ConfigRepo - implementation of IConfigRepository for GORM/SQLite
type ConfigRepo struct {
	db *gorm.DB
}

// NewConfigRepo - create new configuration repository
func NewConfigRepo(db *gorm.DB) *ConfigRepo {
	return &ConfigRepo{db: db}
}

// Load - load configuration from database (singleton with ID=1)
func (r *ConfigRepo) Load(ctx context.Context) (*domain.DetectionConfig, error) {
	var model ProcessDetectionConfigModel

	err := r.db.WithContext(ctx).First(&model, 1).Error
	if err != nil {
		// If record not found, create default
		if err == gorm.ErrRecordNotFound {
			defaultModel := ProcessDetectionConfigModel{
				ID:              1,
				Enabled:         true,
				UseHelperTool:   false,
				HelperInstalled: false,
				CacheEnabled:    true,
				CacheTTLSeconds: 300,
				FallbackEnabled: true,
			}
			if createErr := r.db.WithContext(ctx).Create(&defaultModel).Error; createErr != nil {
				return nil, createErr
			}
			return toDomainConfig(defaultModel), nil
		}
		return nil, err
	}

	return toDomainConfig(model), nil
}

// Save - save configuration to database
func (r *ConfigRepo) Save(ctx context.Context, cfg *domain.DetectionConfig) error {
	model := toModel(cfg)
	model.ID = 1 // singleton

	return r.db.WithContext(ctx).Save(&model).Error
}

// toDomainConfig - convert GORM model to domain entity
func toDomainConfig(m ProcessDetectionConfigModel) *domain.DetectionConfig {
	return &domain.DetectionConfig{
		ID:              m.ID,
		Enabled:         m.Enabled,
		UseHelperTool:   m.UseHelperTool,
		HelperInstalled: m.HelperInstalled,
		CacheEnabled:    m.CacheEnabled,
		CacheTTLSeconds: m.CacheTTLSeconds,
		FallbackEnabled: m.FallbackEnabled,
		UpdatedAt:       m.UpdatedAt,
	}
}

// toModel - convert domain entity to GORM model
func toModel(c *domain.DetectionConfig) ProcessDetectionConfigModel {
	return ProcessDetectionConfigModel{
		ID:              c.ID,
		Enabled:         c.Enabled,
		UseHelperTool:   c.UseHelperTool,
		HelperInstalled: c.HelperInstalled,
		CacheEnabled:    c.CacheEnabled,
		CacheTTLSeconds: c.CacheTTLSeconds,
		FallbackEnabled: c.FallbackEnabled,
	}
}
