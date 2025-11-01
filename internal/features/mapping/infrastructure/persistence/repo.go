package persistence

import (
	"context"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
	"network-debugger/pkg/shared/id"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

var _ mdomain.MapRulesRepository = (*Repo)(nil)

func (r *Repo) List(ctx context.Context) ([]mdomain.MapRule, error) {
	var rows []MapRuleModel
	tx := r.db.WithContext(ctx).Order("priority ASC, updated_at ASC").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}
	out := make([]mdomain.MapRule, 0, len(rows))
	for i := range rows {
		out = append(out, toDomain(&rows[i]))
	}
	return out, nil
}

func (r *Repo) Upsert(ctx context.Context, d mdomain.MapRule) (mdomain.MapRule, error) {
	m := toModel(d)
	if m.ID == "" {
		m.ID = id.New()
		if m.CreatedAt.IsZero() {
			m.CreatedAt = time.Now().UTC()
		}
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now().UTC()
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(m).Error; err != nil {
		return mdomain.MapRule{}, err
	}
	// перечитаем
	var saved MapRuleModel
	if err := r.db.WithContext(ctx).First(&saved, "id = ?", m.ID).Error; err != nil {
		return mdomain.MapRule{}, err
	}
	return toDomain(&saved), nil
}

func (r *Repo) Delete(ctx context.Context, rid string) error {
	return r.db.WithContext(ctx).Delete(&MapRuleModel{ID: rid}).Error
}

func (r *Repo) Reorder(ctx context.Context, ids []string) error {
	// проставим приоритеты с 1..N в переданном порядке
	now := time.Now().UTC()
	for i, idv := range ids {
		prio := i + 1
		if err := r.db.WithContext(ctx).Model(&MapRuleModel{}).Where("id = ?", idv).Updates(map[string]any{
			"priority":   prio,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
