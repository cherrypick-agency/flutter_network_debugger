package usecase

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"network-debugger/internal/domain"
)

// ComposeService coordinates library CRUD and sending requests.
type ComposeService struct {
	libRepo        ComposeLibraryRepository
	histRepo       ComposeHistoryRepository
	sessions       *SessionService
	newClient      func() *http.Client // factory to create configured HTTP client
	maxUploadBytes int
}

func NewComposeService(lr ComposeLibraryRepository, hr ComposeHistoryRepository, sessions *SessionService, clientFactory func() *http.Client) *ComposeService {
	if clientFactory == nil {
		clientFactory = func() *http.Client { return &http.Client{Timeout: 30 * time.Second} }
	}
	return &ComposeService{libRepo: lr, histRepo: hr, sessions: sessions, newClient: clientFactory, maxUploadBytes: 10 << 20}
}

// SetMaxUploadMB sets the limit for single multipart file in megabytes.
func (s *ComposeService) SetMaxUploadMB(mb int) {
	if mb <= 0 {
		return
	}
	s.maxUploadBytes = mb << 20
}

// Library operations

func (s *ComposeService) LoadLibrary(ctx context.Context) (domain.ComposeLibrary, error) {
	lib, err := s.libRepo.LoadLibrary(ctx)
	if err != nil {
		return domain.ComposeLibrary{}, err
	}
	if lib.RequestsByID == nil {
		lib.RequestsByID = map[string]domain.ComposeRequestTemplate{}
	}
	return lib, nil
}

func (s *ComposeService) SaveLibrary(ctx context.Context, lib domain.ComposeLibrary) error {
	if lib.RequestsByID == nil {
		lib.RequestsByID = map[string]domain.ComposeRequestTemplate{}
	}
	return s.libRepo.SaveLibrary(ctx, lib)
}

// Upsert operations are implemented in service to keep repo simple (snapshot store).
func (s *ComposeService) UpsertRequest(ctx context.Context, tpl domain.ComposeRequestTemplate) (domain.ComposeLibrary, error) {
	lib, err := s.LoadLibrary(ctx)
	if err != nil {
		return domain.ComposeLibrary{}, err
	}
	lib.RequestsByID[tpl.ID] = tpl
	if err := s.SaveLibrary(ctx, lib); err != nil {
		return domain.ComposeLibrary{}, err
	}
	return lib, nil
}

func (s *ComposeService) DeleteRequest(ctx context.Context, id string) (domain.ComposeLibrary, error) {
	lib, err := s.LoadLibrary(ctx)
	if err != nil {
		return domain.ComposeLibrary{}, err
	}
	delete(lib.RequestsByID, id)
	// Also purge from folder trees
	for i := range lib.Collections {
		lib.Collections[i].Root = removeRequestFromFolder(lib.Collections[i].Root, id)
	}
	if err := s.SaveLibrary(ctx, lib); err != nil {
		return domain.ComposeLibrary{}, err
	}
	return lib, nil
}

// UpsertCollection adds or replaces a collection by ID.
func (s *ComposeService) UpsertCollection(ctx context.Context, col domain.ComposeCollection) (domain.ComposeLibrary, error) {
	lib, err := s.LoadLibrary(ctx)
	if err != nil {
		return domain.ComposeLibrary{}, err
	}
	found := false
	for i := range lib.Collections {
		if lib.Collections[i].ID == col.ID {
			lib.Collections[i] = col
			found = true
			break
		}
	}
	if !found {
		lib.Collections = append(lib.Collections, col)
	}
	if err := s.SaveLibrary(ctx, lib); err != nil {
		return domain.ComposeLibrary{}, err
	}
	return lib, nil
}

// DeleteCollection removes a collection by ID.
func (s *ComposeService) DeleteCollection(ctx context.Context, id string) (domain.ComposeLibrary, error) {
	lib, err := s.LoadLibrary(ctx)
	if err != nil {
		return domain.ComposeLibrary{}, err
	}
	out := make([]domain.ComposeCollection, 0, len(lib.Collections))
	for _, c := range lib.Collections {
		if c.ID != id {
			out = append(out, c)
		}
	}
	lib.Collections = out
	if err := s.SaveLibrary(ctx, lib); err != nil {
		return domain.ComposeLibrary{}, err
	}
	return lib, nil
}

