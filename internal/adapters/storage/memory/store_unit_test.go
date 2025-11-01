package memory

import (
	"context"
	"testing"

	"network-debugger/internal/domain"
)

func TestStore_ListEvents(t *testing.T) {
	ctx := context.Background()

	t.Run("session not found", func(t *testing.T) {
		store := NewStore(100, 100, 0)
		events, next, err := store.ListEvents(ctx, "non-existent", "", 10)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		if events != nil {
			t.Errorf("ListEvents() events = %v, want nil", events)
		}
		if next != "" {
			t.Errorf("ListEvents() next = %q, want empty", next)
		}
	})

	t.Run("with events", func(t *testing.T) {
		store := NewStore(100, 100, 0)

		// Create session
		sessionID := "test-session-1"
		err := store.CreateSession(ctx, domain.Session{
			ID:     sessionID,
			Target: "test",
			Kind:   "http",
		})
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}

		// Append events
		events := []domain.Event{
			{ID: "event-1", Name: "test1", ArgsPreview: "payload1"},
			{ID: "event-2", Name: "test2", ArgsPreview: "payload2"},
			{ID: "event-3", Name: "test3", ArgsPreview: "payload3"},
			{ID: "event-4", Name: "test4", ArgsPreview: "payload4"},
			{ID: "event-5", Name: "test5", ArgsPreview: "payload5"},
		}
		for _, e := range events {
			store.AppendEvent(ctx, sessionID, e)
		}

		// List all
		result, next, err := store.ListEvents(ctx, sessionID, "", 0)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		if len(result) != 5 {
			t.Errorf("ListEvents() returned %d events, want 5", len(result))
		}
		if next != "" {
			t.Errorf("ListEvents() next = %q, want empty", next)
		}

		// List with limit
		result, next, err = store.ListEvents(ctx, sessionID, "", 2)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		if len(result) != 2 {
			t.Errorf("ListEvents() returned %d events, want 2", len(result))
		}
		if result[0].ID != "event-1" {
			t.Errorf("ListEvents() first event ID = %q, want event-1", result[0].ID)
		}
		if next != "event-2" {
			t.Errorf("ListEvents() next = %q, want event-2", next)
		}

		// List from cursor
		result, next, err = store.ListEvents(ctx, sessionID, "event-2", 2)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		if len(result) != 2 {
			t.Errorf("ListEvents() returned %d events, want 2", len(result))
		}
		if result[0].ID != "event-3" {
			t.Errorf("ListEvents() first event ID = %q, want event-3", result[0].ID)
		}
		if result[1].ID != "event-4" {
			t.Errorf("ListEvents() second event ID = %q, want event-4", result[1].ID)
		}

		// List from last position
		result, next, err = store.ListEvents(ctx, sessionID, "event-4", 10)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		if len(result) != 1 {
			t.Errorf("ListEvents() returned %d events, want 1", len(result))
		}
		if result[0].ID != "event-5" {
			t.Errorf("ListEvents() first event ID = %q, want event-5", result[0].ID)
		}
		if next != "" {
			t.Errorf("ListEvents() next = %q, want empty", next)
		}

		// List from non-existent cursor
		result, next, err = store.ListEvents(ctx, sessionID, "non-existent", 10)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		// Should return all events since cursor not found
		if len(result) != 5 {
			t.Errorf("ListEvents() with non-existent cursor returned %d events, want 5", len(result))
		}
	})

	t.Run("empty events list", func(t *testing.T) {
		store := NewStore(100, 100, 0)

		// Create session without events
		sessionID := "test-session-2"
		err := store.CreateSession(ctx, domain.Session{
			ID:     sessionID,
			Target: "test",
			Kind:   "http",
		})
		if err != nil {
			t.Fatalf("CreateSession() error = %v", err)
		}

		result, next, err := store.ListEvents(ctx, sessionID, "", 10)
		if err != nil {
			t.Fatalf("ListEvents() error = %v", err)
		}
		if len(result) != 0 {
			t.Errorf("ListEvents() returned %d events, want 0", len(result))
		}
		if next != "" {
			t.Errorf("ListEvents() next = %q, want empty", next)
		}
	})
}
