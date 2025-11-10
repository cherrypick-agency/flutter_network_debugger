package persistence

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"network-debugger/internal/features/tags/domain"
)

// Composer 1.
func TestNewRepo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)

	if repo == nil {
		t.Fatal("NewRepo returned nil")
	}

	if repo.db != db {
		t.Error("Database not set correctly")
	}
}

// Composer 1.
func TestRepo_ListPredefinedTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	// Создаем теги
	tag1 := domain.PredefinedTag{
		ID:           "tag-1",
		Name:         "important",
		Color:        "#ff0000",
		Category:     "priority",
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
	}

	tag2 := domain.PredefinedTag{
		ID:           "tag-2",
		Name:         "urgent",
		Color:        "#00ff00",
		Category:     "status",
		DisplayOrder: 2,
		CreatedAt:    time.Now(),
	}

	if err := repo.CreatePredefinedTag(ctx, tag1); err != nil {
		t.Fatalf("CreatePredefinedTag failed: %v", err)
	}

	if err := repo.CreatePredefinedTag(ctx, tag2); err != nil {
		t.Fatalf("CreatePredefinedTag failed: %v", err)
	}

	// Получаем список
	tags, err := repo.ListPredefinedTags(ctx)
	if err != nil {
		t.Fatalf("ListPredefinedTags failed: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("ListPredefinedTags() length = %d, want 2", len(tags))
	}
}

