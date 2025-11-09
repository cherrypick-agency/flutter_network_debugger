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

	// Determine log format: check LOG_FORMAT env var, fallback to DEV_MODE behavior
	logFormat := strings.ToLower(os.Getenv("LOG_FORMAT"))
	if logFormat == "" {
		// Default: console in DEV_MODE, json otherwise
		if os.Getenv("DEV_MODE") == "1" || os.Getenv("DEV_MODE") == "true" {
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
