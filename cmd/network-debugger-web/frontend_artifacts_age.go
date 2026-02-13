package main

import (
	"os"
	"strings"
	"time"

	cfgpkg "network-debugger/internal/infrastructure/config"

	"github.com/rs/zerolog"
)

func logFrontendArtifactsAge(logger *zerolog.Logger, cfg cfgpkg.Config) {
	// Не шумим в проде: выводим только при явном DEV_MODE или при запуске через `go run`.
	if !cfg.DevMode && !looksLikeGoRun(os.Args[0]) {
		return
	}

	const (
		indexPath = "cmd/network-debugger-web/_web/index.html"
		maxAge    = 24 * time.Hour
	)

	fi, err := os.Stat(indexPath)
	if err != nil {
		// В релизном бинаре `_web` на диске может не существовать — это нормально.
		logger.Debug().
			Err(err).
			Str("path", indexPath).
			Msg("frontend build info unavailable")
		return
	}

	buildTime := fi.ModTime()
	age := time.Since(buildTime)
	if age < 0 {
		age = 0
	}

	logger.Info().
		Time("frontend_build_time", buildTime).
		Dur("age", age).
		Msg("frontend artifacts timestamp")

	if age > maxAge {
		logger.Warn().
			Time("frontend_build_time", buildTime).
			Dur("age", age).
			Msg("frontend artifacts look stale; rebuild frontend to avoid mismatches")
	}
}

func looksLikeGoRun(exePath string) bool {
	// Признак `go run`: бинарь в tmp с подстрокой `go-build`.
	return strings.Contains(exePath, "go-build")
}
