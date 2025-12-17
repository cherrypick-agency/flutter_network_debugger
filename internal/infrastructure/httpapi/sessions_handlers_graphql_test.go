package httpapi

import (
	"context"
	"testing"
	"time"

	"network-debugger/internal/domain"
	uc "network-debugger/internal/usecase"
)

type mockRepoForGraphQL struct {
	frames []domain.Frame
	err    error
}

func (m mockRepoForGraphQL) CreateSession(context.Context, domain.Session) error { return nil }
func (m mockRepoForGraphQL) GetSession(context.Context, string) (domain.Session, bool, error) {
	return domain.Session{}, false, nil
}
func (m mockRepoForGraphQL) DeleteSession(context.Context, string) error { return nil }
func (m mockRepoForGraphQL) ListSessions(context.Context, uc.SessionFilter) ([]domain.Session, int, error) {
	return nil, 0, nil
}
func (m mockRepoForGraphQL) IncrementCounters(context.Context, string, domain.Frame) error {
	return nil
}
func (m mockRepoForGraphQL) SetClosed(context.Context, string, time.Time, *string) error { return nil }
func (m mockRepoForGraphQL) ClearAllSessions(context.Context) error                      { return nil }
func (m mockRepoForGraphQL) DeleteImportedSessions(context.Context) error                { return nil }
func (m mockRepoForGraphQL) AppendFrame(context.Context, string, domain.Frame) error     { return nil }
func (m mockRepoForGraphQL) ListFrames(context.Context, string, string, int) ([]domain.Frame, string, error) {
	return m.frames, "", m.err
}
func (m mockRepoForGraphQL) GetFrameByID(context.Context, string, string) (domain.Frame, bool, error) {
	return domain.Frame{}, false, nil
}
func (m mockRepoForGraphQL) UpdateFrameBodyFile(context.Context, string, string, string) error {
	return nil
}
func (m mockRepoForGraphQL) AppendEvent(context.Context, string, domain.Event) error { return nil }
func (m mockRepoForGraphQL) ListEvents(context.Context, string, string, int) ([]domain.Event, string, error) {
	return nil, "", nil
}
func (m mockRepoForGraphQL) AppendHTTPTransaction(context.Context, domain.HTTPTransaction) error {
	return nil
}
func (m mockRepoForGraphQL) ListHTTPTransactions(context.Context, string, string, int) ([]domain.HTTPTransaction, string, error) {
	return nil, "", nil
}

func TestDetectGraphQLByBody_NoFrames(t *testing.T) {
	repo := mockRepoForGraphQL{frames: nil, err: nil}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session")
	if result {
		t.Error("expected false for session with no frames")
	}
}

func TestDetectGraphQLByBody_GraphQLContentType(t *testing.T) {
	preview := `{"type":"http_request","headers":{"content-type":"application/graphql"},"body":""}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session")
	if !result {
		t.Error("expected true for GraphQL content-type")
	}
}

func TestDetectGraphQLByBody_GraphQLBodyWithQuery(t *testing.T) {
	preview := `{"type":"http_request","headers":{},"body":"{\"query\":\"query { user { name } }\"}"}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session")
	if !result {
		t.Error("expected true for body with 'query' field")
	}
}

func TestDetectGraphQLByBody_GraphQLBodyWithOperationName(t *testing.T) {
	preview := `{"type":"http_request","headers":{},"body":"{\"operationName\":\"GetUser\"}"}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session")
	if !result {
		t.Error("expected true for body with 'operationName' field")
	}
}

func TestDetectGraphQLByBody_NonHTTPRequest(t *testing.T) {
	// Only non-http_request frames - should continue past them and return false
	preview := `{"type":"http_response","headers":{},"body":""}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session-nonhttp")
	// The function skips non-http_request frames and has no http_request to check
	// So it caches false result
	if result {
		t.Error("expected false when no http_request frames")
	}
}

func TestDetectGraphQLByBody_InvalidJSON(t *testing.T) {
	// Invalid JSON frames are skipped via continue
	preview := `not valid json`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session-invalid")
	// Invalid JSON is skipped, no valid http_request found
	if result {
		t.Error("expected false when no valid frames")
	}
}

func TestDetectGraphQLByBody_NoGraphQLIndicators(t *testing.T) {
	// This test should check the latest http_request only
	preview := `{"type":"http_request","headers":{"content-type":"application/json"},"body":"{\"data\":\"test\"}"}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session-nogql")
	// Body doesn't have 'query' or 'operationName', headers don't have graphql
	if result {
		t.Error("expected false for regular JSON without GraphQL indicators")
	}
}

func TestDetectGraphQLByBody_MultipleFrames(t *testing.T) {
	// First frame: not GraphQL
	preview1 := `{"type":"http_request","headers":{},"body":"{\"data\":\"test\"}"}`
	// Last frame: GraphQL
	preview2 := `{"type":"http_request","headers":{},"body":"{\"query\":\"{ user }\"}"}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview1},
			{Preview: preview2},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session")
	if !result {
		t.Error("expected true for GraphQL in last frame")
	}
}

func TestDetectGraphQLByBody_CacheHit(t *testing.T) {
	// Clear cache before test
	gqlBodyCache.Delete("cache-test-session")

	preview := `{"type":"http_request","headers":{},"body":"{\"query\":\"{ user }\"}"}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	// First call - should compute and cache
	result1 := detectGraphQLByBody(context.Background(), d, "cache-test-session")
	if !result1 {
		t.Error("first call: expected true")
	}

	// Second call - should hit cache
	result2 := detectGraphQLByBody(context.Background(), d, "cache-test-session")
	if !result2 {
		t.Error("second call: expected true (from cache)")
	}
}

func TestDetectGraphQLByBody_EmptyBody(t *testing.T) {
	preview := `{"type":"http_request","headers":{"content-type":"application/json"},"body":""}`
	repo := mockRepoForGraphQL{
		frames: []domain.Frame{
			{Preview: preview},
		},
		err: nil,
	}
	s := uc.NewSessionService(repo, repo, repo)
	d := &Deps{Svc: s}

	result := detectGraphQLByBody(context.Background(), d, "test-session")
	if result {
		t.Error("expected false for empty body")
	}
}

func TestDetectGraphQLByBody_CaseInsensitiveContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "lowercase graphql",
			contentType: "application/graphql",
			expected:    true,
		},
		{
			name:        "uppercase GRAPHQL",
			contentType: "APPLICATION/GRAPHQL",
			expected:    true,
		},
		{
			name:        "mixed case GraphQL",
			contentType: "application/GraphQL",
			expected:    true,
		},
		{
			name:        "graphql with charset",
			contentType: "application/graphql; charset=utf-8",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preview := `{"type":"http_request","headers":{"Content-Type":"` + tt.contentType + `"},"body":""}`
			repo := mockRepoForGraphQL{
				frames: []domain.Frame{
					{Preview: preview},
				},
				err: nil,
			}
			s := uc.NewSessionService(repo, repo, repo)
			d := &Deps{Svc: s}

			result := detectGraphQLByBody(context.Background(), d, "test-session-"+tt.name)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