// MoveRequest moves a request into a target folder within a collection.
func (s *ComposeService) MoveRequest(ctx context.Context, reqID, collectionID, targetFolderID string, index int) (domain.ComposeLibrary, error) {
	lib, err := s.LoadLibrary(ctx)
	if err != nil {
		return domain.ComposeLibrary{}, err
	}
	// Remove from any existing folder
	for i := range lib.Collections {
		lib.Collections[i].Root = removeRequestFromFolder(lib.Collections[i].Root, reqID)
	}
	// Insert into target
	for i := range lib.Collections {
		if lib.Collections[i].ID == collectionID {
			f, ok := findFolderByID(&lib.Collections[i].Root, targetFolderID)
			if !ok {
				break
			}
			if index < 0 || index > len(f.Requests) {
				index = len(f.Requests)
			}
			newReqs := make([]string, 0, len(f.Requests)+1)
			newReqs = append(newReqs, f.Requests[:index]...)
			newReqs = append(newReqs, reqID)
			newReqs = append(newReqs, f.Requests[index:]...)
			f.Requests = newReqs
			break
		}
	}
	if err := s.SaveLibrary(ctx, lib); err != nil {
		return domain.ComposeLibrary{}, err
	}
	return lib, nil
}

// findFolderByID searches recursively and returns pointer to folder and true if found.
func findFolderByID(root *domain.ComposeFolder, id string) (*domain.ComposeFolder, bool) {
	if root.ID == id {
		return root, true
	}
	for i := range root.Folders {
		if f, ok := findFolderByID(&root.Folders[i], id); ok {
			return f, true
		}
	}
	return nil, false
}

func removeRequestFromFolder(f domain.ComposeFolder, id string) domain.ComposeFolder {
	filtered := make([]string, 0, len(f.Requests))
	for _, rid := range f.Requests {
		if rid != id {
			filtered = append(filtered, rid)
		}
	}
	f.Requests = filtered
	for i := range f.Folders {
		f.Folders[i] = removeRequestFromFolder(f.Folders[i], id)
	}
	return f
}

// ComposeSendCmd is the input for sending an ad-hoc request.
type ComposeSendCmd struct {
	Template domain.ComposeRequestTemplate
	// Optional override of URL/method/headers etc. can be added later.
}

type ComposeSendResult struct {
	SessionID    string                 `json:"sessionId"`
	Transaction  domain.HTTPTransaction `json:"transaction"`
	RespPreview  string                 `json:"respPreview"`
	RespBody     []byte                 `json:"respBody,omitempty"`
	RespBodyFile string                 `json:"respBodyFile,omitempty"`
}

