package httpapi

import "testing"

func TestAddrPort(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want int
	}{
		{
			name: "standard port format",
			addr: ":8888",
			want: 8888,
		},
		{
			name: "localhost with port",
			addr: "localhost:8888",
			want: 8888,
		},
		{
			name: "IP with port",
			addr: "127.0.0.1:8888",
			want: 8888,
		},
		{
			name: "IPv6 with port",
			addr: "[::1]:8888",
			want: 8888,
		},
		{
			name: "no colon",
			addr: "localhost",
			want: 0,
		},
		{
			name: "colon at end",
			addr: "localhost:",
			want: 0,
		},
		{
			name: "empty string",
			addr: "",
			want: 0,
		},
		{
			name: "high port number",
			addr: ":65535",
			want: 65535,
		},
		{
			name: "port 80",
			addr: "example.com:80",
			want: 80,
		},
		{
			name: "port 443",
			addr: "example.com:443",
			want: 443,
		},
		{
			name: "invalid port",
			addr: ":abc",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addrPort(tt.addr)
			if got != tt.want {
				t.Errorf("addrPort(%q) = %d, want %d", tt.addr, got, tt.want)
			}
		})
	}
}
