package httpapi

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCertAuthorityFromPEM_ValidRSAPrivateKey(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)

	if err != nil {
		t.Fatalf("LoadCertAuthorityFromPEM failed: %v", err)
	}
	if ca == nil {
		t.Fatal("ca should not be nil")
	}
	if ca.caCert == nil {
		t.Error("caCert should not be nil")
	}
	if ca.caKey == nil {
		t.Error("caKey should not be nil")
	}
	if ca.cache == nil {
		t.Error("cache should be initialized")
	}
}

func TestLoadCertAuthorityFromPEM_ValidPKCS8PrivateKey(t *testing.T) {
	// Generate a matching cert and key, then convert key to PKCS8
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Decode the RSA key
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		t.Fatal("failed to decode key PEM")
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse RSA key: %v", err)
	}

	// Re-encode as PKCS8
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	if err != nil {
		t.Fatalf("failed to marshal PKCS8 key: %v", err)
	}
	keyPEMPKCS8 := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Key})

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEMPKCS8)

	if err != nil {
		t.Fatalf("LoadCertAuthorityFromPEM failed: %v", err)
	}
	if ca == nil {
		t.Fatal("ca should not be nil")
	}
	if ca.caKey == nil {
		t.Error("caKey should not be nil")
	}
}

func TestLoadCertAuthorityFromPEM_InvalidCertPEM(t *testing.T) {
	_, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	invalidCertPEM := []byte("not a valid PEM")

	_, err = LoadCertAuthorityFromPEM(invalidCertPEM, keyPEM)

	if err == nil {
		t.Error("should fail with invalid cert PEM")
	}
}

func TestLoadCertAuthorityFromPEM_InvalidCertContent(t *testing.T) {
	_, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Valid PEM structure but invalid certificate content
	invalidCertPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: []byte("invalid cert data"),
	})

	_, err = LoadCertAuthorityFromPEM(invalidCertPEM, keyPEM)

	if err == nil {
		t.Error("should fail with invalid cert content")
	}
}

func TestLoadCertAuthorityFromPEM_InvalidKeyPEM(t *testing.T) {
	certPEM, _, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	invalidKeyPEM := []byte("not a valid PEM")

	_, err = LoadCertAuthorityFromPEM(certPEM, invalidKeyPEM)

	if err == nil {
		t.Error("should fail with invalid key PEM")
	}
}

func TestLoadCertAuthorityFromPEM_InvalidRSAKeyContent(t *testing.T) {
	certPEM, _, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Valid PEM structure but invalid RSA key content
	invalidKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: []byte("invalid key data"),
	})

	_, err = LoadCertAuthorityFromPEM(certPEM, invalidKeyPEM)

	if err == nil {
		t.Error("should fail with invalid RSA key content")
	}
}

func TestLoadCertAuthorityFromPEM_InvalidPKCS8KeyContent(t *testing.T) {
	certPEM, _, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Valid PEM structure but invalid PKCS8 key content
	invalidKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: []byte("invalid pkcs8 data"),
	})

	_, err = LoadCertAuthorityFromPEM(certPEM, invalidKeyPEM)

	if err == nil {
		t.Error("should fail with invalid PKCS8 key content")
	}
}

func TestLoadCertAuthorityFromPEM_UnsupportedKeyType(t *testing.T) {
	certPEM, _, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Valid PEM but unsupported key type
	invalidKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: []byte("ec key data"),
	})

	_, err = LoadCertAuthorityFromPEM(certPEM, invalidKeyPEM)

	if err == nil {
		t.Error("should fail with unsupported key type")
	}
}

func TestLoadCertAuthorityFromPEM_MismatchedCertAndKey(t *testing.T) {
	// Generate two different CAs
	certPEM1, _, err := GenerateDevCA("Test CA 1", 1)
	if err != nil {
		t.Fatalf("failed to generate CA1: %v", err)
	}

	_, keyPEM2, err := GenerateDevCA("Test CA 2", 1)
	if err != nil {
		t.Fatalf("failed to generate CA2: %v", err)
	}

	// Try to load with mismatched cert and key
	_, err = LoadCertAuthorityFromPEM(certPEM1, keyPEM2)

	if err == nil {
		t.Error("should fail with mismatched cert and key")
	}
}

