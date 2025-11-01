package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Addr            string
	LogLevel        string
	DevMode         bool
	InsecureTLS     bool
	CORSAllowOrigin string
	DefaultTarget   string
	PreviewMaxBytes int
	// Disable opening browser automatically in binaries
	NoBrowser              bool
	SSEPollIntervalMs      int
	ExposeSensitiveHeaders bool
	// Reverse-proxy cookie handling and stealth
	Cookies        CookiesConfig
	StealthHeaders bool
	// Optional TLS server for REST/reverse-proxy (HTTP/2 enabled by default in net/http)
	TLSAddr     string
	TLSCertFile string
	TLSKeyFile  string
	// HTTP body capture (reverse proxy)
	CaptureBodies     bool
	BodyMaxBytes      int
	BodySpoolDir      string
	PreviewDecompress bool
	// WS preview/capture settings (forward proxy)
	WSPreviewMaxBytes int
	WSCaptureBodies   bool
	WSBodyMaxBytes    int
	WSDeflatePreview  bool
	// Artificial response delay for proxy responses (ms)
	ResponseDelayMs int
	// Optional range support: if set, each response delay will be random in [min,max]
	ResponseDelayMinMs int
	ResponseDelayMaxMs int

	// Network throttling (bandwidth/packet loss/offline)
	// If Enabled, proxy will limit transfer speed and optionally drop data to
	// simulate unreliable networks. Units are kilobits per second (kbit/s).
	ThrottleEnabled    bool
	ThrottleDownKbps   int // downstream (upstream->client)
	ThrottleUpKbps     int // upstream (client->upstream)
	ThrottlePacketLoss int // 0..100 percent of reads/writes dropped (best-effort)
	ThrottleOffline    bool

	// Latency injection (RTT/ping simulation)
	// Adds artificial delay to network operations to simulate high-latency networks.
	// Unlike response delay (which only delays the response), latency injection affects:
	// - Connection establishment (TCP/TLS handshakes)
	// - Request sending
	// - Response receiving
	// This simulates geographic distance, satellite links, or congested networks.
	ThrottleLatencyMs     int // Base latency in milliseconds (e.g., 100 = 100ms RTT)
	ThrottleLatencyJitter int // Random jitter ± ms (e.g., 20 = varies 80-120ms)

	// Forward proxy MITM (HTTPS inspection)
	// If enabled and CA is provided, CONNECT requests will be intercepted and
	// decrypted using dynamically issued certificates for requested hosts.
	MITMEnabled    bool
	MITMCACertFile string
	MITMCAKeyFile  string
	// Comma-separated domain suffix allow/deny lists (e.g. ".example.com,api.test")
	MITMDomainsAllow []string
	MITMDomainsDeny  []string

	// Interception / Breakpoints (MVP)
	InterceptEnabled      bool
	InterceptRequests     bool
	InterceptResponses    bool
	InterceptTimeoutMs    int
	InterceptQueueMax     int
	InterceptBodyMaxBytes int
	InterceptReencode     bool
	InterceptOverflow     string
	AdminToken            string
	InterceptMethods      []string
	InterceptURLContains  []string
	InterceptContentTypes []string

	// Compose/Request Builder limits
	ComposeMaxUploadMB int
}

// CookiesConfig controls reverse-proxy cookie rewriting behavior
type CookiesConfig struct {
	// Mode: "isolate" (default), "auto", or "off"
	Mode string
	// DomainStrategy: "hostOnly" (default) or "proxyHost"
	DomainStrategy string
	// PathStrategy: "prefix" (default) or "root"
	PathStrategy string
}

