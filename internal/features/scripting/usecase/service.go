package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"network-debugger/internal/features/scripting/domain"
)

// ScriptService orchestrates script execution across multiple runtimes
// This is the main use case layer following Clean Architecture
type ScriptService struct {
	executors map[domain.ScriptRuntime]domain.ScriptExecutor
	repo      domain.ScriptRepository
}

// NewScriptService creates a new script service
func NewScriptService(repo domain.ScriptRepository) *ScriptService {
	return &ScriptService{
		executors: make(map[domain.ScriptRuntime]domain.ScriptExecutor),
		repo:      repo,
	}
}

// RegisterExecutor adds a runtime executor (Dependency Inversion Principle)
// This allows us to plug in different executors (Extism, Dart, etc.) without changing the service
func (s *ScriptService) RegisterExecutor(executor domain.ScriptExecutor) {
	s.executors[executor.Runtime()] = executor
	log.Printf("[ScriptService] Registered executor: %s", executor.Runtime())
}

// ExecuteForRequest executes all enabled scripts matching the request
// This is called BEFORE the HTTP request is proxied
func (s *ScriptService) ExecuteForRequest(ctx context.Context, req *domain.HTTPRequest, session *domain.SessionInfo) (*domain.HTTPRequest, error) {
	enabled := true
	scripts, err := s.repo.List(ctx, domain.ScriptFilter{
		Enabled:     &enabled,
		TriggerType: domain.TriggerRequest,
	})
	if err != nil {
		return req, fmt.Errorf("failed to list scripts: %w", err)
	}

	// Filter scripts matching this request
	matchedScripts := s.filterMatchingScripts(scripts, req)

	if len(matchedScripts) == 0 {
		return req, nil
	}

	log.Printf("[ScriptService] Executing %d request scripts", len(matchedScripts))

	currentReq := req
	for _, script := range matchedScripts {
		log.Printf("[ScriptService] Processing script %s with runtime %s", script.Name, script.Runtime)
		executor, ok := s.executors[script.Runtime]
		if !ok {
			log.Printf("[ScriptService] No executor for runtime %s, skipping script %s (available executors: %v)", script.Runtime, script.Name, func() []domain.ScriptRuntime {
				runtimes := make([]domain.ScriptRuntime, 0, len(s.executors))
				for rt := range s.executors {
					runtimes = append(runtimes, rt)
				}
				return runtimes
			}())
			continue
		}
		log.Printf("[ScriptService] Found executor for runtime %s", script.Runtime)

		// Prepare script context
		scriptCtx := domain.ScriptContext{
			Request: currentReq,
			Session: session,
		}
		input, _ := json.Marshal(scriptCtx)

		// Execute script
		result, err := executor.Execute(ctx, domain.ExecutionRequest{
			Script: *script,
			Input:  input,
		})

		if err != nil {
			log.Printf("[ScriptService] Script %s execution failed: %v", script.Name, err)
			continue // Don't break chain on error
		}

		if result.Error != "" {
			log.Printf("[ScriptService] Script %s returned error: %s", script.Name, result.Error)
			continue
		}

		// Parse result
		var scriptResult domain.ScriptResult
		if err := json.Unmarshal(result.Output, &scriptResult); err != nil {
			log.Printf("[ScriptService] Failed to parse script result: %v", err)
			continue
		}

		// Apply modifications
		if scriptResult.Modified && scriptResult.ModifiedRequest != nil {
			log.Printf("[ScriptService] Script %s modified request", script.Name)
			currentReq = scriptResult.ModifiedRequest
		}

		// Log script logs
		for _, logMsg := range result.Logs {
			log.Printf("[Script %s] %s", script.Name, logMsg)
		}

		log.Printf("[ScriptService] Script %s executed in %v", script.Name, result.Duration)
	}

	return currentReq, nil
}

