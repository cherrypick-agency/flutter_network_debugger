package domain

import "time"

// PredefinedTag represents a predefined tag template with styling
type PredefinedTag struct {
	ID           string
	Name         string
	Color        string
	Category     string
	DisplayOrder int
	CreatedAt    time.Time
}

// SessionTag represents a tag attached to a session
type SessionTag struct {
	ID        string
	SessionID string
	TagName   string
	CreatedAt time.Time
}

// SessionAnnotation represents a key-value annotation attached to a session
type SessionAnnotation struct {
	ID        string
	SessionID string
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BulkTagOperation represents a bulk operation on tags
type BulkTagOperation struct {
	Operation  string // "add" or "remove"
	SessionIDs []string
	TagNames   []string
}
