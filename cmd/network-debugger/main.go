package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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
	loadDotEnv()
	cfg := cfgpkg.FromEnv()

	logger := obs.NewLogger(cfg.LogLevel)
	logger.Info().Str("addr", cfg.Addr).Msg("starting network-debugger")

	metrics := obs.NewMetrics()

	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	// init MITM (generate default CA if paths are not provided)
	if cfg.MITMEnabled {
		if cfg.MITMCACertFile == "" || cfg.MITMCAKeyFile == "" {
			base := "data"
			_ = os.MkdirAll(base, 0o755)
			if cfg.MITMCACertFile == "" {
				cfg.MITMCACertFile = filepath.Join(base, "mitm_dev_ca.crt")
			}
			if cfg.MITMCAKeyFile == "" {
				cfg.MITMCAKeyFile = filepath.Join(base, "mitm_dev_ca.key")
			}
			// If files do not exist, generate and persist a dev CA
			if _, err := os.Stat(cfg.MITMCACertFile); os.IsNotExist(err) {
				certPEM, keyPEM, gerr := httpapi.GenerateDevCA("network-debugger dev CA", 5)
				if gerr != nil {
					logger.Error().Err(gerr).Msg("failed to generate default dev CA")
				} else {
					if err := os.WriteFile(cfg.MITMCACertFile, certPEM, 0o644); err != nil {
						logger.Error().Err(err).Msg("failed to write CA cert file")
					}
					if err := os.WriteFile(cfg.MITMCAKeyFile, keyPEM, 0o600); err != nil {
						logger.Error().Err(err).Msg("failed to write CA key file")
					}
					logger.Info().Str("cert", cfg.MITMCACertFile).Str("key", cfg.MITMCAKeyFile).Msg("generated default dev CA")
				}
			}
		}
		if cfg.MITMCACertFile != "" && cfg.MITMCAKeyFile != "" {
			if ca, err := httpapi.LoadCertAuthority(cfg.MITMCACertFile, cfg.MITMCAKeyFile); err != nil {
				logger.Error().Err(err).Msg("mitm init failed")
			} else {
				deps.MITM = &httpapi.MITM{CA: ca, AllowSuffix: cfg.MITMDomainsAllow, DenySuffix: cfg.MITMDomainsDeny}
				logger.Info().Msg("MITM enabled for forward proxy")
			}
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
		if cfg.DevMode || cfg.NoBrowser {
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

// loadDotEnv loads key=value pairs from a local .env file if present.
// Only sets variables that are not already defined in the environment.
func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// allow comments at line end
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		// trim surrounding quotes
		v = strings.Trim(v, "\"'")
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}
