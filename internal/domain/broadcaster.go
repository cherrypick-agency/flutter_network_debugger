package domain

// MonitorEvent represents an event to be broadcasted to all listeners.
type MonitorEvent struct {
	Type  string        `json:"type"`
	ID    string        `json:"id"`
	Ref   string        `json:"ref,omitempty"`
	Error *ErrorDetails `json:"error,omitempty"`
}

// ErrorDetails contains error information for monitor events.
type ErrorDetails struct {
	Code    string `json:"code"`             // Error code: CONNECTION_CLOSED, SERVER_UNAVAILABLE, etc.
	Message string `json:"message"`          // Human-readable error message
	Raw     string `json:"raw,omitempty"`    // Original technical error (for debugging)
	Target  string `json:"target,omitempty"` // Target URL
	Method  string `json:"method,omitempty"` // HTTP method
}

// Broadcaster defines the interface for broadcasting monitor events.
type Broadcaster interface {
	Broadcast(event MonitorEvent)
}
