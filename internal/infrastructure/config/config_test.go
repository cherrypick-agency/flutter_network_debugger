package config

import (
	"os"
	"testing"
)

// Arrange-Act-Assert style tests for FromEnv

func TestFromEnv_Defaults(t *testing.T) {
	t.Parallel()
	// Arrange
	clearEnv(t)

	// Act
	cfg := FromEnv()

	// Assert
	if cfg.Addr != ":9091" {
		t.Fatalf("Addr default: %q", cfg.Addr)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel default: %q", cfg.LogLevel)
	}
	if cfg.CORSAllowOrigin != "*" {
		t.Fatalf("CORSAllowOrigin: %q", cfg.CORSAllowOrigin)
	}
	if cfg.PreviewMaxBytes != 50000 {
		t.Fatalf("PreviewMaxBytes: %d", cfg.PreviewMaxBytes)
	}
	if cfg.SSEPollIntervalMs != 777 {
		t.Fatalf("SSEPollIntervalMs: %d", cfg.SSEPollIntervalMs)
	}
	if cfg.TLSAddr != "" || cfg.TLSCertFile != "" || cfg.TLSKeyFile != "" {
		t.Fatalf("TLS defaults not empty")
	}
	if cfg.CaptureBodies {
		t.Fatalf("CaptureBodies should be false by default")
	}
	if cfg.BodyMaxBytes != 8<<20 {
		t.Fatalf("BodyMaxBytes: %d", cfg.BodyMaxBytes)
	}
	if !cfg.PreviewDecompress {
		t.Fatalf("PreviewDecompress default true")
	}
	if cfg.WSPreviewMaxBytes != cfg.PreviewMaxBytes {
		t.Fatalf("WSPreviewMaxBytes default: %d", cfg.WSPreviewMaxBytes)
	}
	if cfg.WSCaptureBodies {
		t.Fatalf("WSCaptureBodies default false")
	}
	if cfg.WSBodyMaxBytes != 1<<20 {
		t.Fatalf("WSBodyMaxBytes: %d", cfg.WSBodyMaxBytes)
	}
	if !cfg.WSDeflatePreview {
		t.Fatalf("WSDeflatePreview default true")
	}
	if cfg.ResponseDelayMs != 0 {
		t.Fatalf("ResponseDelayMs: %d", cfg.ResponseDelayMs)
	}
	if cfg.InsecureTLS {
		t.Fatalf("InsecureTLS default false")
	}
	if cfg.NoBrowser {
		t.Fatalf("NoBrowser default false")
	}
	if !cfg.ExposeSensitiveHeaders {
		t.Fatalf("ExposeSensitiveHeaders default true")
	}
	if cfg.MITMEnabled {
		t.Fatalf("MITMEnabled default false")
	}
	if len(cfg.MITMDomainsAllow) != 0 || len(cfg.MITMDomainsDeny) != 0 {
		t.Fatalf("MITM domains should be empty")
	}
}

func TestFromEnv_BooleansAndInts(t *testing.T) {
	// cannot use t.Parallel with t.Setenv
	// Arrange
	clearEnv(t)
	t.Setenv("DEV_MODE", "true")
	t.Setenv("CAPTURE_BODIES", "1")
	t.Setenv("PREVIEW_MAX_BYTES", "60000")
	t.Setenv("WS_PREVIEW_MAX_BYTES", "0") // triggers fallback to 4096
	t.Setenv("WS_DEFLATE_PREVIEW", "false")
	t.Setenv("WS_CAPTURE_BODIES", "true")
	t.Setenv("WS_BODY_MAX_BYTES", "524288")
	t.Setenv("RESPONSE_DELAY_MS", "1500")
	t.Setenv("INSECURE_TLS", "1")
	t.Setenv("NO_BROWSER", "true")
	t.Setenv("EXPOSE_SENSITIVE_HEADERS", "0")

	// Act
	cfg := FromEnv()

	// Assert
	if !cfg.DevMode {
		t.Fatalf("DevMode")
	}
	if !cfg.CaptureBodies {
		t.Fatalf("CaptureBodies")
	}
	if cfg.PreviewMaxBytes != 60000 {
		t.Fatalf("PreviewMaxBytes: %d", cfg.PreviewMaxBytes)
	}
	if cfg.WSPreviewMaxBytes != 4096 {
		t.Fatalf("WSPreviewMaxBytes fallback: %d", cfg.WSPreviewMaxBytes)
	}
	if cfg.WSCaptureBodies != true {
		t.Fatalf("WSCaptureBodies")
	}
	if cfg.WSBodyMaxBytes != 524288 {
		t.Fatalf("WSBodyMaxBytes: %d", cfg.WSBodyMaxBytes)
	}
	if cfg.ResponseDelayMs != 1500 {
		t.Fatalf("ResponseDelayMs: %d", cfg.ResponseDelayMs)
	}
	if !cfg.InsecureTLS {
		t.Fatalf("InsecureTLS")
	}
	if !cfg.NoBrowser {
		t.Fatalf("NoBrowser")
	}
	if cfg.ExposeSensitiveHeaders {
		t.Fatalf("ExposeSensitiveHeaders should be false")
	}
}

func TestFromEnv_ResponseDelayRange(t *testing.T) {
	// cannot use t.Parallel with t.Setenv
	// Arrange
	clearEnv(t)
	t.Setenv("RESPONSE_DELAY_MS", "100-300")

	// Act
	cfg := FromEnv()

	// Assert
	if cfg.ResponseDelayMinMs != 100 || cfg.ResponseDelayMaxMs != 300 {
		t.Fatalf("range: %d-%d", cfg.ResponseDelayMinMs, cfg.ResponseDelayMaxMs)
	}
}