// Send executes the provided template, logs transaction into sessions, and returns response summary.
func (s *ComposeService) Send(ctx context.Context, cmd ComposeSendCmd) (ComposeSendResult, error) {
	if cmd.Template.Method == "" || cmd.Template.URL == "" {
		return ComposeSendResult{}, errors.New("invalid template: method and url are required")
	}
	// Create session
	sess := domain.Session{ID: generateID(), Target: cmd.Template.URL, StartedAt: time.Now().UTC(), Kind: "http"}
	_ = s.sessions.Create(ctx, sess)

	// Build request
	req, err := buildHTTPRequestFromTemplate(cmd.Template, s.maxUploadBytes)
	if err != nil {
		errStr := err.Error()
		_ = s.sessions.SetClosed(ctx, sess.ID, time.Now().UTC(), &errStr)
		return ComposeSendResult{}, err
	}

	var tStart, tDNSStart, tConnStart, tTLSStart, tFirstByte time.Time
	var dnsMs, connMs, tlsMs, ttfbMs int64
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) { tDNSStart = time.Now() },
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !tDNSStart.IsZero() {
				dnsMs = time.Since(tDNSStart).Milliseconds()
			}
		},
		ConnectStart: func(_, _ string) { tConnStart = time.Now() },
		ConnectDone: func(string, string, error) {
			if !tConnStart.IsZero() {
				connMs = time.Since(tConnStart).Milliseconds()
			}
		},
		TLSHandshakeStart: func() { tTLSStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !tTLSStart.IsZero() {
				tlsMs = time.Since(tTLSStart).Milliseconds()
			}
		},
		GotFirstResponseByte: func() {
			if !tStart.IsZero() {
				tFirstByte = time.Now()
				ttfbMs = tFirstByte.Sub(tStart).Milliseconds()
			}
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	client := s.newClient()
	tStart = time.Now()
	// Log request frame (synthetic) so compose session has the same frames as proxy-based sessions
	// Build minimal request preview from template
	func() {
		// best-effort, ignore errors
		prev := minimalHTTPRequestPreviewFromTemplate(cmd.Template)
		size := int(req.ContentLength)
		if size < 0 {
			size = 0
		}
		_ = s.sessions.AddFrame(ctx, sess.ID, domain.Frame{
			ID:        generateID(),
			Ts:        tStart.UTC(),
			Direction: domain.DirectionClientToUpstream,
			Opcode:    domain.OpcodeText,
			Size:      size,
			Preview:   prev,
		})
	}()
	resp, err := client.Do(req)
	if err != nil {
		errStr := err.Error()
		_ = s.sessions.SetClosed(ctx, sess.ID, time.Now().UTC(), &errStr)
		return ComposeSendResult{}, err
	}
	defer resp.Body.Close()

	// Read up to 64KB for preview
	const previewCap = 65536
	bodyBuf, _ := io.ReadAll(io.LimitReader(resp.Body, previewCap))

	tx := domain.HTTPTransaction{
		ID:          generateID(),
		SessionID:   sess.ID,
		Method:      cmd.Template.Method,
		URL:         cmd.Template.URL,
		Status:      resp.StatusCode,
		ReqSize:     int(req.ContentLength),
		RespSize:    int(resp.ContentLength),
		StartedAt:   tStart.UTC(),
		EndedAt:     time.Now().UTC(),
		Timings:     domain.HTTPTimings{DNS: dnsMs, Connect: connMs, TLS: tlsMs, TTFB: ttfbMs, Total: ttfbMs},
		ContentType: resp.Header.Get("Content-Type"),
	}
	_ = s.sessions.AddHTTPTransaction(ctx, tx)
	_ = s.sessions.SetClosed(ctx, sess.ID, time.Now().UTC(), nil)

	if s.histRepo != nil {
		_ = s.histRepo.AppendHistory(ctx, domain.ComposeHistoryEntry{ID: generateID(), Template: cmd.Template.ID, Method: cmd.Template.Method, URL: cmd.Template.URL, Status: resp.StatusCode, When: time.Now().UTC()})
	}

	preview := minimalHTTPResponsePreview(resp)
	// Enrich preview with timings and inline body (preview-sized) for better UX in compose
	if preview != "" {
		var m map[string]any
		if json.Unmarshal([]byte(preview), &m) == nil {
			total := time.Since(tStart).Milliseconds()
			m["timings"] = map[string]any{"ttfbMs": ttfbMs, "totalMs": total}
			// attach body preview as string
			// try compact JSON first to keep it readable
			bstr := string(bodyBuf)
			var jj any
			if json.Unmarshal(bodyBuf, &jj) == nil {
				if jb, err := json.Marshal(jj); err == nil {
					bstr = string(jb)
				}
			}
			if bstr != "" {
				m["body"] = bstr
			}
			if b, err := json.Marshal(m); err == nil {
				preview = string(b)
			}
		}
	}
	// Add response frame
	_ = s.sessions.AddFrame(ctx, sess.ID, domain.Frame{
		ID:        generateID(),
		Ts:        time.Now().UTC(),
		Direction: domain.DirectionUpstreamToClient,
		Opcode:    domain.OpcodeText,
		Size:      int(resp.ContentLength),
		Preview:   preview,
	})
	return ComposeSendResult{SessionID: sess.ID, Transaction: tx, RespPreview: preview, RespBody: bodyBuf}, nil
}

// ListHistory returns last N history entries.
func (s *ComposeService) ListHistory(ctx context.Context, from string, limit int) ([]domain.ComposeHistoryEntry, string, error) {
	if s.histRepo == nil {
		return []domain.ComposeHistoryEntry{}, "", nil
	}
	return s.histRepo.ListHistory(ctx, from, limit)
}

// --- helpers ---

