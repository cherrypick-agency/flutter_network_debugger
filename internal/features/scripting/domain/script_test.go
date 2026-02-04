package domain

import (
	"testing"
	"time"
)

// Composer 1.
func TestScript_Validate(t *testing.T) {
	tests := []struct {
		name    string
		script  *Script
		wantErr bool
	}{
		{
			name: "valid script with code",
			script: &Script{
				Name:        "test",
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
				Code:        []byte{1, 2, 3},
			},
			wantErr: false,
		},
		{
			name: "valid script with source code",
			script: &Script{
				Name:        "test",
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
				SourceCode:  "fn main() {}",
			},
			wantErr: false,
		},
		{
			name: "valid script with dependencies",
			script: &Script{
				Name:        "test",
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
				Dependencies: map[string]string{
					"Cargo.toml": "[package]",
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			script: &Script{
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
				Code:        []byte{1, 2, 3},
			},
			wantErr: true,
		},
		{
			name: "missing runtime",
			script: &Script{
				Name:        "test",
				TriggerType: TriggerRequest,
				Code:        []byte{1, 2, 3},
			},
			wantErr: true,
		},
		{
			name: "missing trigger type",
			script: &Script{
				Name:    "test",
				Runtime: RuntimeExtism,
				Code:    []byte{1, 2, 3},
			},
			wantErr: true,
		},
		{
			name: "missing code, source and dependencies",
			script: &Script{
				Name:        "test",
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
			},
			wantErr: true,
		},
		{
			name: "timeout too large",
			script: &Script{
				Name:        "test",
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
				Code:        []byte{1, 2, 3},
				Config: ScriptConfig{
					TimeoutMs: 120000,
				},
			},
			wantErr: true,
		},
		{
			name: "memory limit too large for extism",
			script: &Script{
				Name:        "test",
				Runtime:     RuntimeExtism,
				TriggerType: TriggerRequest,
				Code:        []byte{1, 2, 3},
				Config: ScriptConfig{
					MemoryLimitMB: 10_000,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.script.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Script.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Composer 1.
func TestScript_Validate_SetsDefaults(t *testing.T) {
	script := &Script{
		Name:        "test",
		Runtime:     RuntimeExtism,
		TriggerType: TriggerRequest,
		Code:        []byte{1, 2, 3},
	}

	err := script.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Check that defaults are set
	if script.Config.TimeoutMs != 5000 {
		t.Errorf("Expected default TimeoutMs 5000, got %d", script.Config.TimeoutMs)
	}

	if script.Config.MemoryLimitMB != 10 {
		t.Errorf("Expected default MemoryLimitMB 10, got %d", script.Config.MemoryLimitMB)
	}

	if script.CompilationStatus != CompilationNotCompiled {
		t.Errorf("Expected default CompilationStatus %s, got %s", CompilationNotCompiled, script.CompilationStatus)
	}
}

// Composer 1.
func TestScript_Validate_PreservesExistingValues(t *testing.T) {
	script := &Script{
		Name:        "test",
		Runtime:     RuntimeExtism,
		TriggerType: TriggerRequest,
		Code:        []byte{1, 2, 3},
		Config: ScriptConfig{
			TimeoutMs:     10000,
			MemoryLimitMB: 20,
		},
		CompilationStatus: CompilationSuccess,
	}

	err := script.Validate()
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	// Check that existing values are not overwritten
	if script.Config.TimeoutMs != 10000 {
		t.Errorf("Expected TimeoutMs 10000, got %d", script.Config.TimeoutMs)
	}

	if script.Config.MemoryLimitMB != 20 {
		t.Errorf("Expected MemoryLimitMB 20, got %d", script.Config.MemoryLimitMB)
	}

	if script.CompilationStatus != CompilationSuccess {
		t.Errorf("Expected CompilationStatus %s, got %s", CompilationSuccess, script.CompilationStatus)
	}
}

// Composer 1.
func TestScript_NeedsRecompilation(t *testing.T) {
	tests := []struct {
		name   string
		script *Script
		want   bool
	}{
		{
			name: "no source code",
			script: &Script{
				SourceCode: "",
			},
			want: false,
		},
		{
			name: "source code but not compiled",
			script: &Script{
				SourceCode:        "fn main() {}",
				CompilationStatus: CompilationNotCompiled,
			},
			want: true,
		},
		{
			name: "source code and successfully compiled",
			script: &Script{
				SourceCode:        "fn main() {}",
				CompilationStatus: CompilationSuccess,
				Code:              []byte{1, 2, 3},
				LastCompiledAt:    timePtr(time.Now()),
			},
			want: false,
		},
		{
			name: "source code but no code binary",
			script: &Script{
				SourceCode:        "fn main() {}",
				CompilationStatus: CompilationSuccess,
				Code:              nil,
			},
			want: true,
		},
		{
			name: "source code but no LastCompiledAt",
			script: &Script{
				SourceCode:        "fn main() {}",
				CompilationStatus: CompilationSuccess,
				Code:              []byte{1, 2, 3},
				LastCompiledAt:    nil,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.script.NeedsRecompilation()
			if got != tt.want {
				t.Errorf("Script.NeedsRecompilation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Composer 1.
func TestScript_MarkCompiling(t *testing.T) {
	script := &Script{
		CompilationStatus: CompilationNotCompiled,
		CompilationError:  "previous error",
	}

	script.MarkCompiling()

	if script.CompilationStatus != CompilationCompiling {
		t.Errorf("Expected CompilationStatus %s, got %s", CompilationCompiling, script.CompilationStatus)
	}

	if script.CompilationError != "" {
		t.Errorf("Expected empty CompilationError, got %s", script.CompilationError)
	}
}

// Composer 1.
func TestScript_MarkCompilationSuccess(t *testing.T) {
	script := &Script{
		CompilationStatus: CompilationCompiling,
		CompilationError:  "some error",
	}

	wasm := []byte{1, 2, 3, 4, 5}
	script.MarkCompilationSuccess(wasm)

	if script.CompilationStatus != CompilationSuccess {
		t.Errorf("Expected CompilationStatus %s, got %s", CompilationSuccess, script.CompilationStatus)
	}

	if len(script.Code) != len(wasm) {
		t.Errorf("Expected code length %d, got %d", len(wasm), len(script.Code))
	}

	if script.LastCompiledAt == nil {
		t.Error("Expected LastCompiledAt to be set")
	}

	if script.CompilationError != "" {
		t.Errorf("Expected empty CompilationError, got %s", script.CompilationError)
	}
}

// Composer 1.
func TestScript_MarkCompilationError(t *testing.T) {
	script := &Script{
		CompilationStatus: CompilationCompiling,
	}

	err := &CompilationError{Message: "compilation failed"}
	script.MarkCompilationError(err)

	if script.CompilationStatus != CompilationStatusError {
		t.Errorf("Expected CompilationStatus %s, got %s", CompilationStatusError, script.CompilationStatus)
	}

	if script.CompilationError != err.Error() {
		t.Errorf("Expected CompilationError %q, got %q", err.Error(), script.CompilationError)
	}
}

// Composer 1.
func TestCompilationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CompilationError
		want     string
		contains []string
	}{
		{
			name: "error with line and column",
			err: &CompilationError{
				Message:  "syntax error",
				Line:     10,
				Column:   5,
				Severity: "error",
			},
			contains: []string{"error", "line 10:5", "syntax error"},
		},
		{
			name: "error without line",
			err: &CompilationError{
				Message:  "general error",
				Severity: "error",
			},
			contains: []string{"error", "general error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			for _, contain := range tt.contains {
				if !contains(got, contain) {
					t.Errorf("Error() = %q, should contain %q", got, contain)
				}
			}
		})
	}
}

// Composer 1.
func TestNewCompilationError(t *testing.T) {
	message := "test error"
	err := NewCompilationError(message)

	if err.Message != message {
		t.Errorf("Expected Message %q, got %q", message, err.Message)
	}

	if err.Severity != "error" {
		t.Errorf("Expected Severity %q, got %q", "error", err.Severity)
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
