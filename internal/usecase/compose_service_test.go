package usecase

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/domain"
)

func TestComposeService_Send_JSON_GET(t *testing.T) {
	t.Parallel()
	// upstream server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// minimal session service deps (stub repos)
	sessRepo := &stubSessionRepo{}
	frameRepo := &stubFrameRepo{}
	eventRepo := &stubEventRepo{}
	sess := NewSessionService(sessRepo, frameRepo, eventRepo)
	// in-memory compose repo (json file repo requires fs). Use noop via memory lib
	lib := &noopLibRepo{}
	hist := &noopHistRepo{}
	s := NewComposeService(lib, hist, sess, func() *http.Client { return srv.Client() })

	res, err := s.Send(context.Background(), ComposeSendCmd{Template: domain.ComposeRequestTemplate{Method: "GET", URL: srv.URL, Body: domain.ComposeBody{Mode: domain.ComposeBodyRaw}}})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if res.Transaction.Status != 200 {
		t.Fatalf("status: %d", res.Transaction.Status)
	}
}

type noopLibRepo struct{}

func (noopLibRepo) LoadLibrary(ctx context.Context) (domain.ComposeLibrary, error) {
	return domain.ComposeLibrary{Collections: []domain.ComposeCollection{}, RequestsByID: map[string]domain.ComposeRequestTemplate{}}, nil
}
func (noopLibRepo) SaveLibrary(ctx context.Context, lib domain.ComposeLibrary) error { return nil }

type noopHistRepo struct{}

func (noopHistRepo) AppendHistory(ctx context.Context, e domain.ComposeHistoryEntry) error {
	return nil
}
func (noopHistRepo) ListHistory(ctx context.Context, from string, limit int) ([]domain.ComposeHistoryEntry, string, error) {
	return nil, "", nil
}

// --- stubs for SessionService deps ---
type stubSessionRepo struct{ sessions map[string]domain.Session }

func (s *stubSessionRepo) CreateSession(ctx context.Context, ss domain.Session) error {
	if s.sessions == nil {
		s.sessions = map[string]domain.Session{}
	}
	s.sessions[ss.ID] = ss
	return nil
}
func (s *stubSessionRepo) GetSession(ctx context.Context, id string) (domain.Session, bool, error) {
	v, ok := s.sessions[id]
	return v, ok, nil
}
func (s *stubSessionRepo) DeleteSession(ctx context.Context, id string) error {
	delete(s.sessions, id)
	return nil
}
func (s *stubSessionRepo) ListSessions(ctx context.Context, f SessionFilter) ([]domain.Session, int, error) {
	out := make([]domain.Session, 0, len(s.sessions))
	for _, v := range s.sessions {
		out = append(out, v)
	}
	return out, len(out), nil
}
func (s *stubSessionRepo) IncrementCounters(ctx context.Context, id string, frame domain.Frame) error {
	return nil
}
func (s *stubSessionRepo) SetClosed(ctx context.Context, id string, t time.Time, errMsg *string) error {
	v := s.sessions[id]
	v.ClosedAt = &t
	v.Error = errMsg
	s.sessions[id] = v
	return nil
}
func (s *stubSessionRepo) ClearAllSessions(ctx context.Context) error {
	s.sessions = map[string]domain.Session{}
	return nil
}

type stubFrameRepo struct{}

func (stubFrameRepo) AppendFrame(ctx context.Context, sessionID string, f domain.Frame) error {
	return nil
}
func (stubFrameRepo) ListFrames(ctx context.Context, sessionID string, from string, limit int) ([]domain.Frame, string, error) {
	return nil, "", nil
}

type stubEventRepo struct{}

func (stubEventRepo) AppendEvent(ctx context.Context, sessionID string, e domain.Event) error {
	return nil
}
func (stubEventRepo) ListEvents(ctx context.Context, sessionID string, from string, limit int) ([]domain.Event, string, error) {
	return nil, "", nil
}

// --- Comprehensive ComposeService tests ---

