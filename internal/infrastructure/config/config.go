package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Addr                   string
	LogLevel               string
	DevMode                bool
	InsecureTLS            bool
	CORSAllowOrigin        string
	DefaultTarget          string
	PreviewMaxBytes        int
	SSEPollIntervalMs      int
	ExposeSensitiveHeaders bool
	// Optional TLS server for REST/reverse-proxy (HTTP/2 enabled by default in net/http)
	TLSAddr     string
	TLSCertFile string
	TLSKeyFile  string
	// HTTP body capture (reverse proxy)
	CaptureBodies     bool
	BodyMaxBytes      int
	BodySpoolDir      string
	PreviewDecompress bool
	// Artificial response delay for proxy responses (ms)
	ResponseDelayMs int
	// Optional range support: if set, each response delay will be random in [min,max]
	ResponseDelayMinMs int
	ResponseDelayMaxMs int

	// Forward proxy MITM (HTTPS inspection)
	// If enabled and CA is provided, CONNECT requests will be intercepted and
	// decrypted using dynamically issued certificates for requested hosts.
	MITMEnabled    bool
	MITMCACertFile string
	MITMCAKeyFile  string
	// Comma-separated domain suffix allow/deny lists (e.g. ".example.com,api.test")
	MITMDomainsAllow []string
	MITMDomainsDeny  []string
}

func FromEnv() Config {
	cfg := Config{
		Addr:            getEnv("ADDR", ":9091"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
	}
	if os.Getenv("DEV_MODE") == "1" || os.Getenv("DEV_MODE") == "true" {
		cfg.DevMode = true
	}
	cfg.DefaultTarget = getEnv("DEFAULT_TARGET", "")
	// Align default preview size with e2e expectations: text previews must be truncated to <=4096 bytes
	cfg.PreviewMaxBytes = getEnvInt("PREVIEW_MAX_BYTES", 40096)
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
	// default: expose sensitive headers unless explicitly disabled
	if os.Getenv("EXPOSE_SENSITIVE_HEADERS") == "0" || os.Getenv("EXPOSE_SENSITIVE_HEADERS") == "false" {
		cfg.ExposeSensitiveHeaders = false
	} else {
		cfg.ExposeSensitiveHeaders = true
	}

	// MITM settings
	if os.Getenv("MITM_ENABLE") == "1" || os.Getenv("MITM_ENABLE") == "true" {
		cfg.MITMEnabled = true
	}
	cfg.MITMCACertFile = getEnv("MITM_CA_CERT_FILE", "")
	cfg.MITMCAKeyFile = getEnv("MITM_CA_KEY_FILE", "")
	if v := strings.TrimSpace(os.Getenv("MITM_DOMAINS_ALLOW")); v != "" {
		cfg.MITMDomainsAllow = splitCSV(v)
	}
	if v := strings.TrimSpace(os.Getenv("MITM_DOMAINS_DENY")); v != "" {
		cfg.MITMDomainsDeny = splitCSV(v)
	}
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
