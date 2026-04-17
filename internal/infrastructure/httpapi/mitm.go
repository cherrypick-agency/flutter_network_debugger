package httpapi

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"

	mitmproxy "github.com/777genius/proxykit/mitm"
)

type CertAuthority = mitmproxy.Authority

func LoadCertAuthority(caCertPath, caKeyPath string) (*CertAuthority, error) {
	return mitmproxy.LoadAuthority(caCertPath, caKeyPath)
}

func LoadCertAuthorityFromPEM(certPEM, keyPEM []byte) (*CertAuthority, error) {
	return mitmproxy.LoadAuthorityFromPEM(certPEM, keyPEM)
}

func GenerateDevCA(commonName string, yearsValid int) (certPEM, keyPEM []byte, err error) {
	return mitmproxy.GenerateDevCA(commonName, yearsValid)
}

type MITM struct {
	CA          *CertAuthority
	AllowSuffix []string
	DenySuffix  []string
}

func (m *MITM) shouldIntercept(host string) bool {
	if m == nil {
		return false
	}
	return mitmproxy.Policy{
		Authority:   m.CA,
		AllowSuffix: m.AllowSuffix,
		DenySuffix:  m.DenySuffix,
	}.ShouldIntercept(host)
}

func pemEncodeCert(w http.ResponseWriter, der []byte) error {
	return mitmproxy.EncodeCertificatePEM(w, der)
}

// writeRaw completely dumps resp to w without re-encoding, useful for upgrades.
func writeRaw(w io.Writer, resp *http.Response) error {
	return resp.Write(w)
}

func authorityRootCertificate(ca *CertAuthority) *x509.Certificate {
	if ca == nil {
		return nil
	}
	return ca.RootCertificate()
}

func authorityPrivateKey(ca *CertAuthority) *rsa.PrivateKey {
	if ca == nil {
		return nil
	}
	return ca.PrivateKey()
}

func authorityTLSCertificate(ca *CertAuthority) tls.Certificate {
	if ca == nil {
		return tls.Certificate{}
	}
	return ca.TLSCertificate()
}