func generateID() string { return time.Now().UTC().Format("20060102150405.000000000") }

func buildHTTPRequestFromTemplate(tpl domain.ComposeRequestTemplate, maxUploadBytes int) (*http.Request, error) {
	urlWithQuery := tpl.URL
	if len(tpl.Query) > 0 {
		q := make(url.Values)
		for _, p := range tpl.Query {
			if p.Key != "" {
				q.Add(p.Key, p.Value)
			}
		}
		sep := "?"
		if strings.Contains(urlWithQuery, "?") {
			sep = "&"
		}
		urlWithQuery = urlWithQuery + sep + q.Encode()
	}
	var body io.Reader
	var mContentType string
	switch tpl.Body.Mode {
	case domain.ComposeBodyRaw:
		body = strings.NewReader(tpl.Body.Raw)
	case domain.ComposeBodyJSON:
		body = strings.NewReader(tpl.Body.JSON)
	case domain.ComposeBodyForm:
		data := make(url.Values)
		for _, f := range tpl.Body.Form {
			data.Add(f.Key, f.Value)
		}
		body = strings.NewReader(data.Encode())
	case domain.ComposeBodyMultipart:
		pr, pw := io.Pipe()
		mw := multipart.NewWriter(pw)
		mContentType = mw.FormDataContentType()
		go func() {
			defer pw.Close()
			defer mw.Close()
			for _, part := range tpl.Body.Multipart {
				if part.IsFile {
					b, err := base64.StdEncoding.DecodeString(part.Base64)
					if err != nil {
						_ = pw.CloseWithError(err)
						return
					}
					if maxUploadBytes > 0 && len(b) > maxUploadBytes {
						_ = pw.CloseWithError(errors.New("file too large"))
						return
					}
					w, err := mw.CreateFormFile(part.Name, part.FileName)
					if err != nil {
						_ = pw.CloseWithError(err)
						return
					}
					if _, err := w.Write(b); err != nil {
						_ = pw.CloseWithError(err)
						return
					}
				} else {
					if err := mw.WriteField(part.Name, part.Value); err != nil {
						_ = pw.CloseWithError(err)
						return
					}
				}
			}
		}()
		body = pr
	}
	req, err := http.NewRequest(tpl.Method, urlWithQuery, body)
	if err != nil {
		return nil, err
	}
	hdr := http.Header{}
	for _, h := range tpl.Headers {
		if h.Key != "" {
			hdr.Add(h.Key, h.Value)
		}
	}
	if req.Body != nil {
		if tpl.Body.Mode == domain.ComposeBodyJSON && hdr.Get("Content-Type") == "" {
			hdr.Set("Content-Type", "application/json; charset=utf-8")
		}
		if tpl.Body.Mode == domain.ComposeBodyForm && hdr.Get("Content-Type") == "" {
			hdr.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		if tpl.Body.Mode == domain.ComposeBodyMultipart && hdr.Get("Content-Type") == "" {
			// include boundary from writer
			if mContentType != "" {
				hdr.Set("Content-Type", mContentType)
			} else {
				hdr.Set("Content-Type", "multipart/form-data")
			}
		}
	}
	if tpl.Auth != nil {
		switch strings.ToLower(tpl.Auth.Type) {
		case "basic":
			token := base64.StdEncoding.EncodeToString([]byte(tpl.Auth.Username + ":" + tpl.Auth.Password))
			hdr.Set("Authorization", "Basic "+token)
		case "bearer":
			hdr.Set("Authorization", "Bearer "+tpl.Auth.BearerToken)
		case "apikey":
			name := tpl.Auth.APIKeyHeader
			if name == "" {
				name = "X-Api-Key"
			}
			hdr.Set(name, tpl.Auth.APIKey)
		}
	}
	req.Header = hdr
	return req, nil
}

// minimalHTTPRequestPreviewFromTemplate builds a compact preview JSON for request frames
func minimalHTTPRequestPreviewFromTemplate(tpl domain.ComposeRequestTemplate) string {
	// Merge query params into URL
	urlWithQuery := tpl.URL
	if len(tpl.Query) > 0 {
		q := make(url.Values)
		for _, p := range tpl.Query {
			if p.Key != "" {
				q.Add(p.Key, p.Value)
			}
		}
		sep := "?"
		if strings.Contains(urlWithQuery, "?") {
			sep = "&"
		}
		urlWithQuery = urlWithQuery + sep + q.Encode()
	}
	// Headers (first value only)
	hdr := map[string]string{}
	for _, h := range tpl.Headers {
		if h.Key != "" {
			hdr[h.Key] = h.Value
		}
	}
	m := map[string]any{
		"type":    "http_request",
		"method":  strings.ToUpper(tpl.Method),
		"url":     urlWithQuery,
		"headers": hdr,
		// headersRaw can be added later if needed
	}
	switch tpl.Body.Mode {
	case domain.ComposeBodyRaw:
		if tpl.Body.Raw != "" {
			m["body"] = tpl.Body.Raw
		}
	case domain.ComposeBodyJSON:
		if tpl.Body.JSON != "" {
			m["body"] = tpl.Body.JSON
		}
	case domain.ComposeBodyForm:
		if len(tpl.Body.Form) > 0 {
			fields := make([]map[string]any, 0, len(tpl.Body.Form))
			for _, f := range tpl.Body.Form {
				fields = append(fields, map[string]any{"name": f.Key, "value": f.Value})
			}
			m["form"] = map[string]any{"type": "urlencoded", "fields": fields}
		}
	case domain.ComposeBodyMultipart:
		if len(tpl.Body.Multipart) > 0 {
			fields := []map[string]any{}
			files := []map[string]any{}
			for _, p := range tpl.Body.Multipart {
				if p.IsFile {
					item := map[string]any{"name": p.Name}
					if p.FileName != "" {
						item["filename"] = p.FileName
					}
					if p.ContentType != "" {
						item["contentType"] = p.ContentType
					}
					files = append(files, item)
				} else {
					fields = append(fields, map[string]any{"name": p.Name, "value": p.Value})
				}
			}
			fm := map[string]any{"type": "multipart"}
			if len(fields) > 0 {
				fm["fields"] = fields
			}
			if len(files) > 0 {
				fm["files"] = files
			}
			m["form"] = fm
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func minimalHTTPResponsePreview(resp *http.Response) string {
	m := map[string]any{"type": "http_response", "status": resp.StatusCode, "headers": map[string]string{}}
	hdr := map[string]string{}
	for k, v := range resp.Header {
		if len(v) > 0 {
			hdr[k] = v[0]
		}
	}
	m["headers"] = hdr
	// TLS summary (best-effort)
	if resp.TLS != nil {
		m["tls"] = map[string]any{
			"version":     tlsVersionString(resp.TLS.Version),
			"cipherSuite": cipherSuiteString(resp.TLS.CipherSuite),
			"alpn":        resp.TLS.NegotiatedProtocol,
			"serverName":  resp.TLS.ServerName,
		}
	}
	// Cookie flags summary (without exposing values)
	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) > 0 {
		var secure, httpOnly, lax, strict, none int
		for _, c := range cookies {
			lc := strings.ToLower(c)
			if strings.Contains(lc, "secure") {
				secure++
			}
			if strings.Contains(lc, "httponly") {
				httpOnly++
			}
			if strings.Contains(lc, "samesite=lax") {
				lax++
			}
			if strings.Contains(lc, "samesite=strict") {
				strict++
			}
			if strings.Contains(lc, "samesite=none") {
				none++
			}
		}
		m["cookieSummary"] = map[string]any{
			"setCookieCount": len(cookies),
			"secure":         secure,
			"httpOnly":       httpOnly,
			"sameSiteLax":    lax,
			"sameSiteStrict": strict,
			"sameSiteNone":   none,
		}
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// minimal helpers for pretty TLS/cipher names (subset)
func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return ""
	}
}

func cipherSuiteString(id uint16) string {
	switch id {
	case tls.TLS_AES_128_GCM_SHA256:
		return "TLS_AES_128_GCM_SHA256"
	case tls.TLS_AES_256_GCM_SHA384:
		return "TLS_AES_256_GCM_SHA384"
	case tls.TLS_CHACHA20_POLY1305_SHA256:
		return "TLS_CHACHA20_POLY1305_SHA256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	default:
		return fmt.Sprintf("0x%04x", id)
	}
}
