package persistence

import (
	"encoding/json"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"

	"gorm.io/gorm"
)

// MapRuleModel — GORM model for map_rules table
type MapRuleModel struct {
	ID             string `gorm:"type:text;primaryKey"`
	Enabled        bool
	Priority       int
	Kind           string `gorm:"type:text"`
	StopProcessing bool

	MethodsJSON string `gorm:"type:text"` // JSON array of methods
	HostPattern string `gorm:"type:text"`
	PathPattern string `gorm:"type:text"`
	PatternType string `gorm:"type:text"`

	// Local
	FilePath            *string `gorm:"type:text"`
	BlobPath            *string `gorm:"type:text"`
	StatusOverride      int
	ContentTypeOverride string `gorm:"type:text"`

	// Remote
	TargetURLTemplate string `gorm:"type:text"`
	PreserveHost      bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (MapRuleModel) TableName() string { return "map_rules" }

func toModel(d mdomain.MapRule) *MapRuleModel {
	methods := "[]"
	if len(d.Methods) > 0 {
		if b, err := json.Marshal(d.Methods); err == nil {
			methods = string(b)
		}
	}
	kind := string(d.Kind)
	if kind == "" {
		kind = string(mdomain.KindRemote)
	}
	return &MapRuleModel{
		ID:                  d.ID,
		Enabled:             d.Enabled,
		Priority:            d.Priority,
		Kind:                kind,
		StopProcessing:      d.StopProcessing,
		MethodsJSON:         methods,
		HostPattern:         d.HostPattern,
		PathPattern:         d.PathPattern,
		PatternType:         string(d.PatternType),
		FilePath:            d.FilePath,
		BlobPath:            d.BlobPath,
		StatusOverride:      d.StatusOverride,
		ContentTypeOverride: d.ContentTypeOverride,
		TargetURLTemplate:   d.TargetURLTemplate,
		PreserveHost:        d.PreserveHost,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
}

func toDomain(m *MapRuleModel) mdomain.MapRule {
	var methods []string
	if m.MethodsJSON != "" {
		_ = json.Unmarshal([]byte(m.MethodsJSON), &methods)
	}
	return mdomain.MapRule{
		ID:                  m.ID,
		Enabled:             m.Enabled,
		Priority:            m.Priority,
		Kind:                mdomain.Kind(m.Kind),
		StopProcessing:      m.StopProcessing,
		Methods:             methods,
		HostPattern:         m.HostPattern,
		PathPattern:         m.PathPattern,
		PatternType:         mdomain.PatternType(m.PatternType),
		FilePath:            m.FilePath,
		BlobPath:            m.BlobPath,
		StatusOverride:      m.StatusOverride,
		ContentTypeOverride: m.ContentTypeOverride,
		TargetURLTemplate:   m.TargetURLTemplate,
		PreserveHost:        m.PreserveHost,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
