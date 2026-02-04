package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestNewScriptService(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	if service == nil {
		t.Fatal("NewScriptService returned nil")
	}

	if service.repo != repo {
		t.Error("Repository not set correctly")
	}

	if service.executors == nil {
		t.Error("Executors map not initialized")
	}

	if len(service.executors) != 0 {
		t.Errorf("Expected empty executors map, got %d", len(service.executors))
	}
}

// Composer 1.
func TestScriptService_RegisterExecutor(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	if len(service.executors) != 1 {
		t.Errorf("Expected 1 executor, got %d", len(service.executors))
	}

	if service.executors[domain.RuntimeExtism] != executor {
		t.Error("Executor not registered correctly")
	}

	// Register second executor
	executor2 := &mockExecutor{runtime: domain.RuntimeDart}
	service.RegisterExecutor(executor2)

	if len(service.executors) != 2 {
		t.Errorf("Expected 2 executors, got %d", len(service.executors))
	}
}

// Composer 1.
func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		pattern string
		want    bool
	}{
		{"exact match", "/api/test", "/api/test", true},
		{"wildcard match", "/api/users/123", "/api/*", true},
		{"wildcard middle", "/api/users/123/posts", "/api/*/posts", true},
		{"wildcard end", "/api/users/123", "/api/users/*", true},
		{"no match", "/api/test", "/api/other", false},
		{"empty pattern", "/api/test", "", false},
		{"empty value", "", "/api/*", false},
		{"multiple wildcards", "/api/users/123/posts", "/api/*/*", true},
		{"wildcard at start", "test/123", "*test", false},
		{"wildcard only", "anything", "*", true},
		{"wildcard with empty parts", "/api/test", "/api/*/test", false},
		{"pattern starts with wildcard", "test", "*test", true},
		{"pattern ends with wildcard", "test", "test*", true},
		{"multiple consecutive wildcards", "/api/test", "/api/**", true},
		{"wildcard in middle no match", "/api/users", "/api/*/posts", false},
		{"empty value empty pattern", "", "", true},
		{"value shorter than pattern", "a", "ab*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchWildcard(tt.value, tt.pattern)
			if got != tt.want {
				t.Errorf("matchWildcard(%q, %q) = %v, want %v", tt.value, tt.pattern, got, tt.want)
			}
		})
	}
}

// Composer 1.
func TestScriptService_MatchPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	tests := []struct {
		name        string
		value       string
		pattern     string
		patternType domain.PatternType
		want        bool
	}{
		{"exact_match", "test", "test", domain.PatternExact, true},
		{"exact_no_match", "test", "other", domain.PatternExact, false},
		{"prefix_match", "test123", "test", domain.PatternPrefix, true},
		{"prefix_no_match", "other", "test", domain.PatternPrefix, false},
		{"wildcard_match", "/api/users/123", "/api/*", domain.PatternWildcard, true},
		{"regex_match", "test123", "^test", domain.PatternRegex, true},
		{"regex_no_match", "other", "^test", domain.PatternRegex, false},
		{"invalid_regex", "test", "[invalid", domain.PatternRegex, false},
		{"unknown_pattern_type_defaults_to_exact", "test", "test", domain.PatternType("unknown"), true},
		{"unknown_pattern_type_no_match", "test", "other", domain.PatternType("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.matchPattern(tt.value, tt.pattern, tt.patternType)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q, %v) = %v, want %v", tt.value, tt.pattern, tt.patternType, got, tt.want)
			}
		})
	}
}

// Composer 1.
func TestScriptService_Close(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor1 := &mockExecutor{runtime: domain.RuntimeExtism}
	executor2 := &mockExecutor{runtime: domain.RuntimeDart}

	service.RegisterExecutor(executor1)
	service.RegisterExecutor(executor2)

	err := service.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}

	if !executor1.closed {
		t.Error("Executor 1 was not closed")
	}

	if !executor2.closed {
		t.Error("Executor 2 was not closed")
	}
}

