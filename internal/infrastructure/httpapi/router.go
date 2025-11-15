package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"network-debugger/internal/domain"
	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/internal/usecase"

	"gorm.io/gorm"

	composep "network-debugger/internal/features/compose/infrastructure/persistence"
	mappingp "network-debugger/internal/features/mapping/infrastructure/persistence"
	mappingrt "network-debugger/internal/features/mapping/runtime"
	mappinguc "network-debugger/internal/features/mapping/usecase"
	performanceuc "network-debugger/internal/features/performance/usecase"
	processdetector "network-debugger/internal/features/process/infrastructure/detector"
	processhelper "network-debugger/internal/features/process/infrastructure/helper"
	processicon "network-debugger/internal/features/process/infrastructure/icon"
	processp "network-debugger/internal/features/process/infrastructure/persistence"
	processuc "network-debugger/internal/features/process/usecase"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	proxyuc "network-debugger/internal/features/proxy/usecase"
	scriptinguc "network-debugger/internal/features/scripting/usecase"
	settingsp "network-debugger/internal/features/settings/infrastructure/persistence"
	settingsuc "network-debugger/internal/features/settings/usecase"
	tagsp "network-debugger/internal/features/tags/infrastructure/persistence"
	tagsuc "network-debugger/internal/features/tags/usecase"
	pruntime "network-debugger/internal/infrastructure/proxyruntime"
)

type Deps struct {
	Cfg                     config.Config
	Logger                  *zerolog.Logger
	Metrics                 *obs.Metrics
	Svc                     *usecase.SessionService
	Monitor                 *MonitorHub
	Live                    *LiveSessions
	MITM                    *MITM
	Compose                 *usecase.ComposeService
	Interceptor             *InterceptorManager
	DB                      *gorm.DB
	Settings                *settingsuc.Service
	ProxySvc                *proxyuc.Service
	ProxyRt                 *pruntime.Manager
	Mapping                 *mappinguc.Service
	MapRt                   *mappingrt.Manager
	ProcessSvc              *processuc.Service
	ScriptSvc               *scriptinguc.ScriptService
	CompilationSvc          *scriptinguc.CompilationService
	CompilerInstallationSvc *scriptinguc.CompilerInstallationService
	TagsSvc                 *tagsuc.Service
	PerformanceSvc          *performanceuc.Service
}

// shouldMonitorTarget returns false if the target URL path is a monitoring endpoint
// to prevent recursive monitoring (monitoring endpoints monitoring themselves)
func (d *Deps) shouldMonitorTarget(targetPath string) bool {
	// Exclude monitoring endpoints from being monitored
	return !strings.HasPrefix(targetPath, "/_api/v1/monitor/")
}

// broadcastMonitorEvent safely broadcasts a monitor event, checking for nil Monitor first
func (d *Deps) broadcastMonitorEvent(ev domain.MonitorEvent) {
	if d.Monitor != nil {
		d.Monitor.Broadcast(ev)
	}
}

func NewRouter(cfg config.Config, logger *zerolog.Logger, metrics *obs.Metrics) http.Handler {
	// backward compatibility shim for early main.go; will be removed when deps used everywhere
	d := &Deps{Cfg: cfg, Logger: logger, Metrics: metrics, Monitor: NewMonitorHub(), Live: NewLiveSessions()}
	return NewRouterWithDeps(d)
}