func TestFromEnv_MITMDomainLists(t *testing.T) {
	// cannot use t.Parallel with t.Setenv
	// Arrange
	clearEnv(t)
	t.Setenv("MITM_ENABLE", "1")
	t.Setenv("MITM_DOMAINS_ALLOW", " .example.com,api.test ,, ")
	t.Setenv("MITM_DOMAINS_DENY", "bad.local,  ")

	// Act
	cfg := FromEnv()

	// Assert
	if !cfg.MITMEnabled {
		t.Fatalf("MITMEnabled")
	}
	if len(cfg.MITMDomainsAllow) != 2 || cfg.MITMDomainsAllow[0] != ".example.com" || cfg.MITMDomainsAllow[1] != "api.test" {
		t.Fatalf("allow: %+v", cfg.MITMDomainsAllow)
	}
	if len(cfg.MITMDomainsDeny) != 1 || cfg.MITMDomainsDeny[0] != "bad.local" {
		t.Fatalf("deny: %+v", cfg.MITMDomainsDeny)
	}
}

func TestGetEnv_WithValue(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")

	result := getEnv("TEST_VAR", "default")

	if result != "test_value" {
		t.Errorf("getEnv() = %s, want test_value", result)
	}
}

func TestGetEnv_WithDefault(t *testing.T) {
	// Make sure env var doesn't exist
	os.Unsetenv("NON_EXISTENT_VAR")

	result := getEnv("NON_EXISTENT_VAR", "default_value")

	if result != "default_value" {
		t.Errorf("getEnv() = %s, want default_value", result)
	}
}

func TestGetEnv_EmptyString(t *testing.T) {
	t.Setenv("EMPTY_VAR", "")

	result := getEnv("EMPTY_VAR", "default")

	// Empty string should return default
	if result != "default" {
		t.Errorf("getEnv() with empty string = %s, want default", result)
	}
}

func TestGetEnvInt_WithValidInt(t *testing.T) {
	t.Setenv("TEST_INT", "42")

	result := getEnvInt("TEST_INT", 0)

	if result != 42 {
		t.Errorf("getEnvInt() = %d, want 42", result)
	}
}

func TestGetEnvInt_WithInvalidInt(t *testing.T) {
	t.Setenv("INVALID_INT", "not_a_number")

	result := getEnvInt("INVALID_INT", 99)

	// Invalid int should return default
	if result != 99 {
		t.Errorf("getEnvInt() with invalid int = %d, want 99", result)
	}
}

func TestGetEnvInt_WithDefault(t *testing.T) {
	os.Unsetenv("MISSING_INT")

	result := getEnvInt("MISSING_INT", 123)

	if result != 123 {
		t.Errorf("getEnvInt() = %d, want 123", result)
	}
}

func TestGetEnvInt_EmptyString(t *testing.T) {
	t.Setenv("EMPTY_INT", "")

	result := getEnvInt("EMPTY_INT", 77)

	// Empty string should return default
	if result != 77 {
		t.Errorf("getEnvInt() with empty string = %d, want 77", result)
	}
}

func TestSplitCSV_Simple(t *testing.T) {
	result := splitCSV("a,b,c")

	if len(result) != 3 {
		t.Errorf("len = %d, want 3", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("result = %v", result)
	}
}

func TestSplitCSV_WithSpaces(t *testing.T) {
	result := splitCSV(" a , b , c ")

	if len(result) != 3 {
		t.Errorf("len = %d, want 3", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("result = %v", result)
	}
}

func TestSplitCSV_EmptyStrings(t *testing.T) {
	result := splitCSV("a,,b,  ,c")

	if len(result) != 3 {
		t.Errorf("len = %d, want 3 (empty strings should be skipped)", len(result))
	}
	if result[0] != "a" || result[1] != "b" || result[2] != "c" {
		t.Errorf("result = %v", result)
	}
}

func TestSplitCSV_EmptyInput(t *testing.T) {
	result := splitCSV("")

	if len(result) != 0 {
		t.Errorf("len = %d, want 0", len(result))
	}
}

func TestSplitCSV_OnlySpaces(t *testing.T) {
	result := splitCSV("  ,  ,  ")

	if len(result) != 0 {
		t.Errorf("len = %d, want 0 (only spaces should result in empty list)", len(result))
	}
}

func TestSplitCSV_Single(t *testing.T) {
	result := splitCSV("single")

	if len(result) != 1 {
		t.Errorf("len = %d, want 1", len(result))
	}
	if result[0] != "single" {
		t.Errorf("result[0] = %s, want single", result[0])
	}
}

// helper to clear relevant env vars
func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"ADDR", "LOG_LEVEL", "DEV_MODE", "CORS_ALLOW_ORIGIN",
		"DEFAULT_TARGET", "PREVIEW_MAX_BYTES", "SSE_POLL_INTERVAL_MS",
		"TLS_ADDR", "TLS_CERT_FILE", "TLS_KEY_FILE",
		"CAPTURE_BODIES", "BODY_MAX_BYTES", "BODY_SPOOL_DIR", "PREVIEW_DECOMPRESS",
		"WS_PREVIEW_MAX_BYTES", "WS_CAPTURE_BODIES", "WS_BODY_MAX_BYTES", "WS_DEFLATE_PREVIEW",
		"RESPONSE_DELAY_MS", "INSECURE_TLS", "NO_BROWSER", "EXPOSE_SENSITIVE_HEADERS",
		"MITM_ENABLE", "MITM_CA_CERT_FILE", "MITM_CA_KEY_FILE", "MITM_DOMAINS_ALLOW", "MITM_DOMAINS_DENY",
	}
	for _, k := range keys {
		_ = os.Unsetenv(k)
	}
}
