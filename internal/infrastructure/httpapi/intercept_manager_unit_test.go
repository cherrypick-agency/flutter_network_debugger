package httpapi

import (
	"testing"
)

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{
			name: "localhost IPv4",
			addr: "127.0.0.1",
			want: true,
		},
		{
			name: "localhost IPv4 with port",
			addr: "127.0.0.1:8080",
			want: true,
		},
		{
			name: "localhost IPv6",
			addr: "::1",
			want: true,
		},
		{
			name: "localhost IPv6 with port",
			addr: "[::1]:8080",
			want: true,
		},
		{
			name: "localhost IPv6 with brackets - not a valid IP format without port",
			addr: "[::1]",
			want: false, // net.ParseIP can't parse "[::1]" format
		},
		{
			name: "unspecified IPv4",
			addr: "0.0.0.0",
			want: true,
		},
		{
			name: "unspecified IPv6",
			addr: "::",
			want: true,
		},
		{
			name: "127.0.0.2 is loopback",
			addr: "127.0.0.2",
			want: true,
		},
		{
			name: "127.255.255.255 is loopback",
			addr: "127.255.255.255",
			want: true,
		},
		{
			name: "public IPv4",
			addr: "8.8.8.8",
			want: false,
		},
		{
			name: "public IPv4 with port",
			addr: "8.8.8.8:443",
			want: false,
		},
		{
			name: "private network IPv4",
			addr: "192.168.1.1",
			want: false,
		},
		{
			name: "private network IPv4 with port",
			addr: "192.168.1.1:8080",
			want: false,
		},
		{
			name: "public IPv6",
			addr: "2001:4860:4860::8888",
			want: false,
		},
		{
			name: "public IPv6 with port",
			addr: "[2001:4860:4860::8888]:443",
			want: false,
		},
		{
			name: "invalid IP - hostname",
			addr: "example.com",
			want: false,
		},
		{
			name: "invalid IP - hostname with port",
			addr: "example.com:8080",
			want: false,
		},
		{
			name: "empty string",
			addr: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLoopback(tt.addr)
			if got != tt.want {
				t.Errorf("isLoopback(%q) = %v, want %v", tt.addr, got, tt.want)
			}
		})
	}
}
