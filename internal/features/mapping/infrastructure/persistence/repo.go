package persistence

import (
	"context"
	"errors"
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

func (r *Repo) GetByID(ctx context.Context, id string) (*mdomain.MapRule, error) {
	var m MapRuleModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, mdomain.RuleNotFoundError{ID: id}
		}
		return nil, err
	}
	d := toDomain(&m)
	return &d, nil
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

	// Important: created_at must not be overwritten on updates (otherwise any edit "recreates" the record).
	onConflict := clause.OnConflict{
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"enabled",
			"priority",
			"kind",
			"stop_processing",
			"methods_json",
			"host_pattern",
			"path_pattern",
			"pattern_type",
			"file_path",
			"blob_path",
			"status_override",
			"content_type_override",
			"target_url_template",
			"preserve_host",
			"updated_at",
			"deleted_at",
		}),
	}

	if err := r.db.WithContext(ctx).Clauses(onConflict).Create(m).Error; err != nil {
		return mdomain.MapRule{}, err
	}
	// re-read
	var saved MapRuleModel
	if err := r.db.WithContext(ctx).First(&saved, "id = ?", m.ID).Error; err != nil {
		return mdomain.MapRule{}, err
	}
	return toDomain(&saved), nil
}

func (r *Repo) Delete(ctx context.Context, rid string) error {
	return r.db.WithContext(ctx).Delete(&MapRuleModel{ID: rid}).Error
}

func (r *Repo) Reorder(ctx context.Context, ids []string) (err error) {
	if len(ids) == 0 {
		return nil
	}
	// set priorities 1..N in the given order
	now := time.Now().UTC()
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for i, idv := range ids {
		prio := i + 1
		res := tx.Model(&MapRuleModel{}).Where("id = ?", idv).Updates(map[string]any{
			"priority":   prio,
			"updated_at": now,
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return mdomain.RuleNotFoundError{ID: idv}
		}
	}

	return tx.Commit().Error
}
