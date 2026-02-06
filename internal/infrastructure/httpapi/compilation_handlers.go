package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

// CompilationHandlers handles HTTP requests for script compilation
// This is the adapter layer for HTTP (Clean Architecture)
type CompilationHandlers struct {
	service *usecase.CompilationService
}

// NewCompilationHandlers creates a new compilation handlers instance
func NewCompilationHandlers(service *usecase.CompilationService) *CompilationHandlers {
	return &CompilationHandlers{
		service: service,
	}
}

// CompileScriptRequest is the request body for compiling a script
type CompileScriptRequest struct {
	Optimize bool `json:"optimize"`
}

// CompileScriptResponse is the response body for compilation result
type CompileScriptResponse struct {
	Status          string   `json:"status"` // "success" or "error"
	WASMSize        int64    `json:"wasmSize,omitempty"`
	CompilationTime string   `json:"compilationTime,omitempty"`
	Logs            []string `json:"logs,omitempty"`
	Error           string   `json:"error,omitempty"`
}

// CompileScript compiles a script by ID
// POST /_api/v1/scripts/:id/compile
func (h *CompilationHandlers) CompileScript(w http.ResponseWriter, r *http.Request) {
	// Extract script ID from URL (Go 1.22+ routing)
	scriptID := r.PathValue("id")
	if scriptID == "" {
		writeJSONError(w, "script id required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req CompileScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to optimize=true if no body provided
		req.Optimize = true
	}

	log.Printf("[CompilationHandlers] Compiling script %s (optimize=%v)", scriptID, req.Optimize)

	// Compile script (use case)
	result, err := h.service.CompileScript(r.Context(), scriptID, req.Optimize)
	if err != nil {
		log.Printf("[CompilationHandlers] Compilation failed for script %s: %v", scriptID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CompileScriptResponse{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CompileScriptResponse{
		Status:          "success",
		WASMSize:        result.WASMSize,
		CompilationTime: result.Duration.String(),
		Logs:            result.Logs,
	})
}

// ValidateSyntaxRequest is the request body for syntax validation
type ValidateSyntaxRequest struct {
	SourceCode   string            `json:"sourceCode"`
	Language     string            `json:"language"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// ValidateSyntaxResponse is the response body for syntax validation
type ValidateSyntaxResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// ValidateSyntax validates script syntax without compiling
// POST /_api/v1/scripts/validate
func (h *CompilationHandlers) ValidateSyntax(w http.ResponseWriter, r *http.Request) {
	var req ValidateSyntaxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create temporary script for validation
	script := &domain.Script{
		SourceCode:   req.SourceCode,
		Language:     req.Language,
		Dependencies: req.Dependencies,
	}

	// Validate syntax
	if err := h.service.ValidateSyntax(r.Context(), script); err != nil {
		// Return 400 Bad Request for invalid syntax
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateSyntaxResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	// Syntax is valid - return 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateSyntaxResponse{
		Valid: true,
	})
}

// CompilersResponse is the response body for listing available compilers
type CompilersResponse struct {
	Compilers []string        `json:"compilers"` // Available compilers
	All       map[string]bool `json:"all"`       // All registered compilers with availability status
}

// ListCompilers returns list of available compilers
// GET /_api/v1/scripts/compilers
func (h *CompilationHandlers) ListCompilers(w http.ResponseWriter, r *http.Request) {
	available := h.service.GetAvailableCompilers()
	all := h.service.GetAllCompilers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CompilersResponse{
		Compilers: available,
		All:       all,
	})
}
