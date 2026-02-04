package domain

import (
	"time"
)

// HTTPView — representation of a single HTTP transaction for printing.
type HTTPView struct {
	SessionID string
	StartedAt time.Time
	Method    string
	URL       string
	Status    int
	ReqSize   int
	RespSize  int
	TotalMs   int64 // total duration
}
