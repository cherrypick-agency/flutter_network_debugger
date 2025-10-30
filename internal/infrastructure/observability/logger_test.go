package observability

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantLevel zerolog.Level
	}{
		{"default info", "", zerolog.InfoLevel},
		{"lowercase debug", "debug", zerolog.DebugLevel},
		{"uppercase debug", "DEBUG", zerolog.DebugLevel},
		{"lowercase warn", "warn", zerolog.WarnLevel},
		{"uppercase warn", "WARN", zerolog.WarnLevel},
		{"lowercase error", "error", zerolog.ErrorLevel},
		{"uppercase error", "ERROR", zerolog.ErrorLevel},
		{"lowercase info", "info", zerolog.InfoLevel},
		{"uppercase info", "INFO", zerolog.InfoLevel},
		{"unknown level", "unknown", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.level)
			if logger == nil {
				t.Fatal("NewLogger returned nil")
			}

			// Verify the logger has the correct level
			actualLevel := logger.GetLevel()
			if actualLevel != tt.wantLevel {
				t.Errorf("logger level = %v, want %v", actualLevel, tt.wantLevel)
			}
		})
	}
}

func TestNewLogger_NotNil(t *testing.T) {
	logger := NewLogger("debug")
	if logger == nil {
		t.Error("NewLogger should never return nil")
	}
}

func TestNewLogger_CanLog(t *testing.T) {
	// Verify that the logger can actually be used without panicking
	logger := NewLogger("debug")

	// These should not panic
	logger.Debug().Msg("test debug")
	logger.Info().Msg("test info")
	logger.Warn().Msg("test warn")
	logger.Error().Msg("test error")
}