func TestComposeService_NewComposeService(t *testing.T) {
	t.Parallel()

	t.Run("with custom client factory", func(t *testing.T) {
		t.Parallel()
		customClient := &http.Client{Timeout: 5 * time.Second}
		factory := func() *http.Client { return customClient }

		svc := NewComposeService(&noopLibRepo{}, &noopHistRepo{}, nil, factory)

		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if svc.newClient() != customClient {
			t.Error("custom client factory not used")
		}
		if svc.maxUploadBytes != 10<<20 {
			t.Errorf("maxUploadBytes = %d, want %d", svc.maxUploadBytes, 10<<20)
		}
	})

	t.Run("with nil client factory uses default", func(t *testing.T) {
		t.Parallel()
		svc := NewComposeService(&noopLibRepo{}, &noopHistRepo{}, nil, nil)

		client := svc.newClient()
		if client == nil {
			t.Fatal("expected non-nil default client")
		}
		if client.Timeout != 30*time.Second {
			t.Errorf("default timeout = %v, want 30s", client.Timeout)
		}
	})
}

func TestComposeService_SetMaxUploadMB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"10 MB", 10, 10 << 20},
		{"100 MB", 100, 100 << 20},
		{"zero ignored", 0, 10 << 20}, // default unchanged
		{"negative ignored", -5, 10 << 20},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := NewComposeService(&noopLibRepo{}, &noopHistRepo{}, nil, nil)

			svc.SetMaxUploadMB(tt.input)

			if svc.maxUploadBytes != tt.expected {
				t.Errorf("maxUploadBytes = %d, want %d", svc.maxUploadBytes, tt.expected)
			}
		})
	}
}

func TestComposeService_LoadLibrary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("loads and ensures RequestsByID", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections:  []domain.ComposeCollection{{ID: "c1", Name: "Test"}},
				RequestsByID: nil, // nil should be initialized
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		result, err := svc.LoadLibrary(ctx)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RequestsByID == nil {
			t.Error("RequestsByID should be initialized")
		}
		if len(result.Collections) != 1 {
			t.Errorf("collections count = %d, want 1", len(result.Collections))
		}
	})

	t.Run("error from repo", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{loadErr: context.DeadlineExceeded}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		_, err := svc.LoadLibrary(ctx)

		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestComposeService_SaveLibrary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("saves and ensures RequestsByID", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		input := domain.ComposeLibrary{
			Collections:  []domain.ComposeCollection{{ID: "c1"}},
			RequestsByID: nil,
		}

		err := svc.SaveLibrary(ctx, input)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if lib.lib.RequestsByID == nil {
			t.Error("RequestsByID should be initialized before save")
		}
	})
}

func TestComposeService_UpsertRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("creates new request", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections:  []domain.ComposeCollection{},
				RequestsByID: map[string]domain.ComposeRequestTemplate{},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		tpl := domain.ComposeRequestTemplate{
			ID:     "req1",
			Name:   "Test Request",
			Method: "POST",
			URL:    "https://api.example.com",
		}

		result, err := svc.UpsertRequest(ctx, tpl)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result.RequestsByID["req1"]; !ok {
			t.Error("request not added to RequestsByID")
		}
		if result.RequestsByID["req1"].Method != "POST" {
			t.Error("request data not saved correctly")
		}
	})

	t.Run("updates existing request", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{},
				RequestsByID: map[string]domain.ComposeRequestTemplate{
					"req1": {ID: "req1", Name: "Old", Method: "GET"},
				},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		tpl := domain.ComposeRequestTemplate{
			ID:     "req1",
			Name:   "Updated",
			Method: "PUT",
		}

		result, err := svc.UpsertRequest(ctx, tpl)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RequestsByID["req1"].Name != "Updated" {
			t.Error("request not updated")
		}
		if result.RequestsByID["req1"].Method != "PUT" {
			t.Error("request method not updated")
		}
	})
}

