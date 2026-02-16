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
	if cfg.Addr != ":9092" {
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
	if !cfg.CaptureBodies {
		t.Fatalf("CaptureBodies should be true by default")
	}
	if cfg.BodyMaxBytes != 32<<20 {
		t.Fatalf("BodyMaxBytes: %d", cfg.BodyMaxBytes)
	}
	if !cfg.PreviewDecompress {
		t.Fatalf("PreviewDecompress default true")
	}
	if cfg.WSPreviewMaxBytes != cfg.PreviewMaxBytes {
		t.Fatalf("WSPreviewMaxBytes default: %d", cfg.WSPreviewMaxBytes)
	}
	if !cfg.WSCaptureBodies {
		t.Fatalf("WSCaptureBodies should be true by default")
	}
	if cfg.WSBodyMaxBytes != 4<<20 {
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

func TestFromEnv_IngestAllowRemote(t *testing.T) {
	clearEnv(t)
	t.Setenv("INGEST_ALLOW_REMOTE", "true")

	cfg := FromEnv()

	if !cfg.IngestAllowRemote {
		t.Fatalf("IngestAllowRemote should be true when INGEST_ALLOW_REMOTE=true")
	}
}

func TestFromEnv_DEV_MODE_One(t *testing.T) {
	clearEnv(t)
	t.Setenv("DEV_MODE", "1")

	cfg := FromEnv()

	if !cfg.DevMode {
		t.Errorf("DevMode should be true when DEV_MODE=1")
	}
}

func TestFromEnv_PREVIEW_DECOMPRESS_Zero(t *testing.T) {
	clearEnv(t)
	t.Setenv("PREVIEW_DECOMPRESS", "0")

	cfg := FromEnv()

	if cfg.PreviewDecompress {
		t.Errorf("PreviewDecompress should be false when PREVIEW_DECOMPRESS=0")
	}
}

func TestFromEnv_WS_DEFLATE_PREVIEW_Zero(t *testing.T) {
	clearEnv(t)
	t.Setenv("WS_DEFLATE_PREVIEW", "0")

	cfg := FromEnv()

	if cfg.WSDeflatePreview {
		t.Errorf("WSDeflatePreview should be false when WS_DEFLATE_PREVIEW=0")
	}
}

func TestFromEnv_ResponseDelayRange_Reversed(t *testing.T) {
	clearEnv(t)
	t.Setenv("RESPONSE_DELAY_MS", "300-100")

	cfg := FromEnv()

	if cfg.ResponseDelayMinMs != 100 || cfg.ResponseDelayMaxMs != 300 {
		t.Errorf("ResponseDelay range should be reversed: got %d-%d, want 100-300", cfg.ResponseDelayMinMs, cfg.ResponseDelayMaxMs)
	}
}

func TestFromEnv_ResponseDelayRange_InvalidMin(t *testing.T) {
	clearEnv(t)
	t.Setenv("RESPONSE_DELAY_MS", "invalid-100")

	cfg := FromEnv()

	if cfg.ResponseDelayMinMs != 0 || cfg.ResponseDelayMaxMs != 0 {
		t.Errorf("ResponseDelay range should be 0-0 when min is invalid: got %d-%d", cfg.ResponseDelayMinMs, cfg.ResponseDelayMaxMs)
	}
}

func TestFromEnv_ResponseDelayRange_InvalidMax(t *testing.T) {
	clearEnv(t)
	t.Setenv("RESPONSE_DELAY_MS", "100-invalid")

	cfg := FromEnv()

	if cfg.ResponseDelayMinMs != 0 || cfg.ResponseDelayMaxMs != 0 {
		t.Errorf("ResponseDelay range should be 0-0 when max is invalid: got %d-%d", cfg.ResponseDelayMinMs, cfg.ResponseDelayMaxMs)
	}
}

func TestFromEnv_INTERCEPT_REQUESTS_Zero(t *testing.T) {
	clearEnv(t)
	t.Setenv("INTERCEPT_REQUESTS", "0")

	cfg := FromEnv()

	if cfg.InterceptRequests {
		t.Errorf("InterceptRequests should be false when INTERCEPT_REQUESTS=0")
	}
}

func TestFromEnv_INTERCEPT_RESPONSES_Zero(t *testing.T) {
	clearEnv(t)
	t.Setenv("INTERCEPT_RESPONSES", "0")

	cfg := FromEnv()

	if cfg.InterceptResponses {
		t.Errorf("InterceptResponses should be false when INTERCEPT_RESPONSES=0")
	}
}

func TestFromEnv_INTERCEPT_REENCODE_Zero(t *testing.T) {
	clearEnv(t)
	t.Setenv("INTERCEPT_REENCODE", "0")

	cfg := FromEnv()

	if cfg.InterceptReencode {
		t.Errorf("InterceptReencode should be false when INTERCEPT_REENCODE=0")
	}
}

func TestFromEnv_STEALTH_HEADERS_Zero(t *testing.T) {
	clearEnv(t)
	t.Setenv("STEALTH_HEADERS", "0")

	cfg := FromEnv()

	if cfg.StealthHeaders {
		t.Errorf("StealthHeaders should be false when STEALTH_HEADERS=0")
	}
}

func TestFromEnv_THROTTLE_ENABLE_True(t *testing.T) {
	clearEnv(t)
	t.Setenv("THROTTLE_ENABLE", "true")

	cfg := FromEnv()

	if !cfg.ThrottleEnabled {
		t.Errorf("ThrottleEnabled should be true when THROTTLE_ENABLE=true")
	}
}

func TestFromEnv_THROTTLE_OFFLINE_True(t *testing.T) {
	clearEnv(t)
	t.Setenv("THROTTLE_OFFLINE", "true")

	cfg := FromEnv()

	if !cfg.ThrottleOffline {
		t.Errorf("ThrottleOffline should be true when THROTTLE_OFFLINE=true")
	}
}

func TestFromEnv_INTERCEPT_ENABLE_True(t *testing.T) {
	clearEnv(t)
	t.Setenv("INTERCEPT_ENABLE", "true")

	cfg := FromEnv()

	if !cfg.InterceptEnabled {
		t.Errorf("InterceptEnabled should be true when INTERCEPT_ENABLE=true")
	}
}

func TestFromEnv_InterceptLists(t *testing.T) {
	clearEnv(t)
	t.Setenv("INTERCEPT_METHODS", "GET,POST")
	t.Setenv("INTERCEPT_URL_CONTAINS", "api,test")
	t.Setenv("INTERCEPT_CONTENT_TYPES", "application/json,text/html")

	cfg := FromEnv()

	if len(cfg.InterceptMethods) != 2 {
		t.Errorf("InterceptMethods length = %d, want 2", len(cfg.InterceptMethods))
	}
	if len(cfg.InterceptURLContains) != 2 {
		t.Errorf("InterceptURLContains length = %d, want 2", len(cfg.InterceptURLContains))
	}
	if len(cfg.InterceptContentTypes) != 2 {
		t.Errorf("InterceptContentTypes length = %d, want 2", len(cfg.InterceptContentTypes))
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

func TestGetAPIPort_Default(t *testing.T) {
	os.Unsetenv("API_PORT")
	os.Unsetenv("ADDR")

	result := getAPIPort(":9092")

	if result != ":9092" {
		t.Errorf("getAPIPort() = %q, want %q", result, ":9092")
	}
}

func TestGetAPIPort_FromAPI_PORT_Number(t *testing.T) {
	t.Setenv("API_PORT", "8080")
	os.Unsetenv("ADDR")

	result := getAPIPort(":9092")

	if result != ":8080" {
		t.Errorf("getAPIPort() = %q, want %q", result, ":8080")
	}
}

func TestGetAPIPort_FromAPI_PORT_WithColon(t *testing.T) {
	t.Setenv("API_PORT", ":8080")
	os.Unsetenv("ADDR")

	result := getAPIPort(":9092")

	if result != ":8080" {
		t.Errorf("getAPIPort() = %q, want %q", result, ":8080")
	}
}

func TestGetAPIPort_FromAPI_PORT_InvalidNumber(t *testing.T) {
	t.Setenv("API_PORT", "invalid")
	os.Unsetenv("ADDR")

	result := getAPIPort(":9092")

	if result != "invalid" {
		t.Errorf("getAPIPort() = %q, want %q", result, "invalid")
	}
}

func TestGetAPIPort_FromADDR_Fallback(t *testing.T) {
	os.Unsetenv("API_PORT")
	t.Setenv("ADDR", ":8080")

	result := getAPIPort(":9092")

	if result != ":8080" {
		t.Errorf("getAPIPort() = %q, want %q", result, ":8080")
	}
}

func TestGetAPIPort_API_PORT_Priority(t *testing.T) {
	t.Setenv("API_PORT", "8080")
	t.Setenv("ADDR", ":9999")

	result := getAPIPort(":9092")

	if result != ":8080" {
		t.Errorf("getAPIPort() should prefer API_PORT over ADDR, got %q", result)
	}
}

func TestFilepathJoinSafe(t *testing.T) {
	result := filepathJoinSafe("path", "to", "file")

	expected := "path" + string(os.PathSeparator) + "to" + string(os.PathSeparator) + "file"
	if result != expected {
		t.Errorf("filepathJoinSafe() = %q, want %q", result, expected)
	}
}

func TestFilepathJoinSafe_Empty(t *testing.T) {
	result := filepathJoinSafe()

	if result != "" {
		t.Errorf("filepathJoinSafe() with no args = %q, want empty string", result)
	}
}

func TestFilepathJoinSafe_Single(t *testing.T) {
	result := filepathJoinSafe("single")

	if result != "single" {
		t.Errorf("filepathJoinSafe() = %q, want %q", result, "single")
	}
}

// helper to clear relevant env vars
func clearEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"ADDR", "API_PORT", "LOG_LEVEL", "DEV_MODE", "CORS_ALLOW_ORIGIN",
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
