package httpapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

type scriptCreateRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Runtime      string            `json:"runtime"`
	Code         string            `json:"code"`
	SourceCode   string            `json:"sourceCode"`
	Dependencies map[string]string `json:"dependencies"`
	Language     string            `json:"language"`
	TriggerType  string            `json:"triggerType"`
	Priority     int               `json:"priority"`
	Enabled      *bool             `json:"enabled"`
	MatchRules   *matchRulesDTO    `json:"matchRules,omitempty"`
	Config       *scriptConfigDTO  `json:"config,omitempty"`
}

type scriptUpdateRequest struct {
	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	Code         *string            `json:"code,omitempty"`
	SourceCode   *string            `json:"sourceCode,omitempty"`
	Dependencies *map[string]string `json:"dependencies,omitempty"`
	Language     *string            `json:"language,omitempty"`
	TriggerType  *string            `json:"triggerType,omitempty"`
	Priority     *int               `json:"priority,omitempty"`
	Enabled      *bool              `json:"enabled,omitempty"`
	MatchRules   *matchRulesDTO     `json:"matchRules,omitempty"`
	Config       *scriptConfigDTO   `json:"config,omitempty"`
}

type scriptToggleRequest struct {
	Enabled bool `json:"enabled"`
}

type scriptCreateResult struct {
	Script     *domain.Script
	FallbackID string
}

type scriptAdminError struct {
	Status  int
	Message string
}

type scriptAdminService struct {
	service            *usecase.ScriptService
	compilationService *usecase.CompilationService
}

func newScriptAdminService(h *ScriptHandlers) scriptAdminService {
	return scriptAdminService{
		service:            h.service,
		compilationService: h.compilationService,
	}
}

func (s scriptAdminService) create(ctx context.Context, req scriptCreateRequest) (*scriptCreateResult, *scriptAdminError) {
	if req.SourceCode != "" || len(req.Dependencies) > 0 {
		return s.createFromSource(ctx, req)
	}
	if req.Code == "" {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "either 'code', 'sourceCode', or 'dependencies' must be provided",
		}
	}

	codeBytes, err := base64.StdEncoding.DecodeString(req.Code)
	if err != nil {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "invalid base64 code: " + err.Error(),
		}
	}

	script := &domain.Script{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Runtime:     domain.ScriptRuntime(req.Runtime),
		Code:        codeBytes,
		Language:    req.Language,
		TriggerType: domain.TriggerType(req.TriggerType),
		Priority:    req.Priority,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if apiErr := applyScriptMatchRules(script, req.MatchRules); apiErr != nil {
		return nil, apiErr
	}
	applyScriptConfig(script, req.Config)

	if err := s.service.CreateScript(ctx, script); err != nil {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}

	return &scriptCreateResult{Script: script}, nil
}

func (s scriptAdminService) createFromSource(ctx context.Context, req scriptCreateRequest) (*scriptCreateResult, *scriptAdminError) {
	if req.Language == "" {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "language required when using sourceCode or dependencies",
		}
	}
	if s.compilationService == nil {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "compilation service not available - use 'code' field with base64 WASM instead",
		}
	}

	scriptID := uuid.New().String()
	tempScript := &domain.Script{
		ID:           scriptID,
		Name:         req.Name,
		Description:  req.Description,
		Runtime:      domain.ScriptRuntime(req.Runtime),
		SourceCode:   req.SourceCode,
		Dependencies: req.Dependencies,
		Language:     req.Language,
		TriggerType:  domain.TriggerType(req.TriggerType),
		Priority:     req.Priority,
		Code:         []byte{},
		Enabled:      false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.service.CreateScript(ctx, tempScript); err != nil {
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "failed to create script: " + err.Error(),
		}
	}

	compileCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if _, err := s.compilationService.CompileScript(compileCtx, scriptID, true); err != nil {
		if deleteErr := s.service.DeleteScript(ctx, scriptID); deleteErr != nil {
			log.Printf("[ScriptHandlers] WARNING: Failed to cleanup script %s after compilation error: %v", scriptID, deleteErr)
		}
		return nil, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "compilation failed: " + err.Error(),
		}
	}

	if req.Enabled != nil && *req.Enabled {
		if err := s.service.ToggleScript(ctx, scriptID, true); err != nil {
			return nil, &scriptAdminError{
				Status:  http.StatusInternalServerError,
				Message: "script created but failed to enable: " + err.Error(),
			}
		}
	}

	createdScript, err := s.service.GetScript(ctx, scriptID)
	if err != nil {
		return &scriptCreateResult{FallbackID: scriptID}, nil
	}
	return &scriptCreateResult{Script: createdScript}, nil
}

func (s scriptAdminService) list(ctx context.Context) ([]scriptDTO, *scriptAdminError) {
	scripts, err := s.service.ListScripts(ctx)
	if err != nil {
		return nil, &scriptAdminError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	out := make([]scriptDTO, len(scripts))
	for i := range scripts {
		out[i] = toScriptDTO(scripts[i])
	}
	return out, nil
}

func (s scriptAdminService) get(ctx context.Context, id string) (scriptDTO, *scriptAdminError) {
	if id == "" {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "script id required",
		}
	}
	script, err := s.service.GetScript(ctx, id)
	if err != nil {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusNotFound,
			Message: err.Error(),
		}
	}
	return toScriptDTO(script), nil
}

