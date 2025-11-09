package persistence

import (
	"time"

	"network-debugger/internal/features/tags/domain"
)

// PredefinedTagModel is the GORM model for predefined_tags table
type PredefinedTagModel struct {
	ID           string `gorm:"type:text;primaryKey"`
	Name         string `gorm:"type:text;uniqueIndex;not null"`
	Color        string `gorm:"type:text;not null;default:#808080"`
	Category     string `gorm:"type:text;default:general"`
	DisplayOrder int    `gorm:"default:0"`
	CreatedAt    time.Time
}

func (PredefinedTagModel) TableName() string { return "predefined_tags" }

// SessionTagModel is the GORM model for session_tags table
type SessionTagModel struct {
	ID        string `gorm:"type:text;primaryKey"`
	SessionID string `gorm:"type:text;not null;index:idx_session_tags_session_id"`
	TagName   string `gorm:"type:text;not null;index:idx_session_tags_tag_name"`
	CreatedAt time.Time
}

func (SessionTagModel) TableName() string { return "session_tags" }

// SessionAnnotationModel is the GORM model for session_annotations table
type SessionAnnotationModel struct {
	ID        string `gorm:"type:text;primaryKey"`
	SessionID string `gorm:"type:text;not null;index:idx_session_annotations_session_id"`
	Key       string `gorm:"type:text;not null"`
	Value     string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SessionAnnotationModel) TableName() string { return "session_annotations" }

// Converters

func predefinedTagToModel(d domain.PredefinedTag) *PredefinedTagModel {
	return &PredefinedTagModel{
		ID:           d.ID,
		Name:         d.Name,
		Color:        d.Color,
		Category:     d.Category,
		DisplayOrder: d.DisplayOrder,
		CreatedAt:    d.CreatedAt,
	}
}

func predefinedTagToDomain(m *PredefinedTagModel) domain.PredefinedTag {
	return domain.PredefinedTag{
		ID:           m.ID,
		Name:         m.Name,
		Color:        m.Color,
		Category:     m.Category,
		DisplayOrder: m.DisplayOrder,
		CreatedAt:    m.CreatedAt,
	}
}

func sessionTagToModel(d domain.SessionTag) *SessionTagModel {
	return &SessionTagModel{
		ID:        d.ID,
		SessionID: d.SessionID,
		TagName:   d.TagName,
		CreatedAt: d.CreatedAt,
	}
}

func sessionTagToDomain(m *SessionTagModel) domain.SessionTag {
	return domain.SessionTag{
		ID:        m.ID,
		SessionID: m.SessionID,
		TagName:   m.TagName,
		CreatedAt: m.CreatedAt,
	}
}

func sessionAnnotationToModel(d domain.SessionAnnotation) *SessionAnnotationModel {
	return &SessionAnnotationModel{
		ID:        d.ID,
		SessionID: d.SessionID,
		Key:       d.Key,
		Value:     d.Value,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func sessionAnnotationToDomain(m *SessionAnnotationModel) domain.SessionAnnotation {
	return domain.SessionAnnotation{
		ID:        m.ID,
		SessionID: m.SessionID,
		Key:       m.Key,
		Value:     m.Value,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
