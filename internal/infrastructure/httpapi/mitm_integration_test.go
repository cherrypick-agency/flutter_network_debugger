package httpapi

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestMITM_FullTLSHandshake tests the complete MITM flow:
// client -> MITM proxy -> upstream HTTPS server
func TestMITM_FullTLSHandshake(t *testing.T) {
	// 1. Generate CA certificate
	certPEM, keyPEM, err := GenerateDevCA("Test MITM CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	// 2. Create upstream HTTPS server
	upstreamHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Upstream-Server", "true")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("upstream response"))
	})
	upstream := httptest.NewTLSServer(upstreamHandler)
	defer upstream.Close()

	// 3. Setup MITM proxy
	mitm := &MITM{CA: ca}

	// 4. Start a listener for the MITM proxy
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	proxyAddr := listener.Addr().String()

	// Extract upstream host:port
	upstreamHost := upstream.Listener.Addr().String()

	// 5. Handle CONNECT requests in a goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read CONNECT request
		br := bufio.NewReader(conn)
		req, err := http.ReadRequest(br)
		if err != nil {
			t.Logf("failed to read request: %v", err)
			return
		}

		if req.Method != http.MethodConnect {
			t.Logf("unexpected method: %s", req.Method)
			return
		}

		// Send 200 Connection Established
		_, _ = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

		// Get certificate for the requested host
		leaf, err := mitm.CA.IssueFor(req.Host)
		if err != nil {
			t.Logf("failed to issue certificate: %v", err)
			return
		}

		// Setup TLS server for client
		tlsSrv := tls.Server(conn, &tls.Config{
			Certificates: []tls.Certificate{leaf},
			NextProtos:   []string{"http/1.1"},
		})
		defer tlsSrv.Close()

		if err := tlsSrv.Handshake(); err != nil {
			t.Logf("TLS handshake with client failed: %v", err)
			return
		}

		// Connect to upstream
		upstreamConn, err := net.Dial("tcp", upstreamHost)
		if err != nil {
			t.Logf("failed to dial upstream: %v", err)
			return
		}
		defer upstreamConn.Close()

		// Setup TLS client for upstream
		serverName := req.Host
		if h, _, err := net.SplitHostPort(req.Host); err == nil {
			serverName = h
		}
		tlsCli := tls.Client(upstreamConn, &tls.Config{
			ServerName:         serverName,
			InsecureSkipVerify: true, // upstream is test server
		})
		defer tlsCli.Close()

		if err := tlsCli.Handshake(); err != nil {
			t.Logf("TLS handshake with upstream failed: %v", err)
			return
		}

		// Read HTTP request from client
		clientBR := bufio.NewReader(tlsSrv)
		httpReq, err := http.ReadRequest(clientBR)
		if err != nil {
			t.Logf("failed to read HTTP request: %v", err)
			return
		}

		// Forward to upstream
		httpReq.URL.Scheme = "https"
		httpReq.URL.Host = upstreamHost
		httpReq.RequestURI = ""

		if err := httpReq.Write(tlsCli); err != nil {
			t.Logf("failed to write request to upstream: %v", err)
			return
		}

		// Read response from upstream
		upstreamBR := bufio.NewReader(tlsCli)
		resp, err := http.ReadResponse(upstreamBR, httpReq)
		if err != nil {
			t.Logf("failed to read response from upstream: %v", err)
			return
		}
		defer resp.Body.Close()

		// Forward response to client
		if err := resp.Write(tlsSrv); err != nil {
			t.Logf("failed to write response to client: %v", err)
			return
		}
	}()

	// 6. Create client with custom CA pool
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPEM)

	// Manually establish CONNECT tunnel
	proxyConn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("failed to connect to proxy: %v", err)
	}
	defer proxyConn.Close()

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", upstreamHost, upstreamHost)
	_, err = proxyConn.Write([]byte(connectReq))
	if err != nil {
		t.Fatalf("failed to send CONNECT: %v", err)
	}

	// Read CONNECT response
	br := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(br, &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatalf("failed to read CONNECT response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("CONNECT failed: %s", resp.Status)
	}

	// Upgrade to TLS
	// ServerName should be hostname only, not host:port
	serverName := upstreamHost
	if h, _, err := net.SplitHostPort(upstreamHost); err == nil {
		serverName = h
	}
	tlsConn := tls.Client(proxyConn, &tls.Config{
		ServerName: serverName,
		RootCAs:    certPool,
	})
	defer tlsConn.Close()

	if err := tlsConn.Handshake(); err != nil {
		t.Fatalf("TLS handshake failed: %v", err)
	}

	// Verify certificate
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		t.Fatal("no peer certificates")
	}
	// CN should be hostname only (without port)
	expectedCN := serverName
	if state.PeerCertificates[0].Subject.CommonName != expectedCN {
		t.Errorf("unexpected CN: got %s, want %s",
			state.PeerCertificates[0].Subject.CommonName, expectedCN)
	}

	// Send HTTP request through the tunnel
	req := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\n\r\n", upstreamHost)
	_, err = tlsConn.Write([]byte(req))
	if err != nil {
		t.Fatalf("failed to send HTTP request: %v", err)
	}

	// Read HTTP response
	httpResp, err := http.ReadResponse(bufio.NewReader(tlsConn), &http.Request{Method: "GET"})
	if err != nil {
		t.Fatalf("failed to read HTTP response: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %s", httpResp.Status)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != "upstream response" {
		t.Errorf("unexpected body: got %q, want %q", string(body), "upstream response")
	}

	if httpResp.Header.Get("X-Upstream-Server") != "true" {
		t.Error("upstream header not found")
	}

	// Wait for goroutine to finish
	wg.Wait()
}

