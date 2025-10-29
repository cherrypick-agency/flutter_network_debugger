package httpapi

import (
    "bufio"
    "bytes"
    "testing"
    "network-debugger/internal/domain"
    cfgpkg "network-debugger/internal/infrastructure/config"
)

// helper: build single-frame WS text with optional mask
func wsTextFrame(payload []byte, masked bool) []byte {
    b := &bytes.Buffer{}
    // FIN=1, RSV1-3=0, opcode=0x1 (text)
    b.WriteByte(0x80 | 0x1)
    var maskKey [4]byte
    maskKey = [4]byte{0x11, 0x22, 0x33, 0x44}
    if masked {
        // set mask bit and len
        b.WriteByte(0x80 | byte(len(payload)))
        b.Write(maskKey[:])
        enc := make([]byte, len(payload))
        for i := range payload { enc[i] = payload[i] ^ maskKey[i&3] }
        b.Write(enc)
    } else {
        b.WriteByte(byte(len(payload)))
        b.Write(payload)
    }
    return b.Bytes()
}

func TestForwardOneWSFrame_TextMaskedPreview(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 1024}}
    data := wsTextFrame([]byte("Hi"), true)
    r := bufio.NewReader(bytes.NewReader(data))
    var w bytes.Buffer
    op, size, prev, err := d.forwardOneWSFrame(r, &w)
    if err != nil { t.Fatalf("err: %v", err) }
    if op != domain.OpcodeText || size != 2 || prev != "Hi" { t.Fatalf("unexpected: %v %d %q", op, size, prev) }
    if !bytes.Equal(w.Bytes(), data) { t.Fatalf("must forward raw bytes unchanged") }
}

func TestForwardOneWSMessage_SingleFrame(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{WSPreviewMaxBytes: 1024}}
    data := wsTextFrame([]byte("hello"), false)
    r := bufio.NewReader(bytes.NewReader(data))
    var w bytes.Buffer
    op, size, preview, raw, file, err := d.forwardOneWSMessage(r, &w)
    if err != nil { t.Fatalf("err: %v", err) }
    if op != domain.OpcodeText || size != 5 || preview != "hello" || raw != "hello" || file != "" {
        t.Fatalf("unexpected: %v %d %q %q file=%q", op, size, preview, raw, file)
    }
}

func TestForwardOneWSMessage_Continuation(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{WSPreviewMaxBytes: 1024}}
    // first frame: FIN=0 text, payload "hel"
    b := &bytes.Buffer{}
    b.WriteByte(0x1)              // FIN=0 opcode=1
    b.WriteByte(byte(3))          // no mask, len=3
    b.WriteString("hel")
    // continuation FIN=1, opcode=0
    b.WriteByte(0x80 | 0x0)       // FIN=1, continuation
    b.WriteByte(byte(2))
    b.WriteString("lo")
    r := bufio.NewReader(bytes.NewReader(b.Bytes()))
    var w bytes.Buffer
    op, size, preview, raw, _, err := d.forwardOneWSMessage(r, &w)
    if err != nil { t.Fatalf("err: %v", err) }
    if op != domain.OpcodeText || size != 5 || preview != "hello" || raw != "hello" {
        t.Fatalf("unexpected continuation result: %v %d %q %q", op, size, preview, raw)
    }
}

func TestForwardOneWSMessage_SingleMaskedFrame(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{WSPreviewMaxBytes: 1024}}
    data := wsTextFrame([]byte("world"), true)
    r := bufio.NewReader(bytes.NewReader(data))
    var w bytes.Buffer
    op, size, preview, raw, _, err := d.forwardOneWSMessage(r, &w)
    if err != nil { t.Fatalf("err: %v", err) }
    if op != domain.OpcodeText || size != 5 || preview != "world" || raw != "world" {
        t.Fatalf("unexpected masked: %v %d %q %q", op, size, preview, raw)
    }
}

func TestForwardOneWSFrame_ExtendedLen126(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 1024}}
    payload := bytes.Repeat([]byte{'a'}, 128)
    // FIN=1, opcode text
    var buf bytes.Buffer
    buf.WriteByte(0x80 | 0x1)
    // no mask, len=126 indicator then 2-byte length
    buf.WriteByte(126)
    buf.WriteByte(0)
    buf.WriteByte(128)
    buf.Write(payload)
    r := bufio.NewReader(bytes.NewReader(buf.Bytes()))
    var out bytes.Buffer
    op, size, prev, err := d.forwardOneWSFrame(r, &out)
    if err != nil { t.Fatalf("err: %v", err) }
    if op != domain.OpcodeText || size != 128 || len(prev) != 128 { t.Fatalf("extended len failed: %d %d", size, len(prev)) }
}

func TestForwardOneWSFrame_ExtendedLen127(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 1024}}
    n := 300
    payload := bytes.Repeat([]byte{'b'}, n)
    var buf bytes.Buffer
    buf.WriteByte(0x80 | 0x1)
    buf.WriteByte(127)
    // 8-byte length
    var ln [8]byte
    // big-endian 300
    ln[7] = byte(n)
    ln[6] = byte(n >> 8)
    buf.Write(ln[:])
    buf.Write(payload)
    r := bufio.NewReader(bytes.NewReader(buf.Bytes()))
    var out bytes.Buffer
    op, size, prev, err := d.forwardOneWSFrame(r, &out)
    if err != nil { t.Fatalf("err: %v", err) }
    if op != domain.OpcodeText || size != n || len(prev) != n { t.Fatalf("extended len127 failed: %d %d", size, len(prev)) }
}

func TestForwardOneWSMessage_CaptureBodiesToFile(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{WSPreviewMaxBytes: 1024, WSCaptureBodies: true, WSBodyMaxBytes: 1024}}
    data := wsTextFrame([]byte("abc"), false)
    r := bufio.NewReader(bytes.NewReader(data))
    var w bytes.Buffer
    _, _, _, _, file, err := d.forwardOneWSMessage(r, &w)
    if err != nil { t.Fatalf("err: %v", err) }
    if file == "" { t.Fatalf("expected body file path") }
}


