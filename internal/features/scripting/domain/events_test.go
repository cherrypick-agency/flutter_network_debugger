package domain

import (
	"errors"
	"testing"
	"time"
)

// Composer 1.
func TestScriptExecutedEvent(t *testing.T) {
	now := time.Now()
	event := ScriptExecutedEvent{
		ScriptID:   "script-123",
		ScriptName: "test-script",
		Runtime:    RuntimeExtism,
		Duration:   100 * time.Millisecond,
		Success:    true,
		Timestamp:  now,
	}

	if event.ScriptID != "script-123" {
		t.Errorf("Expected ScriptID 'script-123', got %q", event.ScriptID)
	}

	if event.ScriptName != "test-script" {
		t.Errorf("Expected ScriptName 'test-script', got %q", event.ScriptName)
	}

	if event.Runtime != RuntimeExtism {
		t.Errorf("Expected Runtime %v, got %v", RuntimeExtism, event.Runtime)
	}

	if event.Duration != 100*time.Millisecond {
		t.Errorf("Expected Duration 100ms, got %v", event.Duration)
	}

	if !event.Success {
		t.Error("Expected Success true")
	}

	if !event.Timestamp.Equal(now) {
		t.Errorf("Expected Timestamp %v, got %v", now, event.Timestamp)
	}
}

// Composer 1.
func TestScriptFailedEvent(t *testing.T) {
	now := time.Now()
	testErr := errors.New("execution failed")
	event := ScriptFailedEvent{
		ScriptID:   "script-456",
		ScriptName: "failing-script",
		Runtime:    RuntimeDart,
		Error:      testErr,
		Timestamp:  now,
	}

	if event.ScriptID != "script-456" {
		t.Errorf("Expected ScriptID 'script-456', got %q", event.ScriptID)
	}

	if event.Error != testErr {
		t.Errorf("Expected Error %v, got %v", testErr, event.Error)
	}

	if event.Runtime != RuntimeDart {
		t.Errorf("Expected Runtime %v, got %v", RuntimeDart, event.Runtime)
	}
}

// Composer 1.
func TestScriptCreatedEvent(t *testing.T) {
	now := time.Now()
	event := ScriptCreatedEvent{
		ScriptID:   "script-789",
		ScriptName: "new-script",
		Runtime:    RuntimeExtism,
		CreatedBy:  "user-123",
		Timestamp:  now,
	}

	if event.CreatedBy != "user-123" {
		t.Errorf("Expected CreatedBy 'user-123', got %q", event.CreatedBy)
	}

	if event.Runtime != RuntimeExtism {
		t.Errorf("Expected Runtime %v, got %v", RuntimeExtism, event.Runtime)
	}
}

// Composer 1.
func TestScriptUpdatedEvent(t *testing.T) {
	now := time.Now()
	event := ScriptUpdatedEvent{
		ScriptID:   "script-999",
		ScriptName: "updated-script",
		UpdatedBy:  "user-456",
		Timestamp:  now,
	}

	if event.UpdatedBy != "user-456" {
		t.Errorf("Expected UpdatedBy 'user-456', got %q", event.UpdatedBy)
	}

	if event.ScriptID != "script-999" {
		t.Errorf("Expected ScriptID 'script-999', got %q", event.ScriptID)
	}
}

// Composer 1.
func TestScriptDeletedEvent(t *testing.T) {
	now := time.Now()
	event := ScriptDeletedEvent{
		ScriptID:  "script-111",
		DeletedBy: "user-789",
		Timestamp: now,
	}

	if event.DeletedBy != "user-789" {
		t.Errorf("Expected DeletedBy 'user-789', got %q", event.DeletedBy)
	}

	if event.ScriptID != "script-111" {
		t.Errorf("Expected ScriptID 'script-111', got %q", event.ScriptID)
	}
}
