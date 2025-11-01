package httpapi

import (
	"network-debugger/internal/domain"
	"testing"
)

func TestOpcodeFromWS(t *testing.T) {
	tests := []struct {
		code byte
		want domain.Opcode
	}{
		{0x1, domain.OpcodeText},
		{0x2, domain.OpcodeBinary},
		{0x8, domain.OpcodeClose},
		{0x9, domain.OpcodePing},
		{0xA, domain.OpcodePong},
		{0x0, domain.OpcodeBinary},  // continuation frame defaults to binary
		{0xFF, domain.OpcodeBinary}, // unknown defaults to binary
	}

	for _, tt := range tests {
		got := opcodeFromWS(tt.code)
		if got != tt.want {
			t.Errorf("opcodeFromWS(0x%X) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestHexPreview(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{"empty", []byte{}, ""},
		{"single byte", []byte{0xFF}, "ff"},
		{"multiple bytes", []byte{0x01, 0x02, 0x03}, "010203"},
		{"text data", []byte("hello"), "68656c6c6f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hexPreview(tt.input)
			if got != tt.want {
				t.Errorf("hexPreview(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHexPreview_LargeInput(t *testing.T) {
	// Test that hexPreview limits output to 256 bytes
	input := make([]byte, 300)
	for i := range input {
		input[i] = byte(i % 256)
	}

	got := hexPreview(input)
	// 256 bytes * 2 hex chars = 512 chars
	if len(got) != 512 {
		t.Errorf("hexPreview with 300 bytes input: got length %d, want 512", len(got))
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-1, 5, -1},
		{-10, -5, -10},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
