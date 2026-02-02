package persistence

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// ScriptModel is the GORM model for scripts table
type ScriptModel struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string

	// Runtime configuration
	Runtime  string `gorm:"not null"`
	Code     []byte `gorm:"type:blob"` // Nullable for scripts with only source code
	Language string `gorm:"not null"`

	// Compilation (NEW)
	SourceCode        string    `gorm:"type:text"`
	Dependencies      StringMap `gorm:"type:text"` // JSON: {"filename": "content"}
	CompilationStatus string    `gorm:"default:not_compiled"`
	CompilationError  string    `gorm:"type:text"`
	LastCompiledAt    *time.Time

	// Trigger configuration
	TriggerType string `gorm:"not null"`
	Priority    int    `gorm:"default:0"`

	// Match rules
	MatchMethods     StringArray `gorm:"type:text"`
	MatchHostPattern string
	MatchPathPattern string
	PatternType      string

	// Resource limits
	TimeoutMs     int         `gorm:"default:5000"`
	MemoryLimitMB int         `gorm:"default:10"`
	AllowedHosts  StringArray `gorm:"type:text"`

	// State
	Enabled bool `gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// StringArray is a custom type for JSON arrays of strings
type StringArray []string

// Value implements the driver.Valuer interface for StringArray
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for StringArray
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*s = []string{}
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// StringMap is a custom type for JSON maps of string→string
type StringMap map[string]string

// Value implements the driver.Valuer interface for StringMap
func (s StringMap) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for StringMap
func (s *StringMap) Scan(value interface{}) error {
	if value == nil {
		*s = make(map[string]string)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		*s = make(map[string]string)
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// TableName specifies the table name for GORM
func (ScriptModel) TableName() string {
	return "scripts"
}

// ToDomain converts GORM model to domain entity
func (m *ScriptModel) ToDomain() *domain.Script {
	return &domain.Script{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Runtime:     domain.ScriptRuntime(m.Runtime),
		Code:        m.Code,
		Language:    m.Language,
		// Compilation fields
		SourceCode:        m.SourceCode,
		Dependencies:      map[string]string(m.Dependencies),
		CompilationStatus: domain.CompilationStatus(m.CompilationStatus),
		CompilationError:  m.CompilationError,
		LastCompiledAt:    m.LastCompiledAt,
		// Trigger configuration
		TriggerType: domain.TriggerType(m.TriggerType),
		MatchRules: domain.MatchRules{
			Methods:     m.MatchMethods,
			HostPattern: m.MatchHostPattern,
			PathPattern: m.MatchPathPattern,
			PatternType: domain.PatternType(m.PatternType),
		},
		Priority: m.Priority,
		Config: domain.ScriptConfig{
			TimeoutMs:     m.TimeoutMs,
			MemoryLimitMB: m.MemoryLimitMB,
			AllowedHosts:  m.AllowedHosts,
		},
		Enabled:   m.Enabled,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromDomain converts domain entity to GORM model
func FromDomain(s *domain.Script) *ScriptModel {
	return &ScriptModel{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Runtime:     string(s.Runtime),
		Code:        s.Code,
		Language:    s.Language,
		// Compilation fields
		SourceCode:        s.SourceCode,
		Dependencies:      StringMap(s.Dependencies),
		CompilationStatus: string(s.CompilationStatus),
		CompilationError:  s.CompilationError,
		LastCompiledAt:    s.LastCompiledAt,
		// Trigger configuration
		TriggerType:      string(s.TriggerType),
		Priority:         s.Priority,
		MatchMethods:     s.MatchRules.Methods,
		MatchHostPattern: s.MatchRules.HostPattern,
		MatchPathPattern: s.MatchRules.PathPattern,
		PatternType:      string(s.MatchRules.PatternType),
		TimeoutMs:        s.Config.TimeoutMs,
		MemoryLimitMB:    s.Config.MemoryLimitMB,
		AllowedHosts:     s.Config.AllowedHosts,
		Enabled:          s.Enabled,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
