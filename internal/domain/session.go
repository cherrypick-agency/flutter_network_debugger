package domain

import "time"

type FrameCounters struct {
	Total   int `json:"total"`
	Text    int `json:"text"`
	Binary  int `json:"binary"`
	Control int `json:"control"`
}

type EventCounters struct {
	Total int `json:"total"`
	SIO   int `json:"sio"`
	Raw   int `json:"raw"`
}

type Session struct {
	ID         string        `json:"id"`
	Target     string        `json:"target"`
	ClientAddr string        `json:"clientAddr"`
	StartedAt  time.Time     `json:"startedAt"`
	ClosedAt   *time.Time    `json:"closedAt"`
	Error      *string       `json:"error"`
	Frames     FrameCounters `json:"frames"`
	Events     EventCounters `json:"events"`
	Evicted    bool          `json:"evicted"`
	Kind       string        `json:"kind"` // "ws" | "http"
	CaptureID  *int          `json:"captureId,omitempty"`
}
