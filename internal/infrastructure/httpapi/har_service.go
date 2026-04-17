package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/domain"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
	"network-debugger/pkg/shared/id"
)

type harImportSummary struct {
	Imported   int
	Failed     int
	Total      int
	SessionIDs []string
}

type harService struct {
	d *Deps
}

func newHARService(d *Deps) harService {
	return harService{d: d}
}

func (s harService) exportLog(ctx context.Context, opts HARExportOptions) (harLog, *sessionAPIError) {
	sessionIDs, apiErr := s.resolveExportSessionIDs(ctx, opts)
	if apiErr != nil {
		return harLog{}, apiErr
	}

	entries := make([]harEntry, 0, 256)
	for _, sessionID := range sessionIDs {
		sessionEntries := s.collectSessionEntries(ctx, sessionID, opts)
		entries = append(entries, sessionEntries...)
	}

	return harLog{
		Version: "1.2",
		Creator: harCreator{
			Name:    "network-debugger",
			Version: obs.Version,
			Comment: "Network Debugger - HTTP/WebSocket Inspector",
		},
		Entries: entries,
	}, nil
}

func (s harService) resolveExportSessionIDs(ctx context.Context, opts HARExportOptions) ([]string, *sessionAPIError) {
	if len(opts.SessionIDs) > 0 {
		return opts.SessionIDs, nil
	}

	sessions, _, err := s.d.Svc.List(ctx, usecase.SessionFilter{Limit: 10000})
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "SESSION_LIST_FAILED",
			Message: err.Error(),
		}
	}

	sessionIDs := make([]string, 0, len(sessions))
	for _, session := range sessions {
		if session.Kind == "http" {
			sessionIDs = append(sessionIDs, session.ID)
		}
	}
	return sessionIDs, nil
}

func (s harService) collectSessionEntries(ctx context.Context, sessionID string, opts HARExportOptions) []harEntry {
	session, found, err := s.d.Svc.Get(ctx, sessionID)
	if err != nil || !found {
		return nil
	}

	entries := make([]harEntry, 0, 8)
	from := ""
	for {
		txs, next, err := s.d.Svc.ListHTTPTransactions(ctx, sessionID, from, 1000)
		if err != nil {
			break
		}
		for _, tx := range txs {
			entry, err := convertToHAREntry(ctx, tx, opts, s.d)
			if err == nil {
				entries = append(entries, entry)
			}
		}
		if next == "" {
			break
		}
		from = next
	}

	if session.Kind == "ws" && len(entries) > 0 {
		wsMessages, err := loadWebSocketMessages(ctx, s.d, sessionID)
		if err == nil && len(wsMessages) > 0 {
			entries[len(entries)-1].WebSocketMessages = wsMessages
		}
	}
	return entries
}

func (s harService) importLog(ctx context.Context, log harLog, mode ImportMode) (harImportSummary, *sessionAPIError) {
	if apiErr := validateHARLog(log); apiErr != nil {
		return harImportSummary{}, apiErr
	}
	if apiErr := s.prepareImport(ctx, mode); apiErr != nil {
		return harImportSummary{}, apiErr
	}

	summary := harImportSummary{
		Total:      len(log.Entries),
		SessionIDs: make([]string, 0, len(log.Entries)),
	}

	for _, entry := range log.Entries {
		sessionID, err := s.importEntry(ctx, entry)
		if err != nil {
			summary.Failed++
			s.d.Logger.Warn().Err(err).Msg("Failed to import HAR entry")
			continue
		}
		summary.Imported++
		if sessionID != "" {
			summary.SessionIDs = append(summary.SessionIDs, sessionID)
		}
	}

	return summary, nil
}

func (s harService) prepareImport(ctx context.Context, mode ImportMode) *sessionAPIError {
	switch mode {
	case ImportModeReplaceAll:
		if err := s.d.clearSessionsAndNotify(ctx); err != nil {
			return &sessionAPIError{
				Status:  http.StatusInternalServerError,
				Code:    "DELETE_FAILED",
				Message: "Failed to clear sessions: " + err.Error(),
			}
		}
	case ImportModeReplaceImported:
		if err := s.d.Svc.DeleteImported(ctx); err != nil {
			return &sessionAPIError{
				Status:  http.StatusInternalServerError,
				Code:    "DELETE_FAILED",
				Message: "Failed to delete imported sessions: " + err.Error(),
			}
		}
	case ImportModeMerge:
	default:
	}
	return nil
}

