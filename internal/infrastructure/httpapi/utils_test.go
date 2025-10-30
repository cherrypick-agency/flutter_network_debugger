package httpapi

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestIsAbsoluteURL_MoreCases(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"http absolute", "http://example.com", true},
		{"https absolute", "https://example.com", true},
		{"ws absolute", "ws://localhost:8080", true},
		{"wss absolute", "wss://secure.example.com", true},
		{"relative path", "/path/to/resource", false},
		{"relative with query", "/path?query=1", false},
		{"empty string", "", false},
		{"just scheme", "http://", false},
		{"just host", "example.com", false},
		{"invalid url", "ht tp://bad url", false},
		{"scheme only", "http:", false},
		{"absolute with port", "http://example.com:8080/path", true},
		{"absolute with user", "http://user@example.com", true},
		{"ftp absolute", "ftp://ftp.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAbsoluteURL(tt.url)
			if got != tt.want {
				t.Errorf("isAbsoluteURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestTLSVersionString_AllVersions(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{tls.VersionTLS13, "TLS 1.3"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS10, "TLS 1.0"},
		{0, ""},
		{0xFFFF, ""},
		{0x0300, ""}, // SSL 3.0 - not supported
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tlsVersionString(tt.version)
			if got != tt.want {
				t.Errorf("tlsVersionString(0x%04X) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestCipherSuiteString_AllCases(t *testing.T) {
	tests := []struct {
		name string
		id   uint16
		want string
	}{
		{"AES_128_GCM_SHA256", tls.TLS_AES_128_GCM_SHA256, "TLS_AES_128_GCM_SHA256"},
		{"AES_256_GCM_SHA384", tls.TLS_AES_256_GCM_SHA384, "TLS_AES_256_GCM_SHA384"},
		{"CHACHA20_POLY1305", tls.TLS_CHACHA20_POLY1305_SHA256, "TLS_CHACHA20_POLY1305_SHA256"},
		{"ECDHE_RSA_AES_128", tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"},
		{"ECDHE_RSA_AES_256", tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"},
		{"ECDHE_ECDSA_AES_128", tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"},
		{"ECDHE_ECDSA_AES_256", tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
		{"unknown_small", 0x0001, "0x1"},
		{"unknown_large", 0xABCD, "0xABCD"},
		{"zero", 0x0000, "0x0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cipherSuiteString(tt.id)
			if got != tt.want {
				t.Errorf("cipherSuiteString(0x%04X) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestUseOrFallback_BothSet(t *testing.T) {
	a := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	b := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	result := useOrFallback(a, b)

	if !result.Equal(a) {
		t.Errorf("useOrFallback should return a when both set, got %v", result)
	}
}

func TestUseOrFallback_FirstZero(t *testing.T) {
	a := time.Time{}
	b := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)

	result := useOrFallback(a, b)

	if !result.Equal(b) {
		t.Errorf("useOrFallback should return b when a is zero, got %v", result)
	}
}

func TestUseOrFallback_SecondZero(t *testing.T) {
	a := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	b := time.Time{}

	result := useOrFallback(a, b)

	if !result.Equal(a) {
		t.Errorf("useOrFallback should return a when set, got %v", result)
	}
}

func TestUseOrFallback_BothZero(t *testing.T) {
	a := time.Time{}
	b := time.Time{}

	result := useOrFallback(a, b)

	if !result.IsZero() {
		t.Errorf("useOrFallback should return zero when both zero, got %v", result)
	}
}

func TestDurationMs_Normal(t *testing.T) {
	from := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 10, 0, 1, 500000000, time.UTC) // +1.5 seconds

	ms := durationMs(from, to)

	if ms != 1500 {
		t.Errorf("durationMs = %d, want 1500", ms)
	}
}

func TestDurationMs_FromZero(t *testing.T) {
	from := time.Time{}
	to := time.Now()

	ms := durationMs(from, to)

	if ms != 0 {
		t.Errorf("durationMs with zero from should return 0, got %d", ms)
	}
}

func TestDurationMs_ToZero(t *testing.T) {
	from := time.Now()
	to := time.Time{}

	ms := durationMs(from, to)

	if ms != 0 {
		t.Errorf("durationMs with zero to should return 0, got %d", ms)
	}
}

func TestDurationMs_BothZero(t *testing.T) {
	ms := durationMs(time.Time{}, time.Time{})

	if ms != 0 {
		t.Errorf("durationMs with both zero should return 0, got %d", ms)
	}
}

func TestInt64ToInt_Normal(t *testing.T) {
	tests := []struct {
		input int64
		want  int
	}{
		{100, 100},
		{0, 0},
		{-10, 0},  // negative clamped to 0
		{-100, 0}, // negative clamped to 0
		{12345, 12345},
	}

	for _, tt := range tests {
		got := int64ToInt(tt.input)
		if got != tt.want {
			t.Errorf("int64ToInt(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestInt64ToInt_Overflow(t *testing.T) {
	// Test with a value at the edge (this tests the boundary condition)
	// On 64-bit systems, max int is the same as max int64, so we test the logic
	maxIntAsInt64 := int64(^uint(0) >> 1)

	result := int64ToInt(maxIntAsInt64)

	// Should return max int
	maxInt := int(^uint(0) >> 1)
	if result != maxInt {
		t.Errorf("int64ToInt at boundary = %d, want %d", result, maxInt)
	}
}

func TestInt64ToInt_ExceedsMax(t *testing.T) {
	// On 32-bit systems, max int is smaller than max int64
	// This test is meaningful on 32-bit architectures
	// On 64-bit systems, we can't actually create an int64 larger than max int
	// without overflow, so the logic path is covered by the boundary test

	// Skip this test on 64-bit systems where int64 and int have same range
	maxInt := int(^uint(0) >> 1)
	maxInt64 := int64(maxInt)

	if maxInt64 == (1<<63 - 1) {
		// 64-bit system - can't overflow, skip
		t.Skip("Skipping overflow test on 64-bit system")
	}

	// On 32-bit: max int is ~2 billion, but int64 can hold ~9 quintillion
	overflowValue := int64(maxInt) + 1000

	result := int64ToInt(overflowValue)

	// Should clamp to max int
	if result != maxInt {
		t.Errorf("int64ToInt with overflow = %d, want %d (max int)", result, maxInt)
	}
}
