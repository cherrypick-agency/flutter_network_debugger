package observability

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func NewLogger(level string) *zerolog.Logger {
	lvl := zerolog.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = zerolog.DebugLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "error":
		lvl = zerolog.ErrorLevel
	}

	var logger zerolog.Logger

	// Determine log format: check LOG_FORMAT env var, fallback to terminal/DEV_MODE.
	logFormat := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if logFormat == "" {
		// Default: pretty console for local runs (TTY) and DEV_MODE, json otherwise.
		// This keeps structured JSON logs for non-interactive environments (files, aggregators, CI).
		if isEnvTrue("DEV_MODE") || isTerminalStdout() {
			logFormat = "console"
		} else {
			logFormat = "json"
		}
	}

	if logFormat == "console" || logFormat == "pretty" {
		// Pretty console output with colors
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05", // 24-hour format: HH:MM:SS
			NoColor:    false,      // Enable colors (respects NO_COLOR env var)
		}
		logger = zerolog.New(output).Level(lvl).With().Timestamp().Logger()
	} else {
		// JSON output (default for production)
		logger = zerolog.New(os.Stdout).Level(lvl).With().Timestamp().Logger()
	}

	return &logger
}

func isEnvTrue(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func isTerminalStdout() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