func TestGenerateDevCA_Success(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 5)

	if err != nil {
		t.Fatalf("GenerateDevCA failed: %v", err)
	}
	if certPEM == nil || len(certPEM) == 0 {
		t.Error("certPEM should not be empty")
	}
	if keyPEM == nil || len(keyPEM) == 0 {
		t.Error("keyPEM should not be empty")
	}

	// Verify we can load the generated CA
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load generated CA: %v", err)
	}
	if ca == nil {
		t.Error("ca should not be nil")
	}
}

func TestGenerateDevCA_ZeroYearsValid(t *testing.T) {
	// Should default to 5 years
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 0)

	if err != nil {
		t.Fatalf("GenerateDevCA failed: %v", err)
	}
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		t.Error("should generate valid CA even with 0 years")
	}

	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}
	if ca == nil {
		t.Error("ca should not be nil")
	}
}

func TestGenerateDevCA_NegativeYearsValid(t *testing.T) {
	// Should default to 5 years
	certPEM, keyPEM, err := GenerateDevCA("Test CA", -1)

	if err != nil {
		t.Fatalf("GenerateDevCA failed: %v", err)
	}
	if len(certPEM) == 0 || len(keyPEM) == 0 {
		t.Error("should generate valid CA even with negative years")
	}
}

func TestMITM_ShouldIntercept_NilMITM(t *testing.T) {
	var m *MITM = nil

	if m.shouldIntercept("example.com") {
		t.Error("nil MITM should not intercept")
	}
}

func TestMITM_ShouldIntercept_NoCA(t *testing.T) {
	m := &MITM{CA: nil}

	if m.shouldIntercept("example.com") {
		t.Error("MITM without CA should not intercept")
	}
}

func TestMITM_ShouldIntercept_EmptyHost(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test", 1)
	ca, _ := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	m := &MITM{CA: ca}

	if m.shouldIntercept("") {
		t.Error("should not intercept empty host")
	}
}

func TestMITM_ShouldIntercept_DeniedDomain(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test", 1)
	ca, _ := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	m := &MITM{
		CA:         ca,
		DenySuffix: []string{"bad.com", ".evil.org"},
	}

	if m.shouldIntercept("bad.com") {
		t.Error("should not intercept denied domain")
	}
	if m.shouldIntercept("sub.evil.org") {
		t.Error("should not intercept denied suffix")
	}
	if !m.shouldIntercept("good.com") {
		t.Error("should intercept non-denied domain when no allow list")
	}
}

func TestMITM_ShouldIntercept_AllowedDomain(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test", 1)
	ca, _ := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	m := &MITM{
		CA:          ca,
		AllowSuffix: []string{"example.com", ".test.org"},
	}

	if !m.shouldIntercept("example.com") {
		t.Error("should intercept allowed domain")
	}
	if !m.shouldIntercept("api.test.org") {
		t.Error("should intercept allowed suffix")
	}
	if m.shouldIntercept("other.com") {
		t.Error("should not intercept non-allowed domain")
	}
}

func TestMITM_ShouldIntercept_DenyOverridesAllow(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test", 1)
	ca, _ := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	m := &MITM{
		CA:          ca,
		AllowSuffix: []string{".example.com"},
		DenySuffix:  []string{"bad.example.com"},
	}

	if m.shouldIntercept("bad.example.com") {
		t.Error("deny should override allow")
	}
	if !m.shouldIntercept("good.example.com") {
		t.Error("should intercept allowed domain not in deny list")
	}
}

func TestMITM_ShouldIntercept_HostWithPort(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test", 1)
	ca, _ := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	m := &MITM{
		CA:          ca,
		AllowSuffix: []string{"example.com"},
	}

	if !m.shouldIntercept("example.com:443") {
		t.Error("should intercept host with port")
	}
}

func TestMITM_ShouldIntercept_CaseInsensitive(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test", 1)
	ca, _ := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	m := &MITM{
		CA:          ca,
		AllowSuffix: []string{"Example.COM"},
	}

	if !m.shouldIntercept("example.com") {
		t.Error("should be case insensitive")
	}
	if !m.shouldIntercept("EXAMPLE.COM") {
		t.Error("should be case insensitive")
	}
}

func TestCertAuthority_IssueFor_EmptyHost(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test CA", 1)
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	_, err = ca.IssueFor("")

	if err == nil {
		t.Error("should fail with empty host")
	}
}

