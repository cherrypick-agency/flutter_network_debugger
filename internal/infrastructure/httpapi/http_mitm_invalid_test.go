package httpapi

import "testing"

func TestLoadCertAuthorityFromPEM_Invalid(t *testing.T) {
	if _, err := LoadCertAuthorityFromPEM([]byte("not-pem"), []byte("not-pem")); err == nil {
		t.Fatalf("expected error for invalid PEM")
	}
}
