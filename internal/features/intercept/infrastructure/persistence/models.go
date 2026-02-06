package persistence

import (
	"encoding/json"
	"log"
	"time"

	"network-debugger/internal/features/intercept/domain"
)

// InterceptRuleModel -- GORM model for intercept_rules table
type InterceptRuleModel struct {
	ID             string `gorm:"type:text;primaryKey"`
	Enabled        bool
	Priority       int
	Action         string `gorm:"type:text"`
	Once           bool
	StopProcessing bool
	WhenJSON       string `gorm:"column:when_json;type:text"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TableName returns the table name for GORM
func (InterceptRuleModel) TableName() string { return "intercept_rules" }

// InterceptConfigModel -- GORM model for intercept_config table
type InterceptConfigModel struct {
	ID           uint `gorm:"primaryKey"`
	Enabled      bool
	Requests     bool
	Responses    bool
	TimeoutMs    int
	QueueMax     int
	BodyMaxBytes int
	Reencode     bool
	Overflow     string `gorm:"type:text"`
	UpdatedAt    time.Time
}

// TableName returns the table name for GORM
func (InterceptConfigModel) TableName() string { return "intercept_config" }

func toRuleModel(r domain.InterceptRule) InterceptRuleModel {
	whenJSON := "{}"
	if b, err := json.Marshal(r.When); err == nil {
		whenJSON = string(b)
	}
	return InterceptRuleModel{
		ID:             r.ID,
		Enabled:        r.Enabled,
		Priority:       r.Priority,
		Action:         r.Action,
		Once:           r.Once,
		StopProcessing: r.StopProcessing,
		WhenJSON:       whenJSON,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func toRuleDomain(m InterceptRuleModel) domain.InterceptRule {
	var when domain.InterceptWhen
	if m.WhenJSON != "" {
		if err := json.Unmarshal([]byte(m.WhenJSON), &when); err != nil {
			log.Printf("[InterceptPersistence] corrupt when_json for rule %s: %v", m.ID, err)
		}
	}
	return domain.InterceptRule{
		ID:             m.ID,
		Enabled:        m.Enabled,
		Priority:       m.Priority,
		Action:         m.Action,
		Once:           m.Once,
		StopProcessing: m.StopProcessing,
		When:           when,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func toConfigModel(c domain.InterceptConfig) InterceptConfigModel {
	return InterceptConfigModel{
		ID:           1,
		Enabled:      c.Enabled,
		Requests:     c.Requests,
		Responses:    c.Responses,
		TimeoutMs:    c.TimeoutMs,
		QueueMax:     c.QueueMax,
		BodyMaxBytes: c.BodyMaxBytes,
		Reencode:     c.Reencode,
		Overflow:     c.Overflow,
	}
}

func toConfigDomain(m InterceptConfigModel) domain.InterceptConfig {
	return domain.InterceptConfig{
		Enabled:      m.Enabled,
		Requests:     m.Requests,
		Responses:    m.Responses,
		TimeoutMs:    m.TimeoutMs,
		QueueMax:     m.QueueMax,
		BodyMaxBytes: m.BodyMaxBytes,
		Reencode:     m.Reencode,
		Overflow:     m.Overflow,
	}
}
