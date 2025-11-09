package usecase

import (
	"context"
	"errors"
	"testing"

	"network-debugger/internal/features/tags/domain"
)

// mockTagsRepository is a mock implementation of domain.TagsRepository
type mockTagsRepository struct {
	predefinedTags     []domain.PredefinedTag
	sessionTags        map[string][]domain.SessionTag
	sessionAnnotations map[string][]domain.SessionAnnotation
	errCreate          error
	errGet             error
}

func newMockRepo() *mockTagsRepository {
	return &mockTagsRepository{
		predefinedTags:     []domain.PredefinedTag{},
		sessionTags:        make(map[string][]domain.SessionTag),
		sessionAnnotations: make(map[string][]domain.SessionAnnotation),
	}
}

func (m *mockTagsRepository) ListPredefinedTags(_ context.Context) ([]domain.PredefinedTag, error) {
	return m.predefinedTags, m.errGet
}

func (m *mockTagsRepository) CreatePredefinedTag(_ context.Context, tag domain.PredefinedTag) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	m.predefinedTags = append(m.predefinedTags, tag)
	return nil
}

func (m *mockTagsRepository) DeletePredefinedTag(_ context.Context, id string) error {
	for i, tag := range m.predefinedTags {
		if tag.ID == id {
			m.predefinedTags = append(m.predefinedTags[:i], m.predefinedTags[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockTagsRepository) GetSessionTags(_ context.Context, sessionID string) ([]domain.SessionTag, error) {
	tags, ok := m.sessionTags[sessionID]
	if !ok {
		return []domain.SessionTag{}, nil
	}
	return tags, m.errGet
}

func (m *mockTagsRepository) AddSessionTag(_ context.Context, tag domain.SessionTag) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	m.sessionTags[tag.SessionID] = append(m.sessionTags[tag.SessionID], tag)
	return nil
}

func (m *mockTagsRepository) RemoveSessionTag(_ context.Context, sessionID, tagName string) error {
	tags := m.sessionTags[sessionID]
	for i, tag := range tags {
		if tag.TagName == tagName {
			m.sessionTags[sessionID] = append(tags[:i], tags[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockTagsRepository) BulkAddSessionTags(_ context.Context, sessionIDs, tagNames []string) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	for _, sessionID := range sessionIDs {
		for _, tagName := range tagNames {
			tag := domain.SessionTag{
				ID:        "mock-id",
				SessionID: sessionID,
				TagName:   tagName,
			}
			m.sessionTags[sessionID] = append(m.sessionTags[sessionID], tag)
		}
	}
	return nil
}

func (m *mockTagsRepository) BulkRemoveSessionTags(_ context.Context, sessionIDs, tagNames []string) error {
	for _, sessionID := range sessionIDs {
		for _, tagName := range tagNames {
			tags := m.sessionTags[sessionID]
			for i, tag := range tags {
				if tag.TagName == tagName {
					m.sessionTags[sessionID] = append(tags[:i], tags[i+1:]...)
					break
				}
			}
		}
	}
	return nil
}

func (m *mockTagsRepository) DeleteAllSessionTags(_ context.Context, sessionID string) error {
	delete(m.sessionTags, sessionID)
	return nil
}

func (m *mockTagsRepository) FindSessionIDsByTags(_ context.Context, tagNames []string) ([]string, error) {
	sessionIDSet := make(map[string]bool)
	for sessionID, tags := range m.sessionTags {
		for _, tag := range tags {
			for _, targetTag := range tagNames {
				if tag.TagName == targetTag {
					sessionIDSet[sessionID] = true
					break
				}
			}
		}
	}
	sessionIDs := make([]string, 0, len(sessionIDSet))
	for sessionID := range sessionIDSet {
		sessionIDs = append(sessionIDs, sessionID)
	}
	return sessionIDs, nil
}

func (m *mockTagsRepository) GetSessionAnnotations(_ context.Context, sessionID string) ([]domain.SessionAnnotation, error) {
	annotations, ok := m.sessionAnnotations[sessionID]
	if !ok {
		return []domain.SessionAnnotation{}, nil
	}
	return annotations, m.errGet
}

func (m *mockTagsRepository) UpsertSessionAnnotation(_ context.Context, annotation domain.SessionAnnotation) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	annotations := m.sessionAnnotations[annotation.SessionID]
	found := false
	for i, ann := range annotations {
		if ann.Key == annotation.Key {
			annotations[i] = annotation
			found = true
			break
		}
	}
	if !found {
		m.sessionAnnotations[annotation.SessionID] = append(annotations, annotation)
	}
	return nil
}

func (m *mockTagsRepository) DeleteSessionAnnotation(_ context.Context, sessionID, key string) error {
	annotations := m.sessionAnnotations[sessionID]
	for i, ann := range annotations {
		if ann.Key == key {
			m.sessionAnnotations[sessionID] = append(annotations[:i], annotations[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockTagsRepository) DeleteAllSessionAnnotations(_ context.Context, sessionID string) error {
	delete(m.sessionAnnotations, sessionID)
	return nil
}

// Tests

func TestCreatePredefinedTag(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	err := svc.CreatePredefinedTag(ctx, "Bug", "#ff0000", "issue", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(repo.predefinedTags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(repo.predefinedTags))
	}

	tag := repo.predefinedTags[0]
	if tag.Name != "Bug" || tag.Color != "#ff0000" || tag.Category != "issue" {
		t.Errorf("tag fields mismatch: %+v", tag)
	}
}

func TestCreatePredefinedTag_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.errCreate = errors.New("db error")
	svc := NewService(repo)
	ctx := context.Background()

	err := svc.CreatePredefinedTag(ctx, "Bug", "#ff0000", "issue", 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAddTagToSession(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	err := svc.AddTagToSession(ctx, "session-1", "important")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tags := repo.sessionTags["session-1"]
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(tags))
	}

	if tags[0].TagName != "important" {
		t.Errorf("expected tag name 'important', got %s", tags[0].TagName)
	}
}

func TestAddTagToSession_Validation(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name      string
		sessionID string
		tagName   string
	}{
		{"empty sessionID", "", "tag"},
		{"empty tagName", "session-1", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AddTagToSession(ctx, tt.sessionID, tt.tagName)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestRemoveTagFromSession(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Add tag first
	_ = svc.AddTagToSession(ctx, "session-1", "important")

	// Remove it
	err := svc.RemoveTagFromSession(ctx, "session-1", "important")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tags := repo.sessionTags["session-1"]
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after removal, got %d", len(tags))
	}
}

func TestBulkAddTags(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	sessionIDs := []string{"s1", "s2", "s3"}
	tagNames := []string{"bug", "critical"}

	err := svc.BulkAddTags(ctx, sessionIDs, tagNames)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check each session has both tags
	for _, sessionID := range sessionIDs {
		tags := repo.sessionTags[sessionID]
		if len(tags) != 2 {
			t.Errorf("session %s: expected 2 tags, got %d", sessionID, len(tags))
		}
	}
}

func TestBulkAddTags_Validation(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name       string
		sessionIDs []string
		tagNames   []string
	}{
		{"empty sessionIDs", []string{}, []string{"tag"}},
		{"empty tagNames", []string{"s1"}, []string{}},
		{"both empty", []string{}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.BulkAddTags(ctx, tt.sessionIDs, tt.tagNames)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestBulkRemoveTags(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Add tags first
	sessionIDs := []string{"s1", "s2"}
	tagNames := []string{"bug", "critical"}
	_ = svc.BulkAddTags(ctx, sessionIDs, tagNames)

	// Remove them
	err := svc.BulkRemoveTags(ctx, sessionIDs, tagNames)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify removal
	for _, sessionID := range sessionIDs {
		tags := repo.sessionTags[sessionID]
		if len(tags) != 0 {
			t.Errorf("session %s: expected 0 tags after removal, got %d", sessionID, len(tags))
		}
	}
}

func TestFindSessionIDsByTags(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Add tags to sessions
	_ = svc.AddTagToSession(ctx, "s1", "bug")
	_ = svc.AddTagToSession(ctx, "s2", "bug")
	_ = svc.AddTagToSession(ctx, "s3", "feature")

	// Find sessions with "bug" tag
	sessionIDs, err := svc.FindSessionIDsByTags(ctx, []string{"bug"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sessionIDs) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessionIDs))
	}
}

func TestFindSessionIDsByTags_EmptyTags(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	sessionIDs, err := svc.FindSessionIDsByTags(ctx, []string{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(sessionIDs) != 0 {
		t.Errorf("expected empty result, got %d sessions", len(sessionIDs))
	}
}

func TestUpsertAnnotation(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Create annotation
	err := svc.UpsertAnnotation(ctx, "session-1", "env", "production")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	annotations := repo.sessionAnnotations["session-1"]
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Key != "env" || annotations[0].Value != "production" {
		t.Errorf("annotation mismatch: %+v", annotations[0])
	}

	// Update annotation
	err = svc.UpsertAnnotation(ctx, "session-1", "env", "staging")
	if err != nil {
		t.Fatalf("expected no error on update, got %v", err)
	}

	annotations = repo.sessionAnnotations["session-1"]
	if len(annotations) != 1 {
		t.Errorf("expected 1 annotation after update, got %d", len(annotations))
	}

	if annotations[0].Value != "staging" {
		t.Errorf("expected updated value 'staging', got %s", annotations[0].Value)
	}
}

func TestUpsertAnnotation_Validation(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	tests := []struct {
		name      string
		sessionID string
		key       string
	}{
		{"empty sessionID", "", "key"},
		{"empty key", "session-1", ""},
		{"both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpsertAnnotation(ctx, tt.sessionID, tt.key, "value")
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestDeleteAnnotation(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Create annotation
	_ = svc.UpsertAnnotation(ctx, "session-1", "env", "production")

	// Delete it
	err := svc.DeleteAnnotation(ctx, "session-1", "env")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	annotations := repo.sessionAnnotations["session-1"]
	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations after deletion, got %d", len(annotations))
	}
}

func TestDeleteAllAnnotations(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	// Create multiple annotations
	_ = svc.UpsertAnnotation(ctx, "session-1", "env", "production")
	_ = svc.UpsertAnnotation(ctx, "session-1", "region", "us-east-1")

	// Delete all
	err := svc.DeleteAllAnnotations(ctx, "session-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	annotations := repo.sessionAnnotations["session-1"]
	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations after delete all, got %d", len(annotations))
	}
}