// TestMITM_CertificateValidation tests that client validates MITM certificates correctly
func TestMITM_CertificateValidation(t *testing.T) {
	// Generate CA
	certPEM, keyPEM, err := GenerateDevCA("Test CA Validation", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	// Issue certificate for a specific host
	cert, err := ca.IssueFor("example.com")
	if err != nil {
		t.Fatalf("failed to issue certificate: %v", err)
	}

	// Parse the certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	// Verify certificate properties
	if x509Cert.Subject.CommonName != "example.com" {
		t.Errorf("unexpected CN: got %s, want %s", x509Cert.Subject.CommonName, "example.com")
	}

	if len(x509Cert.DNSNames) != 1 || x509Cert.DNSNames[0] != "example.com" {
		t.Errorf("unexpected DNS names: %v", x509Cert.DNSNames)
	}

	// Verify certificate chain
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPEM)

	opts := x509.VerifyOptions{
		DNSName: "example.com",
		Roots:   certPool,
	}

	if _, err := x509Cert.Verify(opts); err != nil {
		t.Errorf("certificate verification failed: %v", err)
	}
}

// TestMITM_SNIHandling tests Server Name Indication handling
func TestMITM_SNIHandling(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test SNI CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	testCases := []struct {
		name     string
		hostname string
	}{
		{"domain", "example.com"},
		{"subdomain", "api.example.com"},
		{"port", "example.com:8443"},
		{"ip", "192.168.1.1"},
		{"ipv6", "::1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cert, err := ca.IssueFor(tc.hostname)
			if err != nil {
				t.Fatalf("failed to issue cert for %s: %v", tc.hostname, err)
			}

			x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
			if err != nil {
				t.Fatalf("failed to parse certificate: %v", err)
			}

			// Verify certificate was issued for the correct hostname
			expectedHost := tc.hostname
			if h, _, err := net.SplitHostPort(tc.hostname); err == nil {
				expectedHost = h
			}

			// For IP addresses, check IPAddresses field
			if ip := net.ParseIP(expectedHost); ip != nil {
				if len(x509Cert.IPAddresses) == 0 {
					t.Errorf("expected IP address in certificate")
				} else if !x509Cert.IPAddresses[0].Equal(ip) {
					t.Errorf("unexpected IP: got %v, want %v", x509Cert.IPAddresses[0], ip)
				}
			} else {
				// For DNS names, check DNSNames field
				if len(x509Cert.DNSNames) == 0 {
					t.Errorf("expected DNS name in certificate")
				} else if x509Cert.DNSNames[0] != expectedHost {
					t.Errorf("unexpected DNS name: got %s, want %s", x509Cert.DNSNames[0], expectedHost)
				}
			}
		})
	}
}

