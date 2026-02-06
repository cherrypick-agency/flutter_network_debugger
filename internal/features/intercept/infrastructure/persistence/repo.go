package persistence

import (
	"context"
	"time"

	"network-debugger/internal/features/intercept/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormRuleRepository implements domain.RuleRepository using GORM
type GormRuleRepository struct {
	db *gorm.DB
}

// NewRuleRepository creates a new GormRuleRepository
func NewRuleRepository(db *gorm.DB) *GormRuleRepository {
	return &GormRuleRepository{db: db}
}

var _ domain.RuleRepository = (*GormRuleRepository)(nil)

// SaveAll replaces all rules in a single transaction
func (r *GormRuleRepository) SaveAll(ctx context.Context, rules []domain.InterceptRule) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&InterceptRuleModel{}).Error; err != nil {
			return err
		}
		for _, rule := range rules {
			m := toRuleModel(rule)
			if m.CreatedAt.IsZero() {
				m.CreatedAt = now
			}
			m.UpdatedAt = now
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// List returns all rules ordered by priority ASC
func (r *GormRuleRepository) List(ctx context.Context) ([]domain.InterceptRule, error) {
	var rows []InterceptRuleModel
	if err := r.db.WithContext(ctx).Order("priority ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.InterceptRule, 0, len(rows))
	for _, row := range rows {
		out = append(out, toRuleDomain(row))
	}
	return out, nil
}

// GormConfigRepository implements domain.ConfigRepository using GORM
type GormConfigRepository struct {
	db *gorm.DB
}

// NewConfigRepository creates a new GormConfigRepository
func NewConfigRepository(db *gorm.DB) *GormConfigRepository {
	return &GormConfigRepository{db: db}
}

var _ domain.ConfigRepository = (*GormConfigRepository)(nil)

// Load loads the configuration, returning defaults if not found
func (r *GormConfigRepository) Load(ctx context.Context) (domain.InterceptConfig, error) {
	var m InterceptConfigModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", 1).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			cfg := domain.InterceptConfig{}
			cfg.SetDefaults()
			return cfg, nil
		}
		return domain.InterceptConfig{}, err
	}
	return toConfigDomain(m), nil
}

// Save upserts the configuration with id=1
func (r *GormConfigRepository) Save(ctx context.Context, cfg domain.InterceptConfig) error {
	m := toConfigModel(cfg)
	m.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "requests", "responses", "timeout_ms", "queue_max", "body_max_bytes", "reencode", "overflow", "updated_at"}),
	}).Create(&m).Error
}
