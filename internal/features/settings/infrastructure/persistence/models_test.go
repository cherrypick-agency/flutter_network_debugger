package persistence

import (
	"testing"
	"time"

	sdomain "network-debugger/internal/features/settings/domain"
)

func TestToSettingsDomain(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		model *RuntimeSettingsModel
		want  sdomain.RuntimeSettings
	}{
		{
			name:  "nil model returns default",
			model: nil,
			want:  sdomain.RuntimeSettings{ID: 1},
		},
		{
			name: "basic settings",
			model: &RuntimeSettingsModel{
				ID:               1,
				ResponseDelayMs:  100,
				ThrottleEnabled:  true,
				ThrottleDownKbps: 1000,
				ThrottleUpKbps:   500,
				UpdatedAt:        now,
			},
			want: sdomain.RuntimeSettings{
				ID:               1,
				ResponseDelayMs:  100,
				ThrottleEnabled:  true,
				ThrottleDownKbps: 1000,
				ThrottleUpKbps:   500,
				UpdatedAt:        now,
			},
		},
		{
			name: "with all throttle settings",
			model: &RuntimeSettingsModel{
				ID:                    1,
				ThrottleEnabled:       true,
				ThrottleDownKbps:      3000,
				ThrottleUpKbps:        1000,
				ThrottlePacketLoss:    5,
				ThrottleLatencyMs:     50,
				ThrottleLatencyJitter: 10,
				ThrottleOffline:       false,
				UpdatedAt:             now,
			},
			want: sdomain.RuntimeSettings{
				ID:                    1,
				ThrottleEnabled:       true,
				ThrottleDownKbps:      3000,
				ThrottleUpKbps:        1000,
				ThrottlePacketLoss:    5,
				ThrottleLatencyMs:     50,
				ThrottleLatencyJitter: 10,
				ThrottleOffline:       false,
				UpdatedAt:             now,
			},
		},
		{
			name: "with UI settings",
			model: &RuntimeSettingsModel{
				ID:                  1,
				FontScale:           1.2,
				HighlightTheme:      "monokai",
				HighlightThemeLight: "github",
				HighlightThemeDark:  "dracula",
				UpdatedAt:           now,
			},
			want: sdomain.RuntimeSettings{
				ID:                  1,
				FontScale:           1.2,
				HighlightTheme:      "monokai",
				HighlightThemeLight: "github",
				HighlightThemeDark:  "dracula",
				UpdatedAt:           now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSettingsDomain(tt.model)
			if got.ID != tt.want.ID {
				t.Errorf("toSettingsDomain().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.ResponseDelayMs != tt.want.ResponseDelayMs {
				t.Errorf("toSettingsDomain().ResponseDelayMs = %v, want %v", got.ResponseDelayMs, tt.want.ResponseDelayMs)
			}
			if got.ThrottleEnabled != tt.want.ThrottleEnabled {
				t.Errorf("toSettingsDomain().ThrottleEnabled = %v, want %v", got.ThrottleEnabled, tt.want.ThrottleEnabled)
			}
			if got.FontScale != tt.want.FontScale {
				t.Errorf("toSettingsDomain().FontScale = %v, want %v", got.FontScale, tt.want.FontScale)
			}
			if got.HighlightTheme != tt.want.HighlightTheme {
				t.Errorf("toSettingsDomain().HighlightTheme = %v, want %v", got.HighlightTheme, tt.want.HighlightTheme)
			}
			if got.HighlightThemeLight != tt.want.HighlightThemeLight {
				t.Errorf("toSettingsDomain().HighlightThemeLight = %v, want %v", got.HighlightThemeLight, tt.want.HighlightThemeLight)
			}
			if got.HighlightThemeDark != tt.want.HighlightThemeDark {
				t.Errorf("toSettingsDomain().HighlightThemeDark = %v, want %v", got.HighlightThemeDark, tt.want.HighlightThemeDark)
			}
		})
	}
}

func TestToProfileModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		profile sdomain.ThrottleProfile
		want    *ThrottleProfileModel
	}{
		{
			name: "basic profile",
			profile: sdomain.ThrottleProfile{
				ID:        "profile-1",
				Name:      "Fast 3G",
				DownKbps:  750,
				UpKbps:    250,
				CreatedAt: now,
				UpdatedAt: now,
			},
			want: &ThrottleProfileModel{
				ID:        "profile-1",
				Name:      "Fast 3G",
				DownKbps:  750,
				UpKbps:    250,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "with packet loss and latency",
			profile: sdomain.ThrottleProfile{
				ID:            "profile-2",
				Name:          "Slow Network",
				DownKbps:      100,
				UpKbps:        50,
				PacketLossPct: 10,
				LatencyMs:     200,
				LatencyJitter: 50,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			want: &ThrottleProfileModel{
				ID:            "profile-2",
				Name:          "Slow Network",
				DownKbps:      100,
				UpKbps:        50,
				PacketLossPct: 10,
				LatencyMs:     200,
				LatencyJitter: 50,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
		{
			name: "offline profile",
			profile: sdomain.ThrottleProfile{
				ID:        "profile-3",
				Name:      "Offline",
				Offline:   true,
				CreatedAt: now,
				UpdatedAt: now,
			},
			want: &ThrottleProfileModel{
				ID:        "profile-3",
				Name:      "Offline",
				Offline:   true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toProfileModel(tt.profile)
			if got.ID != tt.want.ID {
				t.Errorf("toProfileModel().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("toProfileModel().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.DownKbps != tt.want.DownKbps {
				t.Errorf("toProfileModel().DownKbps = %v, want %v", got.DownKbps, tt.want.DownKbps)
			}
			if got.PacketLossPct != tt.want.PacketLossPct {
				t.Errorf("toProfileModel().PacketLossPct = %v, want %v", got.PacketLossPct, tt.want.PacketLossPct)
			}
		})
	}
}

func TestToProfileDomain(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		model *ThrottleProfileModel
		want  sdomain.ThrottleProfile
	}{
		{
			name: "basic profile",
			model: &ThrottleProfileModel{
				ID:        "profile-1",
				Name:      "4G",
				DownKbps:  10000,
				UpKbps:    5000,
				CreatedAt: now,
				UpdatedAt: now,
			},
			want: sdomain.ThrottleProfile{
				ID:        "profile-1",
				Name:      "4G",
				DownKbps:  10000,
				UpKbps:    5000,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			name: "with latency",
			model: &ThrottleProfileModel{
				ID:            "profile-2",
				Name:          "High Latency",
				DownKbps:      1000,
				UpKbps:        500,
				LatencyMs:     500,
				LatencyJitter: 100,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			want: sdomain.ThrottleProfile{
				ID:            "profile-2",
				Name:          "High Latency",
				DownKbps:      1000,
				UpKbps:        500,
				LatencyMs:     500,
				LatencyJitter: 100,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toProfileDomain(tt.model)
			if got.ID != tt.want.ID {
				t.Errorf("toProfileDomain().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("toProfileDomain().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.DownKbps != tt.want.DownKbps {
				t.Errorf("toProfileDomain().DownKbps = %v, want %v", got.DownKbps, tt.want.DownKbps)
			}
		})
	}
}
