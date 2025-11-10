package usecase

import (
	"context"
	"testing"
	"time"

	"network-debugger/internal/features/tags/domain"
)

// Composer 1.
func TestNewService(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	if service == nil {
		t.Fatal("NewService returned nil")
	}

	// Проверяем что repository установлен (не можем сравнить напрямую из-за интерфейса)
	if service.repo == nil {
		t.Error("Repository not set")
	}
}

// Composer 1.
func TestService_ListPredefinedTags(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем теги в mock
	repo.predefinedTags = []domain.PredefinedTag{
		{ID: "tag-1", Name: "important", Color: "#ff0000"},
		{ID: "tag-2", Name: "urgent", Color: "#00ff00"},
	}

	ctx := context.Background()
	tags, err := service.ListPredefinedTags(ctx)

	if err != nil {
		t.Errorf("ListPredefinedTags() error = %v, want nil", err)
	}

	if len(tags) != 2 {
		t.Errorf("ListPredefinedTags() length = %d, want 2", len(tags))
	}
}

// Composer 1.
func TestService_CreatePredefinedTag(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()
	err := service.CreatePredefinedTag(ctx, "test-tag", "#0000ff", "test", 1)

	if err != nil {
		t.Errorf("CreatePredefinedTag() error = %v, want nil", err)
	}

	if len(repo.predefinedTags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(repo.predefinedTags))
	}

	if repo.predefinedTags[0].Name != "test-tag" {
		t.Errorf("Tag name = %q, want %q", repo.predefinedTags[0].Name, "test-tag")
	}
}

// Composer 1.
func TestService_DeletePredefinedTag(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем тег
	repo.predefinedTags = []domain.PredefinedTag{
		{ID: "tag-to-delete", Name: "delete-me"},
	}

	ctx := context.Background()
	err := service.DeletePredefinedTag(ctx, "tag-to-delete")

	if err != nil {
		t.Errorf("DeletePredefinedTag() error = %v, want nil", err)
	}

	if len(repo.predefinedTags) != 0 {
		t.Errorf("Expected 0 tags after delete, got %d", len(repo.predefinedTags))
	}
}

// Composer 1.
func TestService_GetSessionTags(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем теги для сессии
	repo.sessionTags["session-1"] = []domain.SessionTag{
		{ID: "st-1", SessionID: "session-1", TagName: "important"},
		{ID: "st-2", SessionID: "session-1", TagName: "urgent"},
	}

	ctx := context.Background()
	tags, err := service.GetSessionTags(ctx, "session-1")

	if err != nil {
		t.Errorf("GetSessionTags() error = %v, want nil", err)
	}

	if len(tags) != 2 {
		t.Errorf("GetSessionTags() length = %d, want 2", len(tags))
	}
}

// Composer 1.
func TestService_AddTagToSession(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()
	err := service.AddTagToSession(ctx, "session-1", "important")

	if err != nil {
		t.Errorf("AddTagToSession() error = %v, want nil", err)
	}

	tags := repo.sessionTags["session-1"]
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0].TagName != "important" {
		t.Errorf("TagName = %q, want %q", tags[0].TagName, "important")
	}
}

// Composer 1.
func TestService_AddTagToSession_EmptyParams(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Пустой sessionID
	err := service.AddTagToSession(ctx, "", "tag")
	if err == nil {
		t.Error("AddTagToSession() should fail with empty sessionID")
	}

	// Пустой tagName
	err = service.AddTagToSession(ctx, "session-1", "")
	if err == nil {
		t.Error("AddTagToSession() should fail with empty tagName")
	}
}

// Composer 1.
func TestService_RemoveTagFromSession(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем тег
	repo.sessionTags["session-1"] = []domain.SessionTag{
		{ID: "st-1", SessionID: "session-1", TagName: "important"},
	}

	ctx := context.Background()
	err := service.RemoveTagFromSession(ctx, "session-1", "important")

	if err != nil {
		t.Errorf("RemoveTagFromSession() error = %v, want nil", err)
	}

	tags := repo.sessionTags["session-1"]
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after remove, got %d", len(tags))
	}
}