func (s harService) importEntry(ctx context.Context, entry harEntry) (string, error) {
	kind := "http"
	if len(entry.WebSocketMessages) > 0 {
		kind = "ws"
	}

	sessionID := id.New()
	session := domain.Session{
		ID:         sessionID,
		Target:     entry.Request.URL,
		ClientAddr: "imported",
		StartedAt:  entry.StartedDateTime,
		Kind:       kind,
	}
	if entry.Time > 0 {
		closedAt := entry.StartedDateTime.Add(time.Duration(entry.Time) * time.Millisecond)
		session.ClosedAt = &closedAt
	}
	if err := s.d.Svc.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

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

	tx.ReqHeaders = make(http.Header)
	for _, header := range entry.Request.Headers {
		tx.ReqHeaders.Add(header.Name, header.Value)
	}
	tx.RespHeaders = make(http.Header)
	for _, header := range entry.Response.Headers {
		tx.RespHeaders.Add(header.Name, header.Value)
	}

	tx.Cookies = make([]*http.Cookie, 0, len(entry.Request.Cookies))
	for _, cookie := range entry.Request.Cookies {
		httpCookie := &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			HttpOnly: cookie.HTTPOnly,
			Secure:   cookie.Secure,
		}
		if cookie.Expires != nil {
			httpCookie.Expires = *cookie.Expires
		}
		tx.Cookies = append(tx.Cookies, httpCookie)
	}

	tx.QueryParams = make(map[string][]string)
	for _, query := range entry.Request.QueryString {
		tx.QueryParams[query.Name] = append(tx.QueryParams[query.Name], query.Value)
	}

	if entry.Request.PostData != nil && entry.Request.PostData.Text != "" {
		bodyBytes := []byte(entry.Request.PostData.Text)
		if file, err := s.d.spoolBodyBytes(bodyBytes, "req"); err == nil && file != "" {
			tx.ReqBodyFile = file
			s.d.Svc.AddSpoolFile(ctx, sessionID, file)
		}
	}

	if entry.Response.Content.Text != "" {
		bodyBytes := []byte(entry.Response.Content.Text)
		if entry.Response.Content.Encoding == "base64" {
			if decoded, err := decodeBase64(entry.Response.Content.Text); err == nil {
				bodyBytes = decoded
			}
		}
		if file, err := s.d.spoolBodyBytes(bodyBytes, "resp"); err == nil && file != "" {
			tx.RespBodyFile = file
			s.d.Svc.AddSpoolFile(ctx, sessionID, file)
		}
	}

	if err := s.d.Svc.AddHTTPTransaction(ctx, tx); err != nil {
		return "", fmt.Errorf("add transaction: %w", err)
	}

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
			if len(msg.Data) > 65536 {
				if file, err := s.d.spoolBodyBytes([]byte(msg.Data), "frame"); err == nil && file != "" {
					frame.BodyFile = file
					frame.Preview = msg.Data[:65536] + "..."
					s.d.Svc.AddSpoolFile(ctx, sessionID, file)
				}
			}
			_ = s.d.Svc.AddFrame(ctx, sessionID, frame)
		}
	}

	if session.ClosedAt != nil {
		_ = s.d.Svc.SetClosed(ctx, sessionID, *session.ClosedAt, nil)
	}

	return sessionID, nil
}

func validateHARLog(log harLog) *sessionAPIError {
	if log.Version != "1.2" && log.Version != "1.1" {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "UNSUPPORTED_VERSION",
			Message: fmt.Sprintf("HAR version %s not supported", log.Version),
		}
	}
	return nil
}

func parseHARImportMode(raw string) (ImportMode, *sessionAPIError) {
	if raw == "" {
		return ImportModeMerge, nil
	}
	mode := ImportMode(raw)
	switch mode {
	case ImportModeMerge, ImportModeReplaceAll, ImportModeReplaceImported:
		return mode, nil
	default:
		return "", &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_MODE",
			Message: fmt.Sprintf("Invalid import mode: %s", raw),
		}
	}
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
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")
	return base64.StdEncoding.DecodeString(s)
}