func TestComposeService_DeleteRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("removes from RequestsByID", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{},
				RequestsByID: map[string]domain.ComposeRequestTemplate{
					"req1": {ID: "req1", Name: "To Delete"},
					"req2": {ID: "req2", Name: "Keep"},
				},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		result, err := svc.DeleteRequest(ctx, "req1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := result.RequestsByID["req1"]; ok {
			t.Error("request should be deleted")
		}
		if _, ok := result.RequestsByID["req2"]; !ok {
			t.Error("other request should remain")
		}
	})

	t.Run("removes from folder trees", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{
					{
						ID:   "c1",
						Name: "Collection 1",
						Root: domain.ComposeFolder{
							ID:       "root",
							Name:     "Root",
							Requests: []string{"req1", "req2"},
						},
					},
				},
				RequestsByID: map[string]domain.ComposeRequestTemplate{
					"req1": {ID: "req1"},
					"req2": {ID: "req2"},
				},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		result, err := svc.DeleteRequest(ctx, "req1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Check that req1 is removed from folder
		requests := result.Collections[0].Root.Requests
		if len(requests) != 1 {
			t.Fatalf("requests count = %d, want 1", len(requests))
		}
		if requests[0] != "req2" {
			t.Errorf("remaining request = %s, want req2", requests[0])
		}
	})
}

func TestComposeService_UpsertCollection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("creates new collection", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections:  []domain.ComposeCollection{},
				RequestsByID: map[string]domain.ComposeRequestTemplate{},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		col := domain.ComposeCollection{
			ID:   "c1",
			Name: "New Collection",
			Root: domain.ComposeFolder{ID: "root", Name: "Root"},
		}

		result, err := svc.UpsertCollection(ctx, col)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collections) != 1 {
			t.Fatalf("collections count = %d, want 1", len(result.Collections))
		}
		if result.Collections[0].Name != "New Collection" {
			t.Error("collection not saved")
		}
	})

	t.Run("updates existing collection", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{
					{ID: "c1", Name: "Old Name"},
				},
				RequestsByID: map[string]domain.ComposeRequestTemplate{},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		col := domain.ComposeCollection{
			ID:   "c1",
			Name: "Updated Name",
			Root: domain.ComposeFolder{ID: "root", Name: "Root"},
		}

		result, err := svc.UpsertCollection(ctx, col)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collections) != 1 {
			t.Fatalf("collections count = %d, want 1", len(result.Collections))
		}
		if result.Collections[0].Name != "Updated Name" {
			t.Error("collection not updated")
		}
	})
}

func TestComposeService_DeleteCollection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("removes collection by ID", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{
					{ID: "c1", Name: "Collection 1"},
					{ID: "c2", Name: "Collection 2"},
				},
				RequestsByID: map[string]domain.ComposeRequestTemplate{},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		result, err := svc.DeleteCollection(ctx, "c1")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collections) != 1 {
			t.Fatalf("collections count = %d, want 1", len(result.Collections))
		}
		if result.Collections[0].ID != "c2" {
			t.Error("wrong collection deleted")
		}
	})
}

func TestComposeService_MoveRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("moves request to target folder", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{
					{
						ID:   "c1",
						Name: "Collection",
						Root: domain.ComposeFolder{
							ID:       "root",
							Name:     "Root",
							Requests: []string{"req1", "req2"},
							Folders: []domain.ComposeFolder{
								{ID: "sub1", Name: "Subfolder", Requests: []string{}},
							},
						},
					},
				},
				RequestsByID: map[string]domain.ComposeRequestTemplate{
					"req1": {ID: "req1"},
					"req2": {ID: "req2"},
				},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		result, err := svc.MoveRequest(ctx, "req1", "c1", "sub1", 0)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Check req1 removed from root
		rootReqs := result.Collections[0].Root.Requests
		if len(rootReqs) != 1 || rootReqs[0] != "req2" {
			t.Error("req1 not removed from root")
		}
		// Check req1 added to subfolder
		subReqs := result.Collections[0].Root.Folders[0].Requests
		if len(subReqs) != 1 || subReqs[0] != "req1" {
			t.Error("req1 not added to subfolder")
		}
	})

	t.Run("inserts at specified index", func(t *testing.T) {
		t.Parallel()
		lib := &memoryLibRepo{
			lib: domain.ComposeLibrary{
				Collections: []domain.ComposeCollection{
					{
						ID:   "c1",
						Name: "Collection",
						Root: domain.ComposeFolder{
							ID:       "root",
							Name:     "Root",
							Requests: []string{"req1", "req2", "req3"},
						},
					},
				},
				RequestsByID: map[string]domain.ComposeRequestTemplate{
					"req1": {ID: "req1"},
					"req2": {ID: "req2"},
					"req3": {ID: "req3"},
				},
			},
		}
		svc := NewComposeService(lib, &noopHistRepo{}, nil, nil)

		// Move req1 to position 1 (between req2 and req3)
		result, err := svc.MoveRequest(ctx, "req1", "c1", "root", 1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		reqs := result.Collections[0].Root.Requests
		if len(reqs) != 3 {
			t.Fatalf("requests count = %d, want 3", len(reqs))
		}
		// Expected order: req2, req1, req3
		if reqs[0] != "req2" || reqs[1] != "req1" || reqs[2] != "req3" {
			t.Errorf("order = %v, want [req2, req1, req3]", reqs)
		}
	})
}