func (s scriptAdminService) update(ctx context.Context, id string, req scriptUpdateRequest) (scriptDTO, *scriptAdminError) {
	if id == "" {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "script id required",
		}
	}

	script, err := s.service.GetScript(ctx, id)
	if err != nil {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusNotFound,
			Message: err.Error(),
		}
	}
	if script.CompilationStatus == domain.CompilationCompiling {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusConflict,
			Message: "cannot update script while compilation is in progress",
		}
	}

	if req.Name != nil {
		script.Name = *req.Name
	}
	if req.Description != nil {
		script.Description = *req.Description
	}
	if req.Code != nil {
		decoded, err := base64.StdEncoding.DecodeString(*req.Code)
		if err != nil {
			return scriptDTO{}, &scriptAdminError{
				Status:  http.StatusBadRequest,
				Message: "invalid base64 code: " + err.Error(),
			}
		}
		script.Code = decoded
	}
	if req.Language != nil {
		script.Language = *req.Language
	}
	if req.TriggerType != nil {
		script.TriggerType = domain.TriggerType(*req.TriggerType)
	}
	if req.Priority != nil {
		script.Priority = *req.Priority
	}
	if req.Enabled != nil {
		script.Enabled = *req.Enabled
	}
	applyScriptConfig(script, req.Config)

	if req.Dependencies != nil {
		totalSize := len(script.SourceCode)
		for filename, content := range *req.Dependencies {
			if len(content) > int(maxDependencyFileSize) {
				return scriptDTO{}, &scriptAdminError{
					Status:  http.StatusBadRequest,
					Message: fmt.Sprintf("dependency %s too large (max %dKB)", filename, maxDependencyFileSize/1024),
				}
			}
			totalSize += len(content)
		}
		if totalSize > int(maxImportProjectSize) {
			return scriptDTO{}, &scriptAdminError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("total project size too large (max %dMB)", maxImportProjectSize/1024/1024),
			}
		}
		if !mapsEqual(script.Dependencies, *req.Dependencies) {
			script.Dependencies = *req.Dependencies
			resetScriptCompilationState(script)
		} else {
			script.Dependencies = *req.Dependencies
		}
	}

	if apiErr := applyScriptMatchRules(script, req.MatchRules); apiErr != nil {
		return scriptDTO{}, apiErr
	}

	if req.SourceCode != nil && *req.SourceCode != script.SourceCode {
		log.Printf("[ScriptHandlers] SourceCode changed for script %s, clearing WASM (requires recompilation)", id)
		script.SourceCode = *req.SourceCode
		resetScriptCompilationState(script)
		log.Printf("[ScriptHandlers] Auto-disabled script %s (must be recompiled and re-enabled)", id)
	}

	script.UpdatedAt = time.Now()

	if err := s.service.UpdateScript(ctx, script); err != nil {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}
	return toScriptDTO(script), nil
}

func (s scriptAdminService) delete(ctx context.Context, id string) *scriptAdminError {
	if id == "" {
		return &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "script id required",
		}
	}
	if err := s.service.DeleteScript(ctx, id); err != nil {
		return &scriptAdminError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return nil
}

func (s scriptAdminService) toggle(ctx context.Context, id string, req scriptToggleRequest) (scriptDTO, *scriptAdminError) {
	if id == "" {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: "script id required",
		}
	}

	if req.Enabled {
		script, err := s.service.GetScript(ctx, id)
		if err != nil {
			return scriptDTO{}, &scriptAdminError{
				Status:  http.StatusNotFound,
				Message: "script not found: " + err.Error(),
			}
		}
		if !hasExecutableScriptCode(script) {
			return scriptDTO{}, &scriptAdminError{
				Status:  http.StatusBadRequest,
				Message: "cannot enable script without executable code",
			}
		}
		if script.CompilationStatus == domain.CompilationStatusError {
			return scriptDTO{}, &scriptAdminError{
				Status:  http.StatusBadRequest,
				Message: "cannot enable script with compilation error",
			}
		}
	}

	if err := s.service.ToggleScript(ctx, id, req.Enabled); err != nil {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	updatedScript, err := s.service.GetScript(ctx, id)
	if err != nil {
		return scriptDTO{}, &scriptAdminError{
			Status:  http.StatusInternalServerError,
			Message: "toggle succeeded but failed to reload script: " + err.Error(),
		}
	}
	return toScriptDTO(updatedScript), nil
}

func applyScriptMatchRules(script *domain.Script, dto *matchRulesDTO) *scriptAdminError {
	if dto == nil {
		return nil
	}
	patternType := dto.PatternType
	if patternType == "" {
		patternType = "wildcard"
	}
	script.MatchRules = domain.MatchRules{
		Methods:     append([]string(nil), dto.Methods...),
		HostPattern: dto.HostPattern,
		PathPattern: dto.PathPattern,
		PatternType: domain.PatternType(patternType),
	}
	if err := validateMatchRulesRegex(script.MatchRules); err != nil {
		return &scriptAdminError{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
	}
	return nil
}

func applyScriptConfig(script *domain.Script, dto *scriptConfigDTO) {
	if dto == nil {
		return
	}
	script.Config.TimeoutMs = dto.TimeoutMs
	script.Config.MemoryLimitMB = dto.MemoryLimitMB
	script.Config.AllowedHosts = dto.AllowedHosts
}

func resetScriptCompilationState(script *domain.Script) {
	script.Code = []byte{}
	script.CompilationStatus = domain.CompilationNotCompiled
	script.CompilationError = ""
	script.ValidationStatus = domain.ValidationNotValidated
	script.ValidationError = ""
	script.Enabled = false
}

func hasExecutableScriptCode(script *domain.Script) bool {
	hasExecutableCode := len(script.Code) > 0
	if script.Runtime == domain.RuntimeDart {
		hasExecutableCode = hasExecutableCode || script.SourceCode != ""
	}
	return hasExecutableCode
}
