package httpapi

import (
	"network-debugger/internal/domain"
	"testing"

	"github.com/gorilla/websocket"
)

func TestOpcodeFromType_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		mt       int
		expected domain.Opcode
	}{
		{"text message", websocket.TextMessage, domain.OpcodeText},
		{"binary message", websocket.BinaryMessage, domain.OpcodeBinary},
		{"ping message", websocket.PingMessage, domain.OpcodePing},
		{"pong message", websocket.PongMessage, domain.OpcodePong},
		{"close message", websocket.CloseMessage, domain.OpcodeClose},
		{"unknown type", 99, domain.OpcodeBinary},  // default case
		{"negative type", -1, domain.OpcodeBinary}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := opcodeFromType(tt.mt)
			if result != tt.expected {
				t.Errorf("opcodeFromType(%d) = %v, want %v", tt.mt, result, tt.expected)
			}
		})
	}
}

func TestOpcodeFromType_BoundaryValues(t *testing.T) {
	// Test with very large value
	result := opcodeFromType(10000)
	if result != domain.OpcodeBinary {
		t.Errorf("opcodeFromType with large value should return Binary, got %v", result)
	}

	// Test with zero
	result = opcodeFromType(0)
	if result != domain.OpcodeBinary {
		t.Errorf("opcodeFromType(0) should return Binary, got %v", result)
	}
}