func TestComposeService_ListHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns history from repo", func(t *testing.T) {
		t.Parallel()
		hist := &memoryHistRepo{
			entries: []domain.ComposeHistoryEntry{
				{ID: "h1", Method: "GET", URL: "http://a.com", Status: 200},
				{ID: "h2", Method: "POST", URL: "http://b.com", Status: 201},
			},
		}
		svc := NewComposeService(&noopLibRepo{}, hist, nil, nil)

		result, next, err := svc.ListHistory(ctx, "", 10)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("entries count = %d, want 2", len(result))
		}
		if next != "next-token" {
			t.Errorf("next = %s, want next-token", next)
		}
	})

	t.Run("returns empty when histRepo is nil", func(t *testing.T) {
		t.Parallel()
		svc := NewComposeService(&noopLibRepo{}, nil, nil, nil)

		result, next, err := svc.ListHistory(ctx, "", 10)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 0 {
			t.Error("expected empty result")
		}
		if next != "" {
			t.Error("expected empty next token")
		}
	})
}

func TestTLSVersionString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		version  uint16
		expected string
	}{
		{0x0304, "TLS 1.3"},
		{0x0303, "TLS 1.2"},
		{0x0302, "TLS 1.1"},
		{0x0301, "TLS 1.0"},
		{0x0300, ""},
		{0x9999, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tlsVersionString(tt.version)
			if result != tt.expected {
				t.Errorf("tlsVersionString(%#x) = %s, want %s", tt.version, result, tt.expected)
			}
		})
	}
}

func TestCipherSuiteString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		suite    uint16
		expected string
	}{
		{0x1301, "TLS_AES_128_GCM_SHA256"},
		{0x1302, "TLS_AES_256_GCM_SHA384"},
		{0x1303, "TLS_CHACHA20_POLY1305_SHA256"},
		{0xC02F, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		{0x9999, "0x9999"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := cipherSuiteString(tt.suite)
			if result != tt.expected {
				t.Errorf("cipherSuiteString(%#x) = %s, want %s", tt.suite, result, tt.expected)
			}
		})
	}
}

// --- Memory-based test repositories ---

type memoryLibRepo struct {
	lib     domain.ComposeLibrary
	loadErr error
	saveErr error
}

func (m *memoryLibRepo) LoadLibrary(ctx context.Context) (domain.ComposeLibrary, error) {
	if m.loadErr != nil {
		return domain.ComposeLibrary{}, m.loadErr
	}
	return m.lib, nil
}

func (m *memoryLibRepo) SaveLibrary(ctx context.Context, lib domain.ComposeLibrary) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.lib = lib
	return nil
}

type memoryHistRepo struct {
	entries []domain.ComposeHistoryEntry
}

func (m *memoryHistRepo) AppendHistory(ctx context.Context, e domain.ComposeHistoryEntry) error {
	m.entries = append(m.entries, e)
	return nil
}

func (m *memoryHistRepo) ListHistory(ctx context.Context, from string, limit int) ([]domain.ComposeHistoryEntry, string, error) {
	return m.entries, "next-token", nil
}

// --- Comprehensive tests for buildHTTPRequestFromTemplate ---

