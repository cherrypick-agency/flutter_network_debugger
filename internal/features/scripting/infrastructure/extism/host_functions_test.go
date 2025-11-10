package extism

import (
	"testing"
)

// TestIsHostAllowed verifies the host allowlist security feature
func TestIsHostAllowed(t *testing.T) {
	tests := []struct {
		name         string
		targetHost   string
		allowedHosts []string
		want         bool
	}{
		{
			name:         "empty allowlist allows all hosts",
			targetHost:   "example.com",
			allowedHosts: nil,
			want:         true,
		},
		{
			name:         "exact match allowed",
			targetHost:   "api.example.com",
			allowedHosts: []string{"api.example.com"},
			want:         true,
		},
		{
			name:         "exact match denied",
			targetHost:   "evil.com",
			allowedHosts: []string{"api.example.com"},
			want:         false,
		},
		{
			name:         "wildcard subdomain match",
			targetHost:   "api.example.com",
			allowedHosts: []string{"*.example.com"},
			want:         true,
		},
		{
			name:         "wildcard subdomain match nested",
			targetHost:   "v1.api.example.com",
			allowedHosts: []string{"*.example.com"},
			want:         true,
		},
		{
			name:         "wildcard subdomain denied wrong domain",
			targetHost:   "api.evil.com",
			allowedHosts: []string{"*.example.com"},
			want:         false,
		},
		{
			name:         "wildcard matches apex domain",
			targetHost:   "example.com",
			allowedHosts: []string{"*.example.com"},
			want:         true,
		},
		{
			name:         "host with port - allowed",
			targetHost:   "api.example.com:8080",
			allowedHosts: []string{"api.example.com"},
			want:         true,
		},
		{
			name:         "host with port - denied",
			targetHost:   "evil.com:8080",
			allowedHosts: []string{"api.example.com"},
			want:         false,
		},
		{
			name:         "multiple allowed hosts - first matches",
			targetHost:   "api.example.com",
			allowedHosts: []string{"api.example.com", "other.com"},
			want:         true,
		},
		{
			name:         "multiple allowed hosts - second matches",
			targetHost:   "other.com",
			allowedHosts: []string{"api.example.com", "other.com"},
			want:         true,
		},
		{
			name:         "multiple allowed hosts - none match",
			targetHost:   "evil.com",
			allowedHosts: []string{"api.example.com", "other.com"},
			want:         false,
		},
		{
			name:         "localhost allowed",
			targetHost:   "localhost",
			allowedHosts: []string{"localhost"},
			want:         true,
		},
		{
			name:         "localhost with port allowed",
			targetHost:   "localhost:3000",
			allowedHosts: []string{"localhost"},
			want:         true,
		},
		{
			name:         "IP address allowed",
			targetHost:   "192.168.1.1",
			allowedHosts: []string{"192.168.1.1"},
			want:         true,
		},
		{
			name:         "IP address with port allowed",
			targetHost:   "192.168.1.1:8080",
			allowedHosts: []string{"192.168.1.1"},
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHostAllowed(tt.targetHost, tt.allowedHosts)
			if got != tt.want {
				t.Errorf("isHostAllowed(%q, %v) = %v, want %v",
					tt.targetHost, tt.allowedHosts, got, tt.want)
			}
		})
	}
}
