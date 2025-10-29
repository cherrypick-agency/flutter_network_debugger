package httpapi

import (
    "testing"
    "network-debugger/internal/domain"
    "github.com/gorilla/websocket"
)

func TestBuildPreview_TextJSONRedaction_AndBinary(t *testing.T) {
    old := previewMaxBytes
    previewMaxBytes = 64
    defer func(){ previewMaxBytes = old }()
    // text json with sensitive keys
    p := buildPreview(domain.OpcodeText, []byte(`{"authorization":"Bearer abc","access_token":"x","k":1}`))
    if !strContains(p, `"authorization":"***"`) || !strContains(p, `"access_token":"***"`) {
        t.Fatalf("redaction failed: %s", p)
    }
    // binary
    b := buildPreview(domain.OpcodeBinary, []byte{0xAA, 0xBB})
    if b == "" { t.Fatalf("binary preview empty") }
}

func TestOpcodeFromType_Mapping(t *testing.T) {
    if opcodeFromType(websocket.TextMessage) != domain.OpcodeText { t.Fatalf("text mapping") }
    if opcodeFromType(websocket.CloseMessage) != domain.OpcodeClose { t.Fatalf("close mapping") }
}


