package domain

import "time"

// PatternType описывает способ сопоставления путей/хостов
type PatternType string

const (
	PatternGlob  PatternType = "glob"
	PatternRegex PatternType = "regex"
)

// Kind определяет тип правила
type Kind string

const (
	KindLocal  Kind = "local"
	KindRemote Kind = "remote"
)

// MapRule — доменная сущность правила маппинга
type MapRule struct {
	ID             string
	Enabled        bool
	Priority       int
	Kind           Kind
	StopProcessing bool

	// Условия
	Methods     []string
	HostPattern string
	PathPattern string
	PatternType PatternType

	// Local‑опции
	FilePath            *string
	BlobPath            *string
	StatusOverride      int
	ContentTypeOverride string

	// Remote‑опции
	TargetURLTemplate string
	PreserveHost      bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