// ExecuteForResponse executes all enabled scripts matching the response
// This is called AFTER the HTTP response is received from upstream
func (s *ScriptService) ExecuteForResponse(ctx context.Context, req *domain.HTTPRequest, resp *domain.HTTPResponse, tx *domain.TransactionInfo) (*domain.HTTPResponse, error) {
	enabled := true
	scripts, err := s.repo.List(ctx, domain.ScriptFilter{
		Enabled:     &enabled,
		TriggerType: domain.TriggerResponse,
	})
	if err != nil {
		return resp, fmt.Errorf("failed to list scripts: %w", err)
	}

	matchedScripts := s.filterMatchingScripts(scripts, req)

	if len(matchedScripts) == 0 {
		return resp, nil
	}

	log.Printf("[ScriptService] Executing %d response scripts", len(matchedScripts))

	currentResp := resp
	for _, script := range matchedScripts {
		executor, ok := s.executors[script.Runtime]
		if !ok {
			continue
		}

		scriptCtx := domain.ScriptContext{
			Request:     req,
			Response:    currentResp,
			Transaction: tx,
		}
		input, _ := json.Marshal(scriptCtx)

		result, err := executor.Execute(ctx, domain.ExecutionRequest{
			Script: *script,
			Input:  input,
		})

		if err != nil || result.Error != "" {
			log.Printf("[ScriptService] Script %s failed: %v %s", script.Name, err, result.Error)
			continue
		}

		var scriptResult domain.ScriptResult
		if err := json.Unmarshal(result.Output, &scriptResult); err != nil {
			continue
		}

		if scriptResult.Modified && scriptResult.ModifiedResponse != nil {
			log.Printf("[ScriptService] Script %s modified response", script.Name)
			currentResp = scriptResult.ModifiedResponse
		}

		log.Printf("[ScriptService] Script %s executed in %v", script.Name, result.Duration)
	}

	return currentResp, nil
}

// filterMatchingScripts filters scripts by match rules
func (s *ScriptService) filterMatchingScripts(scripts []*domain.Script, req *domain.HTTPRequest) []*domain.Script {
	var matched []*domain.Script

	for _, script := range scripts {
		// If no match rules defined, script matches everything
		if len(script.MatchRules.Methods) == 0 &&
			script.MatchRules.HostPattern == "" &&
			script.MatchRules.PathPattern == "" {
			matched = append(matched, script)
			continue
		}

		// Check method match
		if len(script.MatchRules.Methods) > 0 {
			methodMatched := false
			for _, m := range script.MatchRules.Methods {
				if m == req.Method {
					methodMatched = true
					break
				}
			}
			if !methodMatched {
				continue
			}
		}

		// Check path pattern match
		if script.MatchRules.PathPattern != "" {
			pathMatched := s.matchPattern(req.URL, script.MatchRules.PathPattern, script.MatchRules.PatternType)
			if !pathMatched {
				continue
			}
		}

		// If we got here, all checks passed
		matched = append(matched, script)
	}

	return matched
}

// matchPattern performs pattern matching based on PatternType
func (s *ScriptService) matchPattern(url string, pattern string, patternType domain.PatternType) bool {
	switch patternType {
	case domain.PatternExact:
		return url == pattern
	case domain.PatternPrefix:
		return len(url) >= len(pattern) && url[:len(pattern)] == pattern
	case "wildcard":
		// Simple wildcard matching: /api/* matches /api/test, /api/users, etc.
		if len(pattern) == 0 {
			return true
		}
		if pattern[len(pattern)-1] == '*' {
			prefix := pattern[:len(pattern)-1]
			return len(url) >= len(prefix) && url[:len(prefix)] == prefix
		}
		return url == pattern
	default:
		return url == pattern
	}
}

// CRUD Operations

