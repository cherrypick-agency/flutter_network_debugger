package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
)

// Mock repository for testing
type mockMapRulesRepository struct {
	listFunc    func(ctx context.Context) ([]mdomain.MapRule, error)
	upsertFunc  func(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error)
	deleteFunc  func(ctx context.Context, id string) error
	reorderFunc func(ctx context.Context, ids []string) error
}

func (m *mockMapRulesRepository) List(ctx context.Context) ([]mdomain.MapRule, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockMapRulesRepository) Upsert(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, r)
	}
	return r, nil
}

func (m *mockMapRulesRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockMapRulesRepository) Reorder(ctx context.Context, ids []string) error {
	if m.reorderFunc != nil {
		return m.reorderFunc(ctx, ids)
	}
	return nil
}

func TestNewService(t *testing.T) {
	repo := &mockMapRulesRepository{}
	svc := NewService(repo)
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
	if svc.repo != repo {
		t.Error("NewService() did not set repository correctly")
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name     string
		mockFunc func(ctx context.Context) ([]mdomain.MapRule, error)
		wantErr  bool
	}{
		{
			name: "success",
			mockFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
				return []mdomain.MapRule{
					{ID: "rule1", Enabled: true},
					{ID: "rule2", Enabled: false},
				}, nil
			},
			wantErr: false,
		},
		{
			name: "error",
			mockFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
				return nil, errors.New("database error")
			},
			wantErr: true,
		},
		{
			name: "empty list",
			mockFunc: func(ctx context.Context) ([]mdomain.MapRule, error) {
				return []mdomain.MapRule{}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMapRulesRepository{
				listFunc: tt.mockFunc,
			}
			svc := NewService(repo)

			rules, err := svc.List(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && rules == nil {
				t.Error("List() returned nil rules without error")
			}
		})
	}
}

func TestService_Upsert(t *testing.T) {
	tests := []struct {
		name         string
		inputRule    mdomain.MapRule
		checkCreated bool
		checkUpdated bool
		wantErr      bool
	}{
		{
			name: "new rule - sets CreatedAt and UpdatedAt",
			inputRule: mdomain.MapRule{
				ID:      "new-rule",
				Enabled: true,
			},
			checkCreated: true,
			checkUpdated: true,
			wantErr:      false,
		},
		{
			name: "existing rule with CreatedAt - sets UpdatedAt",
			inputRule: mdomain.MapRule{
				ID:        "existing-rule",
				Enabled:   true,
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			checkCreated: false,
			checkUpdated: true,
			wantErr:      false,
		},
		{
			name: "existing rule with both dates - keeps both",
			inputRule: mdomain.MapRule{
				ID:        "existing-rule",
				Enabled:   true,
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
			checkCreated: false,
			checkUpdated: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedRule mdomain.MapRule
			repo := &mockMapRulesRepository{
				upsertFunc: func(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
					receivedRule = r
					return r, nil
				},
			}
			svc := NewService(repo)

			result, err := svc.Upsert(context.Background(), tt.inputRule)
			if (err != nil) != tt.wantErr {
				t.Errorf("Upsert() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkCreated {
				if receivedRule.CreatedAt.IsZero() {
					t.Error("Upsert() did not set CreatedAt for new rule")
				}
			}

			if tt.checkUpdated {
				if receivedRule.UpdatedAt.IsZero() {
					t.Error("Upsert() did not set UpdatedAt")
				}
			}

			if !tt.wantErr && result.ID != tt.inputRule.ID {
				t.Errorf("Upsert() returned rule with ID %v, want %v", result.ID, tt.inputRule.ID)
			}
		})
	}
}

func TestService_Upsert_Error(t *testing.T) {
	repo := &mockMapRulesRepository{
		upsertFunc: func(ctx context.Context, r mdomain.MapRule) (mdomain.MapRule, error) {
			return mdomain.MapRule{}, errors.New("repository error")
		},
	}
	svc := NewService(repo)

	_, err := svc.Upsert(context.Background(), mdomain.MapRule{ID: "test"})
	if err == nil {
		t.Error("Upsert() should return error from repository")
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		mockFunc func(ctx context.Context, id string) error
		wantErr  bool
	}{
		{
			name: "success",
			id:   "rule-to-delete",
			mockFunc: func(ctx context.Context, id string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "error - not found",
			id:   "non-existent",
			mockFunc: func(ctx context.Context, id string) error {
				return errors.New("not found")
			},
			wantErr: true,
		},
		{
			name: "empty id",
			id:   "",
			mockFunc: func(ctx context.Context, id string) error {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMapRulesRepository{
				deleteFunc: tt.mockFunc,
			}
			svc := NewService(repo)

			err := svc.Delete(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Reorder(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		mockFunc func(ctx context.Context, ids []string) error
		wantErr  bool
	}{
		{
			name: "success",
			ids:  []string{"rule1", "rule2", "rule3"},
			mockFunc: func(ctx context.Context, ids []string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "error",
			ids:  []string{"rule1", "rule2"},
			mockFunc: func(ctx context.Context, ids []string) error {
				return errors.New("database error")
			},
			wantErr: true,
		},
		{
			name: "empty list",
			ids:  []string{},
			mockFunc: func(ctx context.Context, ids []string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "single item",
			ids:  []string{"rule1"},
			mockFunc: func(ctx context.Context, ids []string) error {
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockMapRulesRepository{
				reorderFunc: tt.mockFunc,
			}
			svc := NewService(repo)

			err := svc.Reorder(context.Background(), tt.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reorder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