// Composer 1.
func TestService_BulkAddTags(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()
	err := service.BulkAddTags(ctx, []string{"session-1", "session-2"}, []string{"tag1", "tag2"})

	if err != nil {
		t.Errorf("BulkAddTags() error = %v, want nil", err)
	}

	// Проверяем что теги добавлены
	if len(repo.sessionTags["session-1"]) != 2 {
		t.Errorf("Expected 2 tags for session-1, got %d", len(repo.sessionTags["session-1"]))
	}
}

// Composer 1.
func TestService_BulkAddTags_EmptyParams(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Пустые sessionIDs
	err := service.BulkAddTags(ctx, []string{}, []string{"tag1"})
	if err == nil {
		t.Error("BulkAddTags() should fail with empty sessionIDs")
	}

	// Пустые tagNames
	err = service.BulkAddTags(ctx, []string{"session-1"}, []string{})
	if err == nil {
		t.Error("BulkAddTags() should fail with empty tagNames")
	}
}

// Composer 1.
func TestService_BulkRemoveTags(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем теги
	repo.sessionTags["session-1"] = []domain.SessionTag{
		{ID: "st-1", SessionID: "session-1", TagName: "tag1"},
		{ID: "st-2", SessionID: "session-1", TagName: "tag2"},
	}

	ctx := context.Background()
	err := service.BulkRemoveTags(ctx, []string{"session-1"}, []string{"tag1"})

	if err != nil {
		t.Errorf("BulkRemoveTags() error = %v, want nil", err)
	}

	if len(repo.sessionTags["session-1"]) != 1 {
		t.Errorf("Expected 1 tag after bulk remove, got %d", len(repo.sessionTags["session-1"]))
	}
}

// Composer 1.
func TestService_DeleteAllSessionTags(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем теги
	repo.sessionTags["session-1"] = []domain.SessionTag{
		{ID: "st-1", SessionID: "session-1", TagName: "tag1"},
		{ID: "st-2", SessionID: "session-1", TagName: "tag2"},
	}

	ctx := context.Background()
	err := service.DeleteAllSessionTags(ctx, "session-1")

	if err != nil {
		t.Errorf("DeleteAllSessionTags() error = %v, want nil", err)
	}

	if len(repo.sessionTags["session-1"]) != 0 {
		t.Errorf("Expected 0 tags after delete all, got %d", len(repo.sessionTags["session-1"]))
	}
}

// Composer 1.
func TestService_FindSessionIDsByTags(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем теги
	repo.sessionTags["session-1"] = []domain.SessionTag{
		{ID: "st-1", SessionID: "session-1", TagName: "important"},
	}
	repo.sessionTags["session-2"] = []domain.SessionTag{
		{ID: "st-2", SessionID: "session-2", TagName: "urgent"},
	}

	ctx := context.Background()
	sessionIDs, err := service.FindSessionIDsByTags(ctx, []string{"important"})

	if err != nil {
		t.Errorf("FindSessionIDsByTags() error = %v, want nil", err)
	}

	if len(sessionIDs) != 1 {
		t.Errorf("FindSessionIDsByTags() length = %d, want 1", len(sessionIDs))
	}
}

// Composer 1.
func TestService_FindSessionIDsByTags_Empty(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()
	sessionIDs, err := service.FindSessionIDsByTags(ctx, []string{})

	if err != nil {
		t.Errorf("FindSessionIDsByTags() error = %v, want nil", err)
	}

	if len(sessionIDs) != 0 {
		t.Errorf("FindSessionIDsByTags() length = %d, want 0", len(sessionIDs))
	}
}

// Composer 1.
func TestService_UpsertAnnotation(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()
	err := service.UpsertAnnotation(ctx, "session-1", "key1", "value1")

	if err != nil {
		t.Errorf("UpsertAnnotation() error = %v, want nil", err)
	}

	annotations := repo.sessionAnnotations["session-1"]
	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Key != "key1" {
		t.Errorf("Key = %q, want %q", annotations[0].Key, "key1")
	}
}

// Composer 1.
func TestService_UpsertAnnotation_EmptyParams(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Пустой sessionID
	err := service.UpsertAnnotation(ctx, "", "key", "value")
	if err == nil {
		t.Error("UpsertAnnotation() should fail with empty sessionID")
	}

	// Пустой key
	err = service.UpsertAnnotation(ctx, "session-1", "", "value")
	if err == nil {
		t.Error("UpsertAnnotation() should fail with empty key")
	}
}