func TestBuildHTTPRequestFromTemplate_QueryParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		url      string
		query    []domain.ComposeQueryParam
		expected string
	}{
		{
			name:     "no query params",
			url:      "http://api.example.com/users",
			query:    nil,
			expected: "http://api.example.com/users",
		},
		{
			name: "add query params to URL without ?",
			url:  "http://api.example.com/users",
			query: []domain.ComposeQueryParam{
				{Key: "page", Value: "1"},
				{Key: "limit", Value: "10"},
			},
			expected: "http://api.example.com/users?page=1&limit=10",
		},
		{
			name: "add query params to URL with existing ?",
			url:  "http://api.example.com/users?foo=bar",
			query: []domain.ComposeQueryParam{
				{Key: "page", Value: "2"},
			},
			expected: "http://api.example.com/users?foo=bar&page=2",
		},
		{
			name: "skip empty keys",
			url:  "http://api.example.com/test",
			query: []domain.ComposeQueryParam{
				{Key: "", Value: "ignored"},
				{Key: "valid", Value: "ok"},
			},
			expected: "http://api.example.com/test?valid=ok",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tpl := domain.ComposeRequestTemplate{
				Method: "GET",
				URL:    tt.url,
				Query:  tt.query,
			}

			req, err := buildHTTPRequestFromTemplate(tpl, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Parse expected URL to compare properly (query param order doesn't matter)
			expectedURL, err := url.Parse(tt.expected)
			if err != nil {
				t.Fatalf("failed to parse expected URL: %v", err)
			}

			if req.URL.Scheme != expectedURL.Scheme {
				t.Errorf("Scheme = %s, want %s", req.URL.Scheme, expectedURL.Scheme)
			}
			if req.URL.Host != expectedURL.Host {
				t.Errorf("Host = %s, want %s", req.URL.Host, expectedURL.Host)
			}
			if req.URL.Path != expectedURL.Path {
				t.Errorf("Path = %s, want %s", req.URL.Path, expectedURL.Path)
			}
			// Compare query params regardless of order
			if req.URL.Query().Encode() != expectedURL.Query().Encode() {
				t.Errorf("Query = %s, want %s", req.URL.Query().Encode(), expectedURL.Query().Encode())
			}
		})
	}
}

func TestBuildHTTPRequestFromTemplate_BodyModes(t *testing.T) {
	t.Parallel()

	t.Run("raw body", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyRaw,
				Raw:  "plain text content",
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		body, _ := io.ReadAll(req.Body)
		if string(body) != "plain text content" {
			t.Errorf("body = %s, want 'plain text content'", string(body))
		}
	})

	t.Run("json body with auto content-type", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyJSON,
				JSON: `{"name":"test"}`,
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		body, _ := io.ReadAll(req.Body)
		if string(body) != `{"name":"test"}` {
			t.Errorf("body = %s, want JSON", string(body))
		}

		ct := req.Header.Get("Content-Type")
		if ct != "application/json; charset=utf-8" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}
	})

	t.Run("form body", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyForm,
				Form: []domain.ComposeFormField{
					{Key: "username", Value: "john"},
					{Key: "password", Value: "secret"},
				},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		body, _ := io.ReadAll(req.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "username=john") || !strings.Contains(bodyStr, "password=secret") {
			t.Errorf("body = %s, missing form fields", bodyStr)
		}

		ct := req.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %s, want application/x-www-form-urlencoded", ct)
		}
	})

	t.Run("multipart with text field", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyMultipart,
				Multipart: []domain.ComposeMultipartPart{
					{Name: "field1", IsFile: false, Value: "text value"},
				},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ct := req.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			t.Errorf("Content-Type = %s, want multipart/form-data", ct)
		}
	})

	t.Run("multipart with file", func(t *testing.T) {
		t.Parallel()
		// Small text file encoded as base64
		content := "Hello, World!"
		encoded := base64.StdEncoding.EncodeToString([]byte(content))

		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyMultipart,
				Multipart: []domain.ComposeMultipartPart{
					{
						Name:        "upload",
						IsFile:      true,
						FileName:    "test.txt",
						ContentType: "text/plain",
						Base64:      encoded,
					},
				},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 1024*1024) // 1MB limit
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ct := req.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data; boundary=") {
			t.Errorf("Content-Type = %s, want multipart with boundary", ct)
		}
	})

	t.Run("multipart file too large", func(t *testing.T) {
		t.Parallel()
		// Create 100 byte content, set limit to 10
		content := make([]byte, 100)
		encoded := base64.StdEncoding.EncodeToString(content)

		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyMultipart,
				Multipart: []domain.ComposeMultipartPart{
					{
						Name:     "upload",
						IsFile:   true,
						FileName: "big.bin",
						Base64:   encoded,
					},
				},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 10) // 10 byte limit
		if err != nil {
			t.Fatalf("unexpected error creating request: %v", err)
		}

		// Try to read body - should get error from pipe
		_, readErr := io.ReadAll(req.Body)
		if readErr == nil {
			t.Error("expected error when reading oversized file")
		}
		if !strings.Contains(readErr.Error(), "file too large") {
			t.Errorf("error = %v, want 'file too large'", readErr)
		}
	})

	t.Run("multipart with invalid base64", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyMultipart,
				Multipart: []domain.ComposeMultipartPart{
					{
						Name:     "upload",
						IsFile:   true,
						FileName: "test.txt",
						Base64:   "!!!invalid-base64!!!",
					},
				},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 1024)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Reading should fail due to base64 decode error
		_, readErr := io.ReadAll(req.Body)
		if readErr == nil {
			t.Error("expected base64 decode error")
		}
	})
}

