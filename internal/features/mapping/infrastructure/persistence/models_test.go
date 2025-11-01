package persistence

import (
	"reflect"
	"testing"

	mdomain "network-debugger/internal/features/mapping/domain"
)

func TestToModel(t *testing.T) {
	tests := []struct {
		name   string
		domain mdomain.MapRule
		want   *MapRuleModel
	}{
		{
			name: "basic mapping rule",
			domain: mdomain.MapRule{
				ID:       "test-1",
				Enabled:  true,
				Priority: 10,
				Kind:     mdomain.KindLocal,
			},
			want: &MapRuleModel{
				ID:          "test-1",
				Enabled:     true,
				Priority:    10,
				Kind:        "local",
				MethodsJSON: "[]",
			},
		},
		{
			name: "with methods",
			domain: mdomain.MapRule{
				ID:       "test-2",
				Enabled:  true,
				Priority: 5,
				Kind:     mdomain.KindRemote,
				Methods:  []string{"GET", "POST"},
			},
			want: &MapRuleModel{
				ID:          "test-2",
				Enabled:     true,
				Priority:    5,
				Kind:        "remote",
				MethodsJSON: `["GET","POST"]`,
			},
		},
		{
			name: "empty kind defaults to remote",
			domain: mdomain.MapRule{
				ID:       "test-3",
				Enabled:  false,
				Priority: 1,
			},
			want: &MapRuleModel{
				ID:          "test-3",
				Enabled:     false,
				Priority:    1,
				Kind:        "remote",
				MethodsJSON: "[]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toModel(tt.domain)
			if got.ID != tt.want.ID {
				t.Errorf("toModel().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Enabled != tt.want.Enabled {
				t.Errorf("toModel().Enabled = %v, want %v", got.Enabled, tt.want.Enabled)
			}
			if got.Priority != tt.want.Priority {
				t.Errorf("toModel().Priority = %v, want %v", got.Priority, tt.want.Priority)
			}
			if got.Kind != tt.want.Kind {
				t.Errorf("toModel().Kind = %v, want %v", got.Kind, tt.want.Kind)
			}
			if got.MethodsJSON != tt.want.MethodsJSON {
				t.Errorf("toModel().MethodsJSON = %v, want %v", got.MethodsJSON, tt.want.MethodsJSON)
			}
		})
	}
}

func TestToDomain(t *testing.T) {
	tests := []struct {
		name  string
		model *MapRuleModel
		want  mdomain.MapRule
	}{
		{
			name: "basic model",
			model: &MapRuleModel{
				ID:          "test-1",
				Enabled:     true,
				Priority:    10,
				Kind:        "local",
				MethodsJSON: "[]",
			},
			want: mdomain.MapRule{
				ID:       "test-1",
				Enabled:  true,
				Priority: 10,
				Kind:     mdomain.Kind("local"),
				Methods:  []string{},
			},
		},
		{
			name: "with methods",
			model: &MapRuleModel{
				ID:          "test-2",
				Enabled:     false,
				Priority:    5,
				Kind:        "remote",
				MethodsJSON: `["GET","POST","PUT"]`,
			},
			want: mdomain.MapRule{
				ID:       "test-2",
				Enabled:  false,
				Priority: 5,
				Kind:     mdomain.Kind("remote"),
				Methods:  []string{"GET", "POST", "PUT"},
			},
		},
		{
			name: "invalid methods json",
			model: &MapRuleModel{
				ID:          "test-3",
				Enabled:     true,
				Priority:    1,
				Kind:        "local",
				MethodsJSON: "invalid json",
			},
			want: mdomain.MapRule{
				ID:       "test-3",
				Enabled:  true,
				Priority: 1,
				Kind:     mdomain.Kind("local"),
				Methods:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDomain(tt.model)
			if got.ID != tt.want.ID {
				t.Errorf("toDomain().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Enabled != tt.want.Enabled {
				t.Errorf("toDomain().Enabled = %v, want %v", got.Enabled, tt.want.Enabled)
			}
			if got.Priority != tt.want.Priority {
				t.Errorf("toDomain().Priority = %v, want %v", got.Priority, tt.want.Priority)
			}
			if got.Kind != tt.want.Kind {
				t.Errorf("toDomain().Kind = %v, want %v", got.Kind, tt.want.Kind)
			}
			if !reflect.DeepEqual(got.Methods, tt.want.Methods) {
				t.Errorf("toDomain().Methods = %v, want %v", got.Methods, tt.want.Methods)
			}
		})
	}
}
