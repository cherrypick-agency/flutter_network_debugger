package domain

import "time"

// ProxyConfig — рантайм‑настройки портов/режимов прокси (singleton с ID=1).
type ProxyConfig struct {
	ID int64

	// HTTP forward‑proxy на отдельном порту
	ForwardEnabled bool
	ForwardAddr    string

	// SOCKS5 сервер
	SocksEnabled bool
	SocksAddr    string

	// Аутентификация SOCKS5 (MVP): none | userpass
	SocksAuthMode string
	SocksUser     string
	SocksPass     string

	UpdatedAt time.Time
}
