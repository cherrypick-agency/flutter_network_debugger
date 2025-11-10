package persistence

import (
	"testing"
	"time"

	"network-debugger/internal/features/tags/domain"
)

// Composer 1.
func TestPredefinedTagToModel(t *testing.T) {
	now := time.Now()
	tag := domain.PredefinedTag{
		ID:           "tag-1",
		Name:         "important",
		Color:        "#ff0000",
		Category:     "priority",
		DisplayOrder: 1,
		CreatedAt:    now,
	}

	model := predefinedTagToModel(tag)

	if model.ID != tag.ID {
		t.Errorf("ID = %q, want %q", model.ID, tag.ID)
	}

	if model.Name != tag.Name {
		t.Errorf("Name = %q, want %q", model.Name, tag.Name)
	}

	if model.Color != tag.Color {
		t.Errorf("Color = %q, want %q", model.Color, tag.Color)
	}

	if model.Category != tag.Category {
		t.Errorf("Category = %q, want %q", model.Category, tag.Category)
	}

	if model.DisplayOrder != tag.DisplayOrder {
		t.Errorf("DisplayOrder = %d, want %d", model.DisplayOrder, tag.DisplayOrder)
	}

	if !model.CreatedAt.Equal(tag.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", model.CreatedAt, tag.CreatedAt)
	}
}

// Composer 1.
func TestPredefinedTagToDomain(t *testing.T) {
	now := time.Now()
	model := &PredefinedTagModel{
		ID:           "tag-2",
		Name:         "urgent",
		Color:        "#00ff00",
		Category:     "status",
		DisplayOrder: 2,
		CreatedAt:    now,
	}

	tag := predefinedTagToDomain(model)

	if tag.ID != model.ID {
		t.Errorf("ID = %q, want %q", tag.ID, model.ID)
	}

	if tag.Name != model.Name {
		t.Errorf("Name = %q, want %q", tag.Name, model.Name)
	}

	if tag.Color != model.Color {
		t.Errorf("Color = %q, want %q", tag.Color, model.Color)
	}
}

// Composer 1.
func TestPredefinedTagRoundTrip(t *testing.T) {
	original := domain.PredefinedTag{
		ID:           "tag-3",
		Name:         "test",
		Color:        "#0000ff",
		Category:     "test",
		DisplayOrder: 3,
		CreatedAt:    time.Now(),
	}

	model := predefinedTagToModel(original)
	converted := predefinedTagToDomain(model)

	if converted.ID != original.ID {
		t.Errorf("RoundTrip ID = %q, want %q", converted.ID, original.ID)
	}

	if converted.Name != original.Name {
		t.Errorf("RoundTrip Name = %q, want %q", converted.Name, original.Name)
	}
}

// Composer 1.
func TestSessionTagToModel(t *testing.T) {
	now := time.Now()
	tag := domain.SessionTag{
		ID:        "session-tag-1",
		SessionID: "session-123",
		TagName:   "important",
		CreatedAt: now,
	}

	model := sessionTagToModel(tag)

	if model.ID != tag.ID {
		t.Errorf("ID = %q, want %q", model.ID, tag.ID)
	}

	if model.SessionID != tag.SessionID {
		t.Errorf("SessionID = %q, want %q", model.SessionID, tag.SessionID)
	}

	if model.TagName != tag.TagName {
		t.Errorf("TagName = %q, want %q", model.TagName, tag.TagName)
	}
}

// Composer 1.
func TestSessionTagToDomain(t *testing.T) {
	now := time.Now()
	model := &SessionTagModel{
		ID:        "session-tag-2",
		SessionID: "session-456",
		TagName:   "urgent",
		CreatedAt: now,
	}

	tag := sessionTagToDomain(model)

	if tag.ID != model.ID {
		t.Errorf("ID = %q, want %q", tag.ID, model.ID)
	}

	if tag.SessionID != model.SessionID {
		t.Errorf("SessionID = %q, want %q", tag.SessionID, model.SessionID)
	}

	if tag.TagName != model.TagName {
		t.Errorf("TagName = %q, want %q", tag.TagName, model.TagName)
	}
}

// Composer 1.
func TestSessionTagRoundTrip(t *testing.T) {
	original := domain.SessionTag{
		ID:        "session-tag-3",
		SessionID: "session-789",
		TagName:   "test",
		CreatedAt: time.Now(),
	}

	model := sessionTagToModel(original)
	converted := sessionTagToDomain(model)

	if converted.ID != original.ID {
		t.Errorf("RoundTrip ID = %q, want %q", converted.ID, original.ID)
	}

	if converted.SessionID != original.SessionID {
		t.Errorf("RoundTrip SessionID = %q, want %q", converted.SessionID, original.SessionID)
	}
}

// Composer 1.
func TestSessionAnnotationToModel(t *testing.T) {
	now := time.Now()
	annotation := domain.SessionAnnotation{
		ID:        "annotation-1",
		SessionID: "session-123",
		Key:       "note",
		Value:     "test value",
		CreatedAt: now,
		UpdatedAt: now,
	}

	model := sessionAnnotationToModel(annotation)

	if model.ID != annotation.ID {
		t.Errorf("ID = %q, want %q", model.ID, annotation.ID)
	}

	if model.SessionID != annotation.SessionID {
		t.Errorf("SessionID = %q, want %q", model.SessionID, annotation.SessionID)
	}

	if model.Key != annotation.Key {
		t.Errorf("Key = %q, want %q", model.Key, annotation.Key)
	}

	if model.Value != annotation.Value {
		t.Errorf("Value = %q, want %q", model.Value, annotation.Value)
	}
}

// Composer 1.
func TestSessionAnnotationToDomain(t *testing.T) {
	now := time.Now()
	model := &SessionAnnotationModel{
		ID:        "annotation-2",
		SessionID: "session-456",
		Key:       "status",
		Value:     "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	annotation := sessionAnnotationToDomain(model)

	if annotation.ID != model.ID {
		t.Errorf("ID = %q, want %q", annotation.ID, model.ID)
	}

	if annotation.Key != model.Key {
		t.Errorf("Key = %q, want %q", annotation.Key, model.Key)
	}

	if annotation.Value != model.Value {
		t.Errorf("Value = %q, want %q", annotation.Value, model.Value)
	}
}

// Composer 1.
func TestSessionAnnotationRoundTrip(t *testing.T) {
	now := time.Now()
	original := domain.SessionAnnotation{
		ID:        "annotation-3",
		SessionID: "session-789",
		Key:       "test",
		Value:     "test value",
		CreatedAt: now,
		UpdatedAt: now,
	}

	model := sessionAnnotationToModel(original)
	converted := sessionAnnotationToDomain(model)

	if converted.ID != original.ID {
		t.Errorf("RoundTrip ID = %q, want %q", converted.ID, original.ID)
	}

	if converted.Key != original.Key {
		t.Errorf("RoundTrip Key = %q, want %q", converted.Key, original.Key)
	}
}

// Composer 1.
func TestPredefinedTagModel_TableName(t *testing.T) {
	model := PredefinedTagModel{}
	if model.TableName() != "predefined_tags" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "predefined_tags")
	}
}

// Composer 1.
func TestSessionTagModel_TableName(t *testing.T) {
	model := SessionTagModel{}
	if model.TableName() != "session_tags" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "session_tags")
	}
}

// Composer 1.
func TestSessionAnnotationModel_TableName(t *testing.T) {
	model := SessionAnnotationModel{}
	if model.TableName() != "session_annotations" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "session_annotations")
	}
}
