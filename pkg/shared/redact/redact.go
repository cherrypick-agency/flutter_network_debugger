package redact

import (
    "encoding/json"
    "strings"
)

var sensitiveKeys = []string{"authorization", "cookie", "access_token", "id_token", "session", "apikey"}

// RedactJSON masks sensitive fields in a JSON string best-effort.
func RedactJSON(s string) string {
    var v any
    if err := json.Unmarshal([]byte(s), &v); err != nil {
        return s
    }
    redactNode(&v)
    b, err := json.Marshal(v)
    if err != nil { return s }
    return string(b)
}

func redactNode(n *any) {
    switch t := (*n).(type) {
    case map[string]any:
        for k, v := range t {
            if isSensitiveKey(k) {
                t[k] = "***"
                continue
            }
            vv := any(v)
            redactNode(&vv)
            t[k] = vv
        }
    case []any:
        for i := range t {
            vv := any(t[i])
            redactNode(&vv)
            t[i] = vv
        }
    }
}

func isSensitiveKey(k string) bool {
    k = strings.ToLower(k)
    for _, s := range sensitiveKeys {
        if k == s { return true }
    }
    return false
}