// Composer 1.
func TestService_DeleteAnnotation(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем аннотацию
	repo.sessionAnnotations["session-1"] = []domain.SessionAnnotation{
		{ID: "ann-1", SessionID: "session-1", Key: "key1", Value: "value1"},
	}

	ctx := context.Background()
	err := service.DeleteAnnotation(ctx, "session-1", "key1")

	if err != nil {
		t.Errorf("DeleteAnnotation() error = %v, want nil", err)
	}

	annotations := repo.sessionAnnotations["session-1"]
	if len(annotations) != 0 {
		t.Errorf("Expected 0 annotations after delete, got %d", len(annotations))
	}
}

// Composer 1.
func TestService_DeleteAllAnnotations(t *testing.T) {
	repo := newMockRepo()
	service := NewService(repo)

	// Добавляем аннотации
	repo.sessionAnnotations["session-1"] = []domain.SessionAnnotation{
		{ID: "ann-1", SessionID: "session-1", Key: "key1", Value: "value1"},
		{ID: "ann-2", SessionID: "session-1", Key: "key2", Value: "value2"},
	}

	ctx := context.Background()
	err := service.DeleteAllAnnotations(ctx, "session-1")

	if err != nil {
		t.Errorf("DeleteAllAnnotations() error = %v, want nil", err)
	}

	if len(repo.sessionAnnotations["session-1"]) != 0 {
		t.Errorf("Expected 0 annotations after delete all, got %d", len(repo.sessionAnnotations["session-1"]))
	}
}

// Mock repository implementation
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

func (m *mockTagsRepository) GetSessionAnnotations(_ context.Context, sessionID string) ([]domain.SessionAnnotation, error) {
	annotations, ok := m.sessionAnnotations[sessionID]
	if !ok {
		return []domain.SessionAnnotation{}, nil
	}
	return annotations, m.errGet
}

func (m *mockTagsRepository) SetSessionAnnotation(_ context.Context, annotation domain.SessionAnnotation) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	annotations := m.sessionAnnotations[annotation.SessionID]
	for i, ann := range annotations {
		if ann.Key == annotation.Key {
			annotations[i] = annotation
			m.sessionAnnotations[annotation.SessionID] = annotations
			return nil
		}
	}
	m.sessionAnnotations[annotation.SessionID] = append(annotations, annotation)
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

func (m *mockTagsRepository) BulkAddTags(_ context.Context, sessionIDs []string, tagNames []string) error {
	if m.errCreate != nil {
		return m.errCreate
	}
	for _, sessionID := range sessionIDs {
		for _, tagName := range tagNames {
			tag := domain.SessionTag{
				ID:        "bulk-" + sessionID + "-" + tagName,
				SessionID: sessionID,
				TagName:   tagName,
				CreatedAt: time.Now(),
			}
			m.sessionTags[sessionID] = append(m.sessionTags[sessionID], tag)
		}
	}
	return nil
}

func (m *mockTagsRepository) BulkRemoveTags(_ context.Context, sessionIDs []string, tagNames []string) error {
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

func (m *mockTagsRepository) BulkAddSessionTags(_ context.Context, sessionIDs []string, tagNames []string) error {
	return m.BulkAddTags(context.Background(), sessionIDs, tagNames)
}

func (m *mockTagsRepository) BulkRemoveSessionTags(_ context.Context, sessionIDs []string, tagNames []string) error {
	return m.BulkRemoveTags(context.Background(), sessionIDs, tagNames)
}

func (m *mockTagsRepository) DeleteAllSessionTags(_ context.Context, sessionID string) error {
	delete(m.sessionTags, sessionID)
	return nil
}

func (m *mockTagsRepository) FindSessionIDsByTags(_ context.Context, tagNames []string) ([]string, error) {
	var sessionIDs []string
	for sessionID, tags := range m.sessionTags {
		for _, tag := range tags {
			for _, tagName := range tagNames {
				if tag.TagName == tagName {
					sessionIDs = append(sessionIDs, sessionID)
					break
				}
			}
		}
	}
	return sessionIDs, nil
}

func (m *mockTagsRepository) UpsertSessionAnnotation(_ context.Context, annotation domain.SessionAnnotation) error {
	return m.SetSessionAnnotation(context.Background(), annotation)
}

func (m *mockTagsRepository) DeleteAllSessionAnnotations(_ context.Context, sessionID string) error {
	delete(m.sessionAnnotations, sessionID)
	return nil
}
