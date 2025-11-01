package runtime

import (
	"net/http"
	"net/url"
	"testing"

	mdomain "network-debugger/internal/features/mapping/domain"
)

func TestManager_GlobLocal(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "api.example.com", PathPattern: "/bundle.js", PatternType: mdomain.PatternGlob,
		StatusOverride: 200, ContentTypeOverride: "application/javascript",
	}}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/bundle.js")
	req := &http.Request{Method: http.MethodGet, URL: u}
	dec, ok := m.EvalRequest(req)
	if !ok || dec.Kind != mdomain.KindLocal || dec.RuleID != "r1" {
		t.Fatalf("expected local r1, got %+v ok=%v", dec, ok)
	}
}

func TestManager_RegexRemote_PreserveHost(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r2", Enabled: true, Priority: 1, Kind: mdomain.KindRemote,
		HostPattern: "^prod\\.example\\.com$", PathPattern: "/v1/.*", PatternType: mdomain.PatternRegex,
		TargetURLTemplate: "https://staging.example.com$0", PreserveHost: true,
	}}
	m.Update(rules)
	u, _ := url.Parse("https://prod.example.com/v1/users")
	req := &http.Request{Method: http.MethodGet, URL: u}
	dec, ok := m.EvalRequest(req)
	if !ok || dec.Kind != mdomain.KindRemote || dec.RuleID != "r2" || !dec.PreserveHost {
		t.Fatalf("expected remote r2 preserveHost, got %+v ok=%v", dec, ok)
	}
}

func TestManager_DisabledRule(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: false, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "api.example.com", PathPattern: "/test", PatternType: mdomain.PatternGlob,
	}}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/test")
	req := &http.Request{Method: http.MethodGet, URL: u}
	_, ok := m.EvalRequest(req)
	if ok {
		t.Fatal("disabled rule should not match")
	}
}

func TestManager_MethodFilter(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "api.example.com", PathPattern: "/test", PatternType: mdomain.PatternGlob,
		Methods: []string{"POST", "PUT"},
	}}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/test")

	// GET should not match
	req := &http.Request{Method: http.MethodGet, URL: u}
	_, ok := m.EvalRequest(req)
	if ok {
		t.Fatal("GET should not match POST/PUT rule")
	}

	// POST should match
	req = &http.Request{Method: http.MethodPost, URL: u}
	dec, ok := m.EvalRequest(req)
	if !ok || dec.RuleID != "r1" {
		t.Fatal("POST should match")
	}
}

func TestManager_NilURL(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "api.example.com", PathPattern: "/test", PatternType: mdomain.PatternGlob,
	}}
	m.Update(rules)
	req := &http.Request{Method: http.MethodGet, URL: nil}
	_, ok := m.EvalRequest(req)
	if ok {
		t.Fatal("nil URL should not match")
	}
}

func TestManager_InvalidRegex(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "[invalid(regex", PathPattern: "/test", PatternType: mdomain.PatternRegex,
	}}
	m.Update(rules) // Should not panic with invalid regex
	u, _ := url.Parse("https://api.example.com/test")
	req := &http.Request{Method: http.MethodGet, URL: u}
	dec, ok := m.EvalRequest(req)
	// Invalid regex is treated as no pattern, so it should match
	if !ok || dec.RuleID != "r1" {
		t.Fatal("invalid regex is treated as no pattern and should match")
	}
}

func TestManager_GlobWildcard(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "*.example.com", PathPattern: "/api/*", PatternType: mdomain.PatternGlob,
	}}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/api/users")
	req := &http.Request{Method: http.MethodGet, URL: u}
	dec, ok := m.EvalRequest(req)
	if !ok || dec.RuleID != "r1" {
		t.Fatal("glob wildcard should match")
	}
}

func TestManager_InvalidTargetURL(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindRemote,
		HostPattern: "api.example.com", PathPattern: "/test", PatternType: mdomain.PatternGlob,
		TargetURLTemplate: "://invalid-url",
	}}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/test")
	req := &http.Request{Method: http.MethodGet, URL: u}
	_, ok := m.EvalRequest(req)
	if ok {
		t.Fatal("invalid target URL should not produce a decision")
	}
}

func TestManager_MultipleRulesPriority(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{
		{
			ID: "r1", Enabled: true, Priority: 2, Kind: mdomain.KindLocal,
			HostPattern: "api.example.com", PathPattern: "/test", PatternType: mdomain.PatternGlob,
		},
		{
			ID: "r2", Enabled: true, Priority: 1, Kind: mdomain.KindRemote,
			HostPattern: "api.example.com", PathPattern: "/test", PatternType: mdomain.PatternGlob,
			TargetURLTemplate: "https://target.com/test",
		},
	}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/test")
	req := &http.Request{Method: http.MethodGet, URL: u}
	dec, ok := m.EvalRequest(req)
	if !ok || dec.RuleID != "r1" {
		t.Fatal("first rule should match (rules processed in order)")
	}
}

func TestManager_EmptyHostPath(t *testing.T) {
	m := New()
	rules := []mdomain.MapRule{{
		ID: "r1", Enabled: true, Priority: 1, Kind: mdomain.KindLocal,
		HostPattern: "", PathPattern: "", PatternType: mdomain.PatternGlob,
	}}
	m.Update(rules)
	u, _ := url.Parse("https://api.example.com/test")
	req := &http.Request{Method: http.MethodGet, URL: u}
	dec, ok := m.EvalRequest(req)
	if !ok || dec.RuleID != "r1" {
		t.Fatal("empty patterns should match everything")
	}
}

func TestNonZero(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		fallback int
		want     int
	}{
		{"zero value uses fallback", 0, 200, 200},
		{"non-zero value preserved", 404, 200, 404},
		{"both zero uses first", 0, 0, 0},
		{"negative value preserved", -1, 200, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nonZero(tt.val, tt.fallback)
			if got != tt.want {
				t.Errorf("nonZero(%d, %d) = %d, want %d", tt.val, tt.fallback, got, tt.want)
			}
		})
	}
}
