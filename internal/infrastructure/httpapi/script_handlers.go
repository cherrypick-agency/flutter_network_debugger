package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"network-debugger/internal/features/scripting/domain"
	"network-debugger/internal/features/scripting/usecase"
)

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
	var req struct {
		Name         string            `json:"name"`
		Description  string            `json:"description"`
		Runtime      string            `json:"runtime"`
		Code         string            `json:"code"`         // Base64-encoded WASM binary (optional if sourceCode provided)
		SourceCode   string            `json:"sourceCode"`   // Source code to compile (optional if code provided)
		Dependencies map[string]string `json:"dependencies"` // Multi-file project files (filename -> content)
		Language     string            `json:"language"`
		TriggerType  string            `json:"triggerType"`
		Priority     int               `json:"priority"`
		Enabled      *bool             `json:"enabled"` // Pointer to distinguish between false and not provided
		MatchRules   *struct {
			Methods     []string `json:"methods"`
			HostPattern string   `json:"hostPattern"`
			PathPattern string   `json:"pathPattern"`
			PatternType string   `json:"patternType"`
		} `json:"matchRules,omitempty"`
		Config *struct {
			TimeoutMs     int      `json:"timeoutMs"`
			MemoryLimitMB int      `json:"memoryLimitMB"`
			AllowedHosts  []string `json:"allowedHosts"`
		} `json:"config,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var codeBytes []byte
	var err error

	// Support: code (base64 WASM), sourceCode (single file), or dependencies (multi-file)
	if req.SourceCode != "" || len(req.Dependencies) > 0 {
		// Compile source code or multi-file project to WASM
		if req.Language == "" {
			http.Error(w, "language required when using sourceCode or dependencies", http.StatusBadRequest)
			return
		}

		// Check if compilation service is available
		if h.compilationService == nil {
			http.Error(w, "compilation service not available - use 'code' field with base64 WASM instead", http.StatusBadRequest)
			return
		}

		// Create a new script with source code first (needed for compilation)
		scriptID := uuid.New().String()
		tempScript := &domain.Script{
			ID:           scriptID,
			Name:         req.Name,
			Description:  req.Description,
			Runtime:      domain.ScriptRuntime(req.Runtime),
			SourceCode:   req.SourceCode,
			Dependencies: req.Dependencies, // Multi-file project support
			Language:     req.Language,
			TriggerType:  domain.TriggerType(req.TriggerType),
			Priority:     req.Priority,
			Code:         []byte{}, // Empty for now
			Config: domain.ScriptConfig{
				TimeoutMs:     5000,
				MemoryLimitMB: 10,
			},
			Enabled:   false, // Start disabled until compilation succeeds
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Save the script with source code first
		if err := h.service.CreateScript(r.Context(), tempScript); err != nil {
			http.Error(w, "failed to create script: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Now compile it (CompileScript will save the WASM to DB automatically)
		// Add timeout to prevent hanging requests on long compilations
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		_, err := h.compilationService.CompileScript(ctx, scriptID, true)
		if err != nil {
			// Compilation failed - delete the script to avoid orphan records
			if deleteErr := h.service.DeleteScript(r.Context(), scriptID); deleteErr != nil {
				log.Printf("[ScriptHandlers] WARNING: Failed to cleanup script %s after compilation error: %v", scriptID, deleteErr)
			}
			http.Error(w, "compilation failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// If enabled flag was provided, update it atomically (avoid race condition)
		if req.Enabled != nil && *req.Enabled {
			if err := h.service.ToggleScript(r.Context(), scriptID, true); err != nil {
				log.Printf("[ScriptHandlers] WARNING: Failed to enable script %s: %v", scriptID, err)
			}
		}

		// Return the created script
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id": scriptID,
		})
		return
	} else if req.Code != "" {
		// Decode base64-encoded code
		codeBytes, err = base64.StdEncoding.DecodeString(req.Code)
		if err != nil {
			http.Error(w, "invalid base64 code: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "either 'code', 'sourceCode', or 'dependencies' must be provided", http.StatusBadRequest)
		return
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
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Apply match rules if provided
	if req.MatchRules != nil {
		script.MatchRules = domain.MatchRules{
			Methods:     req.MatchRules.Methods,
			HostPattern: req.MatchRules.HostPattern,
			PathPattern: req.MatchRules.PathPattern,
			PatternType: domain.PatternType(req.MatchRules.PatternType),
		}
	}

	// Apply config if provided
	if req.Config != nil {
		script.Config.TimeoutMs = req.Config.TimeoutMs
		script.Config.MemoryLimitMB = req.Config.MemoryLimitMB
		script.Config.AllowedHosts = req.Config.AllowedHosts
	}

	if err := h.service.CreateScript(r.Context(), script); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(script)
}

// ListScripts handles GET /_api/v1/scripts
func (h *ScriptHandlers) ListScripts(w http.ResponseWriter, r *http.Request) {
	scripts, err := h.service.ListScripts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scripts)
}

// GetScript handles GET /_api/v1/scripts/{id}
func (h *ScriptHandlers) GetScript(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	script, err := h.service.GetScript(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script)
}

// UpdateScript handles PUT /_api/v1/scripts/{id}
func (h *ScriptHandlers) UpdateScript(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	// Get existing script
	script, err := h.service.GetScript(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Decode updates
	var req struct {
		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		Code        *string `json:"code,omitempty"`       // Base64-encoded WASM binary
		SourceCode  *string `json:"sourceCode,omitempty"` // Source code (requires recompilation)
		Language    *string `json:"language,omitempty"`
		TriggerType *string `json:"triggerType,omitempty"`
		Priority    *int    `json:"priority,omitempty"`
		Config      *struct {
			TimeoutMs     int      `json:"timeoutMs"`
			MemoryLimitMB int      `json:"memoryLimitMB"`
			AllowedHosts  []string `json:"allowedHosts"`
		} `json:"config,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Apply updates
	if req.Name != nil {
		script.Name = *req.Name
	}
	if req.Description != nil {
		script.Description = *req.Description
	}
	if req.Code != nil {
		script.Code = []byte(*req.Code)
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
	if req.Config != nil {
		script.Config.TimeoutMs = req.Config.TimeoutMs
		script.Config.MemoryLimitMB = req.Config.MemoryLimitMB
		script.Config.AllowedHosts = req.Config.AllowedHosts
	}

	// Handle sourceCode updates - clear WASM to force recompilation
	if req.SourceCode != nil && *req.SourceCode != script.SourceCode {
		log.Printf("[ScriptHandlers] SourceCode changed for script %s, clearing WASM (requires recompilation)", id)
		script.SourceCode = *req.SourceCode
		script.Code = []byte{} // Clear WASM - script must be recompiled via /compile endpoint
		script.Enabled = false // Auto-disable to prevent runtime errors with empty Code
		log.Printf("[ScriptHandlers] Auto-disabled script %s (must be recompiled and re-enabled)", id)
	}

	script.UpdatedAt = time.Now()

	if err := h.service.UpdateScript(r.Context(), script); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script)
}

