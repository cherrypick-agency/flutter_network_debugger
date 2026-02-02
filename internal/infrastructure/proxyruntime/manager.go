package proxyruntime

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"syscall"
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
		if errors.Is(err, syscall.EADDRINUSE) {
			health := probeHTTPHealth(addrsForPortProbe(addr))
			if health.ok {
				return fmt.Errorf("forward listen %s: %w (port is already in use; existing service answers /healthz: %s)", addr, err, health.summary)
			}
			if health.summary != "" {
				return fmt.Errorf("forward listen %s: %w (port is already in use; existing service does NOT answer /healthz: %s)", addr, err, health.summary)
			}
			return fmt.Errorf("forward listen %s: %w (port is already in use; existing service does NOT answer /healthz)", addr, err)
		}
		return fmt.Errorf("forward listen %s: %w", addr, err)
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
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			if m.log != nil {
				m.log.Error().Err(err).Msg("forward proxy server error")
			}
			// Если Serve внезапно завершился, порт может остаться "слушать" без accept loop,
			// и клиенты будут успешно коннектиться по TCP, но никогда не получат HTTP-ответ.
			// Закрываем листенер и чистим состояние, чтобы проблема была заметна и лечилась рестартом.
			_ = ln.Close()
			m.mu.Lock()
			if m.fwdLn == ln {
				m.fwdLn = nil
			}
			if m.fwdSrv == srv {
				m.fwdSrv = nil
			}
			m.mu.Unlock()
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
		if errors.Is(err, syscall.EADDRINUSE) {
			health := probeHTTPHealth(addrsForPortProbe(addr))
			if health.ok {
				return fmt.Errorf("socks listen %s: %w (port is already in use; existing service answers /healthz: %s)", addr, err, health.summary)
			}
			if health.summary != "" {
				return fmt.Errorf("socks listen %s: %w (port is already in use; existing service does NOT answer /healthz: %s)", addr, err, health.summary)
			}
			return fmt.Errorf("socks listen %s: %w (port is already in use; existing service does NOT answer /healthz)", addr, err)
		}
		return fmt.Errorf("socks listen %s: %w", addr, err)
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

type healthProbeResult struct {
	ok      bool
	summary string
}

func addrsForPortProbe(listenAddr string) []string {
	// listenAddr can be ":9091", "127.0.0.1:9091", "[::]:9091"
	// For probes we try loopback v4/v6 with the same port.
	port := ""
	if strings.HasPrefix(listenAddr, ":") {
		port = strings.TrimPrefix(listenAddr, ":")
	} else {
		_, p, err := net.SplitHostPort(listenAddr)
		if err == nil {
			port = p
		}
	}
	if port == "" {
		return nil
	}
	return []string{
		net.JoinHostPort("127.0.0.1", port),
		net.JoinHostPort("::1", port),
	}
}

func probeHTTPHealth(addrs []string) healthProbeResult {
	const deadline = 300 * time.Millisecond
	const req = "GET /healthz HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
	if len(addrs) == 0 {
		return healthProbeResult{ok: false, summary: "no probe addr"}
	}
	for _, a := range addrs {
		c, err := net.DialTimeout("tcp", a, deadline)
		if err != nil {
			continue
		}
		_ = c.SetDeadline(time.Now().Add(deadline))
		_, _ = c.Write([]byte(req))
		br := bufio.NewReader(c)
		line, _ := br.ReadString('\n')
		_ = c.Close()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Example: "HTTP/1.1 200 OK"
		if strings.Contains(line, " 200 ") {
			return healthProbeResult{ok: true, summary: fmt.Sprintf("%s -> %s", a, line)}
		}
		return healthProbeResult{ok: false, summary: fmt.Sprintf("%s -> %s", a, line)}
	}
	return healthProbeResult{ok: false, summary: "dial ok, but no http status line (or timeout)"}
}
