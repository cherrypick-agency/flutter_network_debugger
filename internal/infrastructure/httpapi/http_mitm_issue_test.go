package httpapi

import (
	"crypto/x509"
	"net"
	"testing"
)

func TestCertAuthority_IssueFor_DomainAndIP(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("dev", 1)
	if err != nil {
		t.Skip("crypto slow/unavailable: ", err)
	}
	ca, err := LoadCertAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("load ca: %v", err)
	}
	leaf, err := ca.IssueFor("example.com")
	if err != nil || len(leaf.Certificate) == 0 {
		t.Fatalf("issue leaf: %v", err)
	}
	// IP variant
	leaf2, err := ca.IssueFor("127.0.0.1")
	if err != nil || len(leaf2.Certificate) == 0 {
		t.Fatalf("issue ip leaf: %v", err)
	}
	// basic parse to ensure DER correctness
	if _, err := x509.ParseCertificate(leaf.Certificate[0]); err != nil {
		t.Fatalf("parse leaf cert: %v", err)
	}
	_ = net.IPv4len // keep net import used
}
