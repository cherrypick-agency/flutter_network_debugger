package domain

import "time"

type Frame struct {
    ID        string    `json:"id"`
    Ts        time.Time `json:"ts"`
    Direction Direction `json:"direction"`
    Opcode    Opcode    `json:"opcode"`
    Size      int       `json:"size"`
    Preview   string    `json:"preview"`
}