func NewRouterWithDeps(d *Deps) http.Handler {
	// Initialize Compose service if not provided
	// Initialize Settings service if DB is provided and not yet set
	if d.DB != nil && d.Settings == nil {
		sr := settingsp.NewSettingsRepo(d.DB)
		pr := settingsp.NewThrottleProfilesRepo(d.DB)
		d.Settings = settingsuc.NewService(sr, pr)
		// Применим сохранённые настройки поверх env-конфига
		if cur, err := d.Settings.Load(contextWithNoCancel()); err == nil {
			settingsuc.ApplyOverlay(&d.Cfg, cur)
		}
	}

	if d.Compose == nil {
		// Только GORM-репозитории для Compose
		libRepo := composep.NewLibraryRepo(d.DB)
		histRepo := composep.NewHistoryRepo(d.DB)
		clientFactory := func() *http.Client { return &http.Client{Transport: newTransport(d.Cfg), Timeout: 30 * time.Second} }
		d.Compose = usecase.NewComposeService(libRepo, histRepo, d.Svc, d.Monitor, clientFactory)
		d.Compose.SetMaxUploadMB(d.Cfg.ComposeMaxUploadMB)
	}

	// Инициализация ProxySvc/ProxyRt
	if d.DB != nil && d.ProxySvc == nil {
		repo := proxyp.NewRepo(d.DB)
		d.ProxySvc = proxyuc.NewService(repo)
	}
	// Инициализация Mapping сервиса
	if d.DB != nil && d.Mapping == nil {
		mrepo := mappingp.NewRepo(d.DB)
		d.Mapping = mappinguc.NewService(mrepo)
	}
	// Инициализация Tags сервиса
	if d.DB != nil && d.TagsSvc == nil {
		tagsRepo := tagsp.NewRepo(d.DB)
		d.TagsSvc = tagsuc.NewService(tagsRepo)
	}
	// Инициализация Performance сервиса
	if d.Svc != nil && d.PerformanceSvc == nil {
		d.PerformanceSvc = performanceuc.NewService(d.Svc)
	}
	// Инициализация Process Detection сервиса
	if d.DB != nil && d.ProcessSvc == nil {
		processConfigRepo := processp.NewConfigRepo(d.DB)
		iconCacheRepo := processp.NewIconCacheRepo(d.DB)
		localDetector, _ := processdetector.NewDetector(false)
		iconExtractor, _ := processicon.NewExtractor()

		// Helper client и installer
		helperClient := processhelper.NewHelperClient()
		helperInstaller := processhelper.NewInstaller()

		// Путь к helper binary - используем абсолютный путь
		// Попробуем найти binary относительно executable или в стандартных местах
		helperBinaryPath := findHelperBinary(d.Logger)

		d.ProcessSvc = processuc.NewService(
			processConfigRepo,
			iconCacheRepo,
			localDetector,
			helperClient,
			iconExtractor,
			helperInstaller,
			helperBinaryPath,
			d.Logger,
		)
	}
	// Связать ProcessSvc с SessionService для автоматической детекции
	if d.ProcessSvc != nil && d.Svc != nil {
		d.Svc.SetProcessService(d.ProcessSvc)
	}
	// Scripting service initialization is now in main.go
	// This allows better control of lifecycle and dependency injection
	if d.MapRt == nil {
		d.MapRt = mappingrt.New()
		if d.Mapping != nil {
			if rules, err := d.Mapping.List(contextWithNoCancel()); err == nil {
				d.MapRt.Update(rules)
			}
		}
		d.MapRt.SetOnFileChange(func(ruleID string, path string) {
			if d.Monitor != nil {
				d.broadcastMonitorEvent(domain.MonitorEvent{Type: "mapping_file_changed", ID: "", Ref: ruleID})
			}
		})
	}
	if d.ProxyRt == nil {
		// создадим рантайм-менеджер
		var zl *zerolog.Logger = d.Logger
		if zl == nil {
			l := obs.NewLogger("info")
			zl = l
		}
		d.ProxyRt = pruntime.New(zl)
	}
	// Применим конфигурацию портов/режимов
	if d.ProxySvc != nil && d.ProxyRt != nil {
		if pc, err := d.ProxySvc.Load(contextWithNoCancel()); err == nil {
			// На порту прокси нужно уметь и forward, и reverse (/httpproxy) для SDK‑пакетов.
			forwardHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/httpproxy") || strings.HasPrefix(r.URL.Path, "/proxy") {
					d.handleHTTPProxy(w, r)
					return
				}
				if strings.HasPrefix(r.URL.Path, "/wsproxy") || strings.HasPrefix(r.URL.Path, "/_ws") {
					d.handleWSProxy(w, r)
					return
				}
				d.handleForwardOrNotFound(w, r)
			})
			_ = d.ProxyRt.Apply(contextWithNoCancel(), pruntime.ApplyConfig{
				ForwardEnabled: pc.ForwardEnabled,
				ForwardAddr:    pc.ForwardAddr,
				SocksEnabled:   pc.SocksEnabled,
				SocksAddr:      pc.SocksAddr,
				SocksAuthMode:  pc.SocksAuthMode,
				SocksUser:      pc.SocksUser,
				SocksPass:      pc.SocksPass,
			}, forwardHandler)
			if d.Logger != nil {
				d.Logger.Info().Str("addr", pc.ForwardAddr).Msg("reverse proxy endpoints enabled (/httpproxy,/wsproxy)")
			}
		}
	}
	mux := buildBaseMux(d)
	// Start background GC for spool files (best-effort)
	go startSpoolGC(d)
	// В этой сборке не перехватываем forward‑proxy на порту UI — он работает на отдельном листенере через ProxyRt.
	return withCORS(d.Cfg, mux)
}