// Composer 1.
func TestScriptService_Close_NoExecutors(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	err := service.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// Composer 1.
func TestScriptService_GetScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	script := &domain.Script{
		ID:          "test-id",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}
	repo.scripts = map[string]*domain.Script{
		"test-id": script,
	}

	ctx := context.Background()
	result, err := service.GetScript(ctx, "test-id")

	if err != nil {
		t.Errorf("GetScript() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("GetScript() result should not be nil")
	}
}

// Composer 1.
func TestScriptService_ListScripts(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	ctx := context.Background()
	scripts, err := service.ListScripts(ctx)

	if err != nil {
		t.Errorf("ListScripts() error = %v, want nil", err)
	}

	if scripts == nil {
		t.Error("ListScripts() returned nil slice")
	}
}

// Composer 1.
func TestScriptService_ToggleScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	ctx := context.Background()
	err := service.ToggleScript(ctx, "test-id", true)

	if err != nil {
		t.Errorf("ToggleScript() error = %v, want nil", err)
	}

	err = service.ToggleScript(ctx, "test-id", false)
	if err != nil {
		t.Errorf("ToggleScript() error = %v, want nil", err)
	}
}

// Composer 1.
func TestScriptService_CreateScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.CreateScript(ctx, script)

	if err != nil {
		t.Errorf("CreateScript() error = %v, want nil", err)
	}
}

// Composer 1.
func TestScriptService_CreateScript_NoExecutor(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.CreateScript(ctx, script)

	if err == nil {
		t.Error("CreateScript() should return error when no executor registered")
	}
}

// Composer 1.
func TestScriptService_UpdateScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Updated Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.UpdateScript(ctx, script)

	if err != nil {
		t.Errorf("UpdateScript() error = %v, want nil", err)
	}
}

// Composer 1.
func TestScriptService_DeleteScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	repo.scripts = map[string]*domain.Script{
		"test-script": script,
	}

	ctx := context.Background()
	err := service.DeleteScript(ctx, "test-script")

	if err != nil {
		t.Errorf("DeleteScript() error = %v, want nil", err)
	}
}

// Composer 1.
func TestScriptService_FilterMatchingScripts_NoRules(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules:  domain.MatchRules{},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// Composer 1.
func TestScriptService_FilterMatchingScripts_MethodMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods: []string{"GET", "POST"},
			},
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods: []string{"PUT"},
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}

	if matched[0].ID != "script-1" {
		t.Errorf("Matched script ID = %q, want %q", matched[0].ID, "script-1")
	}
}

// Composer 1.
func TestScriptService_FilterMatchingScripts_PathMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/api/users",
				PatternType: domain.PatternExact,
			},
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/api/posts",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/users",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}

	if matched[0].ID != "script-1" {
		t.Errorf("Matched script ID = %q, want %q", matched[0].ID, "script-1")
	}
}

// Composer 1.
func TestScriptService_FilterMatchingScripts_HostMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "example.com",
				PatternType: domain.PatternExact,
			},
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "other.com",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}

	if matched[0].ID != "script-1" {
		t.Errorf("Matched script ID = %q, want %q", matched[0].ID, "script-1")
	}
}

// Composer 1.
func TestScriptService_TestScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	testReq := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	ctx := context.Background()
	result, logs, err := service.TestScript(ctx, script, testReq)

	if err != nil {
		t.Errorf("TestScript() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("TestScript() result should not be nil")
	}

	if logs == nil {
		t.Error("TestScript() logs should not be nil")
	}
}

// Composer 1.
func TestScriptService_TestScript_NoExecutor(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	testReq := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	ctx := context.Background()
	_, _, err := service.TestScript(ctx, script, testReq)

	if err == nil {
		t.Error("TestScript() should return error when no executor registered")
	}
}

// Mock implementations for testing
type mockScriptRepository struct {
	scripts   map[string]*domain.Script
	listFunc  func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error)
	saveError error
}

func (m *mockScriptRepository) Save(ctx context.Context, script *domain.Script) error {
	if m.saveError != nil {
		return m.saveError
	}
	if m.scripts == nil {
		m.scripts = make(map[string]*domain.Script)
	}
	m.scripts[script.ID] = script
	return nil
}

func (m *mockScriptRepository) Get(ctx context.Context, id string) (*domain.Script, error) {
	if m.scripts == nil {
		return nil, fmt.Errorf("script not found")
	}
	script := m.scripts[id]
	if script == nil {
		return nil, fmt.Errorf("script not found")
	}
	return script, nil
}

