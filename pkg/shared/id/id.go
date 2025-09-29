package id

import (
    "crypto/rand"
    "encoding/hex"
)

func New() string {
    var b [12]byte
    _, _ = rand.Read(b[:])
    return hex.EncodeToString(b[:])
}


