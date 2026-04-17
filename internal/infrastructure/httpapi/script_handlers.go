package httpapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

const (
	maxUploadFormSize     = 10 << 20        // 10MB for multipart form parsing
	maxSingleFileSize     = 500 * 1024      // 500KB per file in ZIP
	maxTotalProjectSize   = 5 * 1024 * 1024 // 5MB total project size
	maxSourceCodeSize     = 500_000         // 500KB max source code
	maxDependencyFileSize = 500_000         // 500KB max dependency file
	maxImportProjectSize  = 5_000_000       // 5MB max total import size
	maxFilesInZip         = 1000            // max files allowed in ZIP
	maxCompressionRatio   = 100             // max compression ratio (detect ZIP bombs)
)

// DTOs for consistent JSON on frontend (camelCase)
type scriptConfigDTO struct {
	TimeoutMs     int      `json:"timeoutMs"`
	MemoryLimitMB int      `json:"memoryLimitMB"`
	AllowedHosts  []string `json:"allowedHosts"`
}

type matchRulesDTO struct {
	Methods     []string `json:"methods"`
	HostPattern string   `json:"hostPattern"`
	PathPattern string   `json:"pathPattern"`
	PatternType string   `json:"patternType"`
}

type scriptDTO struct {
	ID                string                   `json:"id"`
	Name              string                   `json:"name"`
	Description       string                   `json:"description"`
	Runtime           domain.ScriptRuntime     `json:"runtime"`
	Code              string                   `json:"code"`
	Language          string                   `json:"language"`
	TriggerType       domain.TriggerType       `json:"triggerType"`
	Priority          int                      `json:"priority"`
	Enabled           bool                     `json:"enabled"`
	MatchRules        matchRulesDTO            `json:"matchRules"`
	Config            scriptConfigDTO          `json:"config"`
	CreatedAt         time.Time                `json:"createdAt"`
	UpdatedAt         time.Time                `json:"updatedAt"`
	SourceCode        string                   `json:"sourceCode"`
	Dependencies      map[string]string        `json:"dependencies"`
	CompilationStatus domain.CompilationStatus `json:"compilationStatus"`
	CompilationError  string                   `json:"compilationError"`
	LastCompiledAt    *time.Time               `json:"lastCompiledAt"`
	ValidationStatus  domain.ValidationStatus  `json:"validationStatus"`
	ValidationError   string                   `json:"validationError,omitempty"`
}

func toScriptDTO(s *domain.Script) scriptDTO {
	var codeStr string
	if len(s.Code) > 0 {
		codeStr = base64.StdEncoding.EncodeToString(s.Code)
	} else {
		codeStr = ""
	}
	// Ensure correct default values
	pt := string(s.MatchRules.PatternType)
	if pt == "" {
		pt = "wildcard"
	}
	methods := []string{}
	if len(s.MatchRules.Methods) > 0 {
		methods = append(methods, s.MatchRules.Methods...)
	}
	allowedHosts := []string{}
	if len(s.Config.AllowedHosts) > 0 {
		allowedHosts = append(allowedHosts, s.Config.AllowedHosts...)
	}
	return scriptDTO{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Runtime:     s.Runtime,
		Code:        codeStr,
		Language:    s.Language,
		TriggerType: s.TriggerType,
		Priority:    s.Priority,
		Enabled:     s.Enabled,
		MatchRules: matchRulesDTO{
			Methods:     methods,
			HostPattern: s.MatchRules.HostPattern,
			PathPattern: s.MatchRules.PathPattern,
			PatternType: pt,
		},
		Config: scriptConfigDTO{
			TimeoutMs:     s.Config.TimeoutMs,
			MemoryLimitMB: s.Config.MemoryLimitMB,
			AllowedHosts:  allowedHosts,
		},
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
		SourceCode:        s.SourceCode,
		Dependencies:      s.Dependencies,
		CompilationStatus: s.CompilationStatus,
		CompilationError:  s.CompilationError,
		LastCompiledAt:    s.LastCompiledAt,
		ValidationStatus:  s.ValidationStatus,
		ValidationError:   s.ValidationError,
	}
}