// Composer 1.
func TestRepo_CreatePredefinedTag(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	tag := domain.PredefinedTag{
		ID:           "tag-new",
		Name:         "new-tag",
		Color:        "#0000ff",
		Category:     "test",
		DisplayOrder: 0,
		CreatedAt:    time.Now(),
	}

	if err := repo.CreatePredefinedTag(ctx, tag); err != nil {
		t.Fatalf("CreatePredefinedTag failed: %v", err)
	}

	// Проверяем что тег создан
	tags, err := repo.ListPredefinedTags(ctx)
	if err != nil {
		t.Fatalf("ListPredefinedTags failed: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}

	if tags[0].Name != tag.Name {
		t.Errorf("Tag name = %q, want %q", tags[0].Name, tag.Name)
	}
}

// Composer 1.
func TestRepo_DeletePredefinedTag(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	tag := domain.PredefinedTag{
		ID:           "tag-to-delete",
		Name:         "delete-me",
		Color:        "#ff0000",
		Category:     "test",
		DisplayOrder: 0,
		CreatedAt:    time.Now(),
	}

	if err := repo.CreatePredefinedTag(ctx, tag); err != nil {
		t.Fatalf("CreatePredefinedTag failed: %v", err)
	}

	// Удаляем
	if err := repo.DeletePredefinedTag(ctx, tag.ID); err != nil {
		t.Fatalf("DeletePredefinedTag failed: %v", err)
	}

	// Проверяем что тег удален
	tags, err := repo.ListPredefinedTags(ctx)
	if err != nil {
		t.Fatalf("ListPredefinedTags failed: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after delete, got %d", len(tags))
	}
}

// Composer 1.
func TestRepo_GetSessionTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	tag := domain.SessionTag{
		ID:        "st-1",
		SessionID: "session-123",
		TagName:   "important",
		CreatedAt: time.Now(),
	}

	if err := repo.AddSessionTag(ctx, tag); err != nil {
		t.Fatalf("AddSessionTag failed: %v", err)
	}

	// Получаем теги сессии
	tags, err := repo.GetSessionTags(ctx, "session-123")
	if err != nil {
		t.Fatalf("GetSessionTags failed: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("GetSessionTags() length = %d, want 1", len(tags))
	}

	if tags[0].TagName != tag.TagName {
		t.Errorf("TagName = %q, want %q", tags[0].TagName, tag.TagName)
	}
}

// Composer 1.
func TestRepo_AddSessionTag(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	tag := domain.SessionTag{
		ID:        "st-new",
		SessionID: "session-456",
		TagName:   "new-tag",
		CreatedAt: time.Now(),
	}

	if err := repo.AddSessionTag(ctx, tag); err != nil {
		t.Fatalf("AddSessionTag failed: %v", err)
	}

	// Проверяем что тег добавлен
	tags, err := repo.GetSessionTags(ctx, "session-456")
	if err != nil {
		t.Fatalf("GetSessionTags failed: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
}

// Composer 1.
func TestRepo_RemoveSessionTag(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	tag := domain.SessionTag{
		ID:        "st-remove",
		SessionID: "session-789",
		TagName:   "to-remove",
		CreatedAt: time.Now(),
	}

	if err := repo.AddSessionTag(ctx, tag); err != nil {
		t.Fatalf("AddSessionTag failed: %v", err)
	}

	// Удаляем тег
	if err := repo.RemoveSessionTag(ctx, "session-789", "to-remove"); err != nil {
		t.Fatalf("RemoveSessionTag failed: %v", err)
	}

	// Проверяем что тег удален
	tags, err := repo.GetSessionTags(ctx, "session-789")
	if err != nil {
		t.Fatalf("GetSessionTags failed: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after remove, got %d", len(tags))
	}
}

// Composer 1.
func TestRepo_BulkAddSessionTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	sessionIDs := []string{"session-1", "session-2"}
	tagNames := []string{"tag1", "tag2"}

	if err := repo.BulkAddSessionTags(ctx, sessionIDs, tagNames); err != nil {
		t.Fatalf("BulkAddSessionTags failed: %v", err)
	}

	// Проверяем что теги добавлены
	for _, sessionID := range sessionIDs {
		tags, err := repo.GetSessionTags(ctx, sessionID)
		if err != nil {
			t.Fatalf("GetSessionTags failed: %v", err)
		}

		if len(tags) != len(tagNames) {
			t.Errorf("Expected %d tags for session %s, got %d", len(tagNames), sessionID, len(tags))
		}
	}
}

// Composer 1.
func TestRepo_BulkAddSessionTags_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	// Пустые списки не должны вызывать ошибку
	if err := repo.BulkAddSessionTags(ctx, []string{}, []string{"tag1"}); err != nil {
		t.Fatalf("BulkAddSessionTags with empty sessionIDs should not fail: %v", err)
	}

	if err := repo.BulkAddSessionTags(ctx, []string{"session-1"}, []string{}); err != nil {
		t.Fatalf("BulkAddSessionTags with empty tagNames should not fail: %v", err)
	}
}

// Composer 1.
func TestRepo_BulkRemoveSessionTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	// Добавляем теги
	sessionIDs := []string{"session-1", "session-2"}
	tagNames := []string{"tag1", "tag2"}

	if err := repo.BulkAddSessionTags(ctx, sessionIDs, tagNames); err != nil {
		t.Fatalf("BulkAddSessionTags failed: %v", err)
	}

	// Удаляем один тег из всех сессий
	if err := repo.BulkRemoveSessionTags(ctx, sessionIDs, []string{"tag1"}); err != nil {
		t.Fatalf("BulkRemoveSessionTags failed: %v", err)
	}

	// Проверяем что tag1 удален, но tag2 остался
	for _, sessionID := range sessionIDs {
		tags, err := repo.GetSessionTags(ctx, sessionID)
		if err != nil {
			t.Fatalf("GetSessionTags failed: %v", err)
		}

		if len(tags) != 1 {
			t.Errorf("Expected 1 tag for session %s, got %d", sessionID, len(tags))
		}

		if tags[0].TagName != "tag2" {
			t.Errorf("Expected tag2, got %q", tags[0].TagName)
		}
	}
}

// Composer 1.
func TestRepo_DeleteAllSessionTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	// Добавляем несколько тегов
	tag1 := domain.SessionTag{
		ID:        "st-1",
		SessionID: "session-all",
		TagName:   "tag1",
		CreatedAt: time.Now(),
	}

	tag2 := domain.SessionTag{
		ID:        "st-2",
		SessionID: "session-all",
		TagName:   "tag2",
		CreatedAt: time.Now(),
	}

	if err := repo.AddSessionTag(ctx, tag1); err != nil {
		t.Fatalf("AddSessionTag failed: %v", err)
	}

	if err := repo.AddSessionTag(ctx, tag2); err != nil {
		t.Fatalf("AddSessionTag failed: %v", err)
	}

	// Удаляем все теги сессии
	if err := repo.DeleteAllSessionTags(ctx, "session-all"); err != nil {
		t.Fatalf("DeleteAllSessionTags failed: %v", err)
	}

	// Проверяем что все теги удалены
	tags, err := repo.GetSessionTags(ctx, "session-all")
	if err != nil {
		t.Fatalf("GetSessionTags failed: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after DeleteAllSessionTags, got %d", len(tags))
	}
}

// Composer 1.
func TestRepo_FindSessionIDsByTags(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	// Добавляем теги для разных сессий
	repo.AddSessionTag(ctx, domain.SessionTag{
		ID:        "st-1",
		SessionID: "session-1",
		TagName:   "important",
		CreatedAt: time.Now(),
	})

	repo.AddSessionTag(ctx, domain.SessionTag{
		ID:        "st-2",
		SessionID: "session-2",
		TagName:   "urgent",
		CreatedAt: time.Now(),
	})

	repo.AddSessionTag(ctx, domain.SessionTag{
		ID:        "st-3",
		SessionID: "session-1",
		TagName:   "urgent",
		CreatedAt: time.Now(),
	})

	// Ищем сессии с тегом "important"
	sessionIDs, err := repo.FindSessionIDsByTags(ctx, []string{"important"})
	if err != nil {
		t.Fatalf("FindSessionIDsByTags failed: %v", err)
	}

	if len(sessionIDs) != 1 {
		t.Errorf("Expected 1 session, got %d", len(sessionIDs))
	}

	if sessionIDs[0] != "session-1" {
		t.Errorf("Expected session-1, got %q", sessionIDs[0])
	}
}

