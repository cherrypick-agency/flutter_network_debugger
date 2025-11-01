package httpapi

import (
	"github.com/gorilla/websocket"
	"network-debugger/internal/domain"
	"testing"
)

func TestBuildPreview_TextJSONRedaction_AndBinary(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(64)
	defer func() { previewMaxBytes.Store(old) }()
	// text json with sensitive keys
	p := buildPreview(domain.OpcodeText, []byte(`{"authorization":"Bearer abc","access_token":"x","k":1}`))
	if !strContains(p, `"authorization":"***"`) || !strContains(p, `"access_token":"***"`) {
		t.Fatalf("redaction failed: %s", p)
	}
	// binary
	b := buildPreview(domain.OpcodeBinary, []byte{0xAA, 0xBB})
	if b == "" {
		t.Fatalf("binary preview empty")
	}
}

func TestOpcodeFromType_Mapping(t *testing.T) {
	if opcodeFromType(websocket.TextMessage) != domain.OpcodeText {
		t.Fatalf("text mapping")
	}
	if opcodeFromType(websocket.CloseMessage) != domain.OpcodeClose {
		t.Fatalf("close mapping")
	}
}

func TestBuildPreview_TextNonJSON(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(64)
	defer func() { previewMaxBytes.Store(old) }()

	// Plain text (not JSON)
	p := buildPreview(domain.OpcodeText, []byte("Hello, World!"))
	if p != "Hello, World!" {
		t.Errorf("plain text preview = %s, want 'Hello, World!'", p)
	}
}

func TestBuildPreview_TextExceedsMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(10)
	defer func() { previewMaxBytes.Store(old) }()

	// Text longer than max
	data := []byte("This is a very long text that exceeds the maximum preview size")
	p := buildPreview(domain.OpcodeText, data)
	if len(p) > 10 {
		t.Errorf("preview length = %d, want <= 10", len(p))
	}
}

func TestBuildPreview_TextJSONExceedsMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(20)
	defer func() { previewMaxBytes.Store(old) }()

	// JSON that exceeds max after marshaling
	data := []byte(`{"key":"value","another":"data","more":"fields"}`)
	p := buildPreview(domain.OpcodeText, data)
	if len(p) > 20 {
		t.Errorf("preview length = %d, want <= 20", len(p))
	}
}

func TestBuildPreview_TextZeroMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(0)
	defer func() { previewMaxBytes.Store(old) }()

	// With zero max, should use full data length
	data := []byte("Test")
	p := buildPreview(domain.OpcodeText, data)
	if p != "Test" {
		t.Errorf("preview = %s, want 'Test'", p)
	}
}

func TestBuildPreview_TextNegativeMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(-1)
	defer func() { previewMaxBytes.Store(old) }()

	// With negative max, should use full data length
	data := []byte("Test")
	p := buildPreview(domain.OpcodeText, data)
	if p != "Test" {
		t.Errorf("preview = %s, want 'Test'", p)
	}
}

func TestBuildPreview_BinarySmall(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(64)
	defer func() { previewMaxBytes.Store(old) }()

	// Small binary data
	data := []byte{0x01, 0x02, 0x03}
	p := buildPreview(domain.OpcodeBinary, data)
	if p == "" {
		t.Error("binary preview should not be empty")
	}
}

func TestBuildPreview_BinaryZeroMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(0)
	defer func() { previewMaxBytes.Store(old) }()

	// With zero max, should use full data length
	data := []byte{0xAA, 0xBB, 0xCC}
	p := buildPreview(domain.OpcodeBinary, data)
	if p == "" {
		t.Error("binary preview should not be empty")
	}
}

func TestBuildPreview_BinaryExceedsMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(300) // > 256
	defer func() { previewMaxBytes.Store(old) }()

	// Large binary data - should cap at 256
	data := make([]byte, 300)
	for i := range data {
		data[i] = byte(i)
	}
	p := buildPreview(domain.OpcodeBinary, data)
	if p == "" {
		t.Error("binary preview should not be empty")
	}
	// The preview should be limited to 256 bytes of input data
}

func TestBuildPreview_BinaryLargerThanMax(t *testing.T) {
	old := previewMaxBytes.Load()
	previewMaxBytes.Store(10)
	defer func() { previewMaxBytes.Store(old) }()

	// Binary data larger than max
	data := make([]byte, 100)
	p := buildPreview(domain.OpcodeBinary, data)
	if p == "" {
		t.Error("binary preview should not be empty")
	}
}
