package httpapi

import (
	"net/textproto"
)

// Simple in-memory implementation of HeaderLike for unit tests
type headerFake struct {
	m map[string][]string
}

func newHeaderFake() *headerFake {
	return &headerFake{m: map[string][]string{}}
}

func (h *headerFake) key(k string) string {
	return textproto.CanonicalMIMEHeaderKey(k)
}

func (h *headerFake) Values(key string) []string {
	k := h.key(key)
	v := h.m[k]
	// copy to prevent tests from accidentally mutating internal state
	out := make([]string, len(v))
	copy(out, v)
	return out
}

func (h *headerFake) Set(key, value string) {
	k := h.key(key)
	h.m[k] = []string{value}
}

func (h *headerFake) Add(key, value string) {
	k := h.key(key)
	h.m[k] = append(h.m[k], value)
}

func (h *headerFake) Del(key string) {
	k := h.key(key)
	delete(h.m, k)
}
