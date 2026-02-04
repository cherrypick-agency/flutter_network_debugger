package domain

import "time"

// PatternType describes how paths/hosts are matched
type PatternType string

const (
	PatternGlob  PatternType = "glob"
	PatternRegex PatternType = "regex"
)

// Kind defines the rule type
type Kind string

const (
	KindLocal  Kind = "local"
	KindRemote Kind = "remote"
)

// MapRule — domain entity for mapping rule
type MapRule struct {
	ID             string
	Enabled        bool
	Priority       int
	Kind           Kind
	StopProcessing bool

	// Conditions
	Methods     []string
	HostPattern string
	PathPattern string
	PatternType PatternType

	// Local options
	FilePath            *string
	BlobPath            *string
	StatusOverride      int
	ContentTypeOverride string

	// Remote options
	TargetURLTemplate string
	PreserveHost      bool

	CreatedAt time.Time
	UpdatedAt time.Time
}
