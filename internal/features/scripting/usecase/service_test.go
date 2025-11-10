package usecase

import (
	"context"
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

	// Регистрация второго executor
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

	ctx := context.Background()
	script, err := service.GetScript(ctx, "test-id")

	// Mock возвращает nil, но метод должен работать без ошибок
	if err != nil {
		t.Errorf("GetScript() error = %v, want nil", err)
	}

	// script может быть nil из-за mock реализации
	_ = script
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
	scripts  map[string]*domain.Script
	listFunc func(ctx context.Context, filter domain.ScriptFilter) ([]*domain.Script, error)
}

func (m *mockScriptRepository) Save(ctx context.Context, script *domain.Script) error {
	if m.scripts == nil {
		m.scripts = make(map[string]*domain.Script)
	}
	m.scripts[script.ID] = script
	return nil
}

func (m *mockScriptRepository) Get(ctx context.Context, id string) (*domain.Script, error) {
	if m.scripts == nil {
		return nil, nil
	}
	return m.scripts[id], nil
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
	// Возвращаем валидный JSON результат для тестов
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

	// Устанавливаем функцию List
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

	// Не должно паниковать если executor не зарегистрирован
	service.invalidatePluginCache("test-script-id", domain.RuntimeExtism)
}

// Composer 1.
func TestScriptService_InvalidatePluginCache_NoRemoveMethod(t *testing.T) {
	repo := &mockScriptRepository{}
	service := NewScriptService(repo)

	// Executor без метода RemovePlugin
	executor := &mockExecutor{runtime: domain.RuntimeExtism}
	service.RegisterExecutor(executor)

	// Не должно паниковать
	service.invalidatePluginCache("test-script-id", domain.RuntimeExtism)
}
