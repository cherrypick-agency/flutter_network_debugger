package extism

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"
)

// TestIsHostAllowed verifies the host allowlist security feature
func TestIsHostAllowed(t *testing.T) {
	tests := []struct {
		name         string
		targetHost   string
		allowedHosts []string
		want         bool
	}{
		{
			name:         "empty allowlist allows all hosts",
			targetHost:   "example.com",
			allowedHosts: nil,
			want:         true,
		},
		{
			name:         "exact match allowed",
			targetHost:   "api.example.com",
			allowedHosts: []string{"api.example.com"},
			want:         true,
		},
		{
			name:         "exact match denied",
			targetHost:   "evil.com",
			allowedHosts: []string{"api.example.com"},
			want:         false,
		},
		{
			name:         "wildcard subdomain match",
			targetHost:   "api.example.com",
			allowedHosts: []string{"*.example.com"},
			want:         true,
		},
		{
			name:         "wildcard subdomain match nested",
			targetHost:   "v1.api.example.com",
			allowedHosts: []string{"*.example.com"},
			want:         true,
		},
		{
			name:         "wildcard subdomain denied wrong domain",
			targetHost:   "api.evil.com",
			allowedHosts: []string{"*.example.com"},
			want:         false,
		},
		{
			name:         "wildcard matches apex domain",
			targetHost:   "example.com",
			allowedHosts: []string{"*.example.com"},
			want:         true,
		},
		{
			name:         "host with port - allowed",
			targetHost:   "api.example.com:8080",
			allowedHosts: []string{"api.example.com"},
			want:         true,
		},
		{
			name:         "host with port - denied",
			targetHost:   "evil.com:8080",
			allowedHosts: []string{"api.example.com"},
			want:         false,
		},
		{
			name:         "multiple allowed hosts - first matches",
			targetHost:   "api.example.com",
			allowedHosts: []string{"api.example.com", "other.com"},
			want:         true,
		},
		{
			name:         "multiple allowed hosts - second matches",
			targetHost:   "other.com",
			allowedHosts: []string{"api.example.com", "other.com"},
			want:         true,
		},
		{
			name:         "multiple allowed hosts - none match",
			targetHost:   "evil.com",
			allowedHosts: []string{"api.example.com", "other.com"},
			want:         false,
		},
		{
			name:         "localhost allowed",
			targetHost:   "localhost",
			allowedHosts: []string{"localhost"},
			want:         true,
		},
		{
			name:         "localhost with port allowed",
			targetHost:   "localhost:3000",
			allowedHosts: []string{"localhost"},
			want:         true,
		},
		{
			name:         "IP address allowed",
			targetHost:   "192.168.1.1",
			allowedHosts: []string{"192.168.1.1"},
			want:         true,
		},
		{
			name:         "IP address with port allowed",
			targetHost:   "192.168.1.1:8080",
			allowedHosts: []string{"192.168.1.1"},
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHostAllowed(tt.targetHost, tt.allowedHosts)
			if got != tt.want {
				t.Errorf("isHostAllowed(%q, %v) = %v, want %v",
					tt.targetHost, tt.allowedHosts, got, tt.want)
			}
		})
	}
}

// TestCreateLogFunction tests createLogFunction
func TestCreateLogFunction(t *testing.T) {
	hostFn := createLogFunction()
	if hostFn.Name == "" {
		t.Error("createLogFunction() should return function with name")
	}
}

// TestCreateHTTPFetchFunction tests createHTTPFetchFunction
func TestCreateHTTPFetchFunction(t *testing.T) {
	allowedHosts := []string{"api.example.com"}
	hostFn := createHTTPFetchFunction(allowedHosts)
	if hostFn.Name == "" {
		t.Error("createHTTPFetchFunction() should return function with name")
	}
}

// TestCreateHTTPFetchFunction_EmptyAllowedHosts tests createHTTPFetchFunction with empty allowedHosts
func TestCreateHTTPFetchFunction_EmptyAllowedHosts(t *testing.T) {
	hostFn := createHTTPFetchFunction(nil)
	if hostFn.Name == "" {
		t.Error("createHTTPFetchFunction() should return function with name")
	}
}

// TestCreateHostFunctions tests createHostFunctions
func TestCreateHostFunctions(t *testing.T) {
	allowedHosts := []string{"api.example.com"}
	hostFns := createHostFunctions(allowedHosts)

	if len(hostFns) == 0 {
		t.Error("createHostFunctions() should return at least one function")
	}

	if len(hostFns) < 2 {
		t.Error("createHostFunctions() should return log and http_fetch functions")
	}
}

// TestCreateHostFunctions_NilAllowedHosts tests createHostFunctions with nil allowedHosts
func TestCreateHostFunctions_NilAllowedHosts(t *testing.T) {
	hostFns := createHostFunctions(nil)

	if len(hostFns) == 0 {
		t.Error("createHostFunctions() should return at least one function")
	}
}

func TestNewHTTPFetchClient_CheckRedirectBlocksDisallowedHost(t *testing.T) {
	client := newHTTPFetchClient([]string{"example.com"})

	req := &http.Request{
		URL: mustParseURL(t, "https://evil.com/"),
	}

	if err := client.CheckRedirect(req, []*http.Request{{URL: mustParseURL(t, "https://example.com/")}}); err == nil {
		t.Fatal("expected redirect check to fail for disallowed host")
	}
}

func TestNewHTTPFetchClient_CheckRedirectAllowsAllowedHost(t *testing.T) {
	client := newHTTPFetchClient([]string{"example.com"})

	req := &http.Request{
		URL: mustParseURL(t, "https://example.com/next"),
	}

	if err := client.CheckRedirect(req, []*http.Request{{URL: mustParseURL(t, "https://example.com/")}}); err != nil {
		t.Fatalf("expected redirect check to pass, got: %v", err)
	}
}

func TestWithMaxTimeoutDoesNotExtendDeadline(t *testing.T) {
	ctx, cancel := withMaxTimeout(withDeadline(t, 100*time.Millisecond), 30*time.Second)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}
	if time.Until(deadline) > time.Second {
		t.Fatal("deadline was unexpectedly extended")
	}
}

func withDeadline(t *testing.T, d time.Duration) (ctx context.Context) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

func mustParseURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u
}
