package socketio

import proxykit "github.com/777genius/proxykit/socketio"

var ParseEvent = proxykit.ParseEvent

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