func TestBuildHTTPRequestFromTemplate_Headers(t *testing.T) {
	t.Parallel()

	t.Run("custom headers", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Headers: []domain.ComposeHeader{
				{Key: "X-Custom", Value: "value1"},
				{Key: "X-Another", Value: "value2"},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if req.Header.Get("X-Custom") != "value1" {
			t.Error("X-Custom header not set")
		}
		if req.Header.Get("X-Another") != "value2" {
			t.Error("X-Another header not set")
		}
	})

	t.Run("skip headers with empty keys", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Headers: []domain.ComposeHeader{
				{Key: "", Value: "ignored"},
				{Key: "Valid", Value: "ok"},
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if req.Header.Get("Valid") != "ok" {
			t.Error("Valid header not set")
		}
	})

	t.Run("custom content-type overrides auto", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Headers: []domain.ComposeHeader{
				{Key: "Content-Type", Value: "custom/type"},
			},
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyJSON,
				JSON: `{}`,
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		ct := req.Header.Get("Content-Type")
		if ct != "custom/type" {
			t.Errorf("Content-Type = %s, want custom/type", ct)
		}
	})
}

func TestBuildHTTPRequestFromTemplate_Auth(t *testing.T) {
	t.Parallel()

	t.Run("basic auth", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Auth: &domain.ComposeAuth{
				Type:     "Basic",
				Username: "user",
				Password: "pass",
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		auth := req.Header.Get("Authorization")
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
		if auth != expected {
			t.Errorf("Authorization = %s, want %s", auth, expected)
		}
	})

	t.Run("bearer auth", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Auth: &domain.ComposeAuth{
				Type:        "Bearer",
				BearerToken: "my-token-123",
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		auth := req.Header.Get("Authorization")
		if auth != "Bearer my-token-123" {
			t.Errorf("Authorization = %s, want Bearer my-token-123", auth)
		}
	})

	t.Run("api key with default header", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Auth: &domain.ComposeAuth{
				Type:   "ApiKey",
				APIKey: "secret-key",
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		key := req.Header.Get("X-Api-Key")
		if key != "secret-key" {
			t.Errorf("X-Api-Key = %s, want secret-key", key)
		}
	})

	t.Run("api key with custom header", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Auth: &domain.ComposeAuth{
				Type:         "apikey",
				APIKeyHeader: "X-Custom-Key",
				APIKey:       "my-key",
			},
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		key := req.Header.Get("X-Custom-Key")
		if key != "my-key" {
			t.Errorf("X-Custom-Key = %s, want my-key", key)
		}
	})

	t.Run("nil auth", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com",
			Auth:   nil,
		}

		req, err := buildHTTPRequestFromTemplate(tpl, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if req.Header.Get("Authorization") != "" {
			t.Error("Authorization header should not be set")
		}
	})
}

