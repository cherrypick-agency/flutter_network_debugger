package httpapi

import (
    "testing"
    "network-debugger/internal/domain"
    cfgpkg "network-debugger/internal/infrastructure/config"
    "compress/flate"
    "bytes"
    "os"
)

func TestBuildWSPreview_BinaryAndText(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: false, WSPreviewMaxBytes: 16, PreviewMaxBytes: 16}}
    p, raw := d.buildWSPreview(domain.OpcodeBinary, false, []byte{0x01, 0x02}, nil)
    if p != "0102" || raw != "" { t.Fatalf("binary preview: %q raw:%q", p, raw) }
    p, raw = d.buildWSPreview(domain.OpcodeText, false, []byte("txt"), nil)
    if p != "txt" || raw != "txt" { t.Fatalf("text preview") }
}

func TestBuildWSPreview_TextDeflatePreview(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: true, WSPreviewMaxBytes: 64, PreviewMaxBytes: 64}}
    // raw DEFLATE-compressed "hello"
    var buf bytes.Buffer
    zw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
    _, _ = zw.Write([]byte("hello"))
    _ = zw.Close()
    p, raw := d.buildWSPreview(domain.OpcodeText, true, nil, buf.Bytes())
    if p != "hello" || raw != "hello" { t.Fatalf("deflate preview failed: %q %q", p, raw) }
}

func TestOpenWSFile_CreatesTemp(t *testing.T) {
    d := &Deps{Cfg: cfgpkg.Config{BodySpoolDir: ""}}
    f, path := d.openWSFile()
    if f == nil || path == "" { t.Fatalf("openWSFile") }
    _ = f.Close()
    if _, err := os.Stat(path); err != nil { t.Fatalf("file missing: %v", err) }
    _ = os.Remove(path)
}


