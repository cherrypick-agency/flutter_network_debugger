package httpapi

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRuleStringMatch_Matches(t *testing.T) {
	tests := []struct {
		name  string
		match *RuleStringMatch
		input string
		want  bool
	}{
		{
			name:  "nil match - always true",
			match: nil,
			input: "anything",
			want:  true,
		},
		{
			name:  "empty match - always true",
			match: &RuleStringMatch{},
			input: "anything",
			want:  true,
		},
		{
			name: "equals match - success",
			match: &RuleStringMatch{
				Equals: "exact",
			},
			input: "exact",
			want:  true,
		},
		{
			name: "equals match - fail",
			match: &RuleStringMatch{
				Equals: "exact",
			},
			input: "different",
			want:  false,
		},
		{
			name: "prefix match - success",
			match: &RuleStringMatch{
				Prefix: "test",
			},
			input: "testing",
			want:  true,
		},
		{
			name: "prefix match - fail",
			match: &RuleStringMatch{
				Prefix: "test",
			},
			input: "failing",
			want:  false,
		},
		{
			name: "suffix match - success",
			match: &RuleStringMatch{
				Suffix: ".json",
			},
			input: "data.json",
			want:  true,
		},
		{
			name: "suffix match - fail",
			match: &RuleStringMatch{
				Suffix: ".json",
			},
			input: "data.xml",
			want:  false,
		},
		{
			name: "contains match - success",
			match: &RuleStringMatch{
				Contains: "api",
			},
			input: "https://example.com/api/v1",
			want:  true,
		},
		{
			name: "contains match - fail",
			match: &RuleStringMatch{
				Contains: "api",
			},
			input: "https://example.com/web/v1",
			want:  false,
		},
		{
			name: "anyOf match - success first",
			match: &RuleStringMatch{
				AnyOf: []string{"foo", "bar", "baz"},
			},
			input: "foo",
			want:  true,
		},
		{
			name: "anyOf match - success middle",
			match: &RuleStringMatch{
				AnyOf: []string{"foo", "bar", "baz"},
			},
			input: "bar",
			want:  true,
		},
		{
			name: "anyOf match - fail",
			match: &RuleStringMatch{
				AnyOf: []string{"foo", "bar", "baz"},
			},
			input: "qux",
			want:  false,
		},
		{
			name: "regex match - success",
			match: &RuleStringMatch{
				Regex: "^test.*[0-9]+$",
			},
			input: "test123",
			want:  true,
		},
		{
			name: "regex match - fail",
			match: &RuleStringMatch{
				Regex: "^test.*[0-9]+$",
			},
			input: "testing",
			want:  false,
		},
		{
			name: "regex match - invalid regex",
			match: &RuleStringMatch{
				Regex: "[invalid(",
			},
			input: "anything",
			want:  false,
		},
		{
			name: "multiple conditions - first matches",
			match: &RuleStringMatch{
				Equals:   "exact",
				Contains: "api",
			},
			input: "exact",
			want:  true,
		},
		{
			name: "multiple conditions - second matches",
			match: &RuleStringMatch{
				Equals:   "exact",
				Contains: "api",
			},
			input: "contains-api-string",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.matches(tt.input)
			if got != tt.want {
				t.Errorf("matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsEmptyRuleStringMatch(t *testing.T) {
	tests := []struct {
		name  string
		match RuleStringMatch
		want  bool
	}{
		{
			name:  "completely empty",
			match: RuleStringMatch{},
			want:  true,
		},
		{
			name: "has equals",
			match: RuleStringMatch{
				Equals: "value",
			},
			want: false,
		},
		{
			name: "has prefix",
			match: RuleStringMatch{
				Prefix: "pre",
			},
			want: false,
		},
		{
			name: "has suffix",
			match: RuleStringMatch{
				Suffix: "suf",
			},
			want: false,
		},
		{
			name: "has contains",
			match: RuleStringMatch{
				Contains: "con",
			},
			want: false,
		},
		{
			name: "has anyOf",
			match: RuleStringMatch{
				AnyOf: []string{"a", "b"},
			},
			want: false,
		},
		{
			name: "has regex",
			match: RuleStringMatch{
				Regex: ".*",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEmptyRuleStringMatch(tt.match)
			if got != tt.want {
				t.Errorf("isEmptyRuleStringMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleHeaderMatch_Matches(t *testing.T) {
	tests := []struct {
		name    string
		match   *RuleHeaderMatch
		headers http.Header
		want    bool
	}{
		{
			name:  "nil match - always true",
			match: nil,
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: true,
		},
		{
			name: "header name matches, empty value",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Content-Type"},
				Value: RuleStringMatch{},
			},
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: true,
		},
		{
			name: "header name and value match",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Content-Type"},
				Value: RuleStringMatch{Contains: "json"},
			},
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: true,
		},
		{
			name: "header name matches, value doesn't",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Content-Type"},
				Value: RuleStringMatch{Contains: "xml"},
			},
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: false,
		},
		{
			name: "header name doesn't match",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Authorization"},
				Value: RuleStringMatch{},
			},
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: false,
		},
		{
			name: "matches one of multiple values",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Accept"},
				Value: RuleStringMatch{Contains: "json"},
			},
			headers: http.Header{
				"Accept": []string{"text/html", "application/json", "text/xml"},
			},
			want: true,
		},
		{
			name: "no matching value in multiple",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Accept"},
				Value: RuleStringMatch{Contains: "pdf"},
			},
			headers: http.Header{
				"Accept": []string{"text/html", "application/json", "text/xml"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.matches(tt.headers)
			if got != tt.want {
				t.Errorf("matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleStatusMatch_Matches(t *testing.T) {
	tests := []struct {
		name  string
		match *RuleStatusMatch
		code  int
		want  bool
	}{
		{
			name:  "nil match - always true",
			match: nil,
			code:  404,
			want:  true,
		},
		{
			name:  "empty match - always true",
			match: &RuleStatusMatch{},
			code:  404,
			want:  true,
		},
		{
			name: "equals match - success",
			match: &RuleStatusMatch{
				Equals: []int{200, 201, 204},
			},
			code: 200,
			want: true,
		},
		{
			name: "equals match - fail",
			match: &RuleStatusMatch{
				Equals: []int{200, 201, 204},
			},
			code: 404,
			want: false,
		},
		{
			name: "range match - success",
			match: &RuleStatusMatch{
				From: 200,
				To:   299,
			},
			code: 250,
			want: true,
		},
		{
			name: "range match - fail",
			match: &RuleStatusMatch{
				From: 200,
				To:   299,
			},
			code: 404,
			want: false,
		},
		{
			name: "range with only From",
			match: &RuleStatusMatch{
				From: 400,
			},
			code: 500,
			want: true,
		},
		{
			name: "range with only To",
			match: &RuleStatusMatch{
				To: 299,
			},
			code: 200,
			want: true,
		},
		{
			name: "range edge case - exactly From",
			match: &RuleStatusMatch{
				From: 200,
				To:   299,
			},
			code: 200,
			want: true,
		},
		{
			name: "range edge case - exactly To",
			match: &RuleStatusMatch{
				From: 200,
				To:   299,
			},
			code: 299,
			want: true,
		},
		{
			name: "Is4xx - success",
			match: &RuleStatusMatch{
				Is4xx: true,
			},
			code: 404,
			want: true,
		},
		{
			name: "Is4xx - fail",
			match: &RuleStatusMatch{
				Is4xx: true,
			},
			code: 200,
			want: false,
		},
		{
			name: "Is5xx - success",
			match: &RuleStatusMatch{
				Is5xx: true,
			},
			code: 500,
			want: true,
		},
		{
			name: "Is5xx - fail",
			match: &RuleStatusMatch{
				Is5xx: true,
			},
			code: 200,
			want: false,
		},
		{
			name: "4xx edge cases - 400",
			match: &RuleStatusMatch{
				Is4xx: true,
			},
			code: 400,
			want: true,
		},
		{
			name: "4xx edge cases - 499",
			match: &RuleStatusMatch{
				Is4xx: true,
			},
			code: 499,
			want: true,
		},
		{
			name: "5xx edge cases - 500",
			match: &RuleStatusMatch{
				Is5xx: true,
			},
			code: 500,
			want: true,
		},
		{
			name: "5xx edge cases - 599",
			match: &RuleStatusMatch{
				Is5xx: true,
			},
			code: 599,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.matches(tt.code)
			if got != tt.want {
				t.Errorf("matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterceptWhen_RequestMatch(t *testing.T) {
	tests := []struct {
		name        string
		when        *InterceptWhen
		method      string
		urlStr      string
		headers     http.Header
		bodyPreview string
		want        bool
	}{
		{
			name:   "nil when - always true",
			when:   nil,
			method: "GET",
			urlStr: "http://example.com/",
			want:   true,
		},
		{
			name: "method match - success",
			when: &InterceptWhen{
				Method: []string{"GET", "POST"},
			},
			method: "GET",
			urlStr: "http://example.com/",
			want:   true,
		},
		{
			name: "method match - fail",
			when: &InterceptWhen{
				Method: []string{"POST", "PUT"},
			},
			method: "GET",
			urlStr: "http://example.com/",
			want:   false,
		},
		{
			name: "scheme match - success",
			when: &InterceptWhen{
				Scheme: []string{"https"},
			},
			method: "GET",
			urlStr: "https://example.com/",
			want:   true,
		},
		{
			name: "scheme match - fail",
			when: &InterceptWhen{
				Scheme: []string{"https"},
			},
			method: "GET",
			urlStr: "http://example.com/",
			want:   false,
		},
		{
			name: "host match - success",
			when: &InterceptWhen{
				Host: &RuleStringMatch{Equals: "example.com"},
			},
			method: "GET",
			urlStr: "http://example.com/path",
			want:   true,
		},
		{
			name: "host match - fail",
			when: &InterceptWhen{
				Host: &RuleStringMatch{Equals: "other.com"},
			},
			method: "GET",
			urlStr: "http://example.com/path",
			want:   false,
		},
		{
			name: "path match - success",
			when: &InterceptWhen{
				Path: &RuleStringMatch{Prefix: "/api"},
			},
			method: "GET",
			urlStr: "http://example.com/api/users",
			want:   true,
		},
		{
			name: "path match - fail",
			when: &InterceptWhen{
				Path: &RuleStringMatch{Prefix: "/api"},
			},
			method: "GET",
			urlStr: "http://example.com/web/users",
			want:   false,
		},
		{
			name: "content-type match - success",
			when: &InterceptWhen{
				ContentType: &RuleStringMatch{Contains: "json"},
			},
			method: "POST",
			urlStr: "http://example.com/",
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: true,
		},
		{
			name: "content-type match - fail",
			when: &InterceptWhen{
				ContentType: &RuleStringMatch{Contains: "json"},
			},
			method: "POST",
			urlStr: "http://example.com/",
			headers: http.Header{
				"Content-Type": []string{"text/xml"},
			},
			want: false,
		},
		{
			name: "header match - success",
			when: &InterceptWhen{
				Header: &RuleHeaderMatch{
					Name:  RuleStringMatch{Equals: "Authorization"},
					Value: RuleStringMatch{Prefix: "Bearer"},
				},
			},
			method: "GET",
			urlStr: "http://example.com/",
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
			},
			want: true,
		},
		{
			name: "body contains - success",
			when: &InterceptWhen{
				BodyContains: "search-term",
			},
			method:      "POST",
			urlStr:      "http://example.com/",
			bodyPreview: "this is search-term in body",
			want:        true,
		},
		{
			name: "body contains - fail",
			when: &InterceptWhen{
				BodyContains: "search-term",
			},
			method:      "POST",
			urlStr:      "http://example.com/",
			bodyPreview: "this is different body",
			want:        false,
		},
		{
			name: "port match - success",
			when: &InterceptWhen{
				Port: &RuleStringMatch{Equals: "8080"},
			},
			method: "GET",
			urlStr: "http://example.com:8080/",
			want:   true,
		},
		{
			name: "port match - fail",
			when: &InterceptWhen{
				Port: &RuleStringMatch{Equals: "8080"},
			},
			method: "GET",
			urlStr: "http://example.com:9090/",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, _ := url.Parse(tt.urlStr)
			req := &http.Request{
				Method: tt.method,
				URL:    parsedURL,
				Header: tt.headers,
			}
			if tt.headers == nil {
				req.Header = http.Header{}
			}

			got := tt.when.requestMatch(req, tt.bodyPreview)
			if got != tt.want {
				t.Errorf("requestMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterceptWhen_ResponseMatch(t *testing.T) {
	tests := []struct {
		name        string
		when        *InterceptWhen
		statusCode  int
		headers     http.Header
		bodyPreview string
		want        bool
	}{
		{
			name:       "nil when - always true",
			when:       nil,
			statusCode: 200,
			want:       true,
		},
		{
			name: "status match - success",
			when: &InterceptWhen{
				Status: &RuleStatusMatch{
					Is4xx: true,
				},
			},
			statusCode: 404,
			want:       true,
		},
		{
			name: "status match - fail",
			when: &InterceptWhen{
				Status: &RuleStatusMatch{
					Is4xx: true,
				},
			},
			statusCode: 200,
			want:       false,
		},
		{
			name: "content-type match - success",
			when: &InterceptWhen{
				ContentType: &RuleStringMatch{Contains: "json"},
			},
			statusCode: 200,
			headers: http.Header{
				"Content-Type": []string{"application/json"},
			},
			want: true,
		},
		{
			name: "content-type match - fail",
			when: &InterceptWhen{
				ContentType: &RuleStringMatch{Contains: "json"},
			},
			statusCode: 200,
			headers: http.Header{
				"Content-Type": []string{"text/html"},
			},
			want: false,
		},
		{
			name: "header match - success",
			when: &InterceptWhen{
				Header: &RuleHeaderMatch{
					Name:  RuleStringMatch{Equals: "X-Custom"},
					Value: RuleStringMatch{Equals: "value"},
				},
			},
			statusCode: 200,
			headers: http.Header{
				"X-Custom": []string{"value"},
			},
			want: true,
		},
		{
			name: "body contains - success",
			when: &InterceptWhen{
				BodyContains: "error",
			},
			statusCode:  500,
			bodyPreview: "internal error occurred",
			want:        true,
		},
		{
			name: "body contains - fail",
			when: &InterceptWhen{
				BodyContains: "error",
			},
			statusCode:  200,
			bodyPreview: "success message",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     tt.headers,
			}
			if tt.headers == nil {
				resp.Header = http.Header{}
			}

			got := tt.when.responseMatch(resp, tt.bodyPreview)
			if got != tt.want {
				t.Errorf("responseMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
