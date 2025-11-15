package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
)

// handleExportHAR handles POST /_api/v1/export/har
// Accepts JSON body with export options
func (d *Deps) handleExportHAR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var opts HARExportOptions

	if r.Method == http.MethodPost {
		// Parse JSON body
		var body struct {
			SessionIDs       []string `json:"sessionIds"`
			IncludeBodies    *bool    `json:"includeBodies"`
			IncludeSensitive *bool    `json:"includeSensitive"`
			Minify           *bool    `json:"minify"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
			return
		}

		opts.SessionIDs = body.SessionIDs
		opts.IncludeBodies = body.IncludeBodies != nil && *body.IncludeBodies
		opts.IncludeSensitive = body.IncludeSensitive != nil && *body.IncludeSensitive
		opts.Minify = body.Minify != nil && *body.Minify
	} else {
		// Parse query parameters (GET fallback)
		q := r.URL.Query()

		if ids := q.Get("sessionIds"); ids != "" {
			opts.SessionIDs = strings.Split(ids, ",")
		}

		opts.IncludeBodies = parseBoolParam(q.Get("includeBodies"), true)
		opts.IncludeSensitive = parseBoolParam(q.Get("includeSensitive"), false)
		opts.Minify = parseBoolParam(q.Get("minify"), false)
	}

	exportHARWithOptions(w, r, d, opts)
}

// handleImportHAR handles POST /_api/v1/import/har
// Accepts HAR file upload and creates sessions/transactions
func (d *Deps) handleImportHAR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse import mode from query parameter (default: merge)
	modeParam := r.URL.Query().Get("mode")
	mode := ImportModeMerge
	if modeParam != "" {
		mode = ImportMode(modeParam)
		// Validate mode
		if mode != ImportModeMerge && mode != ImportModeReplaceAll && mode != ImportModeReplaceImported {
			writeError(w, http.StatusBadRequest, "INVALID_MODE", fmt.Sprintf("Invalid import mode: %s", modeParam), nil)
			return
		}
	}

	// Limit request body to 500MB
	r.Body = http.MaxBytesReader(w, r.Body, 500*1024*1024)

	// Parse HAR from body
	var harData struct {
		Log harLog `json:"log"`
	}

	if err := json.NewDecoder(r.Body).Decode(&harData); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HAR", "Failed to parse HAR format: "+err.Error(), nil)
		return
	}

	// Validate HAR version
	if harData.Log.Version != "1.2" && harData.Log.Version != "1.1" {
		writeError(w, http.StatusBadRequest, "UNSUPPORTED_VERSION", fmt.Sprintf("HAR version %s not supported", harData.Log.Version), nil)
		return
	}

	// Handle session cleanup based on mode
	ctx := r.Context()
	switch mode {
	case ImportModeReplaceAll:
		if err := d.Svc.ClearAll(ctx); err != nil {
			writeError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to clear sessions: "+err.Error(), nil)
			return
		}
	case ImportModeReplaceImported:
		if err := d.Svc.DeleteImported(ctx); err != nil {
			writeError(w, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete imported sessions: "+err.Error(), nil)
			return
		}
	case ImportModeMerge:
		// No cleanup needed
	}

	// Import entries
	imported := 0
	failed := 0
	sessionIDs := make([]string, 0)

	for _, entry := range harData.Log.Entries {
		sessionID, err := importHAREntry(ctx, d, entry)
		if err != nil {
			failed++
			d.Logger.Warn().Err(err).Msg("Failed to import HAR entry")
			continue
		}
		imported++
		if sessionID != "" {
			sessionIDs = append(sessionIDs, sessionID)
		}
	}

	// Return summary
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"imported":   imported,
		"failed":     failed,
		"total":      len(harData.Log.Entries),
		"sessionIds": sessionIDs,
	})
}

func importHAREntry(ctx context.Context, d *Deps, entry harEntry) (string, error) {
	// Determine session kind based on WebSocket messages
	kind := "http"
	if len(entry.WebSocketMessages) > 0 {
		kind = "ws"
	}

	// Create session for this entry
	sessionID := id.New()

	session := domain.Session{
		ID:         sessionID,
		Target:     entry.Request.URL,
		ClientAddr: "imported",
		StartedAt:  entry.StartedDateTime,
		Kind:       kind,
	}

	// Set closed time if we have end time
	if entry.Time > 0 {
		closedAt := entry.StartedDateTime.Add(time.Duration(entry.Time) * time.Millisecond)
		session.ClosedAt = &closedAt
	}

	if err := d.Svc.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	// Create HTTP transaction
	tx := domain.HTTPTransaction{
		ID:              id.New(),
		SessionID:       sessionID,
		Method:          entry.Request.Method,
		URL:             entry.Request.URL,
		Status:          entry.Response.Status,
		StartedAt:       entry.StartedDateTime,
		EndedAt:         entry.StartedDateTime.Add(time.Duration(entry.Time) * time.Millisecond),
		ContentType:     entry.Response.Content.MimeType,
		ReqSize:         int(entry.Request.BodySize),
		RespSize:        int(entry.Response.BodySize),
		ReqHTTPVersion:  entry.Request.HTTPVersion,
		RespHTTPVersion: entry.Response.HTTPVersion,
		Timings: domain.HTTPTimings{
			DNS:     entry.Timings.DNS,
			Connect: entry.Timings.Connect,
			TLS:     entry.Timings.SSL,
			TTFB:    entry.Timings.Wait,
			Total:   entry.Time,
		},
	}

	// Convert headers
	tx.ReqHeaders = make(http.Header)
	for _, h := range entry.Request.Headers {
		tx.ReqHeaders.Add(h.Name, h.Value)
	}

	tx.RespHeaders = make(http.Header)
	for _, h := range entry.Response.Headers {
		tx.RespHeaders.Add(h.Name, h.Value)
	}

	// Convert cookies
	tx.Cookies = make([]*http.Cookie, 0, len(entry.Request.Cookies))
	for _, c := range entry.Request.Cookies {
		cookie := &http.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			HttpOnly: c.HTTPOnly,
			Secure:   c.Secure,
		}
		if c.Expires != nil {
			cookie.Expires = *c.Expires
		}
		tx.Cookies = append(tx.Cookies, cookie)
	}

	// Convert query params
	tx.QueryParams = make(map[string][]string)
	for _, q := range entry.Request.QueryString {
		tx.QueryParams[q.Name] = append(tx.QueryParams[q.Name], q.Value)
	}

	// Store request body if available
	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		bodyBytes := []byte(entry.Request.PostData.Text)
		if f, err := d.spoolBodyBytes(bodyBytes, "req"); err == nil && f != "" {
			tx.ReqBodyFile = f
			d.Svc.AddSpoolFile(ctx, sessionID, f)
		}
	}

	// Store response body if available
	if entry.Response.Content.Text != "" {
		bodyBytes := []byte(entry.Response.Content.Text)
		// Decode base64 if needed
		if entry.Response.Content.Encoding == "base64" {
			if decoded, err := decodeBase64(entry.Response.Content.Text); err == nil {
				bodyBytes = decoded
			}
		}

		if f, err := d.spoolBodyBytes(bodyBytes, "resp"); err == nil && f != "" {
			tx.RespBodyFile = f
			d.Svc.AddSpoolFile(ctx, sessionID, f)
		}
	}

	// Add transaction
	if err := d.Svc.AddHTTPTransaction(ctx, tx); err != nil {
		return "", fmt.Errorf("add transaction: %w", err)
	}

	// Import WebSocket messages if present
	if len(entry.WebSocketMessages) > 0 {
		for _, msg := range entry.WebSocketMessages {
			opcode := domain.OpcodeText
			if msg.Opcode == 2 {
				opcode = domain.OpcodeBinary
			}

			direction := domain.DirectionClientToUpstream
			if msg.Type == "receive" {
				direction = domain.DirectionUpstreamToClient
			}

			frame := domain.Frame{
				ID:        id.New(),
				Ts:        time.Unix(0, int64(msg.Time*1e9)),
				Direction: direction,
				Opcode:    opcode,
				Size:      len(msg.Data),
				Preview:   msg.Data,
			}

			// Store full body if large
			if len(msg.Data) > 65536 {
				if f, err := d.spoolBodyBytes([]byte(msg.Data), "frame"); err == nil && f != "" {
					frame.BodyFile = f
					frame.Preview = msg.Data[:65536] + "..."
					d.Svc.AddSpoolFile(ctx, sessionID, f)
				}
			}

			_ = d.Svc.AddFrame(ctx, sessionID, frame)
		}
	}

	// Close session
	if session.ClosedAt != nil {
		_ = d.Svc.SetClosed(ctx, sessionID, *session.ClosedAt, nil)
	}

	return sessionID, nil
}

func parseBoolParam(val string, defaultVal bool) bool {
	if val == "" {
		return defaultVal
	}
	if b, err := strconv.ParseBool(val); err == nil {
		return b
	}
	return defaultVal
}

func decodeBase64(s string) ([]byte, error) {
	// Remove whitespace
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")

	// Decode using standard library
	return base64.StdEncoding.DecodeString(s)
}
