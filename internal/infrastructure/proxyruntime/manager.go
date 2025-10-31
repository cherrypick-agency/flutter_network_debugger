package proxyruntime

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	socks5 "github.com/armon/go-socks5"
	"github.com/rs/zerolog"
)

// Manager управляет отдельными листенерами forward‑proxy и SOCKS5.
// Перезапуск листенеров происходит graceful через Shutdown/Close.
type Manager struct {
	log *zerolog.Logger

	mu sync.Mutex

	// HTTP forward‑proxy
	fwdSrv *http.Server
	fwdLn  net.Listener

	// SOCKS5
	socksSrv *socks5.Server
	socksLn  net.Listener
}

func New(log *zerolog.Logger) *Manager { return &Manager{log: log} }

// StartForward запускает HTTP сервер на addr с переданным handler.
func (m *Manager) StartForward(addr string, handler http.Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fwdLn != nil {
		// уже запущен — сначала останавливаем
		_ = m.stopForwardLocked(context.Background())
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.fwdLn = ln
	m.fwdSrv = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Захватим ссылку локально, чтобы избежать гонки при StopForward()
	srv := m.fwdSrv
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			if m.log != nil {
				m.log.Error().Err(err).Msg("forward proxy server error")
			}
		}
	}()
	if m.log != nil {
		m.log.Info().Str("addr", addr).Msg("forward proxy started")
	}
	return nil
}

func (m *Manager) StopForward(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopForwardLocked(ctx)
}

func (m *Manager) stopForwardLocked(ctx context.Context) error {
	if m.fwdSrv == nil {
		return nil
	}
	srv := m.fwdSrv
	ln := m.fwdLn
	m.fwdSrv = nil
	m.fwdLn = nil
	if ln != nil {
		_ = ln.Close()
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	if m.log != nil {
		m.log.Info().Msg("forward proxy stopped")
	}
	return nil
}

// StartSocks запускает SOCKS5 сервер на addr с опциональной аутентификацией.
// authMode: "none" | "userpass"; user/pass используются только для userpass.
func (m *Manager) StartSocks(addr, authMode, user, pass string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.socksLn != nil {
		_ = m.stopSocksLocked()
	}

	conf := &socks5.Config{}
	switch authMode {
	case "userpass":
		creds := socks5.StaticCredentials{user: pass}
		auth := socks5.UserPassAuthenticator{Credentials: creds}
		conf.AuthMethods = []socks5.Authenticator{auth}
	default:
		// no-auth
		conf.AuthMethods = []socks5.Authenticator{}
	}
	srv, err := socks5.New(conf)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.socksSrv = srv
	m.socksLn = ln
	go func() {
		if err := srv.Serve(ln); err != nil && m.log != nil {
			m.log.Error().Err(err).Msg("socks server error")
		}
	}()
	if m.log != nil {
		m.log.Info().Str("addr", addr).Msg("socks5 started")
	}
	return nil
}

func (m *Manager) StopSocks(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopSocksLocked()
}

func (m *Manager) stopSocksLocked() error {
	if m.socksLn == nil {
		return nil
	}
	_ = m.socksLn.Close()
	m.socksLn = nil
	m.socksSrv = nil
	if m.log != nil {
		m.log.Info().Msg("socks5 stopped")
	}
	return nil
}

// Apply включает/перезапускает листенеры согласно предоставленной конфигурации.
type ApplyConfig struct {
	ForwardEnabled bool
	ForwardAddr    string

	SocksEnabled  bool
	SocksAddr     string
	SocksAuthMode string
	SocksUser     string
	SocksPass     string
}

// ForwardAddr возвращает фактический адрес forward‑proxy (если запущен).
func (m *Manager) ForwardAddr() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fwdLn == nil {
		return ""
	}
	return m.fwdLn.Addr().String()
}

// SocksAddr возвращает фактический адрес SOCKS5 (если запущен).
func (m *Manager) SocksAddr() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.socksLn == nil {
		return ""
	}
	return m.socksLn.Addr().String()
}

// Apply принимает handler для forward‑proxy.
func (m *Manager) Apply(ctx context.Context, cfg ApplyConfig, forwardHandler http.Handler) error {
	// Forward
	if cfg.ForwardEnabled {
		if err := m.StartForward(cfg.ForwardAddr, forwardHandler); err != nil {
			return err
		}
	} else {
		_ = m.StopForward(ctx)
	}
	// SOCKS
	if cfg.SocksEnabled {
		if err := m.StartSocks(cfg.SocksAddr, cfg.SocksAuthMode, cfg.SocksUser, cfg.SocksPass); err != nil {
			return err
		}
	} else {
		_ = m.StopSocks(ctx)
	}
	return nil
}
