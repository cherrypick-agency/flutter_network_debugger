package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	cfgpkg "network-debugger/internal/infrastructure/config"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

//go:embed _web/*
var webDist embed.FS

func main() {
	cfg := cfgpkg.FromEnv()

	logger := obs.NewLogger(cfg.LogLevel)
	logger.Info().Str("addr", cfg.Addr).Msg("starting wsapp (api + embedded web)")

	metrics := obs.NewMetrics()

	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}

	// API handler (no forward proxy for static)
	apiRouter := httpapi.NewRouterWithDeps(deps)

	// Sub FS to web root
	sub, err := fs.Sub(webDist, "_web")
	if err != nil {
		logger.Error().Err(err).Msg("failed to mount embedded web FS")
		os.Exit(1)
	}
	mux := httpapi.NewEmbeddedWebMux(apiRouter, sub)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("server error")
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("server shutdown error")
	}
	logger.Info().Msg("wsapp stopped")
}
