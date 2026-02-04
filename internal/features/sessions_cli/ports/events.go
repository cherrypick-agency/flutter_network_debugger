package ports

import "context"

// EventType describes the type of monitoring event relevant to CLI.
type EventType string

const (
	EventSessionStarted EventType = "session_started"
	EventSessionEnded   EventType = "session_ended"
	EventHTTPTxAdded    EventType = "http_tx_added"
)

// Event — simplified event model for CLI.
type Event struct {
	Type      EventType
	SessionID string
	Ref       string // e.g., http transaction id
}

// EventStream — port for subscribing to runtime events.
//
// SRP/DIP: CLI depends only on the abstraction, not on the concrete MonitorHub.
type EventStream interface {
	Subscribe(ctx context.Context) (<-chan Event, func(), error)
}
