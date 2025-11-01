package httpapi

import (
	"net/url"
	"testing"
)

func TestUrlString(t *testing.T) {
	tests := []struct {
		name string
		u    *url.URL
		want string
	}{
		{
			name: "nil url",
			u:    nil,
			want: "",
		},
		{
			name: "basic http url",
			u: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/test",
			},
			want: "http://example.com/test",
		},
		{
			name: "https url with query",
			u: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/users",
				RawQuery: "page=1&limit=10",
			},
			want: "https://api.example.com/v1/users?page=1&limit=10",
		},
		{
			name: "url with port",
			u: &url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
				Path:   "/api",
			},
			want: "http://localhost:8080/api",
		},
		{
			name: "url with fragment",
			u: &url.URL{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/page",
				Fragment: "section",
			},
			want: "https://example.com/page#section",
		},
		{
			name: "empty path",
			u: &url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
			want: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := urlString(tt.u)
			if got != tt.want {
				t.Errorf("urlString() = %q, want %q", got, tt.want)
			}
		})
	}
}
