package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	sdomain "network-debugger/internal/features/settings/domain"
	"network-debugger/internal/infrastructure/config"
)

func TestNewService(t *testing.T) {
	t.Parallel()
	sr := &mockSettingsRepo{}
	pr := &mockProfilesRepo{}
	svc := NewService(sr, pr)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.settingsRepo != sr {
		t.Error("settingsRepo not assigned")
	}
	if svc.profilesRepo != pr {
		t.Error("profilesRepo not assigned")
	}
}

func TestService_Load(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		expected := sdomain.RuntimeSettings{
			ID:               1,
			ResponseDelayMs:  100,
			ThrottleEnabled:  true,
			ThrottleDownKbps: 1000,
			UpdatedAt:        time.Now().UTC(),
		}
		sr := &mockSettingsRepo{settings: expected}
		svc := NewService(sr, &mockProfilesRepo{})

		result, err := svc.Load(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != expected.ID {
			t.Errorf("ID: got %d, want %d", result.ID, expected.ID)
		}
		if result.ResponseDelayMs != expected.ResponseDelayMs {
			t.Errorf("ResponseDelayMs: got %d, want %d", result.ResponseDelayMs, expected.ResponseDelayMs)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		sr := &mockSettingsRepo{loadErr: errors.New("db error")}
		svc := NewService(sr, &mockProfilesRepo{})

		_, err := svc.Load(ctx)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_SaveRuntime(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("sets UpdatedAt when zero", func(t *testing.T) {
		t.Parallel()
		sr := &mockSettingsRepo{settings: sdomain.RuntimeSettings{ID: 1}}
		svc := NewService(sr, &mockProfilesRepo{})

		input := sdomain.RuntimeSettings{
			ID:              1,
			ResponseDelayMs: 200,
		}

		before := time.Now().UTC()
		result, err := svc.SaveRuntime(ctx, input)
		after := time.Now().UTC()

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if sr.savedSettings.UpdatedAt.IsZero() {
			t.Error("UpdatedAt not set")
		}
		if sr.savedSettings.UpdatedAt.Before(before) || sr.savedSettings.UpdatedAt.After(after) {
			t.Errorf("UpdatedAt out of range")
		}
		if result.ID != 1 {
			t.Errorf("result ID: got %d, want 1", result.ID)
		}
	})

	t.Run("preserves existing UpdatedAt", func(t *testing.T) {
		t.Parallel()
		sr := &mockSettingsRepo{settings: sdomain.RuntimeSettings{ID: 1}}
		svc := NewService(sr, &mockProfilesRepo{})

		existingTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		input := sdomain.RuntimeSettings{
			ID:        1,
			UpdatedAt: existingTime,
		}

		_, err := svc.SaveRuntime(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !sr.savedSettings.UpdatedAt.Equal(existingTime) {
			t.Errorf("UpdatedAt changed")
		}
	})

	t.Run("save error", func(t *testing.T) {
		t.Parallel()
		sr := &mockSettingsRepo{saveErr: errors.New("save failed")}
		svc := NewService(sr, &mockProfilesRepo{})

		_, err := svc.SaveRuntime(ctx, sdomain.RuntimeSettings{ID: 1})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_ListProfiles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		profiles := []sdomain.ThrottleProfile{
			{ID: "p1", Name: "Slow 3G", DownKbps: 400},
			{ID: "p2", Name: "Fast 3G", DownKbps: 1600},
		}
		pr := &mockProfilesRepo{profiles: profiles}
		svc := NewService(&mockSettingsRepo{}, pr)

		result, err := svc.ListProfiles(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("len: got %d, want 2", len(result))
		}
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		pr := &mockProfilesRepo{listErr: errors.New("db error")}
		svc := NewService(&mockSettingsRepo{}, pr)

		_, err := svc.ListProfiles(ctx)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_UpsertProfile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		profile := sdomain.ThrottleProfile{ID: "p1", Name: "Custom", DownKbps: 500}
		pr := &mockProfilesRepo{upsertedProfile: profile}
		svc := NewService(&mockSettingsRepo{}, pr)

		result, err := svc.UpsertProfile(ctx, profile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != "p1" {
			t.Errorf("ID: got %s, want p1", result.ID)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		pr := &mockProfilesRepo{upsertErr: errors.New("upsert failed")}
		svc := NewService(&mockSettingsRepo{}, pr)

		_, err := svc.UpsertProfile(ctx, sdomain.ThrottleProfile{ID: "p1"})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestService_DeleteProfile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		pr := &mockProfilesRepo{}
		svc := NewService(&mockSettingsRepo{}, pr)

		err := svc.DeleteProfile(ctx, "p1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pr.deletedID != "p1" {
			t.Errorf("deletedID: got %s, want p1", pr.deletedID)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()
		pr := &mockProfilesRepo{deleteErr: errors.New("delete failed")}
		svc := NewService(&mockSettingsRepo{}, pr)

		err := svc.DeleteProfile(ctx, "p1")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestApplyOverlay(t *testing.T) {
	t.Parallel()

	t.Run("nil config no panic", func(t *testing.T) {
		t.Parallel()
		ApplyOverlay(nil, sdomain.RuntimeSettings{})
	})

	t.Run("range delay applied", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{}
		rs := sdomain.RuntimeSettings{
			ResponseDelayMinMs: 100,
			ResponseDelayMaxMs: 500,
		}

		ApplyOverlay(cfg, rs)

		if cfg.ResponseDelayMinMs != 100 {
			t.Errorf("ResponseDelayMinMs: got %d, want 100", cfg.ResponseDelayMinMs)
		}
		if cfg.ResponseDelayMaxMs != 500 {
			t.Errorf("ResponseDelayMaxMs: got %d, want 500", cfg.ResponseDelayMaxMs)
		}
		if cfg.ResponseDelayMs != 0 {
			t.Errorf("ResponseDelayMs: got %d, want 0", cfg.ResponseDelayMs)
		}
	})

	t.Run("fixed delay applied", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{}
		rs := sdomain.RuntimeSettings{
			ResponseDelayMs: 300,
		}

		ApplyOverlay(cfg, rs)

		if cfg.ResponseDelayMs != 300 {
			t.Errorf("ResponseDelayMs: got %d, want 300", cfg.ResponseDelayMs)
		}
		if cfg.ResponseDelayMinMs != 0 {
			t.Errorf("ResponseDelayMinMs: got %d, want 0", cfg.ResponseDelayMinMs)
		}
		if cfg.ResponseDelayMaxMs != 0 {
			t.Errorf("ResponseDelayMaxMs: got %d, want 0", cfg.ResponseDelayMaxMs)
		}
	})

	t.Run("zero fixed delay allowed", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{ResponseDelayMs: 100}
		rs := sdomain.RuntimeSettings{
			ResponseDelayMs: 0,
		}

		ApplyOverlay(cfg, rs)

		if cfg.ResponseDelayMs != 0 {
			t.Errorf("ResponseDelayMs: got %d, want 0", cfg.ResponseDelayMs)
		}
	})

	t.Run("throttle settings applied", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{}
		rs := sdomain.RuntimeSettings{
			ThrottleEnabled:    true,
			ThrottleDownKbps:   1000,
			ThrottleUpKbps:     500,
			ThrottlePacketLoss: 5,
			ThrottleOffline:    false,
		}

		ApplyOverlay(cfg, rs)

		if cfg.ThrottleEnabled != true {
			t.Errorf("ThrottleEnabled: got %v, want true", cfg.ThrottleEnabled)
		}
		if cfg.ThrottleDownKbps != 1000 {
			t.Errorf("ThrottleDownKbps: got %d, want 1000", cfg.ThrottleDownKbps)
		}
		if cfg.ThrottleUpKbps != 500 {
			t.Errorf("ThrottleUpKbps: got %d, want 500", cfg.ThrottleUpKbps)
		}
		if cfg.ThrottlePacketLoss != 5 {
			t.Errorf("ThrottlePacketLoss: got %d, want 5", cfg.ThrottlePacketLoss)
		}
		if cfg.ThrottleOffline != false {
			t.Errorf("ThrottleOffline: got %v, want false", cfg.ThrottleOffline)
		}
	})
}

// Mock implementations

type mockSettingsRepo struct {
	settings      sdomain.RuntimeSettings
	savedSettings sdomain.RuntimeSettings
	loadErr       error
	saveErr       error
}

func (m *mockSettingsRepo) Load(ctx context.Context) (sdomain.RuntimeSettings, error) {
	if m.loadErr != nil {
		return sdomain.RuntimeSettings{}, m.loadErr
	}
	return m.settings, nil
}

func (m *mockSettingsRepo) Save(ctx context.Context, rs sdomain.RuntimeSettings) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedSettings = rs
	return nil
}

type mockProfilesRepo struct {
	profiles        []sdomain.ThrottleProfile
	upsertedProfile sdomain.ThrottleProfile
	deletedID       string
	listErr         error
	upsertErr       error
	deleteErr       error
}

func (m *mockProfilesRepo) List(ctx context.Context) ([]sdomain.ThrottleProfile, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.profiles, nil
}

func (m *mockProfilesRepo) Upsert(ctx context.Context, p sdomain.ThrottleProfile) (sdomain.ThrottleProfile, error) {
	if m.upsertErr != nil {
		return sdomain.ThrottleProfile{}, m.upsertErr
	}
	return m.upsertedProfile, nil
}

func (m *mockProfilesRepo) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedID = id
	return nil
}