func (m *mockScriptRepository) List(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return []*domain.Script{}, nil
}

func (m *mockScriptRepository) Delete(ctx context.Context, id string) error {
	if m.scripts != nil {
		delete(m.scripts, id)
	}
	return nil
}

func (m *mockScriptRepository) UpdateEnabled(ctx context.Context, id string, enabled bool) error {
	if m.scripts != nil && m.scripts[id] != nil {
		m.scripts[id].Enabled = enabled
	}
	return nil
}

type mockExecutor struct {
	runtime domain.ScriptRuntime
	closed  bool
}

func (m *mockExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	// Return valid JSON result for tests
	output := []byte(`{"modified":false}`)
	return domain.ExecutionResult{
		Output: output,
		Logs:   []string{"test log"},
	}, nil
}

func (m *mockExecutor) Runtime() domain.ScriptRuntime {
	return m.runtime
}

func (m *mockExecutor) Validate(ctx context.Context, script domain.Script) error {
	return nil
}

func (m *mockExecutor) Close() error {
	m.closed = true
	return nil
}

type mockExecutorWithPluginRemover struct {
	mockExecutor
	removedPlugins []string
}

func (m *mockExecutorWithPluginRemover) RemovePlugin(scriptID string) {
	m.removedPlugins = append(m.removedPlugins, scriptID)
}

// Composer 1.
func TestScriptService_ExecuteForRequest(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.scripts = map[string]*domain.Script{
		"script-1": script,
	}

	// Set List function
	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// Composer 1.
func TestScriptService_ExecuteForRequest_NoMatchingScripts(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result != req {
		t.Error("ExecuteForRequest() should return original request when no scripts match")
	}
}

// Composer 1.
func TestScriptService_ExecuteForResponse(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerResponse,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerResponse {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForResponse() result should not be nil")
	}
}

// Composer 1.
func TestScriptService_ExecuteForResponse_NoMatchingScripts(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result != resp {
		t.Error("ExecuteForResponse() should return original response when no scripts match")
	}
}

// Composer 1.
func TestScriptService_InvalidatePluginCache(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithPluginRemover{
		mockExecutor: mockExecutor{runtime: domain.RuntimeExtism},
	}
	service.RegisterExecutor(executor)

	scriptID := "test-script-id"
	runtime := domain.RuntimeExtism

	service.invalidatePluginCache(scriptID, runtime)

	if len(executor.removedPlugins) != 1 {
		t.Errorf("Expected 1 removed plugin, got %d", len(executor.removedPlugins))
	}

	if executor.removedPlugins[0] != scriptID {
		t.Errorf("Removed plugin ID = %q, want %q", executor.removedPlugins[0], scriptID)
	}
}

// Composer 1.
func TestScriptService_InvalidatePluginCache_NoExecutor(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	// Should not panic if executor is not registered
	service.invalidatePluginCache("test-script-id", domain.RuntimeExtism)
}

// Composer 1.
func TestScriptService_InvalidatePluginCache_NoRemoveMethod(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	// Executor without RemovePlugin method
	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	// Should not panic
	service.invalidatePluginCache("test-script-id", domain.RuntimeExtism)
}

// TestScriptService_ExecuteForRequest_RepositoryError tests ExecuteForRequest with repository error
func TestScriptService_ExecuteForRequest_RepositoryError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		return nil, fmt.Errorf("repository error")
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	_, err := service.ExecuteForRequest(ctx, req, session)

	if err == nil {
		t.Error("ExecuteForRequest() should return error when repository fails")
	}
}

// TestScriptService_ExecuteForRequest_NoExecutor tests ExecuteForRequest with script without executor
func TestScriptService_ExecuteForRequest_NoExecutor(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// TestScriptService_ExecuteForRequest_ExecutionError tests ExecuteForRequest with execution error
func TestScriptService_ExecuteForRequest_ExecutionError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// TestScriptService_ExecuteForRequest_ScriptError tests ExecuteForRequest with script returning error
func TestScriptService_ExecuteForRequest_ScriptError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithScriptError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// TestScriptService_ExecuteForRequest_ModifiedRequest tests ExecuteForRequest with script modifying request
func TestScriptService_ExecuteForRequest_ModifiedRequest(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	modifiedReq := &domain.HTTPRequest{
		Method: "POST",
		URL:    "https://example.com/api/modified",
		Headers: map[string][]string{
			"X-Custom": {"test"},
		},
		Body: []byte("modified"),
	}

	executor := &mockExecutorWithModifiedRequest{
		mockExecutor: mockExecutor{runtime: domain.RuntimeExtism},
		modifiedReq:  modifiedReq,
	}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}

	if result.Method != modifiedReq.Method {
		t.Errorf("ExecuteForRequest() result.Method = %q, want %q", result.Method, modifiedReq.Method)
	}
}