// NewRouterWithoutForwardProxy returns the same routes but without the forward-proxy wrapper.
// Useful for TLS server where we want HTTP/2 for REST/reverse, while keeping CONNECT on the plain server.
func NewRouterWithoutForwardProxy(d *Deps) http.Handler {
	// Build mux and apply CORS, but skip withForwardProxy
	return withCORS(d.Cfg, buildBaseMux(d))
}

// buildBaseMux constructs the mux with all routes, without wrappers.
func buildBaseMux(d *Deps) *http.ServeMux {
	mux := http.NewServeMux()

	// Lazy init interceptor
	if d.Interceptor == nil {
		d.Interceptor = NewInterceptorManager(&d.Cfg, d.Monitor, d.Metrics)
		// Seed simple rules from env config (MVP convenience)
		if d.Cfg.InterceptEnabled {
			existing := d.Interceptor.ListRules()
			if len(existing) == 0 {
				rules := make([]InterceptRule, 0, 4)
				prio := 10
				mkRule := func(urlContains string, ct string) InterceptRule {
					w := InterceptWhen{Method: d.Cfg.InterceptMethods}
					if strings.TrimSpace(urlContains) != "" {
						w.Path = &RuleStringMatch{Contains: urlContains}
					}
					if strings.TrimSpace(ct) != "" {
						w.ContentType = &RuleStringMatch{Prefix: strings.ToLower(ct)}
					}
					return InterceptRule{ID: "", Enabled: true, Priority: prio, Action: "both", Once: false, StopProcessing: true, When: w}
				}
				if len(d.Cfg.InterceptURLContains) > 0 {
					for _, u := range d.Cfg.InterceptURLContains {
						if len(d.Cfg.InterceptContentTypes) > 0 {
							for _, c := range d.Cfg.InterceptContentTypes {
								rules = append(rules, mkRule(u, c))
								prio++
							}
						} else {
							rules = append(rules, mkRule(u, ""))
							prio++
						}
					}
				} else if len(d.Cfg.InterceptContentTypes) > 0 {
					for _, c := range d.Cfg.InterceptContentTypes {
						rules = append(rules, mkRule("", c))
						prio++
					}
				} else if len(d.Cfg.InterceptMethods) > 0 {
					rules = append(rules, mkRule("", ""))
				}
				if len(rules) > 0 {
					d.Interceptor.UpdateRules(rules)
				}
			}
		}
	}

	// Apply preview limit from config (<=0 disables truncation)
	previewMaxBytes.Store(int32(d.Cfg.PreviewMaxBytes))
	// Apply sensitive headers exposure flag
	exposeSensitiveHeaders.Store(d.Cfg.ExposeSensitiveHeaders)
	previewDecompress.Store(d.Cfg.PreviewDecompress)

	healthHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/_health", healthHandler) // Alias for desktop app
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	mux.Handle("/metrics", promhttp.HandlerFor(d.Metrics.Registry(), promhttp.HandlerOpts{}))

	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":    "network-debugger",
			"version": "0.1.0-mvp",
			"time":    time.Now().UTC(),
		})
	})

	// REST sessions (legacy base)
	mux.HandleFunc("/api/sessions", d.handleListSessions)
	// Single handler for /api/sessions/* to avoid duplicate registrations
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ws/send") {
			d.handleWSSendText(w, r)
			return
		}
		d.handleSessionByID(w, r)
	})

	// Monitor WS (legacy)
	mux.HandleFunc("/api/monitor/ws", d.Monitor.HandleWS)

	// WS Proxy
	mux.HandleFunc("/wsproxy", d.handleWSProxy)
	mux.HandleFunc("/wsproxy/", d.handleWSProxy)

	// HTTP Reverse Proxy (prefix-based)
	// Usage examples:
	//  - GET /httpproxy?_target=https://api.example.com               -> proxies to https://api.example.com/
	//  - GET /httpproxy/v1/users?_target=https://api.example.com      -> proxies to https://api.example.com/v1/users
	//  Query params except `_target` are forwarded to upstream.
	mux.HandleFunc("/httpproxy", d.handleHTTPProxy)
	mux.HandleFunc("/httpproxy/", d.handleHTTPProxy)
	// Unified endpoint for both HTTP reverse and WebSocket proxy
	// If the request is an upgrade to websocket, it will be handled by handleWSProxy-like flow.
	mux.HandleFunc("/proxy", d.handleUnifiedProxy)
	mux.HandleFunc("/proxy/", d.handleUnifiedProxy)

	// === V1 API ===
	mux.HandleFunc("/_api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":    "network-debugger",
			"version": "v1",
			"time":    time.Now().UTC(),
		})
	})
	mux.HandleFunc("/_api/v1/sessions", d.handleV1ListSessions)
	mux.HandleFunc("/_api/v1/sessions/aggregate", d.handleV1SessionsAggregate)
	// HAR export/import
	mux.HandleFunc("/_api/v1/export/har", d.handleExportHAR)
	mux.HandleFunc("/_api/v1/import/har", d.handleImportHAR)
	// Capture controls
	mux.HandleFunc("/_api/v1/capture", d.handleV1Capture)
	mux.HandleFunc("/_api/v1/captures", d.handleV1Captures)
	mux.HandleFunc("/_api/v1/capture/reset", d.handleV1CaptureReset)
	// Runtime settings (response delay, etc.)
	mux.HandleFunc("/_api/v1/settings", d.handleV1Settings)
	mux.HandleFunc("/_api/v1/monitor/ws", d.Monitor.HandleWS)
	// Socket.IO endpoint для push‑обновлений (sessions/aggregate)
	// StripPrefix убирает /_api/v1/monitor/io перед передачей в Socket.IO
	// Клиент подключается: /_api/v1/monitor/io/socket.io/ (path="/socket.io/")
	mux.Handle("/_api/v1/monitor/io/", http.StripPrefix("/_api/v1/monitor/io", NewSocketIOServer(d)))
	mux.HandleFunc("/_api/v1/httpproxy", d.handleHTTPProxy)
	mux.HandleFunc("/_api/v1/httpproxy/", d.handleHTTPProxy)

	// Proxy (ports, SOCKS/forward) config
	mux.HandleFunc("/_api/v1/proxy/config", d.handleV1ProxyConfig)

	// Compose / Request Builder
	mux.HandleFunc("/_api/v1/compose/send", d.handleComposeSend)
	mux.HandleFunc("/_api/v1/compose/library", d.handleComposeGetLibrary)
	mux.HandleFunc("/_api/v1/compose/library/requests", d.handleComposeUpsertRequest)
	mux.HandleFunc("/_api/v1/compose/library/requests/", d.handleComposeDeleteRequest)
	// Compose config (limits)
	mux.HandleFunc("/_api/v1/compose/config", d.handleComposeConfig)
	// Compose history
	mux.HandleFunc("/_api/v1/compose/history", d.handleComposeHistory)
	// Throttle profiles management
	mux.HandleFunc("/_api/v1/throttle/profiles", d.handleV1ThrottleProfiles)
	mux.HandleFunc("/_api/v1/throttle/profiles/", d.handleV1ThrottleProfiles)
	// collections
	mux.HandleFunc("/_api/v1/compose/library/collections", d.handleComposeUpsertCollection)
	mux.HandleFunc("/_api/v1/compose/library/collections/", d.handleComposeDeleteCollection)
	// move request between folders
	mux.HandleFunc("/_api/v1/compose/library/requests_move/", d.handleComposeMoveRequest)

	// Network throttling runtime API
	mux.HandleFunc("/_api/v1/throttle", d.handleV1Throttle)

	// MITM helpers (dev tooling)
	mux.HandleFunc("/_api/v1/mitm/status", d.handleV1MITMStatus)
	mux.HandleFunc("/_api/v1/mitm/ca", d.handleV1MITMGetCA)
	mux.HandleFunc("/_api/v1/mitm/ca/generate", d.handleV1MITMGenerate)
	// generate new dev CA, persist to files and swap runtime CA
	mux.HandleFunc("/_api/v1/mitm/ca/regenerate", d.handleV1MITMRegeneratePersist)

	// Intercept/Breakpoints API (MVP)
	mux.HandleFunc("/_api/v1/intercept/config", d.handleInterceptConfig)
	mux.HandleFunc("/_api/v1/intercept/rules", d.handleInterceptRules)
	mux.HandleFunc("/_api/v1/intercept/pending", d.handleInterceptPending)
	mux.HandleFunc("/_api/v1/intercept/items/", d.handleInterceptItem)

	// Mapping API (Map Local/Remote)
	mux.HandleFunc("/_api/v1/mapping/config", d.handleMappingConfig)
	mux.HandleFunc("/_api/v1/mapping/rules", d.handleMappingRules)
	mux.HandleFunc("/_api/v1/mapping/rules/reorder", d.handleMappingRules)
	mux.HandleFunc("/_api/v1/mapping/rules/", d.handleMappingRuleByID)
	mux.HandleFunc("/_api/v1/mapping/upload", d.handleMappingUpload)

	// Process Detection API
	mux.HandleFunc("/_api/v1/process/config", d.handleV1ProcessConfig)
	mux.HandleFunc("/_api/v1/process/helper/status", d.handleV1ProcessHelperStatus)
	mux.HandleFunc("/_api/v1/process/helper/install", d.handleV1ProcessHelperInstall)

	// Tags & Annotations API
	mux.HandleFunc("/_api/v1/tags/predefined", d.handlePredefinedTags)
	mux.HandleFunc("/_api/v1/tags/predefined/", d.handlePredefinedTagByID)
	mux.HandleFunc("/_api/v1/tags/bulk", d.handleBulkTags)

	// Session-specific tags & annotations
	mux.HandleFunc("/_api/v1/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.Contains(path, "/tags") {
			d.handleSessionTags(w, r)
		} else if strings.Contains(path, "/annotations") {
			d.handleSessionAnnotations(w, r)
		} else {
			d.handleV1SessionByID(w, r)
		}
	})

	// Performance Insights API
	mux.HandleFunc("/_api/v1/performance/overview", d.handlePerformanceOverview)

	// Scripting API
	if d.ScriptSvc != nil {
		scriptHandlers := NewScriptHandlers(d.ScriptSvc)
		// Set compilation service if available (for sourceCode compilation support)
		if d.CompilationSvc != nil {
			scriptHandlers.SetCompilationService(d.CompilationSvc)
		}
		mux.HandleFunc("POST /_api/v1/scripts", scriptHandlers.CreateScript)
		mux.HandleFunc("GET /_api/v1/scripts", scriptHandlers.ListScripts)
		mux.HandleFunc("GET /_api/v1/scripts/{id}", scriptHandlers.GetScript)
		mux.HandleFunc("PUT /_api/v1/scripts/{id}", scriptHandlers.UpdateScript)
		mux.HandleFunc("DELETE /_api/v1/scripts/{id}", scriptHandlers.DeleteScript)
		mux.HandleFunc("PATCH /_api/v1/scripts/{id}/toggle", scriptHandlers.ToggleScript)
		mux.HandleFunc("POST /_api/v1/scripts/test", scriptHandlers.TestScript)
		mux.HandleFunc("GET /_api/v1/scripts/examples", scriptHandlers.ListExamples)

		// Project Upload/Download API (multi-file support for IDE plugins)
		mux.HandleFunc("POST /_api/v1/scripts/{id}/upload-project", scriptHandlers.UploadProject)
		mux.HandleFunc("GET /_api/v1/scripts/{id}/download-project", scriptHandlers.DownloadProject)
		mux.HandleFunc("GET /_api/v1/scripts/{id}/files", scriptHandlers.ListProjectFiles)

		// Import/Export complete scripts as ZIP (with metadata + WASM)
		mux.HandleFunc("GET /_api/v1/scripts/{id}/export-zip", scriptHandlers.ExportScriptAsZip)
		mux.HandleFunc("POST /_api/v1/scripts/import-zip", scriptHandlers.ImportScriptFromZip)

		// Compilation API
		if d.CompilationSvc != nil {
			compilationHandlers := NewCompilationHandlers(d.CompilationSvc)
			mux.HandleFunc("POST /_api/v1/scripts/{id}/compile", compilationHandlers.CompileScript)
			mux.HandleFunc("POST /_api/v1/scripts/validate", compilationHandlers.ValidateSyntax)
			mux.HandleFunc("GET /_api/v1/scripts/compilers", compilationHandlers.ListCompilers)
		}

		// Compiler Installation API (download-on-demand for Zig, Kotlin, Swift)
		if d.CompilerInstallationSvc != nil {
			installHandlers := NewCompilerInstallationHandlers(d.CompilerInstallationSvc, *d.Logger)
			mux.HandleFunc("GET /_api/v1/compilers/status", installHandlers.GetCompilersStatus)
			mux.HandleFunc("GET /_api/v1/compilers/supported", installHandlers.GetSupportedLanguages)
			mux.HandleFunc("GET /_api/v1/compilers/{language}/status", installHandlers.GetCompilerStatus)
			mux.HandleFunc("POST /_api/v1/compilers/{language}/install", installHandlers.InstallCompiler)
			mux.HandleFunc("DELETE /_api/v1/compilers/{language}", installHandlers.UninstallCompiler)
			mux.HandleFunc("GET /_api/v1/compilers/{language}/progress", installHandlers.GetInstallationProgress)
			mux.HandleFunc("GET /_api/v1/compilers/cache/size", installHandlers.GetCacheSize)
			mux.HandleFunc("DELETE /_api/v1/compilers/cache", installHandlers.ClearCache)
		}
	}

	return mux
}

