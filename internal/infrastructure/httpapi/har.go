package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	stdurl "net/url"
	"os"
	"strings"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
)

// HAR 1.2 specification structures

// ImportMode defines how HAR import should handle existing sessions
type ImportMode string

const (
	// ImportModeMerge adds imported sessions to existing ones (default)
	ImportModeMerge ImportMode = "merge"
	// ImportModeReplaceAll deletes all sessions before import
	ImportModeReplaceAll ImportMode = "replace_all"
	// ImportModeReplaceImported deletes only previously imported sessions
	ImportModeReplaceImported ImportMode = "replace_imported"
)

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Entries []harEntry `json:"entries"`
	Comment string     `json:"comment,omitempty"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

type harEntry struct {
	StartedDateTime   time.Time      `json:"startedDateTime"`
	Time              int64          `json:"time"`
	Request           harRequest     `json:"request"`
	Response          harResponse    `json:"response"`
	Timings           harTimings     `json:"timings"`
	Cache             harCache       `json:"cache"`
	ServerIPAddress   string         `json:"serverIPAddress,omitempty"`
	Connection        string         `json:"connection,omitempty"`
	Comment           string         `json:"comment,omitempty"`
	WebSocketMessages []harWSMessage `json:"_webSocketMessages,omitempty"` // Chrome extension
}

type harRequest struct {
	Method      string         `json:"method"`
	URL         string         `json:"url"`
	HTTPVersion string         `json:"httpVersion"`
	Headers     []harNameValue `json:"headers"`
	QueryString []harNameValue `json:"queryString"`
	Cookies     []harCookie    `json:"cookies"`
	PostData    *harPostData   `json:"postData,omitempty"`
	HeadersSize int64          `json:"headersSize"`
	BodySize    int64          `json:"bodySize"`
	Comment     string         `json:"comment,omitempty"`
}

type harResponse struct {
	Status      int            `json:"status"`
	StatusText  string         `json:"statusText"`
	HTTPVersion string         `json:"httpVersion"`
	Headers     []harNameValue `json:"headers"`
	Cookies     []harCookie    `json:"cookies"`
	Content     harContent     `json:"content"`
	RedirectURL string         `json:"redirectURL"`
	HeadersSize int64          `json:"headersSize"`
	BodySize    int64          `json:"bodySize"`
	Comment     string         `json:"comment,omitempty"`
}

type harNameValue struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

type harCookie struct {
	Name     string     `json:"name"`
	Value    string     `json:"value"`
	Path     string     `json:"path,omitempty"`
	Domain   string     `json:"domain,omitempty"`
	Expires  *time.Time `json:"expires,omitempty"`
	HTTPOnly bool       `json:"httpOnly,omitempty"`
	Secure   bool       `json:"secure,omitempty"`
	Comment  string     `json:"comment,omitempty"`
}

type harPostData struct {
	MimeType string     `json:"mimeType"`
	Params   []harParam `json:"params,omitempty"`
	Text     string     `json:"text"`
	Comment  string     `json:"comment,omitempty"`
}

type harParam struct {
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type harContent struct {
	Size        int64  `json:"size"`
	Compression int64  `json:"compression,omitempty"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

