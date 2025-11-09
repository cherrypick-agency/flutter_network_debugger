package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"network-debugger/internal/features/tags/domain"
)

// Service provides use cases for tags and annotations
type Service struct {
	repo domain.TagsRepository
}

// NewService creates a new tags service
func NewService(repo domain.TagsRepository) *Service {
	return &Service{repo: repo}
}

// Predefined Tags

// ListPredefinedTags returns all predefined tags
func (s *Service) ListPredefinedTags(ctx context.Context) ([]domain.PredefinedTag, error) {
	return s.repo.ListPredefinedTags(ctx)
}

// CreatePredefinedTag creates a new predefined tag
func (s *Service) CreatePredefinedTag(ctx context.Context, name, color, category string, displayOrder int) error {
	tag := domain.PredefinedTag{
		ID:           uuid.NewString(),
		Name:         name,
		Color:        color,
		Category:     category,
		DisplayOrder: displayOrder,
	}
	return s.repo.CreatePredefinedTag(ctx, tag)
}

// DeletePredefinedTag deletes a predefined tag
func (s *Service) DeletePredefinedTag(ctx context.Context, id string) error {
	return s.repo.DeletePredefinedTag(ctx, id)
}

// Session Tags

// GetSessionTags returns all tags for a session
func (s *Service) GetSessionTags(ctx context.Context, sessionID string) ([]domain.SessionTag, error) {
	return s.repo.GetSessionTags(ctx, sessionID)
}

// AddTagToSession adds a tag to a session
func (s *Service) AddTagToSession(ctx context.Context, sessionID, tagName string) error {
	if sessionID == "" || tagName == "" {
		return fmt.Errorf("sessionID and tagName are required")
	}

	tag := domain.SessionTag{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		TagName:   tagName,
	}
	return s.repo.AddSessionTag(ctx, tag)
}

// RemoveTagFromSession removes a tag from a session
func (s *Service) RemoveTagFromSession(ctx context.Context, sessionID, tagName string) error {
	if sessionID == "" || tagName == "" {
		return fmt.Errorf("sessionID and tagName are required")
	}
	return s.repo.RemoveSessionTag(ctx, sessionID, tagName)
}

// BulkAddTags adds tags to multiple sessions
func (s *Service) BulkAddTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	if len(sessionIDs) == 0 || len(tagNames) == 0 {
		return fmt.Errorf("sessionIDs and tagNames cannot be empty")
	}
	return s.repo.BulkAddSessionTags(ctx, sessionIDs, tagNames)
}

// BulkRemoveTags removes tags from multiple sessions
func (s *Service) BulkRemoveTags(ctx context.Context, sessionIDs []string, tagNames []string) error {
	if len(sessionIDs) == 0 || len(tagNames) == 0 {
		return fmt.Errorf("sessionIDs and tagNames cannot be empty")
	}
	return s.repo.BulkRemoveSessionTags(ctx, sessionIDs, tagNames)
}

// DeleteAllSessionTags removes all tags from a session
func (s *Service) DeleteAllSessionTags(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}
	return s.repo.DeleteAllSessionTags(ctx, sessionID)
}

// FindSessionIDsByTags finds sessions by tags
func (s *Service) FindSessionIDsByTags(ctx context.Context, tagNames []string) ([]string, error) {
	if len(tagNames) == 0 {
		return []string{}, nil
	}
	return s.repo.FindSessionIDsByTags(ctx, tagNames)
}

// Session Annotations

// GetSessionAnnotations returns all annotations for a session
func (s *Service) GetSessionAnnotations(ctx context.Context, sessionID string) ([]domain.SessionAnnotation, error) {
	return s.repo.GetSessionAnnotations(ctx, sessionID)
}

// UpsertAnnotation creates or updates a session annotation
func (s *Service) UpsertAnnotation(ctx context.Context, sessionID, key, value string) error {
	if sessionID == "" || key == "" {
		return fmt.Errorf("sessionID and key are required")
	}

	annotation := domain.SessionAnnotation{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		Key:       key,
		Value:     value,
	}
	return s.repo.UpsertSessionAnnotation(ctx, annotation)
}

// DeleteAnnotation deletes a session annotation
func (s *Service) DeleteAnnotation(ctx context.Context, sessionID, key string) error {
	if sessionID == "" || key == "" {
		return fmt.Errorf("sessionID and key are required")
	}
	return s.repo.DeleteSessionAnnotation(ctx, sessionID, key)
}

// DeleteAllAnnotations deletes all annotations for a session
func (s *Service) DeleteAllAnnotations(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("sessionID is required")
	}
	return s.repo.DeleteAllSessionAnnotations(ctx, sessionID)
}
