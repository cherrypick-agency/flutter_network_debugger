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
	processp "network-debugger/internal/features/process/infrastructure/persistence"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	scriptingdomain "network-debugger/internal/features/scripting/domain"
	scriptingcache "network-debugger/internal/features/scripting/infrastructure/cache"
	scriptingcompilers "network-debugger/internal/features/scripting/infrastructure/compilers"
	scriptingdart "network-debugger/internal/features/scripting/infrastructure/dart"
	scriptingdownloaders "network-debugger/internal/features/scripting/infrastructure/downloaders"
	scriptingextism "network-debugger/internal/features/scripting/infrastructure/extism"
	scriptingp "network-debugger/internal/features/scripting/infrastructure/persistence"
	scriptingusecase "network-debugger/internal/features/scripting/usecase"
	sessionscli "network-debugger/internal/features/sessions_cli"
	cliopts "network-debugger/internal/features/sessions_cli/domain"
	setp "network-debugger/internal/features/settings/infrastructure/persistence"
	tagsp "network-debugger/internal/features/tags/infrastructure/persistence"
	cfgpkg "network-debugger/internal/infrastructure/config"
	dbinfra "network-debugger/internal/infrastructure/db"
	httpapi "network-debugger/internal/infrastructure/httpapi"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"
)

func main() {
	loadDotEnv()
	// CLI mode session output flags
	var cliMode bool
	var cliPreset string
	var cliFields string
	var cliBodyBytes int
	var cliColor string
	var cliFilter string
	var openBrowserOnStart bool
	// Port flags (for desktop application)
	var apiPort int
	var proxyPort int
	var dataDir string

	flag.BoolVar(&cliMode, "cli", false, "enable CLI sessions output mode")
	flag.StringVar(&cliPreset, "cli-preset", "basic", "preset: minimal|basic|advanced|full")
	flag.StringVar(&cliFields, "cli-fields", "", "comma-separated sections to show (overrides preset)")
	flag.IntVar(&cliBodyBytes, "cli-body-bytes", 0, "body preview limit (bytes); 0 = use PREVIEW_MAX_BYTES")
	flag.StringVar(&cliColor, "cli-color", "auto", "color mode: auto|always|never")
	flag.StringVar(&cliFilter, "cli-filter", "", "simple substring filter (URL/method/status)")
	flag.BoolVar(&openBrowserOnStart, "open-browser", false, "open browser on start (opt-in)")
	flag.IntVar(&apiPort, "api-port", 0, "API/UI server port (overrides ADDR)")
	flag.IntVar(&proxyPort, "proxy-port", 0, "forward proxy port (overrides default)")
	flag.StringVar(&dataDir, "data-dir", "", "data directory for database and files")
	flag.Parse()

	// Apply CLI flags via env variables BEFORE config initialization
	if dataDir != "" {
		// Override database and files path
		dbPath := filepath.Join(dataDir, "network_debugger.db")
		os.Setenv("DB_PATH", dbPath)
		// Also create directory if it doesn't exist
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create data directory: %v\n", err)
			os.Exit(1)
		}
	}

	if proxyPort > 0 {
		// Set default port for forward proxy (used during first DB initialization)
		os.Setenv("FORWARD_PROXY_DEFAULT_PORT", fmt.Sprintf("%d", proxyPort))
	}

	cfg := cfgpkg.FromEnv()
	if cliMode {
		cfg.NoBrowser = true
	}

	// Apply CLI flags on top of configuration
	if apiPort > 0 {
		cfg.Addr = fmt.Sprintf(":%d", apiPort)
	}

	logger := obs.NewLogger(cfg.LogLevel)
	logger.Info().Str("addr", cfg.Addr).Msg("starting network-debugger")

	metrics := obs.NewMetrics()

	store := memory.NewStore(500, 10000, 2*time.Hour)
	svc := usecase.NewSessionService(store, store, store)
	deps := &httpapi.Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Svc: svc, Monitor: httpapi.NewMonitorHub()}

	// Init SQLite (GORM) - will be used by features (settings/compose)
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
			// Automatic table creation on first run
			if err := gdb.AutoMigrate(
				&setp.RuntimeSettingsModel{},
				&setp.ThrottleProfileModel{},
				&compp.ComposeLibraryModel{},
				&compp.ComposeHistoryEntryModel{},
				&proxyp.ProxyConfigModel{},
				&mappingp.MapRuleModel{},
				&processp.ProcessDetectionConfigModel{},
				&processp.IconCacheModel{},
				// Tags & Annotations
				&tagsp.PredefinedTagModel{},
				&tagsp.SessionTagModel{},
				&tagsp.SessionAnnotationModel{},
				&scriptingp.ScriptModel{},
			); err != nil {
				logger.Error().Err(err).Msg("db automigrate failed")
			}
			// Additional indexes/constraints not present in GORM models (composite UNIQUE)
			// This fixes SQLite error: "ON CONFLICT clause does not match any PRIMARY KEY or UNIQUE constraint".
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

		// Initialize Scripting feature if DB is available
		if deps.DB != nil {
			// Create script repository
			scriptRepo := scriptingp.NewGormScriptRepository(deps.DB)

			// Create script service
			scriptService := scriptingusecase.NewScriptService(scriptRepo)

			// Register Extism executor (WASM) with compilation cache
			extismExecutor, err := scriptingextism.NewExtismExecutor("./cache/wasm")
			if err != nil {
				logger.Error().Err(err).Msg("failed to create Extism executor")
			} else {
				scriptService.RegisterExecutor(extismExecutor)

				// Pre-warm plugins at startup to eliminate cold start
				ctx := context.Background()
				if scriptPtrs, err := scriptRepo.List(ctx, scriptingdomain.ScriptFilter{}); err == nil {
					// Convert []*Script to []Script for WarmUpPlugins
					scripts := make([]scriptingdomain.Script, len(scriptPtrs))
					for i, s := range scriptPtrs {
						scripts[i] = *s
					}
					if err := extismExecutor.WarmUpPlugins(ctx, scripts); err != nil {
						logger.Warn().Err(err).Msg("failed to warm up some plugins")
					}
				}
			}

			// Register Dart executor (subprocess) - optional, only if Dart SDK is available
			if dartExecutor, err := scriptingdart.NewDartExecutor(1, "scripts/dart/script_runner.dart"); err == nil {
				scriptService.RegisterExecutor(dartExecutor)
				logger.Info().Msg("Scripting feature initialized with Extism and Dart executors")
			} else {
				logger.Info().Msg("Scripting feature initialized with Extism executor (Dart not available)")
			}

			// Assign to deps
			deps.ScriptSvc = scriptService

			// Initialize Compilation service with WASM validator
			wasmValidator := scriptingextism.NewExtismWASMValidator()
			compilationService := scriptingusecase.NewCompilationService(scriptRepo, wasmValidator)

			// Register compilers (ADAPTERS)

			// 1. Python (Python → Rust+RustPython → WASM)
			pythonCompiler := scriptingcompilers.NewPythonCompiler()
			compilationService.RegisterCompiler(pythonCompiler)
			if pythonCompiler.IsAvailable() {
				logger.Info().Msg("registered Python compiler (via RustPython)")
			} else {
				logger.Warn().Msg("Python compiler not available (requires Rust toolchain)")
			}

			// Assign to deps
			deps.CompilationSvc = compilationService

			// =================================================================
			// Compiler Download-on-Demand System (Zig, Kotlin, Swift)
			// =================================================================

			// Initialize Cache Manager
			cacheManager, err := scriptingcache.NewFileSystemCache("")
			if err != nil {
				logger.Error().Err(err).Msg("failed to initialize compiler cache")
			} else {
				logger.Info().Str("path", cacheManager.GetCacheDir()).Msg("compiler cache initialized")
			}

			// Initialize Compiler Installation Service
			installationService := scriptingusecase.NewCompilerInstallationService(cacheManager)

			// Register Downloaders (ADAPTERS)

			// 1. Zig Downloader
			zigDownloader := scriptingdownloaders.NewZigDownloader()
			installationService.RegisterDownloader("zig", zigDownloader)
			logger.Info().Msg("registered Zig downloader")

			// 2. Kotlin Downloader
			kotlinDownloader := scriptingdownloaders.NewKotlinDownloader()
			installationService.RegisterDownloader("kotlin", kotlinDownloader)
			logger.Info().Msg("registered Kotlin downloader")

			// 3. Swift Downloader
			swiftDownloader := scriptingdownloaders.NewSwiftDownloader()
			installationService.RegisterDownloader("swift", swiftDownloader)
			logger.Info().Msg("registered Swift downloader")

			// 4. TinyGo Downloader
			tinygoDownloader := scriptingdownloaders.NewTinyGoDownloader()
			installationService.RegisterDownloader("tinygo", tinygoDownloader)
			logger.Info().Msg("registered TinyGo downloader")

			// 5. Rust Downloader
			rustDownloader := scriptingdownloaders.NewRustDownloader()
			installationService.RegisterDownloader("rust", rustDownloader)
			logger.Info().Msg("registered Rust downloader")

			// 6. AssemblyScript Downloader
			assemblyScriptDownloader := scriptingdownloaders.NewAssemblyScriptDownloader()
			installationService.RegisterDownloader("assemblyscript", assemblyScriptDownloader)
			logger.Info().Msg("registered AssemblyScript downloader")

			// 7. WASI-SDK Downloader
			wasisdkDownloader := scriptingdownloaders.NewWASISDKDownloader()
			installationService.RegisterDownloader("wasi-sdk", wasisdkDownloader)
			logger.Info().Msg("registered WASI-SDK downloader")

			// Register Download-on-Demand Compilers (with cache integration)

			// 3. Zig (Zig → WASM)
			zigCompiler := scriptingcompilers.NewZigCompiler(cacheManager)
			compilationService.RegisterCompiler(zigCompiler)
			if zigCompiler.IsAvailable() {
				logger.Info().Msg("registered Zig compiler (from cache or system)")
			} else {
				logger.Warn().Msg("Zig compiler not available (not installed, use download API)")
			}

			// 4. Kotlin (Kotlin → WASM)
			kotlinCompiler := scriptingcompilers.NewKotlinCompiler(cacheManager)
			compilationService.RegisterCompiler(kotlinCompiler)
			if kotlinCompiler.IsAvailable() {
				logger.Info().Msg("registered Kotlin compiler (from cache or system)")
			} else {
				logger.Warn().Msg("Kotlin compiler not available (not installed, use download API)")
			}

			// 5. Swift (Swift → WASM via SwiftWasm SDK)
			swiftCompiler := scriptingcompilers.NewSwiftCompiler(cacheManager)
			compilationService.RegisterCompiler(swiftCompiler)
			if swiftCompiler.IsAvailable() {
				logger.Info().Msg("registered Swift compiler (system Swift + SwiftWasm SDK from cache)")
			} else {
				logger.Warn().Msg("Swift compiler not available (system Swift or SwiftWasm SDK not installed)")
			}

			// 6. TinyGo (Go → WASM)
			tinygoCompiler := scriptingcompilers.NewTinyGoCompiler(cacheManager)
			compilationService.RegisterCompiler(tinygoCompiler)
			if tinygoCompiler.IsAvailable() {
				logger.Info().Msg("registered TinyGo compiler (from cache or system)")
			} else {
				logger.Warn().Msg("TinyGo compiler not available (not installed, use download API)")
			}

			// 7. Rust (Rust → WASM via Cargo)
			rustCompiler := scriptingcompilers.NewRustCompiler(cacheManager)
			compilationService.RegisterCompiler(rustCompiler)
			if rustCompiler.IsAvailable() {
				logger.Info().Msg("registered Rust compiler (from cache or system)")
			} else {
				logger.Warn().Msg("Rust compiler not available (not installed, use download API)")
			}

			// 8. AssemblyScript (TypeScript/JavaScript → WASM)
			assemblyScriptCompiler := scriptingcompilers.NewAssemblyScriptCompiler(cacheManager)
			compilationService.RegisterCompiler(assemblyScriptCompiler)
			if assemblyScriptCompiler.IsAvailable() {
				logger.Info().Msg("registered AssemblyScript compiler (from cache or system)")
			} else {
				logger.Warn().Msg("AssemblyScript compiler not available (not installed, use download API)")
			}

			// 9. C/C++ (C/C++ → WASM via WASI-SDK/clang)
			cCompiler := scriptingcompilers.NewCCPPCompiler(cacheManager)
			compilationService.RegisterCompiler(cCompiler)
			if cCompiler.IsAvailable() {
				logger.Info().Msg("registered C/C++ compiler (WASI-SDK from cache or system clang)")
			} else {
				logger.Warn().Msg("C/C++ compiler not available (not installed, use download API)")
			}

			// Assign to deps
			deps.CompilerInstallationSvc = installationService
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

	// CLI mode: start printing sessions after server starts
	var cliCancel context.CancelFunc
	if cliMode {
		// Prepare options
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
	// Important: `network-debugger` is a backend/proxy without embedded Web UI.
	// Therefore auto-open is opt-in only (flag or env).
	if shouldOpenBrowser(openBrowserOnStart) && !cfg.NoBrowser {
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
	}

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

	// Close script service and plugin pools
	if deps.ScriptSvc != nil {
		if err := deps.ScriptSvc.Close(); err != nil {
			logger.Error().Err(err).Msg("script service shutdown error")
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

func shouldOpenBrowser(flagValue bool) bool {
	if flagValue {
		return true
	}
	v := strings.ToLower(strings.TrimSpace(os.Getenv("OPEN_BROWSER")))
	return v == "1" || v == "true" || v == "yes"
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
