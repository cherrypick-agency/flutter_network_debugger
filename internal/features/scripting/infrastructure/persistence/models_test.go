package persistence

import (
	"encoding/json"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// Composer 1.
func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name    string
		array   StringArray
		wantNil bool
	}{
		{
			name:    "empty array",
			array:   StringArray{},
			wantNil: true,
		},
		{
			name:    "nil array",
			array:   nil,
			wantNil: true,
		},
		{
			name:    "single element",
			array:   StringArray{"test"},
			wantNil: false,
		},
		{
			name:    "multiple elements",
			array:   StringArray{"a", "b", "c"},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.array.Value()
			if err != nil {
				t.Fatalf("Value() error = %v", err)
			}

			if tt.wantNil && value != nil {
				t.Errorf("Value() = %v, want nil", value)
			}

			if !tt.wantNil && value == nil {
				t.Error("Value() = nil, want non-nil")
			}

			if !tt.wantNil {
				// Проверяем что это валидный JSON
				bytes, ok := value.([]byte)
				if !ok {
					t.Fatalf("Value() = %T, want []byte", value)
				}

				var result []string
				if err := json.Unmarshal(bytes, &result); err != nil {
					t.Fatalf("Value() returned invalid JSON: %v", err)
				}

				if len(result) != len(tt.array) {
					t.Errorf("Value() JSON length = %d, want %d", len(result), len(tt.array))
				}
			}
		})
	}
}

// Composer 1.
func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantLen int
		wantErr bool
	}{
		{
			name:    "nil value",
			value:   nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "valid JSON bytes",
			value:   []byte(`["a","b","c"]`),
			wantLen: 3,
			wantErr: false,
		},
		{
			name:    "empty JSON array",
			value:   []byte(`[]`),
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "valid JSON string",
			value:   `["a","b"]`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "invalid type",
			value:   123,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			value:   []byte(`invalid json`),
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var array StringArray
			err := array.Scan(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(array) != tt.wantLen {
				t.Errorf("Scan() length = %d, want %d", len(array), tt.wantLen)
			}
		})
	}
}

// Composer 1.
func TestStringMap_Value(t *testing.T) {
	tests := []struct {
		name    string
		m       StringMap
		wantNil bool
	}{
		{
			name:    "empty map",
			m:       StringMap{},
			wantNil: true,
		},
		{
			name:    "nil map",
			m:       nil,
			wantNil: true,
		},
		{
			name:    "single entry",
			m:       StringMap{"key": "value"},
			wantNil: false,
		},
		{
			name:    "multiple entries",
			m:       StringMap{"a": "1", "b": "2", "c": "3"},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.m.Value()
			if err != nil {
				t.Fatalf("Value() error = %v", err)
			}

			if tt.wantNil && value != nil {
				t.Errorf("Value() = %v, want nil", value)
			}

			if !tt.wantNil && value == nil {
				t.Error("Value() = nil, want non-nil")
			}

			if !tt.wantNil {
				bytes, ok := value.([]byte)
				if !ok {
					t.Fatalf("Value() = %T, want []byte", value)
				}

				var result map[string]string
				if err := json.Unmarshal(bytes, &result); err != nil {
					t.Fatalf("Value() returned invalid JSON: %v", err)
				}

				if len(result) != len(tt.m) {
					t.Errorf("Value() JSON length = %d, want %d", len(result), len(tt.m))
				}
			}
		})
	}
}

// Composer 1.
func TestStringMap_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantLen int
		wantErr bool
	}{
		{
			name:    "nil value",
			value:   nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "valid JSON bytes",
			value:   []byte(`{"a":"1","b":"2"}`),
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "empty JSON object",
			value:   []byte(`{}`),
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "valid JSON string",
			value:   `{"a":"1","b":"2"}`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "invalid type",
			value:   123,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			value:   []byte(`invalid json`),
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m StringMap
			err := m.Scan(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(m) != tt.wantLen {
				t.Errorf("Scan() length = %d, want %d", len(m), tt.wantLen)
			}
		})
	}
}

// Composer 1.
func TestScriptModel_TableName(t *testing.T) {
	model := ScriptModel{}
	if model.TableName() != "scripts" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "scripts")
	}
}