// TestMITM_CertificateCaching tests that certificates are properly cached
func TestMITM_CertificateCaching(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test Cache CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	hostname := "cache-test.example.com"

	// Issue certificate first time
	cert1, err := ca.IssueFor(hostname)
	if err != nil {
		t.Fatalf("failed to issue cert: %v", err)
	}

	// Issue certificate second time (should come from cache)
	cert2, err := ca.IssueFor(hostname)
	if err != nil {
		t.Fatalf("failed to issue cert: %v", err)
	}

	// Compare certificate bytes to ensure they're identical
	if !bytes.Equal(cert1.Certificate[0], cert2.Certificate[0]) {
		t.Error("cached certificate differs from original")
	}

	// Verify cache contains the entry
	if !ca.HasCached(hostname) {
		t.Error("certificate not found in cache")
	}
	cacheSize := ca.CacheSize()

	if cacheSize != 1 {
		t.Errorf("unexpected cache size: got %d, want 1", cacheSize)
	}

	// Issue for different hostname
	cert3, err := ca.IssueFor("other.example.com")
	if err != nil {
		t.Fatalf("failed to issue cert for other host: %v", err)
	}

	// Should be different from the first one
	if bytes.Equal(cert1.Certificate[0], cert3.Certificate[0]) {
		t.Error("certificates for different hosts should differ")
	}

	// Cache should now have 2 entries
	cacheSize = ca.CacheSize()

	if cacheSize != 2 {
		t.Errorf("unexpected cache size: got %d, want 2", cacheSize)
	}
}

// TestMITM_ConcurrentCertificateIssuance tests concurrent certificate generation
func TestMITM_ConcurrentCertificateIssuance(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test Concurrent CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	hostname := "concurrent.example.com"
	numGoroutines := 10

	var wg sync.WaitGroup
	certs := make([]tls.Certificate, numGoroutines)
	errors := make([]error, numGoroutines)

	// Issue certificates concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cert, err := ca.IssueFor(hostname)
			certs[idx] = cert
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Errorf("goroutine %d failed: %v", i, err)
		}
	}

	// All certificates should be identical (from cache)
	for i := 1; i < numGoroutines; i++ {
		if !bytes.Equal(certs[0].Certificate[0], certs[i].Certificate[0]) {
			t.Errorf("certificate %d differs from certificate 0", i)
		}
	}

	// Cache should have only 1 entry
	cacheSize := ca.CacheSize()

	if cacheSize != 1 {
		t.Errorf("unexpected cache size: got %d, want 1", cacheSize)
	}
}

// TestMITM_CertificateExpiry tests certificate validity periods
func TestMITM_CertificateExpiry(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test Expiry CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	cert, err := ca.IssueFor("expiry-test.example.com")
	if err != nil {
		t.Fatalf("failed to issue cert: %v", err)
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		t.Fatalf("failed to parse certificate: %v", err)
	}

	// Check NotBefore is in the past (with 5 minute buffer)
	if x509Cert.NotBefore.After(time.Now()) {
		t.Error("certificate NotBefore is in the future")
	}

	// Check NotAfter is in the future
	if x509Cert.NotAfter.Before(time.Now()) {
		t.Error("certificate has already expired")
	}

	// Check validity period is approximately 24 hours (leafTTL)
	duration := x509Cert.NotAfter.Sub(x509Cert.NotBefore)
	expectedDuration := 24 * time.Hour
	tolerance := 10 * time.Minute

	if duration < expectedDuration-tolerance || duration > expectedDuration+tolerance {
		t.Errorf("unexpected validity period: got %v, want ~%v", duration, expectedDuration)
	}
}
