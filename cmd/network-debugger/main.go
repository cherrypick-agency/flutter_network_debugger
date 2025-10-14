package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	cfgpkg "network-debugger/internal/infrastructure/config"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

func main() {
	cfg := cfgpkg.FromEnv()

	logger := obs.NewLogger(cfg.LogLevel)
	logger.Info().Str("addr", cfg.Addr).Msg("starting network-debugger")

	metrics := obs.NewMetrics()

	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	// init MITM if configured
	if cfg.MITMEnabled && cfg.MITMCACertFile != "" && cfg.MITMCAKeyFile != "" {
		if ca, err := httpapi.LoadCertAuthority(cfg.MITMCACertFile, cfg.MITMCAKeyFile); err != nil {
			logger.Error().Err(err).Msg("mitm init failed")
		} else {
			deps.MITM = &httpapi.MITM{CA: ca, AllowSuffix: cfg.MITMDomainsAllow, DenySuffix: cfg.MITMDomainsDeny}
			logger.Info().Msg("MITM enabled for forward proxy")
		}
	}

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
		if tlsAddr == "" {
			tlsAddr = ":9443"
		}
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

	// Launch browser to downloads page on start (best-effort)
	go func() {
		time.Sleep(300 * time.Millisecond)
		if cfg.DevMode {
			return
		}
		addr := cfg.Addr
		if strings.HasPrefix(addr, ":") {
			addr = "http://localhost" + addr
		} else if !strings.HasPrefix(addr, "http") {
			addr = fmt.Sprintf("http://%s", addr)
		}
		_ = openBrowser(addr + "/")
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
	logger.Info().Msg("network-debugger stopped")
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