// Composer 1.
func TestScriptModel_ToDomain(t *testing.T) {
	now := time.Now()
	model := &ScriptModel{
		ID:                "script-1",
		Name:              "Test Script",
		Description:       "Test Description",
		Runtime:           "extism",
		Code:              []byte{1, 2, 3},
		Language:          "rust",
		SourceCode:        "fn main() {}",
		Dependencies:      StringMap{"Cargo.toml": "[package]"},
		CompilationStatus: "success",
		CompilationError:  "",
		LastCompiledAt:    &now,
		TriggerType:       "request",
		Priority:          10,
		MatchMethods:      StringArray{"GET", "POST"},
		MatchHostPattern:  "api.example.com",
		MatchPathPattern:  "/users/*",
		PatternType:       "wildcard",
		TimeoutMs:         5000,
		MemoryLimitMB:     10,
		AllowedHosts:      StringArray{"example.com"},
		Enabled:           true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	script := model.ToDomain()

	if script.ID != model.ID {
		t.Errorf("ToDomain() ID = %q, want %q", script.ID, model.ID)
	}

	if script.Runtime != domain.RuntimeExtism {
		t.Errorf("ToDomain() Runtime = %v, want %v", script.Runtime, domain.RuntimeExtism)
	}

	if len(script.MatchRules.Methods) != 2 {
		t.Errorf("ToDomain() MatchRules.Methods length = %d, want 2", len(script.MatchRules.Methods))
	}

	if script.Priority != 10 {
		t.Errorf("ToDomain() Priority = %d, want 10", script.Priority)
	}
}

// Composer 1.
func TestFromDomain(t *testing.T) {
	now := time.Now()
	script := &domain.Script{
		ID:                "script-2",
		Name:              "Test Script 2",
		Description:       "Description 2",
		Runtime:           domain.RuntimeDart,
		Code:              []byte{4, 5, 6},
		Language:          "dart",
		SourceCode:        "void main() {}",
		Dependencies:      map[string]string{"pubspec.yaml": "name: test"},
		CompilationStatus: domain.CompilationSuccess,
		CompilationError:  "",
		LastCompiledAt:    &now,
		TriggerType:       domain.TriggerBoth,
		Priority:          20,
		MatchRules: domain.MatchRules{
			Methods:     []string{"PUT", "DELETE"},
			HostPattern: "api.test.com",
			PathPattern: "/api/*",
			PatternType: domain.PatternRegex,
		},
		Config: domain.ScriptConfig{
			TimeoutMs:     10000,
			MemoryLimitMB: 20,
			AllowedHosts:  []string{"test.com"},
		},
		Enabled:   false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	model := FromDomain(script)

	if model.ID != script.ID {
		t.Errorf("FromDomain() ID = %q, want %q", model.ID, script.ID)
	}

	if model.Runtime != "dart" {
		t.Errorf("FromDomain() Runtime = %q, want %q", model.Runtime, "dart")
	}

	if len(model.MatchMethods) != 2 {
		t.Errorf("FromDomain() MatchMethods length = %d, want 2", len(model.MatchMethods))
	}

	if model.Priority != 20 {
		t.Errorf("FromDomain() Priority = %d, want 20", model.Priority)
	}

	if model.Enabled {
		t.Error("FromDomain() Enabled = true, want false")
	}
}

// Composer 1.
func TestStringArray_RoundTrip(t *testing.T) {
	original := StringArray{"a", "b", "c"}

	// Value -> JSON
	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	// JSON -> Scan
	var scanned StringArray
	err = scanned.Scan(value)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(scanned) != len(original) {
		t.Errorf("RoundTrip length = %d, want %d", len(scanned), len(original))
	}

	for i, v := range original {
		if scanned[i] != v {
			t.Errorf("RoundTrip[%d] = %q, want %q", i, scanned[i], v)
		}
	}
}

// Composer 1.
func TestStringMap_RoundTrip(t *testing.T) {
	original := StringMap{"key1": "value1", "key2": "value2"}

	// Value -> JSON
	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}

	// JSON -> Scan
	var scanned StringMap
	err = scanned.Scan(value)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(scanned) != len(original) {
		t.Errorf("RoundTrip length = %d, want %d", len(scanned), len(original))
	}

	for k, v := range original {
		if scanned[k] != v {
			t.Errorf("RoundTrip[%q] = %q, want %q", k, scanned[k], v)
		}
	}
}
