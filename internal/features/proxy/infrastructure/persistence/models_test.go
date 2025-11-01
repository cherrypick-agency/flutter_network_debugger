package persistence

import (
	"testing"
	"time"

	pdomain "network-debugger/internal/features/proxy/domain"
)

func TestToDomain(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		model *ProxyConfigModel
		want  pdomain.ProxyConfig
	}{
		{
			name:  "nil model",
			model: nil,
			want:  pdomain.ProxyConfig{ID: 1},
		},
		{
			name: "basic config",
			model: &ProxyConfigModel{
				ID:             1,
				ForwardEnabled: true,
				ForwardAddr:    ":8080",
				SocksEnabled:   false,
				UpdatedAt:      now,
			},
			want: pdomain.ProxyConfig{
				ID:             1,
				ForwardEnabled: true,
				ForwardAddr:    ":8080",
				SocksEnabled:   false,
				UpdatedAt:      now,
			},
		},
		{
			name: "with socks config",
			model: &ProxyConfigModel{
				ID:             1,
				ForwardEnabled: false,
				SocksEnabled:   true,
				SocksAddr:      ":1080",
				SocksAuthMode:  "basic",
				SocksUser:      "testuser",
				SocksPass:      "testpass",
				UpdatedAt:      now,
			},
			want: pdomain.ProxyConfig{
				ID:             1,
				ForwardEnabled: false,
				SocksEnabled:   true,
				SocksAddr:      ":1080",
				SocksAuthMode:  "basic",
				SocksUser:      "testuser",
				SocksPass:      "testpass",
				UpdatedAt:      now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toDomain(tt.model)
			if got.ID != tt.want.ID {
				t.Errorf("toDomain().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.ForwardEnabled != tt.want.ForwardEnabled {
				t.Errorf("toDomain().ForwardEnabled = %v, want %v", got.ForwardEnabled, tt.want.ForwardEnabled)
			}
			if got.ForwardAddr != tt.want.ForwardAddr {
				t.Errorf("toDomain().ForwardAddr = %v, want %v", got.ForwardAddr, tt.want.ForwardAddr)
			}
			if got.SocksEnabled != tt.want.SocksEnabled {
				t.Errorf("toDomain().SocksEnabled = %v, want %v", got.SocksEnabled, tt.want.SocksEnabled)
			}
			if got.SocksAddr != tt.want.SocksAddr {
				t.Errorf("toDomain().SocksAddr = %v, want %v", got.SocksAddr, tt.want.SocksAddr)
			}
		})
	}
}

func TestToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		config pdomain.ProxyConfig
		want   *ProxyConfigModel
	}{
		{
			name: "basic config",
			config: pdomain.ProxyConfig{
				ID:             999, // Should be overridden to 1
				ForwardEnabled: true,
				ForwardAddr:    ":9000",
				UpdatedAt:      now,
			},
			want: &ProxyConfigModel{
				ID:             1, // Always 1
				ForwardEnabled: true,
				ForwardAddr:    ":9000",
				UpdatedAt:      now,
			},
		},
		{
			name: "with socks",
			config: pdomain.ProxyConfig{
				SocksEnabled:  true,
				SocksAddr:     ":1080",
				SocksAuthMode: "none",
				UpdatedAt:     now,
			},
			want: &ProxyConfigModel{
				ID:            1,
				SocksEnabled:  true,
				SocksAddr:     ":1080",
				SocksAuthMode: "none",
				UpdatedAt:     now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toModel(tt.config)
			if got.ID != tt.want.ID {
				t.Errorf("toModel().ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.ForwardEnabled != tt.want.ForwardEnabled {
				t.Errorf("toModel().ForwardEnabled = %v, want %v", got.ForwardEnabled, tt.want.ForwardEnabled)
			}
			if got.ForwardAddr != tt.want.ForwardAddr {
				t.Errorf("toModel().ForwardAddr = %v, want %v", got.ForwardAddr, tt.want.ForwardAddr)
			}
			if got.SocksEnabled != tt.want.SocksEnabled {
				t.Errorf("toModel().SocksEnabled = %v, want %v", got.SocksEnabled, tt.want.SocksEnabled)
			}
		})
	}
}
