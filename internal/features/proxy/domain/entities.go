package domain

import "time"

// ProxyConfig - runtime settings for proxy ports/modes (singleton with ID=1).
type ProxyConfig struct {
	ID int64

	// HTTP forward-proxy on separate port
	ForwardEnabled bool
	ForwardAddr    string

	// SOCKS5 server
	SocksEnabled bool
	SocksAddr    string

	// SOCKS5 authentication (MVP): none | userpass
	SocksAuthMode string
	SocksUser     string
	SocksPass     string

	UpdatedAt time.Time
}
