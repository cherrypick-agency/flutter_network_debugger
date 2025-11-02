package main

import (
	"context"
	"errors"
	"flag"
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
	compp "network-debugger/internal/features/compose/infrastructure/persistence"
	mappingp "network-debugger/internal/features/mapping/infrastructure/persistence"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	sessionscli "network-debugger/internal/features/sessions_cli"
	cliopts "network-debugger/internal/features/sessions_cli/domain"
	setp "network-debugger/internal/features/settings/infrastructure/persistence"
	cfgpkg "network-debugger/internal/infrastructure/config"
	dbinfra "network-debugger/internal/infrastructure/db"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

func main() {
	loadDotEnv()
	// Флаги CLI режима вывода сессий
	var cliMode bool
	var cliPreset string
	var cliFields string
	var cliBodyBytes int
	var cliColor string
	var cliFilter string

	flag.BoolVar(&cliMode, "cli", false, "enable CLI sessions output mode")
	flag.StringVar(&cliPreset, "cli-preset", "basic", "preset: minimal|basic|advanced|full")
	flag.StringVar(&cliFields, "cli-fields", "", "comma-separated sections to show (overrides preset)")
	flag.IntVar(&cliBodyBytes, "cli-body-bytes", 0, "body preview limit (bytes); 0 = use PREVIEW_MAX_BYTES")
	flag.StringVar(&cliColor, "cli-color", "auto", "color mode: auto|always|never")
	flag.StringVar(&cliFilter, "cli-filter", "", "simple substring filter (URL/method/status)")
	flag.Parse()

	cfg := cfgpkg.FromEnv()
	if cliMode {
		cfg.NoBrowser = true
	}

	logger := obs.NewLogger(cfg.LogLevel)
	logger.Info().Str("addr", cfg.Addr).Msg("starting network-debugger")

	metrics := obs.NewMetrics()

	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}

	// Init SQLite (GORM) — будет использовано фичами (settings/compose)
	if dbPath := dbinfra.PathFromEnv(); dbPath != "" {
		// detect first-run db file creation (best effort)
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
			// миграции фич
			if cfg.DevMode {
				if err := gdb.AutoMigrate(&setp.RuntimeSettingsModel{}, &setp.ThrottleProfileModel{}, &compp.ComposeLibraryModel{}, &compp.ComposeHistoryEntryModel{}, &proxyp.ProxyConfigModel{}, &mappingp.MapRuleModel{}); err != nil {
					logger.Error().Err(err).Msg("db automigrate failed")
				}
			} else {
				logger.Info().Msg("auto-migrate disabled (non-dev). Apply SQL migrations via goose/migrate in CI/CD")
			}
			if created {
				logger.Info().Str("db", dbPath).Msg("db not found, created new (sqlite)")
			} else {
				logger.Info().Str("db", dbPath).Msg("db connected (sqlite)")
			}
		}
	}
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

	// CLI режим: запускаем печать сессий после старта сервера
	var cliCancel context.CancelFunc
	if cliMode {
		// Подготовим опции
		opts := cliopts.Options{Preset: cliPreset, BodyPreviewBytes: cliBodyBytes, Filter: cliFilter}
		switch strings.ToLower(cliColor) {
		case "always":
			opts.Color = cliopts.ColorAlways
		case "never":
			opts.Color = cliopts.ColorNever
		default:
			opts.Color = cliopts.ColorAuto
		}
		if fields := strings.TrimSpace(cliFields); fields != "" {
			opts.Fields = map[string]bool{}
			for _, f := range strings.Split(fields, ",") {
				f = strings.TrimSpace(f)
				if f != "" {
					opts.Fields[f] = true
				}
			}
		}
		if opts.BodyPreviewBytes <= 0 {
			opts.BodyPreviewBytes = cfg.PreviewMaxBytes
		}
		var ctx context.Context
		ctx, cliCancel = context.WithCancel(context.Background())
		go func() {
			if err := sessionscli.Run(ctx, deps, opts, os.Stdout); err != nil {
				logger.Error().Err(err).Msg("cli sessions printer stopped with error")
			}
		}()
	}

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
	if cliCancel != nil {
		cliCancel()
	}
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