func TestCertAuthority_IssueFor_ValidHost(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test CA", 1)
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	cert, err := ca.IssueFor("example.com")

	if err != nil {
		t.Fatalf("IssueFor failed: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Error("certificate should not be empty")
	}
}

func TestCertAuthority_IssueFor_Caching(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test CA", 1)
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	cert1, err := ca.IssueFor("example.com")
	if err != nil {
		t.Fatalf("IssueFor failed: %v", err)
	}

	cert2, err := ca.IssueFor("example.com")
	if err != nil {
		t.Fatalf("IssueFor failed: %v", err)
	}

	// Should return the same certificate from cache
	if len(cert1.Certificate) != len(cert2.Certificate) {
		t.Error("cached certificate should be the same")
	}
}

func TestCertAuthority_IssueFor_HostWithPort(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test CA", 1)
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	cert, err := ca.IssueFor("example.com:443")

	if err != nil {
		t.Fatalf("IssueFor failed: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Error("certificate should not be empty")
	}
}

func TestCertAuthority_IssueFor_IPAddress(t *testing.T) {
	certPEM, keyPEM, _ := GenerateDevCA("Test CA", 1)
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("failed to load CA: %v", err)
	}

	cert, err := ca.IssueFor("192.168.1.1")

	if err != nil {
		t.Fatalf("IssueFor failed for IP: %v", err)
	}
	if len(cert.Certificate) == 0 {
		t.Error("certificate should not be empty for IP")
	}
}

func TestLoadCertAuthority_Success(t *testing.T) {
	// Generate test cert and key
	certPEM, keyPEM, err := GenerateDevCA("Test LoadCA", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Write to temporary files
	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "cert.pem")
	keyPath := filepath.Join(tempDir, "key.pem")

	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	// Load from files
	ca, err := LoadCertAuthority(certPath, keyPath)

	if err != nil {
		t.Fatalf("LoadCertAuthority failed: %v", err)
	}
	if ca == nil {
		t.Fatal("ca should not be nil")
	}
	if ca.caCert == nil {
		t.Error("caCert should not be nil")
	}
	if ca.caKey == nil {
		t.Error("caKey should not be nil")
	}
}

func TestLoadCertAuthority_CertFileNotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, err := LoadCertAuthority(
		filepath.Join(tempDir, "nonexistent.pem"),
		filepath.Join(tempDir, "key.pem"),
	)

	if err == nil {
		t.Error("should fail when cert file doesn't exist")
	}
}

func TestLoadCertAuthority_KeyFileNotFound(t *testing.T) {
	// Generate test cert
	certPEM, _, err := GenerateDevCA("Test", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "cert.pem")
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}

	_, err = LoadCertAuthority(certPath, filepath.Join(tempDir, "nonexistent.pem"))

	if err == nil {
		t.Error("should fail when key file doesn't exist")
	}
}

func TestLoadCertAuthority_InvalidCertFile(t *testing.T) {
	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "cert.pem")
	keyPath := filepath.Join(tempDir, "key.pem")

	// Write invalid cert
	if err := os.WriteFile(certPath, []byte("invalid cert"), 0o600); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("invalid key"), 0o600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	_, err := LoadCertAuthority(certPath, keyPath)

	if err == nil {
		t.Error("should fail with invalid cert file")
	}
}

func TestLoadCertAuthority_PKCS8Key(t *testing.T) {
	// Generate test cert and key, then convert key to PKCS8
	certPEM, keyPEM, err := GenerateDevCA("Test", 1)
	if err != nil {
		t.Fatalf("failed to generate CA: %v", err)
	}

	// Decode the RSA key
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		t.Fatal("failed to decode key PEM")
	}
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse RSA key: %v", err)
	}

	// Re-encode as PKCS8
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	if err != nil {
		t.Fatalf("failed to marshal PKCS8 key: %v", err)
	}
	keyPEMPKCS8 := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Key})

	// Write to temporary files
	tempDir := t.TempDir()
	certPath := filepath.Join(tempDir, "cert.pem")
	keyPath := filepath.Join(tempDir, "key.pem")

	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		t.Fatalf("failed to write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, keyPEMPKCS8, 0o600); err != nil {
		t.Fatalf("failed to write key: %v", err)
	}

	// Load from files
	ca, err := LoadCertAuthority(certPath, keyPath)

	if err != nil {
		t.Fatalf("LoadCertAuthority failed: %v", err)
	}
	if ca == nil {
		t.Fatal("ca should not be nil")
	}
}