func TestBuildHTTPRequestFromTemplate_InvalidURL(t *testing.T) {
	t.Parallel()

	tpl := domain.ComposeRequestTemplate{
		Method: "GET",
		URL:    "://invalid-url",
	}

	_, err := buildHTTPRequestFromTemplate(tpl, 0)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestMinimalHTTPRequestPreviewFromTemplate(t *testing.T) {
	t.Parallel()

	t.Run("basic GET", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com/users",
		}

		preview := minimalHTTPRequestPreviewFromTemplate(tpl)

		if !strings.Contains(preview, `"type":"http_request"`) {
			t.Error("preview missing type")
		}
		if !strings.Contains(preview, `"method":"GET"`) {
			t.Error("preview missing method")
		}
		if !strings.Contains(preview, "http://api.example.com/users") {
			t.Error("preview missing URL")
		}
	})

	t.Run("with query params", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "GET",
			URL:    "http://api.example.com/search",
			Query: []domain.ComposeQueryParam{
				{Key: "q", Value: "test"},
			},
		}

		preview := minimalHTTPRequestPreviewFromTemplate(tpl)

		if !strings.Contains(preview, "q=test") {
			t.Error("preview missing query param")
		}
	})

	t.Run("with raw body", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyRaw,
				Raw:  "plain text",
			},
		}

		preview := minimalHTTPRequestPreviewFromTemplate(tpl)

		if !strings.Contains(preview, "plain text") {
			t.Error("preview missing raw body")
		}
	})

	t.Run("with json body", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyJSON,
				JSON: `{"key":"value"}`,
			},
		}

		preview := minimalHTTPRequestPreviewFromTemplate(tpl)

		// JSON body is escaped when marshaled, so check for escaped version
		if !strings.Contains(preview, `"body":"{\"key\":\"value\"}"`) && !strings.Contains(preview, `"body":"{"key":"value"}"`) {
			t.Errorf("preview missing JSON body, got: %s", preview)
		}
	})

	t.Run("with form body", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyForm,
				Form: []domain.ComposeFormField{
					{Key: "field1", Value: "value1"},
				},
			},
		}

		preview := minimalHTTPRequestPreviewFromTemplate(tpl)

		if !strings.Contains(preview, `"form"`) {
			t.Error("preview missing form")
		}
		if !strings.Contains(preview, "urlencoded") {
			t.Error("preview missing form type")
		}
	})

	t.Run("with multipart", func(t *testing.T) {
		t.Parallel()
		tpl := domain.ComposeRequestTemplate{
			Method: "POST",
			URL:    "http://api.example.com",
			Body: domain.ComposeBody{
				Mode: domain.ComposeBodyMultipart,
				Multipart: []domain.ComposeMultipartPart{
					{Name: "field1", IsFile: false, Value: "text"},
					{Name: "file1", IsFile: true, FileName: "test.txt"},
				},
			},
		}

		preview := minimalHTTPRequestPreviewFromTemplate(tpl)

		if !strings.Contains(preview, `"multipart"`) {
			t.Error("preview missing multipart")
		}
		if !strings.Contains(preview, `"files"`) {
			t.Error("preview missing files array")
		}
	})
}

func TestMinimalHTTPResponsePreview(t *testing.T) {
	t.Parallel()

	t.Run("basic response", func(t *testing.T) {
		t.Parallel()
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}

		preview := minimalHTTPResponsePreview(resp)

		if !strings.Contains(preview, `"type":"http_response"`) {
			t.Error("preview missing type")
		}
		if !strings.Contains(preview, `"status":200`) {
			t.Error("preview missing status")
		}
		if !strings.Contains(preview, "application/json") {
			t.Error("preview missing content-type")
		}
	})

	t.Run("with cookies", func(t *testing.T) {
		t.Parallel()
		resp := &http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Set-Cookie": []string{
					"session=abc; Secure; HttpOnly; SameSite=Lax",
					"token=xyz; SameSite=Strict",
				},
			},
		}

		preview := minimalHTTPResponsePreview(resp)

		if !strings.Contains(preview, `"cookieSummary"`) {
			t.Error("preview missing cookie summary")
		}
		if !strings.Contains(preview, `"setCookieCount":2`) {
			t.Error("preview missing cookie count")
		}
		if !strings.Contains(preview, `"secure":1`) {
			t.Error("preview missing secure count")
		}
		if !strings.Contains(preview, `"httpOnly":1`) {
			t.Error("preview missing httpOnly count")
		}
		if !strings.Contains(preview, `"sameSiteLax":1`) {
			t.Error("preview missing sameSiteLax count")
		}
		if !strings.Contains(preview, `"sameSiteStrict":1`) {
			t.Error("preview missing sameSiteStrict count")
		}
	})
}
