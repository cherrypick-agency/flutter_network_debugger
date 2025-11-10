package domain

import (
	"testing"
	"time"
)

// Composer 1.
func TestPredefinedTag(t *testing.T) {
	now := time.Now()
	tag := PredefinedTag{
		ID:           "tag-1",
		Name:         "important",
		Color:        "#ff0000",
		Category:     "priority",
		DisplayOrder: 1,
		CreatedAt:    now,
	}

	if tag.ID != "tag-1" {
		t.Errorf("Expected ID 'tag-1', got %q", tag.ID)
	}

	if tag.Name != "important" {
		t.Errorf("Expected Name 'important', got %q", tag.Name)
	}

	if tag.Color != "#ff0000" {
		t.Errorf("Expected Color '#ff0000', got %q", tag.Color)
	}

	if tag.Category != "priority" {
		t.Errorf("Expected Category 'priority', got %q", tag.Category)
	}

	if tag.DisplayOrder != 1 {
		t.Errorf("Expected DisplayOrder 1, got %d", tag.DisplayOrder)
	}

	if !tag.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, tag.CreatedAt)
	}
}

// Composer 1.
func TestSessionTag(t *testing.T) {
	now := time.Now()
	tag := SessionTag{
		ID:        "session-tag-1",
		SessionID: "session-123",
		TagName:   "important",
		CreatedAt: now,
	}

	if tag.SessionID != "session-123" {
		t.Errorf("Expected SessionID 'session-123', got %q", tag.SessionID)
	}

	if tag.TagName != "important" {
		t.Errorf("Expected TagName 'important', got %q", tag.TagName)
	}
}

// Composer 1.
func TestSessionAnnotation(t *testing.T) {
	now := time.Now()
	annotation := SessionAnnotation{
		ID:        "annotation-1",
		SessionID: "session-456",
		Key:       "note",
		Value:     "test value",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if annotation.Key != "note" {
		t.Errorf("Expected Key 'note', got %q", annotation.Key)
	}

	if annotation.Value != "test value" {
		t.Errorf("Expected Value 'test value', got %q", annotation.Value)
	}

	if annotation.SessionID != "session-456" {
		t.Errorf("Expected SessionID 'session-456', got %q", annotation.SessionID)
	}
}

// Composer 1.
func TestBulkTagOperation(t *testing.T) {
	op := BulkTagOperation{
		Operation:  "add",
		SessionIDs: []string{"session-1", "session-2"},
		TagNames:   []string{"tag1", "tag2"},
	}

	if op.Operation != "add" {
		t.Errorf("Expected Operation 'add', got %q", op.Operation)
	}

	if len(op.SessionIDs) != 2 {
		t.Errorf("Expected 2 SessionIDs, got %d", len(op.SessionIDs))
	}

	if len(op.TagNames) != 2 {
		t.Errorf("Expected 2 TagNames, got %d", len(op.TagNames))
	}

	// Тест операции remove
	opRemove := BulkTagOperation{
		Operation:  "remove",
		SessionIDs: []string{"session-3"},
		TagNames:   []string{"tag1"},
	}

	if opRemove.Operation != "remove" {
		t.Errorf("Expected Operation 'remove', got %q", opRemove.Operation)
	}
}