func withForwardProxy(d *Deps, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Intercept standard proxy patterns: CONNECT and absolute-URI
		if r.Method == http.MethodConnect || (r.URL != nil && r.URL.Scheme != "" && r.URL.Host != "") {
			d.handleForwardProxy(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func withCORS(cfg config.Config, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", cfg.CORSAllowOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Cookie, Sec-WebSocket-Protocol, X-Admin-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// findHelperBinary - найти helper binary используя абсолютные пути
func findHelperBinary(logger *zerolog.Logger) string {
	// Попробовать найти в стандартных местах
	candidates := []string{}

	// 1. Получить путь к текущему executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		// Helper должен быть в той же директории что и main binary
		candidates = append(candidates, filepath.Join(exeDir, "process-helper"))
		// Или в ../bin относительно executable
		candidates = append(candidates, filepath.Join(exeDir, "..", "bin", "process-helper"))
	}

	// 2. Относительно текущей рабочей директории (но как абсолютный путь)
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "bin", "process-helper"))
		candidates = append(candidates, filepath.Join(wd, "..", "bin", "process-helper"))
	}

	// 3. Проверить каждый candidate
	for _, path := range candidates {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			if logger != nil {
				logger.Info().Str("path", absPath).Msg("Found helper binary")
			}
			return absPath
		}
	}

	// Fallback - вернуть первый candidate как есть (установщик может скопировать binary позже)
	if len(candidates) > 0 {
		absPath, _ := filepath.Abs(candidates[0])
		if logger != nil {
			logger.Warn().Str("path", absPath).Msg("Helper binary not found, using fallback path")
		}
		return absPath
	}

	// Last resort
	return "bin/process-helper"
}
