package httpapi

import (
	"testing"
)

func TestNewLiveSessions(t *testing.T) {
	ls := NewLiveSessions()

	if ls == nil {
		t.Fatal("NewLiveSessions returned nil")
	}
	if ls.m == nil {
		t.Error("map should be initialized")
	}
}

func TestLiveSessions_Register(t *testing.T) {
	ls := NewLiveSessions()

	ls.Register("session1", nil, nil)

	ls.mu.RLock()
	if ls.m["session1"] == nil {
		t.Error("session should be registered")
	}
	ls.mu.RUnlock()
}

func TestLiveSessions_Register_EmptySessionID(t *testing.T) {
	ls := NewLiveSessions()

	ls.Register("", nil, nil)

	ls.mu.RLock()
	if len(ls.m) != 0 {
		t.Error("should not register with empty session ID")
	}
	ls.mu.RUnlock()
}

func TestLiveSessions_Unregister(t *testing.T) {
	ls := NewLiveSessions()

	ls.Register("session1", nil, nil)
	ls.Unregister("session1")

	ls.mu.RLock()
	if ls.m["session1"] != nil {
		t.Error("session should be unregistered")
	}
	ls.mu.RUnlock()
}

func TestLiveSessions_Unregister_EmptySessionID(t *testing.T) {
	ls := NewLiveSessions()

	ls.Register("session1", nil, nil)
	ls.Unregister("")

	ls.mu.RLock()
	if len(ls.m) != 1 {
		t.Error("should not unregister with empty session ID")
	}
	ls.mu.RUnlock()
}

func TestLiveSessions_CloseAll_Empty(t *testing.T) {
	ls := NewLiveSessions()

	// Should not panic with empty map
	ls.CloseAll()

	ls.mu.RLock()
	if len(ls.m) != 0 {
		t.Error("map should be empty after CloseAll")
	}
	ls.mu.RUnlock()
}

func TestLiveSessions_CloseAll_NilConnections(t *testing.T) {
	ls := NewLiveSessions()

	ls.Register("session1", nil, nil)
	ls.Register("session2", nil, nil)

	// Should not panic with nil connections
	ls.CloseAll()

	ls.mu.RLock()
	if len(ls.m) != 0 {
		t.Error("all sessions should be removed")
	}
	ls.mu.RUnlock()
}

func TestLiveSessions_CloseAll_WithMultipleSessions(t *testing.T) {
	ls := NewLiveSessions()

	// Register multiple sessions with nil connections
	ls.Register("session1", nil, nil)
	ls.Register("session2", nil, nil)
	ls.Register("session3", nil, nil)

	// Verify they're all registered
	ls.mu.RLock()
	count := len(ls.m)
	ls.mu.RUnlock()
	if count != 3 {
		t.Errorf("expected 3 sessions, got %d", count)
	}

	// Close all
	ls.CloseAll()

	// Verify all are removed
	ls.mu.RLock()
	if len(ls.m) != 0 {
		t.Error("all sessions should be removed after CloseAll")
	}
	ls.mu.RUnlock()
}
