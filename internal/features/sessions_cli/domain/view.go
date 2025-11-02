package domain

import (
	"time"
)

// HTTPView — представление одной HTTP‑транзакции для печати.
type HTTPView struct {
	SessionID string
	StartedAt time.Time
	Method    string
	URL       string
	Status    int
	ReqSize   int
	RespSize  int
	TotalMs   int64 // общая длительность
}