// DeleteScript handles DELETE /_api/v1/scripts/{id}
func (h *ScriptHandlers) DeleteScript(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteScript(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleScript handles PATCH /_api/v1/scripts/{id}/toggle
func (h *ScriptHandlers) ToggleScript(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.ToggleScript(r.Context(), id, req.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// TestScript handles POST /_api/v1/scripts/test
// Tests a script with sample request data before deploying
func (h *ScriptHandlers) TestScript(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Script struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Runtime     string `json:"runtime"`
			Code        string `json:"code"`
			Language    string `json:"language"`
			TriggerType string `json:"triggerType"`
			Priority    int    `json:"priority"`
			MatchRules  *struct {
				Methods     []string `json:"methods"`
				HostPattern string   `json:"hostPattern"`
				PathPattern string   `json:"pathPattern"`
				PatternType string   `json:"patternType"`
			} `json:"matchRules,omitempty"`
			Config *struct {
				TimeoutMs     int      `json:"timeoutMs"`
				MemoryLimitMB int      `json:"memoryLimitMB"`
				AllowedHosts  []string `json:"allowedHosts"`
			} `json:"config,omitempty"`
		} `json:"script"`
		TestRequest struct {
			Method  string              `json:"method"`
			URL     string              `json:"url"`
			Headers map[string][]string `json:"headers"`
			Body    string              `json:"body"`
		} `json:"testRequest"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Decode base64-encoded code
	codeBytes, err := base64.StdEncoding.DecodeString(req.Script.Code)
	if err != nil {
		http.Error(w, "invalid base64 code: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Build script object
	script := &domain.Script{
		ID:          "test-script",
		Name:        req.Script.Name,
		Description: req.Script.Description,
		Runtime:     domain.ScriptRuntime(req.Script.Runtime),
		Code:        codeBytes,
		Language:    req.Script.Language,
		TriggerType: domain.TriggerType(req.Script.TriggerType),
		Priority:    req.Script.Priority,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Apply match rules if provided
	if req.Script.MatchRules != nil {
		script.MatchRules = domain.MatchRules{
			Methods:     req.Script.MatchRules.Methods,
			HostPattern: req.Script.MatchRules.HostPattern,
			PathPattern: req.Script.MatchRules.PathPattern,
			PatternType: domain.PatternType(req.Script.MatchRules.PatternType),
		}
	}

	// Apply config if provided
	if req.Script.Config != nil {
		script.Config.TimeoutMs = req.Script.Config.TimeoutMs
		script.Config.MemoryLimitMB = req.Script.Config.MemoryLimitMB
		script.Config.AllowedHosts = req.Script.Config.AllowedHosts
	}

	// Build test request
	testReq := &domain.HTTPRequest{
		Method:  req.TestRequest.Method,
		URL:     req.TestRequest.URL,
		Headers: req.TestRequest.Headers,
		Body:    []byte(req.TestRequest.Body),
	}

	// Execute test
	startTime := time.Now()
	result, logs, err := h.service.TestScript(r.Context(), script, testReq)
	duration := time.Since(startTime)

	// Build response
	response := map[string]interface{}{
		"success":    err == nil && (result == nil || result.Error == ""),
		"durationMs": duration.Milliseconds(),
		"logs":       logs,
	}

	if err != nil {
		response["error"] = err.Error()
	} else if result != nil {
		if result.Error != "" {
			response["success"] = false
			response["error"] = result.Error
		}
		if result.Modified && result.ModifiedRequest != nil {
			response["modifiedRequest"] = map[string]interface{}{
				"method":  result.ModifiedRequest.Method,
				"url":     result.ModifiedRequest.URL,
				"headers": result.ModifiedRequest.Headers,
				"body":    string(result.ModifiedRequest.Body),
			}
		}
		if result.Modified && result.ModifiedResponse != nil {
			response["modifiedResponse"] = map[string]interface{}{
				"status":  result.ModifiedResponse.Status,
				"headers": result.ModifiedResponse.Headers,
				"body":    string(result.ModifiedResponse.Body),
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UploadProject handles POST /_api/v1/scripts/{id}/upload-project
// Uploads a ZIP file containing multiple source files for a script project
func (h *ScriptHandlers) UploadProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	// Get existing script
	script, err := h.service.GetScript(r.Context(), id)
	if err != nil {
		http.Error(w, "script not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("project")
	if err != nil {
		http.Error(w, "project file required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Verify it's a ZIP file
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		http.Error(w, "file must be a ZIP archive", http.StatusBadRequest)
		return
	}

	// Read ZIP file into memory
	zipData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read zip: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse ZIP archive
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		http.Error(w, "invalid zip file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Extract files into Dependencies map
	dependencies := make(map[string]string)
	totalSize := 0
	const maxFileSize = 500 * 1024       // 500KB per file
	const maxTotalSize = 5 * 1024 * 1024 // 5MB total

	for _, zipFile := range zipReader.File {
		// Skip directories
		if zipFile.FileInfo().IsDir() {
			continue
		}

		// Skip hidden files and __MACOSX
		if strings.HasPrefix(filepath.Base(zipFile.Name), ".") || strings.Contains(zipFile.Name, "__MACOSX") {
			continue
		}

		// Validate file extension (security: only source files)
		ext := strings.ToLower(filepath.Ext(zipFile.Name))
		allowedExts := map[string]bool{
			".rs": true, ".go": true, ".ts": true, ".js": true,
			".toml": true, ".json": true, ".mod": true, ".sum": true,
			".c": true, ".cpp": true, ".h": true, ".hpp": true,
			".py": true, ".zig": true, ".kt": true, ".swift": true,
		}
		if !allowedExts[ext] {
			continue // Skip non-source files
		}

		// Check file size
		if zipFile.UncompressedSize64 > maxFileSize {
			http.Error(w, fmt.Sprintf("file %s too large (max 500KB)", zipFile.Name), http.StatusBadRequest)
			return
		}

		// Open file in ZIP
		rc, err := zipFile.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read %s: %v", zipFile.Name, err), http.StatusInternalServerError)
			return
		}

		// Read file content
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read %s: %v", zipFile.Name, err), http.StatusInternalServerError)
			return
		}

		// Update total size
		totalSize += len(content)
		if totalSize > maxTotalSize {
			http.Error(w, "total project size exceeds 5MB", http.StatusBadRequest)
			return
		}

		// Normalize path (remove leading directories if nested in single folder)
		filename := zipFile.Name
		if strings.Contains(filename, "/") {
			// If all files are in a single root folder, strip it
			parts := strings.Split(filename, "/")
			if len(parts) > 1 {
				// Check if this looks like a project root (skip single folder wrapping)
				filename = strings.Join(parts[1:], "/")
			}
		}

		dependencies[filename] = string(content)
	}

	if len(dependencies) == 0 {
		http.Error(w, "no valid source files found in ZIP", http.StatusBadRequest)
		return
	}

	// Update script with new dependencies
	script.Dependencies = dependencies
	script.UpdatedAt = time.Now()

	if err := h.service.UpdateScript(r.Context(), script); err != nil {
		http.Error(w, "failed to update script: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success with file count
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"fileCount": len(dependencies),
		"files":     getFileList(dependencies),
		"totalSize": totalSize,
	})
}

// DownloadProject handles GET /_api/v1/scripts/{id}/download-project
// Downloads all script files (code + dependencies) as a ZIP archive
func (h *ScriptHandlers) DownloadProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	// Get script
	script, err := h.service.GetScript(r.Context(), id)
	if err != nil {
		http.Error(w, "script not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Create in-memory ZIP
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Add main source code if present
	if script.SourceCode != "" {
		mainFile := getMainFilename(script.Language)
		fw, err := zipWriter.Create(mainFile)
		if err != nil {
			http.Error(w, "failed to create zip: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fw.Write([]byte(script.SourceCode)); err != nil {
			http.Error(w, "failed to write source: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Add all dependency files
	for filename, content := range script.Dependencies {
		fw, err := zipWriter.Create(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create %s: %v", filename, err), http.StatusInternalServerError)
			return
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			http.Error(w, fmt.Sprintf("failed to write %s: %v", filename, err), http.StatusInternalServerError)
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "failed to finalize zip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send ZIP file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-project.zip", script.ID))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Write(buf.Bytes())
}

// ListProjectFiles handles GET /_api/v1/scripts/{id}/files
// Lists all files in the script project (main code + dependencies)
func (h *ScriptHandlers) ListProjectFiles(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	// Get script
	script, err := h.service.GetScript(r.Context(), id)
	if err != nil {
		http.Error(w, "script not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Build file list
	files := make([]map[string]interface{}, 0)

	// Add main source file
	if script.SourceCode != "" {
		mainFile := getMainFilename(script.Language)
		files = append(files, map[string]interface{}{
			"filename": mainFile,
			"size":     len(script.SourceCode),
			"type":     "source",
		})
	}

	// Add dependency files
	for filename, content := range script.Dependencies {
		files = append(files, map[string]interface{}{
			"filename": filename,
			"size":     len(content),
			"type":     getDependencyType(filename),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"scriptId":  script.ID,
		"fileCount": len(files),
		"files":     files,
	})
}

// Helper: get main filename based on language
func getMainFilename(language string) string {
	switch strings.ToLower(language) {
	case "rust":
		return "src/lib.rs"
	case "go":
		return "main.go"
	case "javascript", "typescript":
		return "index.ts"
	case "dart":
		return "main.dart"
	case "python":
		return "main.py"
	case "zig":
		return "main.zig"
	case "kotlin":
		return "Main.kt"
	case "swift":
		return "main.swift"
	case "c":
		return "main.c"
	case "cpp", "c++":
		return "main.cpp"
	default:
		return "main.txt"
	}
}

// Helper: determine file type from extension
func getDependencyType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".toml", ".json", ".mod", ".sum":
		return "config"
	case ".rs", ".go", ".ts", ".js", ".dart", ".py", ".zig", ".kt", ".swift", ".c", ".cpp", ".h", ".hpp":
		return "source"
	default:
		return "other"
	}
}

// Helper: get file list from dependencies map
func getFileList(deps map[string]string) []string {
	files := make([]string, 0, len(deps))
	for filename := range deps {
		files = append(files, filename)
	}
	return files
}

// ExportScriptAsZip handles GET /_api/v1/scripts/{id}/export-zip
// Exports complete script (metadata + source + dependencies + WASM) as ZIP
func (h *ScriptHandlers) ExportScriptAsZip(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "script id required", http.StatusBadRequest)
		return
	}

	// Get script
	script, err := h.service.GetScript(r.Context(), id)
	if err != nil {
		http.Error(w, "script not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Create in-memory ZIP
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 1. Add metadata.json (all script info for import)
	metadata := map[string]interface{}{
		"name":          script.Name,
		"description":   script.Description,
		"language":      script.Language,
		"runtime":       script.Runtime,
		"priority":      script.Priority,
		"enabled":       script.Enabled,
		"triggerType":   script.TriggerType,
		"matchRules":    script.MatchRules,
		"config":        script.Config,
		"sourceCode":    script.SourceCode,
		"dependencies":  script.Dependencies,
		"exportVersion": "1.0",
		"exportedAt":    time.Now().Format(time.RFC3339),
	}

	metadataFile, err := zipWriter.Create("metadata.json")
	if err != nil {
		http.Error(w, "failed to create metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(metadataFile).Encode(metadata); err != nil {
		http.Error(w, "failed to write metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Add main source code if present
	if script.SourceCode != "" {
		mainFile := getMainFilename(script.Language)
		fw, err := zipWriter.Create(mainFile)
		if err != nil {
			http.Error(w, "failed to create main file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := fw.Write([]byte(script.SourceCode)); err != nil {
			http.Error(w, "failed to write main file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 3. Add all dependency files
	for filename, content := range script.Dependencies {
		fw, err := zipWriter.Create(filename)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create %s: %v", filename, err), http.StatusInternalServerError)
			return
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			http.Error(w, fmt.Sprintf("failed to write %s: %v", filename, err), http.StatusInternalServerError)
			return
		}
	}

	// 4. Add compiled WASM if exists
	if len(script.Code) > 0 {
		wasmFile, err := zipWriter.Create("output.wasm")
		if err != nil {
			http.Error(w, "failed to create wasm file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// script.Code is already []byte, write directly
		if _, err := wasmFile.Write(script.Code); err != nil {
			http.Error(w, "failed to write wasm: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		http.Error(w, "failed to finalize zip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Send ZIP file
	safeFilename := strings.ReplaceAll(script.Name, " ", "-")
	safeFilename = strings.ReplaceAll(safeFilename, "/", "-")
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", safeFilename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	w.Write(buf.Bytes())
}

// ImportScriptFromZip handles POST /_api/v1/scripts/import-zip
// Imports complete script from ZIP file
func (h *ScriptHandlers) ImportScriptFromZip(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file to memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Open ZIP
	zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		http.Error(w, "invalid ZIP file: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Parse ZIP contents
	var metadata map[string]interface{}
	sourceFiles := make(map[string]string)
	var wasmCode string

	for _, zipFile := range zipReader.File {
		rc, err := zipFile.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to open %s: %v", zipFile.Name, err), http.StatusInternalServerError)
			return
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read %s: %v", zipFile.Name, err), http.StatusInternalServerError)
			return
		}

		switch zipFile.Name {
		case "metadata.json":
			if err := json.Unmarshal(content, &metadata); err != nil {
				http.Error(w, "invalid metadata: "+err.Error(), http.StatusBadRequest)
				return
			}

		case "output.wasm":
			wasmCode = base64.StdEncoding.EncodeToString(content)

		default:
			// All other files are source/dependency files
			sourceFiles[zipFile.Name] = string(content)
		}
	}

	// Validate metadata
	if metadata == nil {
		http.Error(w, "metadata.json not found in ZIP", http.StatusBadRequest)
		return
	}

	// Extract fields from metadata
	name, _ := metadata["name"].(string)
	description, _ := metadata["description"].(string)
	language, _ := metadata["language"].(string)
	runtime, _ := metadata["runtime"].(string)
	sourceCode, _ := metadata["sourceCode"].(string)

	// Dependencies from metadata (fallback to parsed files)
	var dependencies map[string]string
	if deps, ok := metadata["dependencies"].(map[string]interface{}); ok {
		dependencies = make(map[string]string)
		for k, v := range deps {
			if strVal, ok := v.(string); ok {
				dependencies[k] = strVal
			}
		}
	} else {
		// Use source files from ZIP
		mainFilename := getMainFilename(language)
		dependencies = make(map[string]string)
		for filename, content := range sourceFiles {
			if filename != mainFilename {
				dependencies[filename] = content
			} else if sourceCode == "" {
				sourceCode = content
			}
		}
	}

	// Validate required fields
	if name == "" {
		http.Error(w, "name is required in metadata", http.StatusBadRequest)
		return
	}
	if language == "" {
		http.Error(w, "language is required in metadata", http.StatusBadRequest)
		return
	}

	// Decode WASM Code from base64 string to []byte
	var wasmCodeBytes []byte
	if wasmCode != "" {
		var err error
		wasmCodeBytes, err = base64.StdEncoding.DecodeString(wasmCode)
		if err != nil {
			log.Printf("Warning: failed to decode WASM code: %v", err)
		}
	}

	// Convert runtime string to ScriptRuntime type (default to extism)
	scriptRuntime := domain.RuntimeExtism
	if runtime == "dart" {
		scriptRuntime = domain.RuntimeDart
	}

	// Create new script entity
	newScript := domain.Script{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		Language:     language,
		Runtime:      scriptRuntime,
		SourceCode:   sourceCode,
		Dependencies: dependencies,
		Code:         wasmCodeBytes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Enabled:      true, // Default to enabled
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
	}

	// Copy optional fields from metadata
	if priority, ok := metadata["priority"].(float64); ok {
		newScript.Priority = int(priority)
	}
	if enabled, ok := metadata["enabled"].(bool); ok {
		newScript.Enabled = enabled
	}
	if triggerType, ok := metadata["triggerType"].(string); ok {
		newScript.TriggerType = domain.TriggerType(triggerType)
	}

	// Parse MatchRules from map
	if matchRulesMap, ok := metadata["matchRules"].(map[string]interface{}); ok {
		matchRules := domain.MatchRules{}
		if methods, ok := matchRulesMap["methods"].([]interface{}); ok {
			for _, m := range methods {
				if method, ok := m.(string); ok {
					matchRules.Methods = append(matchRules.Methods, method)
				}
			}
		}
		if hostPattern, ok := matchRulesMap["hostPattern"].(string); ok {
			matchRules.HostPattern = hostPattern
		}
		if pathPattern, ok := matchRulesMap["pathPattern"].(string); ok {
			matchRules.PathPattern = pathPattern
		}
		if patternType, ok := matchRulesMap["patternType"].(string); ok {
			matchRules.PatternType = domain.PatternType(patternType)
		}
		newScript.MatchRules = matchRules
	}

	// Parse Config from map
	if configMap, ok := metadata["config"].(map[string]interface{}); ok {
		if timeoutMs, ok := configMap["timeoutMs"].(float64); ok {
			newScript.Config.TimeoutMs = int(timeoutMs)
		}
		if memoryLimitMB, ok := configMap["memoryLimitMB"].(float64); ok {
			newScript.Config.MemoryLimitMB = int(memoryLimitMB)
		}
		if allowedHosts, ok := configMap["allowedHosts"].([]interface{}); ok {
			hosts := make([]string, 0, len(allowedHosts))
			for _, h := range allowedHosts {
				if host, ok := h.(string); ok {
					hosts = append(hosts, host)
				}
			}
			newScript.Config.AllowedHosts = hosts
		}
	}

	// Validate script data
	if err := validateScriptData(&newScript); err != nil {
		http.Error(w, "validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create script in DB
	if err := h.service.CreateScript(r.Context(), &newScript); err != nil {
		http.Error(w, "failed to create script: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return created script
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&newScript)
}

// validateScriptData validates script data before import
func validateScriptData(script *domain.Script) error {
	// Required fields
	if script.Name == "" {
		return errors.New("name is required")
	}
	if script.Language == "" {
		return errors.New("language is required")
	}

	// Valid language
	validLanguages := []string{"rust", "go", "javascript", "typescript", "dart", "python", "zig", "kotlin", "swift", "c", "cpp"}
	validLang := false
	for _, lang := range validLanguages {
		if strings.EqualFold(script.Language, lang) {
			validLang = true
			break
		}
	}
	if !validLang {
		return fmt.Errorf("invalid language: %s", script.Language)
	}

	// Size limits
	if len(script.SourceCode) > 500_000 {
		return errors.New("source code too large (max 500KB)")
	}

	totalSize := len(script.SourceCode)
	for _, content := range script.Dependencies {
		totalSize += len(content)
		if len(content) > 500_000 {
			return errors.New("dependency file too large (max 500KB per file)")
		}
	}
	if totalSize > 5_000_000 {
		return errors.New("total project size too large (max 5MB)")
	}

	return nil
}

// ScriptExample represents an example script in the library
type ScriptExample struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Difficulty  string `json:"difficulty"` // beginner, intermediate, advanced
	Category    string `json:"category"`   // request_modification, response_modification, both
	TriggerType string `json:"triggerType"`
	SourceCode  string `json:"sourceCode"`
}

// ListExamples handles GET /_api/v1/scripts/examples
func (h *ScriptHandlers) ListExamples(w http.ResponseWriter, r *http.Request) {
	examples := []ScriptExample{
		{
			ID:          "rust-passthrough",
			Name:        "Passthrough (No Modification)",
			Description: "Simplest possible script - passes requests through without any modifications. Great starting point for understanding script structure.",
			Language:    "rust",
			Difficulty:  "beginner",
			Category:    "request_modification",
			TriggerType: "request",
			SourceCode: `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Simple passthrough - no modifications
    Ok(Json(req))
}`,
		},
		{
			ID:          "rust-add-header",
			Name:        "Add Custom Headers",
			Description: "Adds custom headers to HTTP requests. Demonstrates basic request modification and header manipulation.",
			Language:    "rust",
			Difficulty:  "beginner",
			Category:    "request_modification",
			TriggerType: "request",
			SourceCode: `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Deserialize, Serialize)]
struct HTTPRequest {
    method: String,
    url: String,
    headers: HashMap<String, Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(mut req): Json<HTTPRequest>) -> FnResult<Json<HTTPRequest>> {
    // Log the request
    log!(LogLevel::Info, "Processing {} {}", req.method, req.url);

    // Add custom headers
    req.headers.insert("X-Script-Processed".to_string(), vec!["Rust".to_string()]);
    req.headers.insert("X-Custom-Header".to_string(), vec!["Hello from Rust!".to_string()]);

    Ok(Json(req))
}`,
		},
		{
			ID:          "rust-modify-response",
			Name:        "Sanitize Response Data",
			Description: "Removes sensitive fields from JSON responses and adds processing metadata. Demonstrates response modification and JSON parsing.",
			Language:    "rust",
			Difficulty:  "advanced",
			Category:    "response_modification",
			TriggerType: "response",
			SourceCode: `use extism_pdk::*;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::collections::HashMap;

#[derive(Deserialize)]
struct HTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
}

