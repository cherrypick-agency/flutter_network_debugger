package domain

import (
	"errors"
	"time"
)

// Script is the aggregate root entity for scripting feature
type Script struct {
	// Identity
	ID          string
	Name        string
	Description string

	// Script Configuration
	Runtime  ScriptRuntime // extism | dart
	Code     []byte        // WASM binary for Extism, source code for Dart
	Language string        // rust | javascript | go | dart | python | etc.

	// Compilation (NEW: for in-app compilation)
	SourceCode        string            // Source code before compilation
	Dependencies      map[string]string // filename → content (Cargo.toml, package.json, go.mod)
	CompilationStatus CompilationStatus // Compilation status
	CompilationError  string            // Compilation error message
	LastCompiledAt    *time.Time        // Last successful compilation time

	// Trigger Configuration
	TriggerType TriggerType // request | response | both
	MatchRules  MatchRules  // Pattern matching
	Priority    int         // Execution order (higher = earlier)

	// Resource Limits
	Config ScriptConfig

	// State
	Enabled bool

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CompilationStatus represents the compilation state
type CompilationStatus string

const (
	CompilationNotCompiled CompilationStatus = "not_compiled" // No compilation attempted
	CompilationPending     CompilationStatus = "pending"      // Queued for compilation
	CompilationCompiling   CompilationStatus = "compiling"    // Currently compiling
	CompilationSuccess     CompilationStatus = "success"      // Successfully compiled
	CompilationStatusError CompilationStatus = "error"        // Compilation failed
)

// Validate performs domain-level validation
func (s *Script) Validate() error {
	if s.Name == "" {
		return errors.New("script name is required")
	}

	// Either Code (compiled WASM), SourceCode (single file), or Dependencies (multi-file) must be present
	if len(s.Code) == 0 && s.SourceCode == "" && len(s.Dependencies) == 0 {
		return errors.New("script code, source code, or dependencies are required")
	}

	if s.Runtime == "" {
		return errors.New("script runtime is required")
	}
	if s.TriggerType == "" {
		return errors.New("trigger type is required")
	}

	// Set defaults
	if s.Config.TimeoutMs <= 0 {
		s.Config.TimeoutMs = 5000 // default 5s
	}
	if s.Config.MemoryLimitMB <= 0 {
		s.Config.MemoryLimitMB = 10 // default 10MB
	}

	// Set default compilation status
	if s.CompilationStatus == "" {
		s.CompilationStatus = CompilationNotCompiled
	}

	return nil
}

// NeedsRecompilation checks if script needs to be compiled
func (s *Script) NeedsRecompilation() bool {
	return s.SourceCode != "" &&
		(s.CompilationStatus != CompilationSuccess ||
			s.LastCompiledAt == nil ||
			len(s.Code) == 0)
}

// MarkCompiling marks script as currently compiling
func (s *Script) MarkCompiling() {
	s.CompilationStatus = CompilationCompiling
	s.CompilationError = ""
}

// MarkCompilationSuccess marks script as successfully compiled
func (s *Script) MarkCompilationSuccess(wasm []byte) {
	s.Code = wasm
	s.CompilationStatus = CompilationSuccess
	now := time.Now()
	s.LastCompiledAt = &now
	s.CompilationError = ""
}

// MarkCompilationError marks script compilation as failed
func (s *Script) MarkCompilationError(err error) {
	s.CompilationStatus = CompilationStatusError
	s.CompilationError = err.Error()
}

// TriggerType defines when a script should execute
type TriggerType string

const (
	TriggerRequest  TriggerType = "request"  // Execute on HTTP request
	TriggerResponse TriggerType = "response" // Execute on HTTP response
	TriggerBoth     TriggerType = "both"     // Execute on both
)

// MatchRules defines when a script matches a request
// Reuses pattern matching logic from mapping feature
type MatchRules struct {
	Methods     []string    // ["GET", "POST"] - empty means all
	HostPattern string      // "api.example.com" - empty means all
	PathPattern string      // "/users/*" - empty means all
	PatternType PatternType // exact | prefix | wildcard | regex
}

// PatternType defines how to match patterns
type PatternType string

const (
	PatternExact    PatternType = "exact"    // Exact string match
	PatternPrefix   PatternType = "prefix"   // Prefix match
	PatternWildcard PatternType = "wildcard" // Wildcard match (supports *)
	PatternRegex    PatternType = "regex"    // Regular expression
)

// ScriptConfig holds runtime-specific configuration
type ScriptConfig struct {
	TimeoutMs     int      // Execution timeout in milliseconds
	MemoryLimitMB int      // Memory limit (for WASM runtimes)
	AllowedHosts  []string // Allowed hosts for http_fetch host function
}

// ScriptContext is passed to scripts during execution (Value Object)
type ScriptContext struct {
	Request     *HTTPRequest     `json:"request,omitempty"`
	Response    *HTTPResponse    `json:"response,omitempty"`
	Session     *SessionInfo     `json:"session,omitempty"`
	Transaction *TransactionInfo `json:"transaction,omitempty"`
}

// HTTPRequest represents an HTTP request for scripts
type HTTPRequest struct {
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

// HTTPResponse represents an HTTP response for scripts
type HTTPResponse struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

// SessionInfo contains session metadata
type SessionInfo struct {
	ID         string `json:"id"`
	ClientAddr string `json:"client_addr"`
}

// TransactionInfo contains transaction metadata
type TransactionInfo struct {
	ID       string        `json:"id"`
	Duration time.Duration `json:"duration"`
}

// ScriptResult is returned by scripts after execution
type ScriptResult struct {
	Modified         bool          `json:"modified"`
	ModifiedRequest  *HTTPRequest  `json:"modifiedRequest,omitempty"`
	ModifiedResponse *HTTPResponse `json:"modifiedResponse,omitempty"`
	Logs             []string      `json:"logs,omitempty"`
	Error            string        `json:"error,omitempty"`
}

// ScriptFilter for querying scripts
type ScriptFilter struct {
	Enabled     *bool
	Runtime     ScriptRuntime
	TriggerType TriggerType
}
