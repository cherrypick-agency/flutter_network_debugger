package httpapi

import (
	"net/http"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"testing"
	"time"
)

func TestNewTransport_Properties(t *testing.T) {
	tr := newTransport(cfgpkg.Config{InsecureTLS: true})
	if tr.ResponseHeaderTimeout != 25*time.Second {
		t.Fatalf("resp header timeout")
	}
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("insecure tls expected")
	}
	// Should be usable to create a request (basic sanity)
	_ = http.Client{Transport: tr}
}