type harTimings struct {
	DNS     int64  `json:"dns"`
	Connect int64  `json:"connect"`
	Blocked int64  `json:"blocked,omitempty"`
	Send    int64  `json:"send"`
	Wait    int64  `json:"wait"`
	Receive int64  `json:"receive"`
	SSL     int64  `json:"ssl,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type harCache struct {
	BeforeRequest *harCacheEntry `json:"beforeRequest,omitempty"`
	AfterRequest  *harCacheEntry `json:"afterRequest,omitempty"`
	Comment       string         `json:"comment,omitempty"`
}

type harCacheEntry struct {
	Expires    string `json:"expires,omitempty"`
	LastAccess string `json:"lastAccess"`
	ETag       string `json:"eTag"`
	HitCount   int    `json:"hitCount"`
	Comment    string `json:"comment,omitempty"`
}

// Chrome DevTools WebSocket extension
type harWSMessage struct {
	Type   string  `json:"type"`   // "send" or "receive"
	Time   float64 `json:"time"`   // Unix timestamp with fractional seconds
	Opcode int     `json:"opcode"` // 1=text, 2=binary
	Data   string  `json:"data"`
}

// HAR export options
type HARExportOptions struct {
	IncludeBodies    bool
	IncludeSensitive bool
	Minify           bool
	SessionIDs       []string // if empty, export all
}

// exportHARForSession exports a single session to HAR format (legacy endpoint)
func exportHARForSession(w http.ResponseWriter, r *http.Request, d *Deps, sessionID string) {
	opts := HARExportOptions{
		IncludeBodies:    true,
		IncludeSensitive: true,
		Minify:           false,
		SessionIDs:       []string{sessionID},
	}
	exportHARWithOptions(w, r, d, opts)
}

// exportHARWithOptions exports HTTP transactions to HAR format with specified options
func exportHARWithOptions(w http.ResponseWriter, r *http.Request, d *Deps, opts HARExportOptions) {
	ctx := r.Context()

	// Collect all sessions if no specific IDs provided
	sessionIDs := opts.SessionIDs
	if len(sessionIDs) == 0 {
		sessions, _, err := d.Svc.List(ctx, usecase.SessionFilter{Limit: 10000})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_LIST_FAILED", err.Error(), nil)
			return
		}
		for _, s := range sessions {
			if s.Kind == "http" {
				sessionIDs = append(sessionIDs, s.ID)
			}
		}
	}

	entries := make([]harEntry, 0, 256)

	// Process each session
	for _, sessionID := range sessionIDs {
		// Get session for WebSocket frames
		session, found, err := d.Svc.Get(ctx, sessionID)
		if err != nil || !found {
			continue
		}

		// Collect HTTP transactions
		from := ""
		for {
			txs, next, err := d.Svc.ListHTTPTransactions(ctx, sessionID, from, 1000)
			if err != nil {
				break
			}

			for _, tx := range txs {
				entry, err := convertToHAREntry(ctx, tx, opts, d)
				if err != nil {
					continue
				}
				entries = append(entries, entry)
			}

			if next == "" {
				break
			}
			from = next
		}

		// Add WebSocket messages if this is a WS session
		if session.Kind == "ws" && len(entries) > 0 {
			wsMessages, err := loadWebSocketMessages(ctx, d, sessionID)
			if err == nil && len(wsMessages) > 0 {
				// Attach WS messages to the last HTTP entry (the upgrade request)
				entries[len(entries)-1].WebSocketMessages = wsMessages
			}
		}
	}

	har := struct {
		Log harLog `json:"log"`
	}{
		Log: harLog{
			Version: "1.2",
			Creator: harCreator{
				Name:    "network-debugger",
				Version: "0.1.6",
				Comment: "Network Debugger - HTTP/WebSocket Inspector",
			},
			Entries: entries,
		},
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	filename := fmt.Sprintf("network-debugger_%s.har", time.Now().Format("2006-01-02_15-04-05"))
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	// Encode
	encoder := json.NewEncoder(w)
	if !opts.Minify {
		encoder.SetIndent("", "  ")
	}
	_ = encoder.Encode(har)
}

func convertToHAREntry(ctx context.Context, tx domain.HTTPTransaction, opts HARExportOptions, d *Deps) (harEntry, error) {
	// Parse URL for query string
	queryString := make([]harNameValue, 0)
	if tx.QueryParams != nil {
		for name, values := range tx.QueryParams {
			for _, value := range values {
				queryString = append(queryString, harNameValue{Name: name, Value: value})
			}
		}
	} else if parsedURL, err := stdurl.Parse(tx.URL); err == nil && parsedURL.RawQuery != "" {
		for name, values := range parsedURL.Query() {
			for _, value := range values {
				queryString = append(queryString, harNameValue{Name: name, Value: value})
			}
		}
	}

	// Request headers
	reqHeaders := convertHeaders(tx.ReqHeaders, opts.IncludeSensitive)

	// Request cookies
	reqCookies := convertCookies(tx.Cookies)

	// Request body
	var postData *harPostData
	reqBodyText := ""
	reqBodySize := int64(tx.ReqSize)
	if opts.IncludeBodies && tx.ReqBodyFile != "" {
		if bodyBytes, err := readBodyFile(tx.ReqBodyFile); err == nil {
			contentType := getHeaderValue(tx.ReqHeaders, "Content-Type")
			reqBodyText = string(bodyBytes)
			reqBodySize = int64(len(bodyBytes))

			if tx.Method == "POST" || tx.Method == "PUT" || tx.Method == "PATCH" {
				postData = &harPostData{
					MimeType: contentType,
					Text:     reqBodyText,
				}
			}
		}
	}

	// Response headers
	respHeaders := convertHeaders(tx.RespHeaders, opts.IncludeSensitive)

	// Response cookies (from Set-Cookie headers)
	respCookies := make([]harCookie, 0)
	if tx.RespHeaders != nil {
		for _, setCookie := range tx.RespHeaders.Values("Set-Cookie") {
			if cookie := parseSetCookie(setCookie); cookie != nil {
				respCookies = append(respCookies, *cookie)
			}
		}
	}

	// Response content
	respBodyText := ""
	respBodySize := int64(tx.RespSize)
	respEncoding := ""
	if opts.IncludeBodies && tx.RespBodyFile != "" {
		if bodyBytes, err := readBodyFile(tx.RespBodyFile); err == nil {
			respBodySize = int64(len(bodyBytes))
			// Check if binary content
			if isBinaryContent(tx.ContentType) {
				respBodyText = base64.StdEncoding.EncodeToString(bodyBytes)
				respEncoding = "base64"
			} else {
				respBodyText = string(bodyBytes)
			}
		}
	}

	content := harContent{
		Size:     respBodySize,
		MimeType: tx.ContentType,
		Text:     respBodyText,
		Encoding: respEncoding,
	}

	// Redirect URL
	redirectURL := ""
	if tx.Status >= 300 && tx.Status < 400 {
		redirectURL = getHeaderValue(tx.RespHeaders, "Location")
	}

	// Timings
	timings := harTimings{
		DNS:     tx.Timings.DNS,
		Connect: tx.Timings.Connect,
		SSL:     tx.Timings.TLS,
		Send:    0, // Not tracked separately
		Wait:    tx.Timings.TTFB,
		Receive: tx.Timings.Total - tx.Timings.TTFB,
	}

	entry := harEntry{
		StartedDateTime: tx.StartedAt,
		Time:            tx.Timings.Total,
		Request: harRequest{
			Method:      tx.Method,
			URL:         tx.URL,
			HTTPVersion: tx.ReqHTTPVersion,
			Headers:     reqHeaders,
			QueryString: queryString,
			Cookies:     reqCookies,
			PostData:    postData,
			HeadersSize: -1,
			BodySize:    reqBodySize,
		},
		Response: harResponse{
			Status:      tx.Status,
			StatusText:  http.StatusText(tx.Status),
			HTTPVersion: tx.RespHTTPVersion,
			Headers:     respHeaders,
			Cookies:     respCookies,
			Content:     content,
			RedirectURL: redirectURL,
			HeadersSize: -1,
			BodySize:    respBodySize,
		},
		Timings: timings,
		Cache:   harCache{}, // Not implemented
	}

	return entry, nil
}

func loadWebSocketMessages(ctx context.Context, d *Deps, sessionID string) ([]harWSMessage, error) {
	messages := make([]harWSMessage, 0)
	from := ""

	for {
		frames, next, err := d.Svc.ListFrames(ctx, sessionID, from, 1000)
		if err != nil {
			break
		}

		for _, frame := range frames {
			msgType := "send"
			if frame.Direction == domain.DirectionUpstreamToClient {
				msgType = "receive"
			}

			opcode := 1 // text
			if frame.Opcode == domain.OpcodeBinary {
				opcode = 2
			}

			// Load full body if available
			data := frame.Preview
			if frame.BodyFile != "" {
				if bodyBytes, err := readBodyFile(frame.BodyFile); err == nil {
					data = string(bodyBytes)
				}
			}

			messages = append(messages, harWSMessage{
				Type:   msgType,
				Time:   float64(frame.Ts.UnixNano()) / 1e9,
				Opcode: opcode,
				Data:   data,
			})
		}

		if next == "" {
			break
		}
		from = next
	}

	return messages, nil
}

func convertHeaders(headers http.Header, includeSensitive bool) []harNameValue {
	if headers == nil {
		return []harNameValue{}
	}

	result := make([]harNameValue, 0, len(headers))
	sensitiveHeaders := map[string]bool{
		"authorization":       true,
		"cookie":              true,
		"set-cookie":          true,
		"proxy-authorization": true,
		"www-authenticate":    true,
	}

	for name, values := range headers {
		if !includeSensitive && sensitiveHeaders[strings.ToLower(name)] {
			result = append(result, harNameValue{
				Name:  name,
				Value: "[REDACTED]",
			})
			continue
		}

		for _, value := range values {
			result = append(result, harNameValue{
				Name:  name,
				Value: value,
			})
		}
	}

	return result
}

func convertCookies(cookies []*http.Cookie) []harCookie {
	if cookies == nil {
		return []harCookie{}
	}

	result := make([]harCookie, 0, len(cookies))
	for _, c := range cookies {
		hc := harCookie{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Domain:   c.Domain,
			HTTPOnly: c.HttpOnly,
			Secure:   c.Secure,
		}
		if !c.Expires.IsZero() {
			hc.Expires = &c.Expires
		}
		result = append(result, hc)
	}

	return result
}

func parseSetCookie(setCookie string) *harCookie {
	// Simple Set-Cookie parser
	parts := strings.Split(setCookie, ";")
	if len(parts) == 0 {
		return nil
	}

	nameValue := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(nameValue) != 2 {
		return nil
	}

	cookie := &harCookie{
		Name:  nameValue[0],
		Value: nameValue[1],
	}

	for i := 1; i < len(parts); i++ {
		attr := strings.TrimSpace(parts[i])
		if strings.EqualFold(attr, "HttpOnly") {
			cookie.HTTPOnly = true
		} else if strings.EqualFold(attr, "Secure") {
			cookie.Secure = true
		} else if strings.HasPrefix(strings.ToLower(attr), "path=") {
			cookie.Path = strings.TrimPrefix(attr, "path=")
			cookie.Path = strings.TrimPrefix(cookie.Path, "Path=")
		} else if strings.HasPrefix(strings.ToLower(attr), "domain=") {
			cookie.Domain = strings.TrimPrefix(attr, "domain=")
			cookie.Domain = strings.TrimPrefix(cookie.Domain, "Domain=")
		}
	}

	return cookie
}

func readBodyFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Limit to 10MB to prevent memory issues
	return io.ReadAll(io.LimitReader(f, 10*1024*1024))
}

func getHeaderValue(headers http.Header, name string) string {
	if headers == nil {
		return ""
	}
	return headers.Get(name)
}

func isBinaryContent(contentType string) bool {
	// Check if content type indicates binary data
	binaryTypes := []string{
		"image/",
		"video/",
		"audio/",
		"application/octet-stream",
		"application/pdf",
		"application/zip",
		"application/gzip",
		"font/",
	}

	contentType = strings.ToLower(contentType)
	for _, bt := range binaryTypes {
		if strings.HasPrefix(contentType, bt) {
			return true
		}
	}

	return false
}
