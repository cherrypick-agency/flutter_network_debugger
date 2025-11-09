package domain

import "context"

// TagsRepository defines the interface for tags and annotations persistence
type TagsRepository interface {
	// Predefined Tags
	ListPredefinedTags(ctx context.Context) ([]PredefinedTag, error)
	CreatePredefinedTag(ctx context.Context, tag PredefinedTag) error
	DeletePredefinedTag(ctx context.Context, id string) error

	// Session Tags
	GetSessionTags(ctx context.Context, sessionID string) ([]SessionTag, error)
	AddSessionTag(ctx context.Context, tag SessionTag) error
	RemoveSessionTag(ctx context.Context, sessionID, tagName string) error
	BulkAddSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error
	BulkRemoveSessionTags(ctx context.Context, sessionIDs []string, tagNames []string) error
	DeleteAllSessionTags(ctx context.Context, sessionID string) error

	// Find sessions by tags
	FindSessionIDsByTags(ctx context.Context, tagNames []string) ([]string, error)

	// Session Annotations
	GetSessionAnnotations(ctx context.Context, sessionID string) ([]SessionAnnotation, error)
	UpsertSessionAnnotation(ctx context.Context, annotation SessionAnnotation) error
	DeleteSessionAnnotation(ctx context.Context, sessionID, key string) error
	DeleteAllSessionAnnotations(ctx context.Context, sessionID string) error
}
