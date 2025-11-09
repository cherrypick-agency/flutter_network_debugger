package domain

import "time"

// Domain Events for scripting feature
// These can be used for logging, monitoring, and event-driven architecture

// ScriptExecutedEvent is published when a script executes successfully
type ScriptExecutedEvent struct {
	ScriptID   string
	ScriptName string
	Runtime    ScriptRuntime
	Duration   time.Duration
	Success    bool
	Timestamp  time.Time
}

// ScriptFailedEvent is published when a script execution fails
type ScriptFailedEvent struct {
	ScriptID   string
	ScriptName string
	Runtime    ScriptRuntime
	Error      error
	Timestamp  time.Time
}

// ScriptCreatedEvent is published when a new script is created
type ScriptCreatedEvent struct {
	ScriptID   string
	ScriptName string
	Runtime    ScriptRuntime
	CreatedBy  string
	Timestamp  time.Time
}

// ScriptUpdatedEvent is published when a script is modified
type ScriptUpdatedEvent struct {
	ScriptID   string
	ScriptName string
	UpdatedBy  string
	Timestamp  time.Time
}

// ScriptDeletedEvent is published when a script is removed
type ScriptDeletedEvent struct {
	ScriptID   string
	ScriptName string
	DeletedBy  string
	Timestamp  time.Time
}
