package domain

import (
	"testing"
	"time"
)

// Composer 1.
func TestDetectionConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *DetectionConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &DetectionConfig{
				CacheTTLSeconds: 3600,
			},
			wantErr: false,
		},
		{
			name: "zero TTL",
			config: &DetectionConfig{
				CacheTTLSeconds: 0,
			},
			wantErr: false,
		},
		{
			name: "negative TTL",
			config: &DetectionConfig{
				CacheTTLSeconds: -1,
			},
			wantErr: true,
		},
		{
			name: "TTL too large",
			config: &DetectionConfig{
				CacheTTLSeconds: 86401,
			},
			wantErr: true,
		},
		{
			name: "TTL exactly 24 hours",
			config: &DetectionConfig{
				CacheTTLSeconds: 86400,
			},
			wantErr: false,
		},
		{
			name: "TTL one second less than 24 hours",
			config: &DetectionConfig{
				CacheTTLSeconds: 86399,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectionConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Composer 1.
func TestProcessInfo(t *testing.T) {
	now := time.Now()
	icon := &AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3},
	}
	bundleID := "com.example.app"

	info := ProcessInfo{
		PID:            1234,
		Name:           "test-app",
		ExecutablePath: "/usr/bin/test-app",
		BundleID:       &bundleID,
		Icon:           icon,
		DetectedAt:     now,
	}

	if info.PID != 1234 {
		t.Errorf("Expected PID 1234, got %d", info.PID)
	}

	if info.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got %q", info.Name)
	}

	if info.ExecutablePath != "/usr/bin/test-app" {
		t.Errorf("Expected ExecutablePath '/usr/bin/test-app', got %q", info.ExecutablePath)
	}

	if info.BundleID == nil || *info.BundleID != bundleID {
		t.Errorf("Expected BundleID %q, got %v", bundleID, info.BundleID)
	}

	if info.Icon != icon {
		t.Error("Icon should be set")
	}

	if !info.DetectedAt.Equal(now) {
		t.Errorf("Expected DetectedAt %v, got %v", now, info.DetectedAt)
	}
}

// Composer 1.
func TestAppIcon(t *testing.T) {
	icon := AppIcon{
		Format: "png",
		Data:   []byte{1, 2, 3, 4, 5},
		Path:   stringPtr("/path/to/icon.png"),
	}

	if icon.Format != "png" {
		t.Errorf("Expected Format 'png', got %q", icon.Format)
	}

	if len(icon.Data) != 5 {
		t.Errorf("Expected Data length 5, got %d", len(icon.Data))
	}

	if icon.Path == nil || *icon.Path != "/path/to/icon.png" {
		t.Errorf("Expected Path '/path/to/icon.png', got %v", icon.Path)
	}
}

// Composer 1.
func TestAppIcon_WithoutPath(t *testing.T) {
	icon := AppIcon{
		Format: "icns",
		Data:   []byte{1, 2, 3},
		Path:   nil,
	}

	if icon.Path != nil {
		t.Error("Path should be nil")
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
