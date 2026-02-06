package domain

import (
	"testing"
)

func TestInterceptRule_SetDefaults(t *testing.T) {
	r := InterceptRule{}
	r.SetDefaults()
	if r.Priority != 10 {
		t.Errorf("expected priority 10, got %d", r.Priority)
	}

	r2 := InterceptRule{Priority: 5}
	r2.SetDefaults()
	if r2.Priority != 5 {
		t.Errorf("expected priority 5, got %d", r2.Priority)
	}
}

func TestInterceptRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rule    InterceptRule
		wantErr bool
	}{
		{
			name:    "empty action",
			rule:    InterceptRule{},
			wantErr: true,
		},
		{
			name:    "invalid action",
			rule:    InterceptRule{Action: "invalid"},
			wantErr: true,
		},
		{
			name:    "valid action - request",
			rule:    InterceptRule{Action: "request"},
			wantErr: false,
		},
		{
			name:    "valid action - response",
			rule:    InterceptRule{Action: "response"},
			wantErr: false,
		},
		{
			name:    "valid action - both",
			rule:    InterceptRule{Action: "both"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInterceptConfig_SetDefaults(t *testing.T) {
	c := InterceptConfig{}
	c.SetDefaults()

	if c.TimeoutMs != 60000 {
		t.Errorf("expected TimeoutMs 60000, got %d", c.TimeoutMs)
	}
	if c.QueueMax != 200 {
		t.Errorf("expected QueueMax 200, got %d", c.QueueMax)
	}
	if c.BodyMaxBytes != 1<<20 {
		t.Errorf("expected BodyMaxBytes %d, got %d", 1<<20, c.BodyMaxBytes)
	}
	if c.Overflow != "auto-continue-oldest" {
		t.Errorf("expected Overflow auto-continue-oldest, got %s", c.Overflow)
	}
}

func TestInterceptConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     InterceptConfig
		wantErr bool
	}{
		{
			name:    "zero timeout",
			cfg:     InterceptConfig{TimeoutMs: 0},
			wantErr: true,
		},
		{
			name:    "negative queueMax",
			cfg:     InterceptConfig{TimeoutMs: 1000, QueueMax: -1},
			wantErr: true,
		},
		{
			name:    "negative bodyMaxBytes",
			cfg:     InterceptConfig{TimeoutMs: 1000, QueueMax: 0, BodyMaxBytes: -1},
			wantErr: true,
		},
		{
			name:    "invalid overflow",
			cfg:     InterceptConfig{TimeoutMs: 1000, QueueMax: 10, BodyMaxBytes: 1024, Overflow: "invalid"},
			wantErr: true,
		},
		{
			name:    "valid config",
			cfg:     InterceptConfig{TimeoutMs: 60000, QueueMax: 200, BodyMaxBytes: 1 << 20, Overflow: "auto-continue-oldest"},
			wantErr: false,
		},
		{
			name:    "valid config - drop-new",
			cfg:     InterceptConfig{TimeoutMs: 60000, QueueMax: 200, BodyMaxBytes: 1 << 20, Overflow: "drop-new"},
			wantErr: false,
		},
		{
			name:    "valid config - empty overflow",
			cfg:     InterceptConfig{TimeoutMs: 60000, QueueMax: 200, BodyMaxBytes: 1 << 20},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