// TestScriptService_ExecuteForRequest_ParseError tests ExecuteForRequest with parse error
func TestScriptService_ExecuteForRequest_ParseError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithInvalidJSON{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// TestScriptService_ExecuteForRequest_MultipleScripts tests ExecuteForRequest with multiple scripts
func TestScriptService_ExecuteForRequest_MultipleScripts(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script1 := &domain.Script{
		ID:          "script-1",
		Name:        "Script 1",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	script2 := &domain.Script{
		ID:          "script-2",
		Name:        "Script 2",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{4, 5, 6},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script1, script2}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// TestScriptService_ExecuteForResponse_RepositoryError tests ExecuteForResponse with repository error
func TestScriptService_ExecuteForResponse_RepositoryError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		return nil, fmt.Errorf("repository error")
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	_, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err == nil {
		t.Error("ExecuteForResponse() should return error when repository fails")
	}
}

// TestScriptService_ExecuteForResponse_NoExecutor tests ExecuteForResponse with script without executor
func TestScriptService_ExecuteForResponse_NoExecutor(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerResponse,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerResponse {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForResponse() result should not be nil")
	}
}

// TestScriptService_ExecuteForResponse_ExecutionError tests ExecuteForResponse with execution error
func TestScriptService_ExecuteForResponse_ExecutionError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerResponse,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerResponse {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForResponse() result should not be nil")
	}
}

// TestScriptService_ExecuteForResponse_ScriptError tests ExecuteForResponse with script returning error
func TestScriptService_ExecuteForResponse_ScriptError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithScriptError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerResponse,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerResponse {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForResponse() result should not be nil")
	}
}

// TestScriptService_ExecuteForResponse_ModifiedResponse tests ExecuteForResponse with script modifying response
func TestScriptService_ExecuteForResponse_ModifiedResponse(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	modifiedResp := &domain.HTTPResponse{
		Status:  201,
		Headers: map[string][]string{"X-Custom": {"test"}},
		Body:    []byte("modified"),
	}

	executor := &mockExecutorWithModifiedResponse{
		mockExecutor: mockExecutor{runtime: domain.RuntimeExtism},
		modifiedResp: modifiedResp,
	}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerResponse,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerResponse {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForResponse() result should not be nil")
	}

	if result.Status != modifiedResp.Status {
		t.Errorf("ExecuteForResponse() result.Status = %d, want %d", result.Status, modifiedResp.Status)
	}
}

// TestScriptService_ExecuteForResponse_ParseError tests ExecuteForResponse with parse error
func TestScriptService_ExecuteForResponse_ParseError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithInvalidJSON{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerResponse,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerResponse {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	resp := &domain.HTTPResponse{
		Status:  200,
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte("{}"),
	}

	tx := &domain.TransactionInfo{
		ID: "test-tx",
	}

	ctx := context.Background()
	result, err := service.ExecuteForResponse(ctx, req, resp, tx)

	if err != nil {
		t.Errorf("ExecuteForResponse() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForResponse() result should not be nil")
	}
}

// TestScriptService_FilterMatchingScripts_CombinedRules tests filterMatchingScripts with combined rules
func TestScriptService_FilterMatchingScripts_CombinedRules(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods:     []string{"GET"},
				HostPattern: "example.com",
				PathPattern: "/api/users",
				PatternType: domain.PatternExact,
			},
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods:     []string{"POST"},
				HostPattern: "example.com",
				PathPattern: "/api/users",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/users",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}

	if matched[0].ID != "script-1" {
		t.Errorf("Matched script ID = %q, want %q", matched[0].ID, "script-1")
	}
}

