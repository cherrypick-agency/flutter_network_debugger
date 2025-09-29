package main

import (
    "context"
    "errors"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    cfgpkg "go-proxy/internal/infrastructure/config"
    obs "go-proxy/internal/infrastructure/observability"
    httpapi "go-proxy/internal/infrastructure/httpapi"
    "go-proxy/internal/adapters/storage/memory"
    "go-proxy/internal/usecase"
)

func main() {
    // Config
    cfg := cfgpkg.FromEnv()

    // Logger
    logger := obs.NewLogger(cfg.LogLevel)
    logger.Info().Str("addr", cfg.Addr).Msg("starting wsproxy")

    // Observability: metrics registry and health state
    metrics := obs.NewMetrics()

    // Storage and services
    store := memory.NewStore(500, 10000, 2*time.Hour)
    svc := usecase.NewSessionService(store, store, store)
    deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}

    // HTTP server (plain). Keeps forward-proxy CONNECT (HTTP/1.1) behavior.
    srv := &http.Server{
        Addr:              cfg.Addr,
        Handler:           httpapi.NewRouterWithDeps(deps),
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       30 * time.Second,
        WriteTimeout:      30 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    // Optional TLS server for REST/reverse with HTTP/2 (net/http enables h2 by default under TLS).
    var tlsSrv *http.Server
    if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
        tlsAddr := cfg.TLSAddr
        if tlsAddr == "" { tlsAddr = ":9443" }
        tlsSrv = &http.Server{
            Addr:              tlsAddr,
            Handler:           httpapi.NewRouterWithoutForwardProxy(deps),
            ReadHeaderTimeout: 5 * time.Second,
            ReadTimeout:       30 * time.Second,
            WriteTimeout:      30 * time.Second,
            IdleTimeout:       60 * time.Second,
        }
        go func() {
            logger.Info().Str("addr", tlsAddr).Msg("starting TLS server (HTTP/2 enabled)")
            if err := tlsSrv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
                logger.Error().Err(err).Msg("tls server error")
                os.Exit(1)
            }
        }()
    }

    // Graceful shutdown
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
    if tlsSrv != nil {
        if err := tlsSrv.Shutdown(ctx); err != nil {
            logger.Error().Err(err).Msg("tls server shutdown error")
        }
    }
    logger.Info().Msg("wsproxy stopped")
}


