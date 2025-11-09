package persistence

import (
	"context"
	"strings"
	"time"

	"network-debugger/internal/features/tags/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

var _ domain.TagsRepository = (*Repo)(nil)

// Predefined Tags

func (r *Repo) ListPredefinedTags(ctx context.Context) ([]domain.PredefinedTag, error) {
	var rows []PredefinedTagModel
	tx := r.db.WithContext(ctx).Order("display_order ASC, name ASC").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	out := make([]domain.PredefinedTag, 0, len(rows))
	for i := range rows {
		out = append(out, predefinedTagToDomain(&rows[i]))
	}
	return out, nil
}

func (r *Repo) CreatePredefinedTag(ctx context.Context, tag domain.PredefinedTag) error {
	m := predefinedTagToModel(tag)
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(m).Error
}

func (r *Repo) DeletePredefinedTag(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&PredefinedTagModel{}, "id = ?", id).Error
}

// Session Tags

func (r *Repo) GetSessionTags(ctx context.Context, sessionID string) ([]domain.SessionTag, error) {
	var rows []SessionTagModel
	tx := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at DESC").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	out := make([]domain.SessionTag, 0, len(rows))
	for i := range rows {
		out = append(out, sessionTagToDomain(&rows[i]))
	}
	return out, nil
}

func (r *Repo) AddSessionTag(ctx context.Context, tag domain.SessionTag) error {
	m := sessionTagToModel(tag)
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	// Use OnConflict to handle duplicate (session_id, tag_name) pairs
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "tag_name"}},
			DoNothing: true,
		}).
		Create(m).Error
}

func (r *Repo) RemoveSessionTag(ctx context.Context, sessionID, tagName string) error {
	return r.db.WithContext(ctx).
		Where("session_id = ? AND tag_name = ?", sessionID, tagName).
		Delete(&SessionTagModel{}).Error
}

func (r *Repo) BulkAddSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	if len(sessionIDs) == 0 || len(tagNames) == 0 {
		return nil
	}

	// Create all combinations
	now := time.Now().UTC()
	tags := make([]SessionTagModel, 0, len(sessionIDs)*len(tagNames))
	for _, sessionID := range sessionIDs {
		for _, tagName := range tagNames {
			tags = append(tags, SessionTagModel{
				ID:        uuid.NewString(),
				SessionID: sessionID,
				TagName:   tagName,
				CreatedAt: now,
			})
		}
	}

	// Batch insert with conflict handling
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "tag_name"}},
			DoNothing: true,
		}).
		Create(&tags).Error
}

func (r *Repo) BulkRemoveSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	if len(sessionIDs) == 0 || len(tagNames) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("session_id IN ? AND tag_name IN ?", sessionIDs, tagNames).
		Delete(&SessionTagModel{}).Error
}

func (r *Repo) DeleteAllSessionTags(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&SessionTagModel{}).Error
}

func (r *Repo) FindSessionIDsByTags(ctx context.Context, tagNames []string) ([]string, error) {
	if len(tagNames) == 0 {
		return []string{}, nil
	}

	// Sanitize tag names: trim whitespace and filter empty values
	sanitized := make([]string, 0, len(tagNames))
	for _, tag := range tagNames {
		trimmed := strings.TrimSpace(tag)
		if trimmed != "" {
			sanitized = append(sanitized, trimmed)
		}
	}

	if len(sanitized) == 0 {
		return []string{}, nil
	}

	var sessionIDs []string
	tx := r.db.WithContext(ctx).
		Model(&SessionTagModel{}).
		Select("DISTINCT session_id").
		Where("tag_name IN ?", sanitized).
		Pluck("session_id", &sessionIDs)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return sessionIDs, nil
}

// Session Annotations

func (r *Repo) GetSessionAnnotations(ctx context.Context, sessionID string) ([]domain.SessionAnnotation, error) {
	var rows []SessionAnnotationModel
	tx := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("key ASC").Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	out := make([]domain.SessionAnnotation, 0, len(rows))
	for i := range rows {
		out = append(out, sessionAnnotationToDomain(&rows[i]))
	}
	return out, nil
}

func (r *Repo) UpsertSessionAnnotation(ctx context.Context, annotation domain.SessionAnnotation) error {
	m := sessionAnnotationToModel(annotation)
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).
		Create(m).Error
}

func (r *Repo) DeleteSessionAnnotation(ctx context.Context, sessionID, key string) error {
	return r.db.WithContext(ctx).
		Where("session_id = ? AND key = ?", sessionID, key).
		Delete(&SessionAnnotationModel{}).Error
}

func (r *Repo) DeleteAllSessionAnnotations(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Delete(&SessionAnnotationModel{}).Error
}