// CreateScript creates a new script with validation
func (s *ScriptService) CreateScript(ctx context.Context, script *domain.Script) error {
	// Validate domain rules
	if err := script.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate syntax using appropriate executor
	executor, ok := s.executors[script.Runtime]
	if !ok {
		return fmt.Errorf("no executor for runtime: %s", script.Runtime)
	}

	if err := executor.Validate(ctx, *script); err != nil {
		return fmt.Errorf("script validation failed: %w", err)
	}

	// Save to repository
	return s.repo.Save(ctx, script)
}

// GetScript retrieves a script by ID
func (s *ScriptService) GetScript(ctx context.Context, id string) (*domain.Script, error) {
	return s.repo.Get(ctx, id)
}

// ListScripts retrieves all scripts
func (s *ScriptService) ListScripts(ctx context.Context) ([]*domain.Script, error) {
	return s.repo.List(ctx, domain.ScriptFilter{})
}

// UpdateScript updates an existing script
func (s *ScriptService) UpdateScript(ctx context.Context, script *domain.Script) error {
	if err := script.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	executor, ok := s.executors[script.Runtime]
	if !ok {
		return fmt.Errorf("no executor for runtime: %s", script.Runtime)
	}

	if err := executor.Validate(ctx, *script); err != nil {
		return fmt.Errorf("script validation failed: %w", err)
	}

	if err := s.repo.Save(ctx, script); err != nil {
		return err
	}

	// Invalidate plugin pool when script is updated
	// This ensures next execution uses the updated code
	s.invalidatePluginCache(script.ID, script.Runtime)

	return nil
}

// DeleteScript deletes a script
func (s *ScriptService) DeleteScript(ctx context.Context, id string) error {
	// Get script to determine runtime before deletion
	script, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate plugin pool when script is deleted
	s.invalidatePluginCache(id, script.Runtime)

	return nil
}

// ToggleScript enables or disables a script
func (s *ScriptService) ToggleScript(ctx context.Context, id string, enabled bool) error {
	return s.repo.UpdateEnabled(ctx, id, enabled)
}

// TestScript executes a script with test request data and returns the result
// This allows testing scripts before deploying them
func (s *ScriptService) TestScript(ctx context.Context, script *domain.Script, testReq *domain.HTTPRequest) (*domain.ScriptResult, []string, error) {
	// Validate script
	if err := script.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get executor for runtime
	executor, ok := s.executors[script.Runtime]
	if !ok {
		return nil, nil, fmt.Errorf("no executor for runtime: %s", script.Runtime)
	}

	// Create test script context
	scriptCtx := domain.ScriptContext{
		Request: testReq,
		Session: &domain.SessionInfo{
			ID:         "test-session",
			ClientAddr: "127.0.0.1",
		},
	}

	input, err := json.Marshal(scriptCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal context: %w", err)
	}

	// Execute script
	result, err := executor.Execute(ctx, domain.ExecutionRequest{
		Script: *script,
		Input:  input,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("execution failed: %w", err)
	}

	// Parse script result
	var scriptResult domain.ScriptResult
	if err := json.Unmarshal(result.Output, &scriptResult); err != nil {
		return nil, result.Logs, fmt.Errorf("failed to parse result: %w", err)
	}

	return &scriptResult, result.Logs, nil
}

// Close cleans up all executors
func (s *ScriptService) Close() error {
	for runtime, executor := range s.executors {
		if err := executor.Close(); err != nil {
			log.Printf("[ScriptService] Failed to close %s executor: %v", runtime, err)
		}
	}
	return nil
}

// invalidatePluginCache invalidates the cached plugin for a script
// This is called when a script is updated or deleted to ensure fresh execution
func (s *ScriptService) invalidatePluginCache(scriptID string, runtime domain.ScriptRuntime) {
	executor, ok := s.executors[runtime]
	if !ok {
		return
	}

	// Use type switch to check if executor has RemovePlugin method
	type pluginRemover interface {
		RemovePlugin(string)
	}

	if remover, ok := executor.(pluginRemover); ok {
		remover.RemovePlugin(scriptID)
		log.Printf("[ScriptService] Invalidated plugin cache for script %s", scriptID)
	}
}
