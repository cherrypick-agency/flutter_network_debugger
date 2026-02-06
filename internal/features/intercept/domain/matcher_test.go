package domain

import (
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
			name:  "equals match - success",
			match: &RuleStringMatch{Equals: "exact"},
			input: "exact",
			want:  true,
		},
		{
			name:  "equals match - fail",
			match: &RuleStringMatch{Equals: "exact"},
			input: "different",
			want:  false,
		},
		{
			name:  "prefix match - success",
			match: &RuleStringMatch{Prefix: "test"},
			input: "testing",
			want:  true,
		},
		{
			name:  "prefix match - fail",
			match: &RuleStringMatch{Prefix: "test"},
			input: "failing",
			want:  false,
		},
		{
			name:  "suffix match - success",
			match: &RuleStringMatch{Suffix: ".json"},
			input: "data.json",
			want:  true,
		},
		{
			name:  "suffix match - fail",
			match: &RuleStringMatch{Suffix: ".json"},
			input: "data.xml",
			want:  false,
		},
		{
			name:  "contains match - success",
			match: &RuleStringMatch{Contains: "api"},
			input: "https://example.com/api/v1",
			want:  true,
		},
		{
			name:  "contains match - fail",
			match: &RuleStringMatch{Contains: "api"},
			input: "https://example.com/web/v1",
			want:  false,
		},
		{
			name:  "anyOf match - success first",
			match: &RuleStringMatch{AnyOf: []string{"foo", "bar", "baz"}},
			input: "foo",
			want:  true,
		},
		{
			name:  "anyOf match - success middle",
			match: &RuleStringMatch{AnyOf: []string{"foo", "bar", "baz"}},
			input: "bar",
			want:  true,
		},
		{
			name:  "anyOf match - fail",
			match: &RuleStringMatch{AnyOf: []string{"foo", "bar", "baz"}},
			input: "qux",
			want:  false,
		},
		{
			name:  "regex match - success",
			match: &RuleStringMatch{Regex: "^test.*[0-9]+$"},
			input: "test123",
			want:  true,
		},
		{
			name:  "regex match - fail",
			match: &RuleStringMatch{Regex: "^test.*[0-9]+$"},
			input: "testing",
			want:  false,
		},
		{
			name:  "regex match - invalid regex",
			match: &RuleStringMatch{Regex: "[invalid("},
			input: "anything",
			want:  false,
		},
		{
			name:  "multiple conditions - first matches",
			match: &RuleStringMatch{Equals: "exact", Contains: "api"},
			input: "exact",
			want:  true,
		},
		{
			name:  "multiple conditions - second matches",
			match: &RuleStringMatch{Equals: "exact", Contains: "api"},
			input: "contains-api-string",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.Matches(tt.input)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
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
		{name: "completely empty", match: RuleStringMatch{}, want: true},
		{name: "has equals", match: RuleStringMatch{Equals: "value"}, want: false},
		{name: "has prefix", match: RuleStringMatch{Prefix: "pre"}, want: false},
		{name: "has suffix", match: RuleStringMatch{Suffix: "suf"}, want: false},
		{name: "has contains", match: RuleStringMatch{Contains: "con"}, want: false},
		{name: "has anyOf", match: RuleStringMatch{AnyOf: []string{"a", "b"}}, want: false},
		{name: "has regex", match: RuleStringMatch{Regex: ".*"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.IsEmpty()
			if got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleHeaderMatch_MatchesHeaders(t *testing.T) {
	tests := []struct {
		name    string
		match   *RuleHeaderMatch
		headers map[string][]string
		want    bool
	}{
		{
			name:    "nil match - always true",
			match:   nil,
			headers: map[string][]string{"Content-Type": {"application/json"}},
			want:    true,
		},
		{
			name: "header name matches, empty value",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Content-Type"},
				Value: RuleStringMatch{},
			},
			headers: map[string][]string{"Content-Type": {"application/json"}},
			want:    true,
		},
		{
			name: "header name and value match",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Content-Type"},
				Value: RuleStringMatch{Contains: "json"},
			},
			headers: map[string][]string{"Content-Type": {"application/json"}},
			want:    true,
		},
		{
			name: "header name matches, value doesn't",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Content-Type"},
				Value: RuleStringMatch{Contains: "xml"},
			},
			headers: map[string][]string{"Content-Type": {"application/json"}},
			want:    false,
		},
		{
			name: "header name doesn't match",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Authorization"},
				Value: RuleStringMatch{},
			},
			headers: map[string][]string{"Content-Type": {"application/json"}},
			want:    false,
		},
		{
			name: "matches one of multiple values",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Accept"},
				Value: RuleStringMatch{Contains: "json"},
			},
			headers: map[string][]string{"Accept": {"text/html", "application/json", "text/xml"}},
			want:    true,
		},
		{
			name: "no matching value in multiple",
			match: &RuleHeaderMatch{
				Name:  RuleStringMatch{Equals: "Accept"},
				Value: RuleStringMatch{Contains: "pdf"},
			},
			headers: map[string][]string{"Accept": {"text/html", "application/json", "text/xml"}},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.MatchesHeaders(tt.headers)
			if got != tt.want {
				t.Errorf("MatchesHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleStatusMatch_MatchesCode(t *testing.T) {
	tests := []struct {
		name  string
		match *RuleStatusMatch
		code  int
		want  bool
	}{
		{name: "nil match - always true", match: nil, code: 404, want: true},
		{name: "empty match - always true", match: &RuleStatusMatch{}, code: 404, want: true},
		{name: "equals match - success", match: &RuleStatusMatch{Equals: []int{200, 201, 204}}, code: 200, want: true},
		{name: "equals match - fail", match: &RuleStatusMatch{Equals: []int{200, 201, 204}}, code: 404, want: false},
		{name: "range match - success", match: &RuleStatusMatch{From: 200, To: 299}, code: 250, want: true},
		{name: "range match - fail", match: &RuleStatusMatch{From: 200, To: 299}, code: 404, want: false},
		{name: "range with only From", match: &RuleStatusMatch{From: 400}, code: 500, want: true},
		{name: "range with only To", match: &RuleStatusMatch{To: 299}, code: 200, want: true},
		{name: "range edge - exactly From", match: &RuleStatusMatch{From: 200, To: 299}, code: 200, want: true},
		{name: "range edge - exactly To", match: &RuleStatusMatch{From: 200, To: 299}, code: 299, want: true},
		{name: "Is4xx - success", match: &RuleStatusMatch{Is4xx: true}, code: 404, want: true},
		{name: "Is4xx - fail", match: &RuleStatusMatch{Is4xx: true}, code: 200, want: false},
		{name: "Is5xx - success", match: &RuleStatusMatch{Is5xx: true}, code: 500, want: true},
		{name: "Is5xx - fail", match: &RuleStatusMatch{Is5xx: true}, code: 200, want: false},
		{name: "4xx edge - 400", match: &RuleStatusMatch{Is4xx: true}, code: 400, want: true},
		{name: "4xx edge - 499", match: &RuleStatusMatch{Is4xx: true}, code: 499, want: true},
		{name: "5xx edge - 500", match: &RuleStatusMatch{Is5xx: true}, code: 500, want: true},
		{name: "5xx edge - 599", match: &RuleStatusMatch{Is5xx: true}, code: 599, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.match.MatchesCode(tt.code)
			if got != tt.want {
				t.Errorf("MatchesCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterceptWhen_MatchesRequest(t *testing.T) {
	tests := []struct {
		name  string
		when  *InterceptWhen
		input RequestMatchInput
		want  bool
	}{
		{
			name:  "nil when - always true",
			when:  nil,
			input: RequestMatchInput{Method: "GET"},
			want:  true,
		},
		{
			name:  "method match - success",
			when:  &InterceptWhen{Method: []string{"GET", "POST"}},
			input: RequestMatchInput{Method: "GET"},
			want:  true,
		},
		{
			name:  "method match - fail",
			when:  &InterceptWhen{Method: []string{"POST", "PUT"}},
			input: RequestMatchInput{Method: "GET"},
			want:  false,
		},
		{
			name:  "scheme match - success",
			when:  &InterceptWhen{Scheme: []string{"https"}},
			input: RequestMatchInput{Method: "GET", Scheme: "https"},
			want:  true,
		},
		{
			name:  "scheme match - fail",
			when:  &InterceptWhen{Scheme: []string{"https"}},
			input: RequestMatchInput{Method: "GET", Scheme: "http"},
			want:  false,
		},
		{
			name:  "host match - success",
			when:  &InterceptWhen{Host: &RuleStringMatch{Equals: "example.com"}},
			input: RequestMatchInput{Method: "GET", Host: "example.com"},
			want:  true,
		},
		{
			name:  "host match - fail",
			when:  &InterceptWhen{Host: &RuleStringMatch{Equals: "other.com"}},
			input: RequestMatchInput{Method: "GET", Host: "example.com"},
			want:  false,
		},
		{
			name:  "path match - success",
			when:  &InterceptWhen{Path: &RuleStringMatch{Prefix: "/api"}},
			input: RequestMatchInput{Method: "GET", Path: "/api/users"},
			want:  true,
		},
		{
			name:  "path match - fail",
			when:  &InterceptWhen{Path: &RuleStringMatch{Prefix: "/api"}},
			input: RequestMatchInput{Method: "GET", Path: "/web/users"},
			want:  false,
		},
		{
			name:  "content-type match - success",
			when:  &InterceptWhen{ContentType: &RuleStringMatch{Contains: "json"}},
			input: RequestMatchInput{Method: "POST", ContentType: "application/json"},
			want:  true,
		},
		{
			name:  "content-type match - fail",
			when:  &InterceptWhen{ContentType: &RuleStringMatch{Contains: "json"}},
			input: RequestMatchInput{Method: "POST", ContentType: "text/xml"},
			want:  false,
		},
		{
			name: "header match - success",
			when: &InterceptWhen{
				Header: &RuleHeaderMatch{
					Name:  RuleStringMatch{Equals: "Authorization"},
					Value: RuleStringMatch{Prefix: "Bearer"},
				},
			},
			input: RequestMatchInput{
				Method:  "GET",
				Headers: map[string][]string{"Authorization": {"Bearer token123"}},
			},
			want: true,
		},
		{
			name:  "body contains - success",
			when:  &InterceptWhen{BodyContains: "search-term"},
			input: RequestMatchInput{Method: "POST", BodyPreview: "this is search-term in body"},
			want:  true,
		},
		{
			name:  "body contains - fail",
			when:  &InterceptWhen{BodyContains: "search-term"},
			input: RequestMatchInput{Method: "POST", BodyPreview: "this is different body"},
			want:  false,
		},
		{
			name:  "port match - success",
			when:  &InterceptWhen{Port: &RuleStringMatch{Equals: "8080"}},
			input: RequestMatchInput{Method: "GET", Port: "8080"},
			want:  true,
		},
		{
			name:  "port match - fail",
			when:  &InterceptWhen{Port: &RuleStringMatch{Equals: "8080"}},
			input: RequestMatchInput{Method: "GET", Port: "9090"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.when.MatchesRequest(tt.input)
			if got != tt.want {
				t.Errorf("MatchesRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInterceptWhen_MatchesResponse(t *testing.T) {
	tests := []struct {
		name  string
		when  *InterceptWhen
		input ResponseMatchInput
		want  bool
	}{
		{
			name:  "nil when - always true",
			when:  nil,
			input: ResponseMatchInput{StatusCode: 200},
			want:  true,
		},
		{
			name:  "status match - success",
			when:  &InterceptWhen{Status: &RuleStatusMatch{Is4xx: true}},
			input: ResponseMatchInput{StatusCode: 404},
			want:  true,
		},
		{
			name:  "status match - fail",
			when:  &InterceptWhen{Status: &RuleStatusMatch{Is4xx: true}},
			input: ResponseMatchInput{StatusCode: 200},
			want:  false,
		},
		{
			name:  "content-type match - success",
			when:  &InterceptWhen{ContentType: &RuleStringMatch{Contains: "json"}},
			input: ResponseMatchInput{StatusCode: 200, ContentType: "application/json"},
			want:  true,
		},
		{
			name:  "content-type match - fail",
			when:  &InterceptWhen{ContentType: &RuleStringMatch{Contains: "json"}},
			input: ResponseMatchInput{StatusCode: 200, ContentType: "text/html"},
			want:  false,
		},
		{
			name: "header match - success",
			when: &InterceptWhen{
				Header: &RuleHeaderMatch{
					Name:  RuleStringMatch{Equals: "X-Custom"},
					Value: RuleStringMatch{Equals: "value"},
				},
			},
			input: ResponseMatchInput{
				StatusCode: 200,
				Headers:    map[string][]string{"X-Custom": {"value"}},
			},
			want: true,
		},
		{
			name:  "body contains - success",
			when:  &InterceptWhen{BodyContains: "error"},
			input: ResponseMatchInput{StatusCode: 500, BodyPreview: "internal error occurred"},
			want:  true,
		},
		{
			name:  "body contains - fail",
			when:  &InterceptWhen{BodyContains: "error"},
			input: ResponseMatchInput{StatusCode: 200, BodyPreview: "success message"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.when.MatchesResponse(tt.input)
			if got != tt.want {
				t.Errorf("MatchesResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