// writeJSONError writes a JSON error response with the given status code
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("[ScriptHandlers] Failed to encode error response: %v", err)
	}
}

// ScriptHandlers handles HTTP requests for script management
type ScriptHandlers struct {
	service            *usecase.ScriptService
	compilationService *usecase.CompilationService
}

// NewScriptHandlers creates a new script handlers instance
func NewScriptHandlers(service *usecase.ScriptService) *ScriptHandlers {
	return &ScriptHandlers{service: service}
}

// SetCompilationService sets the compilation service for script compilation support
func (h *ScriptHandlers) SetCompilationService(compilationService *usecase.CompilationService) {
	h.compilationService = compilationService
}

// CreateScript handles POST /_api/v1/scripts
func (h *ScriptHandlers) CreateScript(w http.ResponseWriter, r *http.Request) {
	var req scriptCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, apiErr := newScriptAdminService(h).create(r.Context(), req)
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if result.Script != nil {
		json.NewEncoder(w).Encode(toScriptDTO(result.Script))
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": result.FallbackID})
}

// ListScripts handles GET /_api/v1/scripts
func (h *ScriptHandlers) ListScripts(w http.ResponseWriter, r *http.Request) {
	items, apiErr := newScriptAdminService(h).list(r.Context())
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GetScript handles GET /_api/v1/scripts/{id}
func (h *ScriptHandlers) GetScript(w http.ResponseWriter, r *http.Request) {
	script, apiErr := newScriptAdminService(h).get(r.Context(), r.PathValue("id"))
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script)
}

// UpdateScript handles PUT /_api/v1/scripts/{id}
func (h *ScriptHandlers) UpdateScript(w http.ResponseWriter, r *http.Request) {
	var req scriptUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	script, apiErr := newScriptAdminService(h).update(r.Context(), r.PathValue("id"), req)
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script)
}

// DeleteScript handles DELETE /_api/v1/scripts/{id}
func (h *ScriptHandlers) DeleteScript(w http.ResponseWriter, r *http.Request) {
	if apiErr := newScriptAdminService(h).delete(r.Context(), r.PathValue("id")); apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleScript handles PATCH /_api/v1/scripts/{id}/toggle
func (h *ScriptHandlers) ToggleScript(w http.ResponseWriter, r *http.Request) {
	var req scriptToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	script, apiErr := newScriptAdminService(h).toggle(r.Context(), r.PathValue("id"), req)
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script)
}

func (h *ScriptHandlers) TestScript(w http.ResponseWriter, r *http.Request) {
	var req scriptTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	response, apiErr := newScriptTestService(h).execute(r.Context(), req)
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ScriptHandlers) UploadProject(w http.ResponseWriter, r *http.Request) {
	response, apiErr := newScriptProjectService(h).upload(r.Context(), r.PathValue("id"), r)
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ScriptHandlers) DownloadProject(w http.ResponseWriter, r *http.Request) {
	download, apiErr := newScriptProjectService(h).download(r.Context(), r.PathValue("id"))
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+download.Filename)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(download.Data)))
	w.Write(download.Data)
}

func (h *ScriptHandlers) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	response, apiErr := newScriptProjectService(h).listFiles(r.Context(), r.PathValue("id"))
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ScriptHandlers) ExportScriptAsZip(w http.ResponseWriter, r *http.Request) {
	download, apiErr := newScriptArchiveService(h).export(r.Context(), r.PathValue("id"))
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+download.Filename)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(download.Data)))
	w.Write(download.Data)
}

func (h *ScriptHandlers) ImportScriptFromZip(w http.ResponseWriter, r *http.Request) {
	dto, apiErr := newScriptArchiveService(h).importZip(r.Context(), r)
	if apiErr != nil {
		writeJSONError(w, apiErr.Message, apiErr.Status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

func (h *ScriptHandlers) ListExamples(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newScriptExamplesProvider().list())
}
