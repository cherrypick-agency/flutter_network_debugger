package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	pdomain "network-debugger/internal/features/proxy/domain"
	"network-debugger/internal/infrastructure/config"
)

func TestNewService(t *testing.T) {
	t.Parallel()
	repo := &mockProxyRepo{}
	svc := NewService(repo)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.repo != repo {
		t.Error("repo not assigned")
	}
}

func TestService_Load(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		expected := pdomain.ProxyConfig{
			ID:             1,
			ForwardEnabled: true,
			ForwardAddr:    ":8080",
			UpdatedAt:      time.Now().UTC(),
		}
		repo := &mockProxyRepo{cfg: expected}
		svc := NewService(repo)

		result, err := svc.Load(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != expected.ID {
			t.Errorf("ID: got %d, want %d", result.ID, expected.ID)
		}
		if result.ForwardEnabled != expected.ForwardEnabled {
			t.Errorf("ForwardEnabled: got %v, want %v", result.ForwardEnabled, expected.ForwardEnabled)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		repo := &mockProxyRepo{loadErr: errors.New("db error")}
		svc := NewService(repo)

		_, err := svc.Load(ctx)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "db error" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestService_Save(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("sets UpdatedAt when zero", func(t *testing.T) {
		t.Parallel()
		repo := &mockProxyRepo{cfg: pdomain.ProxyConfig{ID: 1}}
		svc := NewService(repo)

		input := pdomain.ProxyConfig{
			ID:             1,
			ForwardEnabled: true,
			ForwardAddr:    ":9090",
		}

		before := time.Now().UTC()
		result, err := svc.Save(ctx, input)
		after := time.Now().UTC()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that UpdatedAt was set
		if repo.savedCfg.UpdatedAt.IsZero() {
			t.Error("UpdatedAt not set")
		}
		if repo.savedCfg.UpdatedAt.Before(before) || repo.savedCfg.UpdatedAt.After(after) {
			t.Errorf("UpdatedAt out of range: %v (expected between %v and %v)", repo.savedCfg.UpdatedAt, before, after)
		}

		// Verify result is from Load
		if result.ID != repo.cfg.ID {
			t.Errorf("result ID: got %d, want %d", result.ID, repo.cfg.ID)
		}
	})

	t.Run("preserves existing UpdatedAt", func(t *testing.T) {
		t.Parallel()
		repo := &mockProxyRepo{cfg: pdomain.ProxyConfig{ID: 1}}
		svc := NewService(repo)

		existingTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		input := pdomain.ProxyConfig{
			ID:        1,
			UpdatedAt: existingTime,
		}

		_, err := svc.Save(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !repo.savedCfg.UpdatedAt.Equal(existingTime) {
			t.Errorf("UpdatedAt changed: got %v, want %v", repo.savedCfg.UpdatedAt, existingTime)
		}
	})

	t.Run("save error", func(t *testing.T) {
		t.Parallel()
		repo := &mockProxyRepo{saveErr: errors.New("save failed")}
		svc := NewService(repo)

		_, err := svc.Save(ctx, pdomain.ProxyConfig{ID: 1})
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "save failed" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("load after save error", func(t *testing.T) {
		t.Parallel()
		repo := &mockProxyRepo{
			cfg:     pdomain.ProxyConfig{ID: 1},
			loadErr: errors.New("load failed"),
		}
		svc := NewService(repo)

		_, err := svc.Save(ctx, pdomain.ProxyConfig{ID: 1})
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "load failed" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestApplyOverlay(t *testing.T) {
	t.Parallel()

	t.Run("nil config no panic", func(t *testing.T) {
		t.Parallel()
		// Should not panic
		ApplyOverlay(nil, pdomain.ProxyConfig{})
	})

	t.Run("empty overlay no changes", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{}
		pc := pdomain.ProxyConfig{}

		ApplyOverlay(cfg, pc)

		// Function is currently a no-op for future compatibility
		// Just verify it doesn't panic
	})
}

// mockProxyRepo is a test double for ProxyConfigRepository
type mockProxyRepo struct {
	cfg      pdomain.ProxyConfig
	savedCfg pdomain.ProxyConfig
	loadErr  error
	saveErr  error
}

func (m *mockProxyRepo) Load(ctx context.Context) (pdomain.ProxyConfig, error) {
	if m.loadErr != nil {
		return pdomain.ProxyConfig{}, m.loadErr
	}
	return m.cfg, nil
}

func (m *mockProxyRepo) Save(ctx context.Context, c pdomain.ProxyConfig) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCfg = c
	return nil
}
