package httpapi

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

type scriptExecutionRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Runtime     string           `json:"runtime"`
	Code        string           `json:"code"`
	Language    string           `json:"language"`
	TriggerType string           `json:"triggerType"`
	Priority    int              `json:"priority"`
	MatchRules  *matchRulesDTO   `json:"matchRules,omitempty"`
	Config      *scriptConfigDTO `json:"config,omitempty"`
}

type scriptTestHTTPRequest struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type scriptTestRequest struct {
	Script      scriptExecutionRequest `json:"script"`
	TestRequest scriptTestHTTPRequest  `json:"testRequest"`
}

type scriptTestService struct {
	service *ScriptHandlers
}

func newScriptTestService(h *ScriptHandlers) scriptTestService {
	return scriptTestService{service: h}
}

func (s scriptTestService) execute(ctx context.Context, req scriptTestRequest) (map[string]any, *scriptAdminError) {
	codeBytes, err := base64.StdEncoding.DecodeString(req.Script.Code)
	if err != nil {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "invalid base64 code: " + err.Error(),
		}
	}

	script := &domain.Script{
		ID:          "test-script",
		Name:        req.Script.Name,
		Description: req.Script.Description,
		Runtime:     domain.ScriptRuntime(req.Script.Runtime),
		Code:        codeBytes,
		Language:    req.Script.Language,
		TriggerType: domain.TriggerType(req.Script.TriggerType),
		Priority:    req.Script.Priority,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if apiErr := applyScriptMatchRules(script, req.Script.MatchRules); apiErr != nil {
		return nil, apiErr
	}
	applyScriptConfig(script, req.Script.Config)

	testReq := &domain.HTTPRequest{
		Method:  req.TestRequest.Method,
		URL:     req.TestRequest.URL,
		Headers: req.TestRequest.Headers,
		Body:    []byte(req.TestRequest.Body),
	}

	startTime := time.Now()
	result, logs, err := s.service.service.TestScript(ctx, script, testReq)
	duration := time.Since(startTime)

	response := map[string]any{
		"success":    err == nil && (result == nil || result.Error == ""),
		"durationMs": duration.Milliseconds(),
		"logs":       logs,
	}
	if err != nil {
		response["error"] = err.Error()
		return response, nil
	}
	if result == nil {
		return response, nil
	}
	if result.Error != "" {
		response["success"] = false
		response["error"] = result.Error
	}
	if result.Modified && result.ModifiedRequest != nil {
		response["modifiedRequest"] = map[string]any{
			"method":  result.ModifiedRequest.Method,
			"url":     result.ModifiedRequest.URL,
			"headers": result.ModifiedRequest.Headers,
			"body":    string(result.ModifiedRequest.Body),
		}
	}
	if result.Modified && result.ModifiedResponse != nil {
		response["modifiedResponse"] = map[string]any{
			"status":  result.ModifiedResponse.Status,
			"headers": result.ModifiedResponse.Headers,
			"body":    string(result.ModifiedResponse.Body),
		}
	}
	return response, nil
}
