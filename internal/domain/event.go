package domain

import "time"

type Event struct {
    ID         string    `json:"id"`
    Ts         time.Time `json:"ts"`
    Namespace  string    `json:"namespace"`
    Name       string    `json:"event"`
    AckID      *int64    `json:"ackId,omitempty"`
    ArgsPreview string   `json:"argsPreview"`
    FrameIDs   []string  `json:"frameIds"`
}


