package httpapi

import (
    "encoding/json"
    "net/http"
    "time"
)

// Minimal HAR 1.2 structs for export
type harLog struct {
    Version string    `json:"version"`
    Creator harName   `json:"creator"`
    Entries []harEntry `json:"entries"`
}
type harName struct{ Name string `json:"name"`; Version string `json:"version"` }
type harEntry struct {
    StartedDateTime time.Time   `json:"startedDateTime"`
    Time            int64       `json:"time"`
    Request         harRequest  `json:"request"`
    Response        harResponse `json:"response"`
}
type harRequest struct {
    Method string `json:"method"`
    URL    string `json:"url"`
    HeadersSize int `json:"headersSize"`
    BodySize    int `json:"bodySize"`
}
type harResponse struct {
    Status int    `json:"status"`
    StatusText string `json:"statusText"`
    HeadersSize int `json:"headersSize"`
    BodySize    int `json:"bodySize"`
}

func exportHARForSession(w http.ResponseWriter, r *http.Request, d *Deps, sessionID string) {
    // collect all http txs
    entries := make([]harEntry, 0, 256)
    from := ""
    for {
        txs, next, err := d.Svc.ListHTTPTransactions(r.Context(), sessionID, from, 1000)
        if err != nil { writeError(w, http.StatusInternalServerError, "HTTP_LIST_FAILED", err.Error(), map[string]any{"id": sessionID}); return }
        for _, tx := range txs {
            entries = append(entries, harEntry{
                StartedDateTime: tx.StartedAt,
                Time:            tx.Timings.Total,
                Request: harRequest{Method: tx.Method, URL: tx.URL, HeadersSize: -1, BodySize: tx.ReqSize},
                Response: harResponse{Status: tx.Status, StatusText: http.StatusText(tx.Status), HeadersSize: -1, BodySize: tx.RespSize},
            })
        }
        if next == "" { break }
        from = next
    }
    har := struct{ Log harLog `json:"log"` }{Log: harLog{Version: "1.2", Creator: harName{Name: "go-proxy", Version: "0.1.0"}, Entries: entries}}
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=wsproxy_session_"+sessionID+".har")
    _ = json.NewEncoder(w).Encode(har)
}


