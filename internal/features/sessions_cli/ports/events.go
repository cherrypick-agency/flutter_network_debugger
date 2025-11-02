package ports

import "context"

// EventType описывает тип события мониторинга, интересующий CLI.
type EventType string

const (
	EventSessionStarted EventType = "session_started"
	EventSessionEnded   EventType = "session_ended"
	EventHTTPTxAdded    EventType = "http_tx_added"
)

// Event — упрощённая модель события для CLI.
type Event struct {
	Type      EventType
	SessionID string
	Ref       string // например, id http-транзакции
}

// EventStream — порт подписки на события из рантайма.
//
// SRP/DIP: CLI зависит только от абстракции, а не от конкретного MonitorHub.
type EventStream interface {
	Subscribe(ctx context.Context) (<-chan Event, func(), error)
}
