package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
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
	compp "network-debugger/internal/features/compose/infrastructure/persistence"
	interceptp "network-debugger/internal/features/intercept/infrastructure/persistence"
	mappingp "network-debugger/internal/features/mapping/infrastructure/persistence"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	setp "network-debugger/internal/features/settings/infrastructure/persistence"
	tagsp "network-debugger/internal/features/tags/infrastructure/persistence"
	cfgpkg "network-debugger/internal/infrastructure/config"
	dbinfra "network-debugger/internal/infrastructure/db"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

//go:embed _web/*
var webDist embed.FS

func main() {
	loadDotEnv()
	cfg := cfgpkg.FromEnv()

	logger := obs.NewLogger(cfg.LogLevel)
	logger.Info().Str("addr", cfg.Addr).Msg("starting wsapp (api + embedded web)")

	// В dev-режиме легко забыть пересобрать фронтенд и смотреть на старую UI.
	// Поэтому при локальном запуске показываем "дату билда" (mtime index.html) и
	// предупреждаем, если артефакты старше суток.
	logFrontendArtifactsAge(logger, cfg)

	metrics := obs.NewMetrics()

	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}
	// Init SQLite (GORM)
	if dbPath := dbinfra.PathFromEnv(); dbPath != "" {
		created := false
		if !strings.HasPrefix(dbPath, "file:") && !strings.Contains(dbPath, ":memory") {
			if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
				created = true
			}
		}
		if gdb, err := dbinfra.NewSQLite(dbPath); err != nil {
			logger.Error().Err(err).Str("path", dbPath).Msg("db init failed")
		} else {
			deps.DB = gdb
			// Automatic table creation on first run
			if err := gdb.AutoMigrate(
				&setp.RuntimeSettingsModel{},
				&setp.ThrottleProfileModel{},
				&compp.ComposeLibraryModel{},
				&compp.ComposeHistoryEntryModel{},
				&proxyp.ProxyConfigModel{},
				&mappingp.MapRuleModel{},
				// Tags & Annotations
				&tagsp.PredefinedTagModel{},
				&tagsp.SessionTagModel{},
				&tagsp.SessionAnnotationModel{},
				&interceptp.InterceptRuleModel{},
				&interceptp.InterceptConfigModel{},
			); err != nil {
				logger.Error().Err(err).Msg("db automigrate failed")
			}
			// Ensure composite UNIQUE indexes for conflict handling
			if err := gdb.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_session_tags_unique ON session_tags(session_id, tag_name)`).Error; err != nil {
				logger.Warn().Err(err).Msg("failed to ensure idx_session_tags_unique")
			}
			if err := gdb.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_session_annotations_unique ON session_annotations(session_id, key)`).Error; err != nil {
				logger.Warn().Err(err).Msg("failed to ensure idx_session_annotations_unique")
			}
			if created {
				logger.Info().Str("db", dbPath).Msg("db not found, created new (sqlite)")
			} else {
				logger.Info().Str("db", dbPath).Msg("db connected (sqlite)")
			}
		}
	}
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
			if _, err := os.Stat(cfg.MITMCACertFile); os.IsNotExist(err) {
				certPEM, keyPEM, gerr := httpapi.GenerateDevCA("network-debugger dev CA", 5)
				if gerr != nil {
					logger.Error().Err(gerr).Msg("failed to generate default dev CA")
				} else {
					_ = os.WriteFile(cfg.MITMCACertFile, certPEM, 0o644)
					_ = os.WriteFile(cfg.MITMCAKeyFile, keyPEM, 0o600)
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

	// Try to open browser with Web UI once server is up
	go func() {
		// small delay to allow server to bind
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
		// open root (Flutter web will route to default page)
		url := addr + "/"
		_ = openBrowser(url)
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
		if i := strings.Index(line, "#"); i >= 0 { // strip trailing comments
			line = strings.TrimSpace(line[:i])
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:i])
		v := strings.TrimSpace(line[i+1:])
		v = strings.Trim(v, "\"'")
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}