func FromEnv() Config {
	cfg := Config{
		Addr:            getEnv("ADDR", ":9092"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
	}
	if os.Getenv("DEV_MODE") == "1" || os.Getenv("DEV_MODE") == "true" {
		cfg.DevMode = true
	}
	cfg.DefaultTarget = getEnv("DEFAULT_TARGET", "")
	// Default preview size; can be overridden via PREVIEW_MAX_BYTES
	cfg.PreviewMaxBytes = getEnvInt("PREVIEW_MAX_BYTES", 50000)
	cfg.SSEPollIntervalMs = getEnvInt("SSE_POLL_INTERVAL_MS", 777)
	// TLS settings (optional). If cert+key provided, a TLS server will start on TLS_ADDR (default :9443)
	cfg.TLSAddr = getEnv("TLS_ADDR", "")
	cfg.TLSCertFile = getEnv("TLS_CERT_FILE", "")
	cfg.TLSKeyFile = getEnv("TLS_KEY_FILE", "")
	// Body capture
	if os.Getenv("CAPTURE_BODIES") == "1" || os.Getenv("CAPTURE_BODIES") == "true" {
		cfg.CaptureBodies = true
	}
	cfg.BodyMaxBytes = getEnvInt("BODY_MAX_BYTES", 8<<20) // 8MB
	cfg.BodySpoolDir = getEnv("BODY_SPOOL_DIR", "")
	if os.Getenv("PREVIEW_DECOMPRESS") == "0" || os.Getenv("PREVIEW_DECOMPRESS") == "false" {
		cfg.PreviewDecompress = false
	} else {
		cfg.PreviewDecompress = true
	}
	// WS preview defaults align with HTTP preview unless overridden
	cfg.WSPreviewMaxBytes = getEnvInt("WS_PREVIEW_MAX_BYTES", cfg.PreviewMaxBytes)
	if cfg.WSPreviewMaxBytes <= 0 {
		cfg.WSPreviewMaxBytes = 4096
	}
	if os.Getenv("WS_CAPTURE_BODIES") == "1" || os.Getenv("WS_CAPTURE_BODIES") == "true" {
		cfg.WSCaptureBodies = true
	}
	cfg.WSBodyMaxBytes = getEnvInt("WS_BODY_MAX_BYTES", 1<<20) // 1MB
	// enable deflate preview by default
	if os.Getenv("WS_DEFLATE_PREVIEW") == "0" || os.Getenv("WS_DEFLATE_PREVIEW") == "false" {
		cfg.WSDeflatePreview = false
	} else {
		cfg.WSDeflatePreview = true
	}
	cfg.ResponseDelayMs = getEnvInt("RESPONSE_DELAY_MS", 0)
	if raw := os.Getenv("RESPONSE_DELAY_MS"); raw != "" && strings.Contains(raw, "-") {
		parts := strings.SplitN(raw, "-", 2)
		if len(parts) == 2 {
			minStr := strings.TrimSpace(parts[0])
			maxStr := strings.TrimSpace(parts[1])
			if min, err1 := strconv.Atoi(minStr); err1 == nil {
				if max, err2 := strconv.Atoi(maxStr); err2 == nil {
					if max < min {
						min, max = max, min
					}
					cfg.ResponseDelayMinMs = min
					cfg.ResponseDelayMaxMs = max
				}
			}
		}
	}
	if os.Getenv("INSECURE_TLS") == "1" || os.Getenv("INSECURE_TLS") == "true" {
		cfg.InsecureTLS = true
	}
	// Browser auto-open control
	if os.Getenv("NO_BROWSER") == "1" || os.Getenv("NO_BROWSER") == "true" {
		cfg.NoBrowser = true
	}
	// default: expose sensitive headers unless explicitly disabled
	if os.Getenv("EXPOSE_SENSITIVE_HEADERS") == "0" || os.Getenv("EXPOSE_SENSITIVE_HEADERS") == "false" {
		cfg.ExposeSensitiveHeaders = false
	} else {
		cfg.ExposeSensitiveHeaders = true
	}

	// Reverse-proxy cookies + stealth defaults and env overrides
	cfg.Cookies = CookiesConfig{
		Mode:           getEnv("COOKIES_MODE", "isolate"),
		DomainStrategy: getEnv("COOKIES_DOMAIN_STRATEGY", "hostOnly"),
		PathStrategy:   getEnv("COOKIES_PATH_STRATEGY", "prefix"),
	}
	if os.Getenv("STEALTH_HEADERS") == "0" || os.Getenv("STEALTH_HEADERS") == "false" {
		cfg.StealthHeaders = false
	} else {
		// по умолчанию скрываем прокси-заголовки в /httpproxy
		cfg.StealthHeaders = true
	}

	// MITM settings
	if os.Getenv("MITM_ENABLE") == "1" || os.Getenv("MITM_ENABLE") == "true" {
		cfg.MITMEnabled = true
	}

	// Throttling settings (runtime can override via API)
	if os.Getenv("THROTTLE_ENABLE") == "1" || os.Getenv("THROTTLE_ENABLE") == "true" {
		cfg.ThrottleEnabled = true
	}
	cfg.ThrottleDownKbps = getEnvInt("THROTTLE_DOWN_KBPS", 0)
	cfg.ThrottleUpKbps = getEnvInt("THROTTLE_UP_KBPS", 0)
	cfg.ThrottlePacketLoss = getEnvInt("THROTTLE_PACKET_LOSS", 0)
	cfg.ThrottleLatencyMs = getEnvInt("THROTTLE_LATENCY_MS", 0)
	cfg.ThrottleLatencyJitter = getEnvInt("THROTTLE_LATENCY_JITTER", 0)
	if os.Getenv("THROTTLE_OFFLINE") == "1" || os.Getenv("THROTTLE_OFFLINE") == "true" {
		cfg.ThrottleOffline = true
	}
	cfg.MITMCACertFile = getEnv("MITM_CA_CERT_FILE", "")
	cfg.MITMCAKeyFile = getEnv("MITM_CA_KEY_FILE", "")
	if v := strings.TrimSpace(os.Getenv("MITM_DOMAINS_ALLOW")); v != "" {
		cfg.MITMDomainsAllow = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv("MITM_DOMAINS_DENY")); v != "" {
		cfg.MITMDomainsDeny = splitCSV(v)
	}

	// Interception (breakpoints)
	if os.Getenv("INTERCEPT_ENABLE") == "1" || os.Getenv("INTERCEPT_ENABLE") == "true" {
		cfg.InterceptEnabled = true
	}
	if os.Getenv("INTERCEPT_REQUESTS") == "0" || os.Getenv("INTERCEPT_REQUESTS") == "false" {
		cfg.InterceptRequests = false
	} else {
		cfg.InterceptRequests = true
	}
	if os.Getenv("INTERCEPT_RESPONSES") == "0" || os.Getenv("INTERCEPT_RESPONSES") == "false" {
		cfg.InterceptResponses = false
	} else {
		cfg.InterceptResponses = true
	}
	cfg.InterceptTimeoutMs = getEnvInt("INTERCEPT_TIMEOUT_MS", 60000)
	cfg.InterceptQueueMax = getEnvInt("INTERCEPT_QUEUE_MAX", 200)
	cfg.InterceptBodyMaxBytes = getEnvInt("INTERCEPT_BODY_MAX_BYTES", 1<<20)
	if os.Getenv("INTERCEPT_REENCODE") == "0" || os.Getenv("INTERCEPT_REENCODE") == "false" {
		cfg.InterceptReencode = false
	} else {
		cfg.InterceptReencode = true
	}
	cfg.InterceptOverflow = getEnv("INTERCEPT_OVERFLOW", "auto-continue-oldest")
	cfg.AdminToken = getEnv("ADMIN_TOKEN", "")
	if v := strings.TrimSpace(os.Getenv("INTERCEPT_METHODS")); v != "" {
		cfg.InterceptMethods = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv("INTERCEPT_URL_CONTAINS")); v != "" {
		cfg.InterceptURLContains = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv("INTERCEPT_CONTENT_TYPES")); v != "" {
		cfg.InterceptContentTypes = splitCSV(v)
	}

	// Compose defaults
	cfg.ComposeMaxUploadMB = getEnvInt("COMPOSE_MAX_UPLOAD_MB", 10)
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// splitCSV splits comma-separated tokens trimming whitespace and skipping empties.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// filepathJoinSafe joins path elements with OS-specific separator.
func filepathJoinSafe(elem ...string) string {
	return filepath.Join(elem...)
}