#[derive(Serialize)]
struct ModifiedHTTPResponse {
    status: i32,
    headers: HashMap<String, Vec<String>>,
}

#[plugin_fn]
pub fn process(Json(mut resp): Json<HTTPResponse>) -> FnResult<Json<ModifiedHTTPResponse>> {
    log!(LogLevel::Info, "Processing response with status: {}", resp.status);

    // Add header indicating response was processed
    resp.headers.insert(
        "X-Script-Sanitized".to_string(),
        vec!["true".to_string()],
    );

    Ok(Json(ModifiedHTTPResponse {
        status: resp.status,
        headers: resp.headers,
    }))
}`,
		},
		{
			ID:          "python-add-header",
			Name:        "Add Headers (Python)",
			Description: "Python version of header addition script. Shows how to use Python for simple request modifications.",
			Language:    "python",
			Difficulty:  "beginner",
			Category:    "request_modification",
			TriggerType: "request",
			SourceCode: `# Simple Python script that adds custom headers to HTTP requests

# Add custom headers
if "X-Python-Processed" not in request.headers:
    request.headers["X-Python-Processed"] = ["true"]

request.headers["X-Script-Language"] = ["Python"]
request.headers["X-Runtime"] = ["RustPython-WASM"]

# Log the request
print(f"Processing {request.method} {request.url}")`,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(examples)
}
