package domain

import "time"

// HTTPTransaction represents a single HTTP reverse-proxied request/response pair captured by the proxy.
type HTTPTransaction struct {
    ID         string    `json:"id"`
    SessionID  string    `json:"sessionId"`
    Method     string    `json:"method"`
    URL        string    `json:"url"`
    Status     int       `json:"status"`
    ReqSize    int       `json:"reqSize"`
    RespSize   int       `json:"respSize"`
    StartedAt  time.Time `json:"startedAt"`
    EndedAt    time.Time `json:"endedAt"`
    Timings    HTTPTimings `json:"timings"`
    ContentType string   `json:"contentType,omitempty"`
    ReqBodyFile string   `json:"reqBodyFile,omitempty"`
    RespBodyFile string  `json:"respBodyFile,omitempty"`
}

// HTTPTimings captures coarse-grained timing milestones for a transaction.
type HTTPTimings struct {
    DNS      int64 `json:"dnsMs"`      // DNS resolve duration in ms
    Connect  int64 `json:"connectMs"`  // TCP connect duration in ms
    TLS      int64 `json:"tlsMs"`      // TLS handshake duration in ms
    TTFB     int64 `json:"ttfbMs"`     // Time to first byte (headers) in ms
    Total    int64 `json:"totalMs"`    // Total duration in ms (start->headers)
}


