package observability

import (
	"io"
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

	logFormat := resolveLogFormat()
	out := resolveLogOutput(logFormat)
	logger := zerolog.New(out).Level(lvl).With().Timestamp().Logger()
	return &logger
}

func resolveLogFormat() string {
	// Приоритет: специфичная переменная для проекта → общая.
	logFormat := strings.ToLower(strings.TrimSpace(os.Getenv("NETWORK_DEBUGGER_LOG_FORMAT")))
	if logFormat == "" {
		logFormat = strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	}
	if logFormat != "" {
		return logFormat
	}

	// По умолчанию: локальный запуск (DEV_MODE или TTY) — читаемый вывод, иначе JSON.
	if isEnvTrue("DEV_MODE") || isTerminalStdout() {
		return "console"
	}
	return "json"
}

func resolveLogOutput(logFormat string) io.Writer {
	switch logFormat {
	case "console", "pretty":
		// Важно: в случае запуска через Dart stdout у Go — пайп,
		// поэтому цвет лучше не отключать только из-за отсутствия TTY.
		// Если нужно без цвета — можно выставить NO_COLOR.
		noColor := isNoColor()
		return zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    noColor,
		}
	default:
		return os.Stdout
	}
}

func isEnvTrue(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}

func isNoColor() bool {
	// https://no-color.org/ — если переменная есть (даже пустая), цвета быть не должно.
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return true
	}
	return false
}

func isTerminalStdout() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
