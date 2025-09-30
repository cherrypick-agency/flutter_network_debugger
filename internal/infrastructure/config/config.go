package config

import (
    "os"
    "strconv"
)

type Config struct {
    Addr           string
    LogLevel       string
    InsecureTLS    bool
    CORSAllowOrigin string
    DefaultTarget  string
    PreviewMaxBytes int
    SSEPollIntervalMs int
    ExposeSensitiveHeaders bool
    // Optional TLS server for REST/reverse-proxy (HTTP/2 enabled by default in net/http)
    TLSAddr       string
    TLSCertFile   string
    TLSKeyFile    string
    // HTTP body capture (reverse proxy)
    CaptureBodies bool
    BodyMaxBytes  int
    BodySpoolDir  string
    PreviewDecompress bool
    // Artificial response delay for proxy responses (ms)
    ResponseDelayMs int
}

func FromEnv() Config {
    cfg := Config{
        Addr:           getEnv("ADDR", ":9091"),
        LogLevel:       getEnv("LOG_LEVEL", "info"),
        CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
    }
    cfg.DefaultTarget = getEnv("DEFAULT_TARGET", "")
    // Align default preview size with e2e expectations: text previews must be truncated to <=4096 bytes
    cfg.PreviewMaxBytes = getEnvInt("PREVIEW_MAX_BYTES", 4096)
    cfg.SSEPollIntervalMs = getEnvInt("SSE_POLL_INTERVAL_MS", 500)
    // TLS settings (optional). If cert+key provided, a TLS server will start on TLS_ADDR (default :9443)
    cfg.TLSAddr = getEnv("TLS_ADDR", "")
    cfg.TLSCertFile = getEnv("TLS_CERT_FILE", "")
    cfg.TLSKeyFile = getEnv("TLS_KEY_FILE", "")
    // Body capture
    if os.Getenv("CAPTURE_BODIES") == "1" || os.Getenv("CAPTURE_BODIES") == "true" { cfg.CaptureBodies = true }
    cfg.BodyMaxBytes = getEnvInt("BODY_MAX_BYTES", 8<<20) // 8MB
    cfg.BodySpoolDir = getEnv("BODY_SPOOL_DIR", "")
    if os.Getenv("PREVIEW_DECOMPRESS") == "0" || os.Getenv("PREVIEW_DECOMPRESS") == "false" {
        cfg.PreviewDecompress = false
    } else {
        cfg.PreviewDecompress = true
    }
    cfg.ResponseDelayMs = getEnvInt("RESPONSE_DELAY_MS", 0)
    if os.Getenv("INSECURE_TLS") == "1" || os.Getenv("INSECURE_TLS") == "true" {
        cfg.InsecureTLS = true
    }
    // default: expose sensitive headers unless explicitly disabled
    if os.Getenv("EXPOSE_SENSITIVE_HEADERS") == "0" || os.Getenv("EXPOSE_SENSITIVE_HEADERS") == "false" {
        cfg.ExposeSensitiveHeaders = false
    } else {
        cfg.ExposeSensitiveHeaders = true
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
        if n, err := strconv.Atoi(v); err == nil { return n }
    }
    return def
}


