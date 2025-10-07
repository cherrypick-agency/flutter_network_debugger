package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
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
	spa := spaHandler{root: sub, index: "index.html"}

	mux := http.NewServeMux()
	// Route API first
	mux.Handle("/_api/", apiRouter)
	mux.Handle("/api/", apiRouter)
	// Forward proxy/compat endpoints
	mux.Handle("/httpproxy", apiRouter)
	mux.Handle("/httpproxy/", apiRouter)
	mux.Handle("/_ws", apiRouter)
	mux.Handle("/_ws/", apiRouter)
	// Static last
	mux.Handle("/", spa)

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

type spaHandler struct {
	root  fs.FS
	index string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if upath == "" || upath == "/" {
		h.serveFile(w, h.index)
		return
	}
	// Trim leading '/'
	p := strings.TrimPrefix(path.Clean(upath), "/")
	// Try asset
	f, err := h.root.Open(p)
	if err == nil {
		_ = f.Close()
		http.FileServer(http.FS(h.root)).ServeHTTP(w, r)
		return
	}
	// Fallback to index for SPA routes
	h.serveFile(w, h.index)
}

func (h spaHandler) serveFile(w http.ResponseWriter, name string) {
	data, err := fs.ReadFile(h.root, name)
	if err != nil {
		http.NotFound(w, &http.Request{})
		return
	}
	// Minimal content-type for index.html
	if strings.HasSuffix(strings.ToLower(name), ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	_, _ = w.Write(data)
}