// Composer 1.
func TestRepo_FindSessionIDsByTags_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	sessionIDs, err := repo.FindSessionIDsByTags(ctx, []string{})
	if err != nil {
		t.Fatalf("FindSessionIDsByTags failed: %v", err)
	}

	if len(sessionIDs) != 0 {
		t.Errorf("Expected empty list, got %d", len(sessionIDs))
	}
}

// Composer 1.
func TestRepo_GetSessionAnnotations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	annotation := domain.SessionAnnotation{
		ID:        "ann-1",
		SessionID: "session-annot",
		Key:       "note",
		Value:     "test value",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.UpsertSessionAnnotation(ctx, annotation); err != nil {
		t.Fatalf("UpsertSessionAnnotation failed: %v", err)
	}

	// Получаем аннотации
	annotations, err := repo.GetSessionAnnotations(ctx, "session-annot")
	if err != nil {
		t.Fatalf("GetSessionAnnotations failed: %v", err)
	}

	if len(annotations) != 1 {
		t.Errorf("GetSessionAnnotations() length = %d, want 1", len(annotations))
	}

	if annotations[0].Key != annotation.Key {
		t.Errorf("Key = %q, want %q", annotations[0].Key, annotation.Key)
	}
}

// Composer 1.
func TestRepo_UpsertSessionAnnotation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	annotation := domain.SessionAnnotation{
		ID:        "ann-new",
		SessionID: "session-upsert",
		Key:       "status",
		Value:     "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.UpsertSessionAnnotation(ctx, annotation); err != nil {
		t.Fatalf("UpsertSessionAnnotation failed: %v", err)
	}

	// Обновляем значение
	annotation.Value = "inactive"
	if err := repo.UpsertSessionAnnotation(ctx, annotation); err != nil {
		t.Fatalf("UpsertSessionAnnotation update failed: %v", err)
	}

	// Проверяем что значение обновлено
	annotations, err := repo.GetSessionAnnotations(ctx, "session-upsert")
	if err != nil {
		t.Fatalf("GetSessionAnnotations failed: %v", err)
	}

	if len(annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(annotations))
	}

	if annotations[0].Value != "inactive" {
		t.Errorf("Value = %q, want %q", annotations[0].Value, "inactive")
	}
}

// Composer 1.
func TestRepo_DeleteSessionAnnotation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	annotation := domain.SessionAnnotation{
		ID:        "ann-delete",
		SessionID: "session-delete",
		Key:       "to-delete",
		Value:     "value",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.UpsertSessionAnnotation(ctx, annotation); err != nil {
		t.Fatalf("UpsertSessionAnnotation failed: %v", err)
	}

	// Удаляем аннотацию
	if err := repo.DeleteSessionAnnotation(ctx, "session-delete", "to-delete"); err != nil {
		t.Fatalf("DeleteSessionAnnotation failed: %v", err)
	}

	// Проверяем что аннотация удалена
	annotations, err := repo.GetSessionAnnotations(ctx, "session-delete")
	if err != nil {
		t.Fatalf("GetSessionAnnotations failed: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("Expected 0 annotations after delete, got %d", len(annotations))
	}
}

// Composer 1.
func TestRepo_DeleteAllSessionAnnotations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepo(db)
	ctx := context.Background()

	// Добавляем несколько аннотаций
	ann1 := domain.SessionAnnotation{
		ID:        "ann-1",
		SessionID: "session-all",
		Key:       "key1",
		Value:     "value1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ann2 := domain.SessionAnnotation{
		ID:        "ann-2",
		SessionID: "session-all",
		Key:       "key2",
		Value:     "value2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := repo.UpsertSessionAnnotation(ctx, ann1); err != nil {
		t.Fatalf("UpsertSessionAnnotation failed: %v", err)
	}

	if err := repo.UpsertSessionAnnotation(ctx, ann2); err != nil {
		t.Fatalf("UpsertSessionAnnotation failed: %v", err)
	}

	// Удаляем все аннотации сессии
	if err := repo.DeleteAllSessionAnnotations(ctx, "session-all"); err != nil {
		t.Fatalf("DeleteAllSessionAnnotations failed: %v", err)
	}

	// Проверяем что все аннотации удалены
	annotations, err := repo.GetSessionAnnotations(ctx, "session-all")
	if err != nil {
		t.Fatalf("GetSessionAnnotations failed: %v", err)
	}

	if len(annotations) != 0 {
		t.Errorf("Expected 0 annotations after DeleteAllSessionAnnotations, got %d", len(annotations))
	}
}

// Helper function for test database
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Auto-migrate
	if err := db.AutoMigrate(&PredefinedTagModel{}, &SessionTagModel{}, &SessionAnnotationModel{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Создаем уникальный индекс для session_tags (session_id, tag_name)
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_session_tags_unique ON session_tags(session_id, tag_name)").Error; err != nil {
		t.Fatalf("Failed to create unique index: %v", err)
	}

	// Создаем уникальный индекс для session_annotations (session_id, key)
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_session_annotations_unique ON session_annotations(session_id, key)").Error; err != nil {
		t.Fatalf("Failed to create unique index for annotations: %v", err)
	}

	return db
}
