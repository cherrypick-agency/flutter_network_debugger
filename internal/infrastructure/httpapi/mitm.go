package httpapi

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// CertAuthority encapsulates CA loading and issuing short-lived certificates for domains.
type CertAuthority struct {
	caCert  *x509.Certificate
	caKey   *rsa.PrivateKey
	tlsCert tls.Certificate
	mu      sync.Mutex
	// simple cache to avoid generating certificates for every request
	cache map[string]tls.Certificate
	// issued certificates validity period
	leafTTL time.Duration
}

func LoadCertAuthority(caCertPath, caKeyPath string) (*CertAuthority, error) {
	certPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}
	keyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("mitm: invalid CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	kblk, _ := pem.Decode(keyPEM)
	if kblk == nil {
		return nil, errors.New("mitm: invalid CA key PEM")
	}
	var caKey *rsa.PrivateKey
	if kblk.Type == "RSA PRIVATE KEY" {
		caKey, err = x509.ParsePKCS1PrivateKey(kblk.Bytes)
		if err != nil {
			return nil, err
		}
	} else if kblk.Type == "PRIVATE KEY" {
		pk, err := x509.ParsePKCS8PrivateKey(kblk.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		caKey, ok = pk.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("mitm: only RSA keys are supported for CA")
		}
	} else {
		return nil, errors.New("mitm: unknown CA key PEM block type")
	}
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &CertAuthority{
		caCert:  caCert,
		caKey:   caKey,
		tlsCert: tlsCert,
		cache:   make(map[string]tls.Certificate),
		leafTTL: 24 * time.Hour,
	}, nil
}

// LoadCertAuthorityFromPEM loads CA from PEM content (without temporary files).
func LoadCertAuthorityFromPEM(certPEM, keyPEM []byte) (*CertAuthority, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("mitm: invalid CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	kblk, _ := pem.Decode(keyPEM)
	if kblk == nil {
		return nil, errors.New("mitм: некорректный PEM ключа CA")
	}
	var caKey *rsa.PrivateKey
	if kblk.Type == "RSA PRIVATE KEY" {
		caKey, err = x509.ParsePKCS1PrivateKey(kblk.Bytes)
		if err != nil {
			return nil, err
		}
	} else if kblk.Type == "PRIVATE KEY" {
		pk, err := x509.ParsePKCS8PrivateKey(kblk.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		caKey, ok = pk.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("mitm: only RSA keys are supported for CA")
		}
	} else {
		return nil, errors.New("mitm: unknown CA key PEM block type")
	}
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &CertAuthority{caCert: caCert, caKey: caKey, tlsCert: tlsCert, cache: make(map[string]tls.Certificate), leafTTL: 24 * time.Hour}, nil
}

// GenerateDevCA generates self-signed root CA (RSA) for development.
func GenerateDevCA(commonName string, yearsValid int) (certPEM, keyPEM []byte, err error) {
	if yearsValid <= 0 {
		yearsValid = 5
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().Add(-5 * time.Minute)
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             now,
		NotAfter:              now.AddDate(yearsValid, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		SubjectKeyId:          []byte{1, 2, 3, 4, 5, 6},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}

// IssueFor issues or takes from cache a certificate for sni/host.
func (ca *CertAuthority) IssueFor(host string) (tls.Certificate, error) {
	h := strings.TrimSpace(host)
	if h == "" {
		return tls.Certificate{}, errors.New("mitm: empty host for certificate issuance")
	}
	// normalize possible host:port
	if strings.Contains(h, ":") {
		if v, _, err := net.SplitHostPort(h); err == nil {
			h = v
		}
	}
	ca.mu.Lock()
	if cert, ok := ca.cache[h]; ok {
		ca.mu.Unlock()
		return cert, nil
	}
	ca.mu.Unlock()

	// Generate key and certificate for specific host
	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	now := time.Now().Add(-5 * time.Minute)
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: h},
		NotBefore:             now,
		NotAfter:              now.Add(ca.leafTTL),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{h},
	}
	// If this is IP, add to SAN IPAddresses
	if ip := net.ParseIP(h); ip != nil {
		tmpl.IPAddresses = []net.IP{ip}
		tmpl.DNSNames = nil
		tmpl.Subject = pkix.Name{CommonName: ip.String()}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca.caCert, &leafKey.PublicKey, ca.caKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(leafKey)})
	leaf, err := tls.X509KeyPair(append(certPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.caCert.Raw})...), keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	ca.mu.Lock()
	ca.cache[h] = leaf
	ca.mu.Unlock()
	return leaf, nil
}

// MITM config and domain checks.
type MITM struct {
	CA          *CertAuthority
	AllowSuffix []string
	DenySuffix  []string
}

// shouldIntercept decides whether to intercept domain based on allow/deny lists.
func (m *MITM) shouldIntercept(host string) bool {
	if m == nil || m.CA == nil {
		return false
	}
	h := host
	if h == "" {
		return false
	}
	if strings.Contains(h, ":") {
		if v, _, err := net.SplitHostPort(h); err == nil {
			h = v
		}
	}
	lh := strings.ToLower(h)
	for _, d := range m.DenySuffix {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" && (lh == d || strings.HasSuffix(lh, d)) {
			return false
		}
	}
	if len(m.AllowSuffix) == 0 {
		return true
	}
	for _, a := range m.AllowSuffix {
		a = strings.ToLower(strings.TrimSpace(a))
		if a != "" && (lh == a || strings.HasSuffix(lh, a)) {
			return true
		}
	}
	return false
}

// writeRaw completely dumps resp to w without re-encoding, useful for upgrades.
func writeRaw(w io.Writer, resp *http.Response) error {
	return resp.Write(w)
}
