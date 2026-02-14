package httpapi

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"
	"testing"
)

func TestHumanizeProxyError_NilError(t *testing.T) {
	code, msg := humanizeProxyError(nil)

	if code != "UNKNOWN_ERROR" {
		t.Errorf("code = %s, want UNKNOWN_ERROR", code)
	}
	if msg == "" {
		t.Error("message should not be empty")
	}
}

func TestHumanizeProxyError_EOF(t *testing.T) {
	code, msg := humanizeProxyError(io.EOF)

	if code != "CONNECTION_CLOSED" {
		t.Errorf("code = %s, want CONNECTION_CLOSED", code)
	}
	if msg == "" {
		t.Error("message should not be empty")
	}
}

func TestHumanizeProxyError_EOFString(t *testing.T) {
	code, _ := humanizeProxyError(errors.New("EOF"))

	if code != "CONNECTION_CLOSED" {
		t.Errorf("code = %s, want CONNECTION_CLOSED", code)
	}
}

func TestHumanizeProxyError_ConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp: connection refused")
	code, msg := humanizeProxyError(err)

	if code != "SERVER_UNAVAILABLE" {
		t.Errorf("code = %s, want SERVER_UNAVAILABLE", code)
	}
	if msg == "" {
		t.Error("message should not be empty")
	}
}

func TestHumanizeProxyError_NoSuchHost(t *testing.T) {
	err := errors.New("dial tcp: lookup example.invalid: no such host")
	code, _ := humanizeProxyError(err)

	if code != "DNS_ERROR" {
		t.Errorf("code = %s, want DNS_ERROR", code)
	}
}

func TestHumanizeProxyError_DNSError(t *testing.T) {
	dnsErr := &net.DNSError{
		Err:  "no such host",
		Name: "invalid.example.com",
	}
	code, msg := humanizeProxyError(dnsErr)

	if code != "DNS_ERROR" {
		t.Errorf("code = %s, want DNS_ERROR", code)
	}
	if msg == "" {
		t.Error("message should not be empty")
	}
}

func TestHumanizeProxyError_DeadlineExceeded(t *testing.T) {
	code, _ := humanizeProxyError(context.DeadlineExceeded)

	if code != "TIMEOUT" {
		t.Errorf("code = %s, want TIMEOUT", code)
	}
}

func TestHumanizeProxyError_ContextCanceled(t *testing.T) {
	code, msg := humanizeProxyError(context.Canceled)
	if code != "CANCELED" {
		t.Errorf("code = %s, want CANCELED", code)
	}
	if msg == "" {
		t.Error("message should not be empty")
	}
}

func TestHumanizeProxyError_TimeoutString(t *testing.T) {
	err := errors.New("i/o timeout")
	code, _ := humanizeProxyError(err)

	if code != "TIMEOUT" {
		t.Errorf("code = %s, want TIMEOUT", code)
	}
}

func TestHumanizeProxyError_ContextDeadline(t *testing.T) {
	err := errors.New("context deadline exceeded")
	code, _ := humanizeProxyError(err)

	if code != "TIMEOUT" {
		t.Errorf("code = %s, want TIMEOUT", code)
	}
}

func TestHumanizeProxyError_TLSError(t *testing.T) {
	err := errors.New("tls: bad certificate")
	code, _ := humanizeProxyError(err)

	if code != "TLS_ERROR" {
		t.Errorf("code = %s, want TLS_ERROR", code)
	}
}

func TestHumanizeProxyError_CertificateError(t *testing.T) {
	err := errors.New("x509: certificate has expired")
	code, _ := humanizeProxyError(err)

	if code != "TLS_ERROR" {
		t.Errorf("code = %s, want TLS_ERROR", code)
	}
}

func TestHumanizeProxyError_NetworkUnreachable(t *testing.T) {
	err := errors.New("connect: network is unreachable")
	code, _ := humanizeProxyError(err)

	if code != "NETWORK_UNREACHABLE" {
		t.Errorf("code = %s, want NETWORK_UNREACHABLE", code)
	}
}

func TestHumanizeProxyError_ConnectionReset(t *testing.T) {
	err := errors.New("read tcp: connection reset by peer")
	code, _ := humanizeProxyError(err)

	if code != "CONNECTION_RESET" {
		t.Errorf("code = %s, want CONNECTION_RESET", code)
	}
}

func TestHumanizeProxyError_GenericError(t *testing.T) {
	err := errors.New("some unknown error")
	code, _ := humanizeProxyError(err)

	// Should have some code, even if generic
	if code == "" {
		t.Error("code should not be empty")
	}
}

func TestTryDecompress_ValidGzip(t *testing.T) {
	original := []byte("hello world test data")

	// Compress with gzip
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(original)
	_ = zw.Close()

	// Decompress
	decompressed, ok := tryDecompress(buf.Bytes(), "gzip")

	if !ok {
		t.Fatal("tryDecompress should succeed for valid gzip")
	}
	if string(decompressed) != string(original) {
		t.Errorf("decompressed = %q, want %q", decompressed, original)
	}
}

func TestTryDecompress_ValidDeflate(t *testing.T) {
	original := []byte("hello world deflate test")

	// Compress with deflate
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	_, _ = zw.Write(original)
	_ = zw.Close()

	// Decompress
	decompressed, ok := tryDecompress(buf.Bytes(), "deflate")

	if !ok {
		t.Fatal("tryDecompress should succeed for valid deflate")
	}
	if string(decompressed) != string(original) {
		t.Errorf("decompressed = %q, want %q", decompressed, original)
	}
}

func TestTryDecompress_UnknownEncoding(t *testing.T) {
	data := []byte("some data")

	_, ok := tryDecompress(data, "brotli")

	if ok {
		t.Error("tryDecompress should fail for unknown encoding")
	}
}

func TestTryDecompress_EmptyEncoding(t *testing.T) {
	data := []byte("some data")

	_, ok := tryDecompress(data, "")

	if ok {
		t.Error("tryDecompress should fail for empty encoding")
	}
}

func TestTryDecompress_InvalidGzipData(t *testing.T) {
	// Already covered in existing tests, but let's be thorough
	_, ok := tryDecompress([]byte("not gzip data"), "gzip")

	if ok {
		t.Error("tryDecompress should fail for invalid gzip data")
	}
}

func TestTryDecompress_InvalidDeflateData(t *testing.T) {
	// Already covered in existing tests
	_, ok := tryDecompress([]byte("not deflate data"), "deflate")

	if ok {
		t.Error("tryDecompress should fail for invalid deflate data")
	}
}

func TestTryDecompress_LargeData(t *testing.T) {
	// Test with data larger than 1MB limit
	largeData := bytes.Repeat([]byte("A"), 2*1024*1024) // 2MB

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, _ = zw.Write(largeData)
	_ = zw.Close()

	decompressed, ok := tryDecompress(buf.Bytes(), "gzip")

	if !ok {
		t.Fatal("tryDecompress should succeed")
	}
	// Should be limited to 1MB
	if len(decompressed) > 1*1024*1024 {
		t.Errorf("decompressed size = %d, should be limited to 1MB", len(decompressed))
	}
}