// TestScriptService_FilterMatchingScripts_PrefixPattern tests filterMatchingScripts with prefix pattern
func TestScriptService_FilterMatchingScripts_PrefixPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/api",
				PatternType: domain.PatternPrefix,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/users/123",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_WildcardPattern tests filterMatchingScripts with wildcard pattern
func TestScriptService_FilterMatchingScripts_WildcardPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/api/*",
				PatternType: domain.PatternWildcard,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/users/123",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_RegexPattern tests filterMatchingScripts with regex pattern
func TestScriptService_FilterMatchingScripts_RegexPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "^/api/users/\\d+$",
				PatternType: domain.PatternRegex,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/users/123",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_InvalidURL tests filterMatchingScripts with invalid URL
func TestScriptService_FilterMatchingScripts_InvalidURL(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "example.com",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "://invalid-url",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 0 {
		t.Errorf("filterMatchingScripts() length = %d, want 0 for invalid URL", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_RelativeURL tests filterMatchingScripts with relative URL
func TestScriptService_FilterMatchingScripts_RelativeURL(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/api/test",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// Composer 1.
func TestScriptService_FilterMatchingScripts_InvalidURL_DoesNotSkipMethodOnlyScriptAfterHostScript(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Host rule",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "example.com",
				PatternType: domain.PatternExact,
			},
		},
		{
			ID:          "script-2",
			Name:        "Method only",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods:     []string{"GET"},
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "://invalid-url",
	}

	matched := service.filterMatchingScripts(scripts, req)
	if len(matched) != 1 {
		t.Fatalf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
	if matched[0].ID != "script-2" {
		t.Fatalf("matched[0].ID = %q, want %q", matched[0].ID, "script-2")
	}
}

// Composer 1.
func TestScriptService_FilterMatchingScripts_HostPattern_IgnoresPort(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Host with port",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "example.com",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com:443/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)
	if len(matched) != 1 {
		t.Fatalf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
	if matched[0].ID != "script-1" {
		t.Fatalf("matched[0].ID = %q, want %q", matched[0].ID, "script-1")
	}
}

// Mock executors for testing
type mockExecutorWithError struct {
	mockExecutor
}

func (m *mockExecutorWithError) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	return domain.ExecutionResult{}, fmt.Errorf("execution error")
}

type mockExecutorWithScriptError struct {
	mockExecutor
}

func (m *mockExecutorWithScriptError) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	return domain.ExecutionResult{
		Output: []byte(`{"modified":false}`),
		Error:  "script error",
	}, nil
}

type mockExecutorWithInvalidJSON struct {
	mockExecutor
}

func (m *mockExecutorWithInvalidJSON) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	return domain.ExecutionResult{
		Output: []byte("invalid json"),
		Logs:   []string{"test log"},
	}, nil
}

type mockExecutorFunc struct {
	runtime  domain.ScriptRuntime
	execFunc func(context.Context, domain.ExecutionRequest) (domain.ExecutionResult, error)
}

func (m *mockExecutorFunc) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	if m.execFunc == nil {
		return domain.ExecutionResult{Output: []byte(`{"modified":false}`)}, nil
	}
	return m.execFunc(ctx, req)
}

func (m *mockExecutorFunc) Runtime() domain.ScriptRuntime { return m.runtime }
func (m *mockExecutorFunc) Validate(ctx context.Context, script domain.Script) error {
	return nil
}
func (m *mockExecutorFunc) Close() error { return nil }

type mockExecutorWithModifiedRequest struct {
	mockExecutor
	modifiedReq *domain.HTTPRequest
}

func (m *mockExecutorWithModifiedRequest) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	resultJSON, _ := json.Marshal(domain.ScriptResult{
		Modified:        true,
		ModifiedRequest: m.modifiedReq,
	})
	return domain.ExecutionResult{
		Output: resultJSON,
		Logs:   []string{"modified request"},
	}, nil
}

// Composer 1.
func TestScriptService_ExecuteForRequest_SkipsHugeOutputAndContinues(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorFunc{
		runtime: domain.RuntimeExtism,
		execFunc: func(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
			if req.Script.ID == "script-1" {
				huge := make([]byte, 2*1024*1024)
				for i := range huge {
					huge[i] = 'a'
				}
				return domain.ExecutionResult{Output: huge}, nil
			}

			resultJSON, _ := json.Marshal(domain.ScriptResult{
				Modified: true,
				ModifiedRequest: &domain.HTTPRequest{
					Method:  "POST",
					URL:     "https://example.com/api/test",
					Headers: map[string][]string{},
				},
			})
			return domain.ExecutionResult{Output: resultJSON}, nil
		},
	}
	service.RegisterExecutor(executor)

	script1 := &domain.Script{
		ID:          "script-1",
		Name:        "Huge output",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}
	script2 := &domain.Script{
		ID:          "script-2",
		Name:        "Modifier",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{2},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script1, script2}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}
	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	result, err := service.ExecuteForRequest(context.Background(), req, session)
	if err != nil {
		t.Fatalf("ExecuteForRequest error: %v", err)
	}
	if result.Method != "POST" {
		t.Fatalf("expected method POST after second script, got %q", result.Method)
	}
}

type mockExecutorWithModifiedResponse struct {
	mockExecutor
	modifiedResp *domain.HTTPResponse
}

func (m *mockExecutorWithModifiedResponse) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	resultJSON, _ := json.Marshal(domain.ScriptResult{
		Modified:         true,
		ModifiedResponse: m.modifiedResp,
	})
	return domain.ExecutionResult{
		Output: resultJSON,
		Logs:   []string{"modified response"},
	}, nil
}

// TestScriptService_CreateScript_ValidationError tests CreateScript with validation error
func TestScriptService_CreateScript_ValidationError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "", // Invalid: empty name
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.CreateScript(ctx, script)

	if err == nil {
		t.Error("CreateScript() should return error when validation fails")
	}
}

// TestScriptService_CreateScript_ExecutorValidationError tests CreateScript with executor validation error
func TestScriptService_CreateScript_ExecutorValidationError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithValidationError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.CreateScript(ctx, script)

	if err == nil {
		t.Error("CreateScript() should return error when executor validation fails")
	}
}

// TestScriptService_CreateScript_RepositoryError tests CreateScript with repository error
func TestScriptService_CreateScript_RepositoryError(t *testing.T) {
	repo := &mockScriptRepositoryWithError{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.CreateScript(ctx, script)

	if err == nil {
		t.Error("CreateScript() should return error when repository fails")
	}
}

// TestScriptService_UpdateScript_ValidationError tests UpdateScript with validation error
func TestScriptService_UpdateScript_ValidationError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "", // Invalid: empty name
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.UpdateScript(ctx, script)

	if err == nil {
		t.Error("UpdateScript() should return error when validation fails")
	}
}

// TestScriptService_UpdateScript_ExecutorValidationError tests UpdateScript with executor validation error
func TestScriptService_UpdateScript_ExecutorValidationError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithValidationError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.UpdateScript(ctx, script)

	if err == nil {
		t.Error("UpdateScript() should return error when executor validation fails")
	}
}

// TestScriptService_UpdateScript_RepositoryError tests UpdateScript with repository error
func TestScriptService_UpdateScript_RepositoryError(t *testing.T) {
	repo := &mockScriptRepositoryWithError{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	ctx := context.Background()
	err := service.UpdateScript(ctx, script)

	if err == nil {
		t.Error("UpdateScript() should return error when repository fails")
	}
}

// TestScriptService_DeleteScript_NotFound tests DeleteScript with script not found
func TestScriptService_DeleteScript_NotFound(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	ctx := context.Background()
	err := service.DeleteScript(ctx, "nonexistent-id")

	if err == nil {
		t.Error("DeleteScript() should return error when script not found")
	}
}

// TestScriptService_DeleteScript_GetError tests DeleteScript with Get error
func TestScriptService_DeleteScript_GetError(t *testing.T) {
	repo := &mockScriptRepositoryWithError{}
	service := NewScriptService(repo)

	ctx := context.Background()
	err := service.DeleteScript(ctx, "test-id")

	if err == nil {
		t.Error("DeleteScript() should return error when Get fails")
	}
}

// TestScriptService_DeleteScript_RepositoryError tests DeleteScript with repository error
func TestScriptService_DeleteScript_RepositoryError(t *testing.T) {
	repo := &mockScriptRepositoryWithError{}
	service := NewScriptService(repo)

	ctx := context.Background()
	err := service.DeleteScript(ctx, "test-id")

	if err == nil {
		t.Error("DeleteScript() should return error when repository fails")
	}
}

// TestScriptService_TestScript_ValidationError tests TestScript with validation error
func TestScriptService_TestScript_ValidationError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "", // Invalid: empty name
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	testReq := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	ctx := context.Background()
	_, _, err := service.TestScript(ctx, script, testReq)

	if err == nil {
		t.Error("TestScript() should return error when validation fails")
	}
}

// TestScriptService_TestScript_ExecutionError tests TestScript with execution error
func TestScriptService_TestScript_ExecutionError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	testReq := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	ctx := context.Background()
	_, _, err := service.TestScript(ctx, script, testReq)

	if err == nil {
		t.Error("TestScript() should return error when execution fails")
	}
}

// TestScriptService_TestScript_ParseError tests TestScript with parse error
func TestScriptService_TestScript_ParseError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithInvalidJSON{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "test-script",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
	}

	testReq := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	ctx := context.Background()
	_, logs, err := service.TestScript(ctx, script, testReq)

	if err == nil {
		t.Error("TestScript() should return error when parse fails")
	}

	if logs == nil {
		t.Error("TestScript() should return logs even on parse error")
	}
}

// TestScriptService_Close_ExecutorError tests Close with executor error
func TestScriptService_Close_ExecutorError(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutorWithCloseError{mockExecutor: mockExecutor{runtime: domain.RuntimeExtism}}
	service.RegisterExecutor(executor)

	err := service.Close()
	if err != nil {
		t.Errorf("Close() should not return error even if executor fails, got: %v", err)
	}
}

// TestScriptService_FilterMatchingScripts_MethodNoMatch tests filterMatchingScripts with method not matching
func TestScriptService_FilterMatchingScripts_MethodNoMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods: []string{"POST"},
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 0 {
		t.Errorf("filterMatchingScripts() length = %d, want 0", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_HostNoMatch tests filterMatchingScripts with host not matching
func TestScriptService_FilterMatchingScripts_HostNoMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "other.com",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 0 {
		t.Errorf("filterMatchingScripts() length = %d, want 0", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_PathNoMatch tests filterMatchingScripts with path not matching
func TestScriptService_FilterMatchingScripts_PathNoMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/api/other",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 0 {
		t.Errorf("filterMatchingScripts() length = %d, want 0", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_AllRulesMatch tests filterMatchingScripts with all rules matching
func TestScriptService_FilterMatchingScripts_AllRulesMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				Methods:     []string{"GET"},
				HostPattern: "example.com",
				PathPattern: "/api/test",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_URLParseErrorAfterMatch tests filterMatchingScripts with URL parse error after some matches
func TestScriptService_FilterMatchingScripts_URLParseErrorAfterMatch(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules:  domain.MatchRules{}, // No rules, matches everything
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "example.com",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "://invalid-url",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1 (script without host/path rules)", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_EmptyPathFallback tests filterMatchingScripts with empty path fallback
func TestScriptService_FilterMatchingScripts_EmptyPathFallback(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				PathPattern: "/relative",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "/relative",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_MultipleScriptsWithBreak tests filterMatchingScripts with URL parse error breaking loop
func TestScriptService_FilterMatchingScripts_MultipleScriptsWithBreak(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules:  domain.MatchRules{}, // No rules, matches everything
		},
		{
			ID:          "script-2",
			Name:        "Script 2",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{2},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "example.com",
				PatternType: domain.PatternExact,
			},
		},
		{
			ID:          "script-3",
			Name:        "Script 3",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{3},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "other.com",
				PatternType: domain.PatternExact,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "://invalid-url",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1 (only script without host/path rules)", len(matched))
	}

	if matched[0].ID != "script-1" {
		t.Errorf("Matched script ID = %q, want %q", matched[0].ID, "script-1")
	}
}

// TestScriptService_ExecuteForRequest_WithLogs tests ExecuteForRequest with script logs
func TestScriptService_ExecuteForRequest_WithLogs(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	script := &domain.Script{
		ID:          "script-1",
		Name:        "Test Script",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1, 2, 3},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}
}

// TestScriptService_ExecuteForRequest_ChainModifications tests ExecuteForRequest with chained modifications
func TestScriptService_ExecuteForRequest_ChainModifications(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	modifiedReq1 := &domain.HTTPRequest{
		Method: "POST",
		URL:    "https://example.com/api/modified1",
	}

	modifiedReq2 := &domain.HTTPRequest{
		Method: "PUT",
		URL:    "https://example.com/api/modified2",
	}

	executor1 := &mockExecutorWithModifiedRequest{
		mockExecutor: mockExecutor{runtime: domain.RuntimeExtism},
		modifiedReq:  modifiedReq1,
	}
	executor2 := &mockExecutorWithModifiedRequest{
		mockExecutor: mockExecutor{runtime: domain.RuntimeExtism},
		modifiedReq:  modifiedReq2,
	}

	service.RegisterExecutor(executor1)
	service.RegisterExecutor(executor2)

	script1 := &domain.Script{
		ID:          "script-1",
		Name:        "Script 1",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{1},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	script2 := &domain.Script{
		ID:          "script-2",
		Name:        "Script 2",
		Runtime:     domain.RuntimeExtism,
		TriggerType: domain.TriggerRequest,
		Code:        []byte{2},
		Language:    "rust",
		Enabled:     true,
		MatchRules:  domain.MatchRules{},
	}

	repo.listFunc = func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error) {
		if filter.TriggerType == domain.TriggerRequest {
			return []*domain.Script{script1, script2}, nil
		}
		return []*domain.Script{}, nil
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://example.com/api/test",
	}

	session := &domain.SessionInfo{
		ID:         "test-session",
		ClientAddr: "127.0.0.1",
	}

	ctx := context.Background()
	result, err := service.ExecuteForRequest(ctx, req, session)

	if err != nil {
		t.Errorf("ExecuteForRequest() error = %v, want nil", err)
	}

	if result == nil {
		t.Error("ExecuteForRequest() result should not be nil")
	}

	if result.Method != modifiedReq2.Method {
		t.Errorf("ExecuteForRequest() result.Method = %q, want %q (last modification)", result.Method, modifiedReq2.Method)
	}
}

// TestScriptService_FilterMatchingScripts_WildcardHostPattern tests filterMatchingScripts with wildcard host pattern
func TestScriptService_FilterMatchingScripts_WildcardHostPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "*.example.com",
				PatternType: domain.PatternWildcard,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://api.example.com/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_PrefixHostPattern tests filterMatchingScripts with prefix host pattern
func TestScriptService_FilterMatchingScripts_PrefixHostPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "api.",
				PatternType: domain.PatternPrefix,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://api.example.com/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// TestScriptService_FilterMatchingScripts_RegexHostPattern tests filterMatchingScripts with regex host pattern
func TestScriptService_FilterMatchingScripts_RegexHostPattern(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	scripts := []*domain.Script{
		{
			ID:          "script-1",
			Name:        "Script 1",
			Runtime:     domain.RuntimeExtism,
			TriggerType: domain.TriggerRequest,
			Code:        []byte{1},
			Language:    "rust",
			Enabled:     true,
			MatchRules: domain.MatchRules{
				HostPattern: "^api\\..*\\.com$",
				PatternType: domain.PatternRegex,
			},
		},
	}

	req := &domain.HTTPRequest{
		Method: "GET",
		URL:    "https://api.example.com/test",
	}

	matched := service.filterMatchingScripts(scripts, req)

	if len(matched) != 1 {
		t.Errorf("filterMatchingScripts() length = %d, want 1", len(matched))
	}
}

// Additional mock types
type mockExecutorWithValidationError struct {
	mockExecutor
}

func (m *mockExecutorWithValidationError) Validate(ctx context.Context, script domain.Script) error {
	return fmt.Errorf("validation error")
}

type mockScriptRepositoryWithError struct {
	mockScriptRepository
}

func (m *mockScriptRepositoryWithError) Save(ctx context.Context, script *domain.Script) error {
	return fmt.Errorf("repository error")
}

func (m *mockScriptRepositoryWithError) Get(ctx context.Context, id string) (*domain.Script, error) {
	return nil, fmt.Errorf("script not found")
}

type mockExecutorWithCloseError struct {
	mockExecutor
}

func (m *mockExecutorWithCloseError) Close() error {
	return fmt.Errorf("close error")
}
