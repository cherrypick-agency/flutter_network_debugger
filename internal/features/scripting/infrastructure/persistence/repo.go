package persistence

import (
	"context"

	"gorm.io/gorm"

	"network-debugger/internal/features/scripting/domain"
)

// GormScriptRepository implements domain.ScriptRepository using GORM
type GormScriptRepository struct {
	db *gorm.DB
}

// NewGormScriptRepository creates a new GORM-based script repository
func NewGormScriptRepository(db *gorm.DB) *GormScriptRepository {
	return &GormScriptRepository{db: db}
}

// Save creates or updates a script
func (r *GormScriptRepository) Save(ctx context.Context, script *domain.Script) error {
	model := FromDomain(script)
	return r.db.WithContext(ctx).Save(model).Error
}

// Get retrieves a script by ID
func (r *GormScriptRepository) Get(ctx context.Context, id string) (*domain.Script, error) {
	var model ScriptModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

// List retrieves scripts matching the filter
func (r *GormScriptRepository) List(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
	var models []ScriptModel

	query := r.db.WithContext(ctx).Order("priority DESC, created_at ASC")

	// Apply filters
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.Runtime != "" {
		query = query.Where("runtime = ?", filter.Runtime)
	}
	if filter.TriggerType != "" {
		query = query.Where("trigger_type = ? OR trigger_type = ?", filter.TriggerType, domain.TriggerBoth)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	scripts := make([]*domain.Script, len(models))
	for i, m := range models {
		scripts[i] = m.ToDomain()
	}
	return scripts, nil
}

// Delete removes a script
func (r *GormScriptRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&ScriptModel{}, "id = ?", id).Error
}

// UpdateEnabled toggles script enabled state
func (r *GormScriptRepository) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	return r.db.WithContext(ctx).
		Model(&ScriptModel{}).
		Where("id = ?", id).
		Update("enabled", enabled).Error
}
