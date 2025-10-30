package httpapi

import (
	"bytes"
	"compress/flate"
	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"os"
	"strings"
	"testing"
)

func TestBuildWSPreview_BinaryAndText(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: false, WSPreviewMaxBytes: 16, PreviewMaxBytes: 16}}
	p, raw := d.buildWSPreview(domain.OpcodeBinary, false, []byte{0x01, 0x02}, nil)
	if p != "0102" || raw != "" {
		t.Fatalf("binary preview: %q raw:%q", p, raw)
	}
	p, raw = d.buildWSPreview(domain.OpcodeText, false, []byte("txt"), nil)
	if p != "txt" || raw != "txt" {
		t.Fatalf("text preview")
	}
}

func TestBuildWSPreview_TextDeflatePreview(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: true, WSPreviewMaxBytes: 64, PreviewMaxBytes: 64}}
	// raw DEFLATE-compressed "hello"
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	_, _ = zw.Write([]byte("hello"))
	_ = zw.Close()
	p, raw := d.buildWSPreview(domain.OpcodeText, true, nil, buf.Bytes())
	if p != "hello" || raw != "hello" {
		t.Fatalf("deflate preview failed: %q %q", p, raw)
	}
}

func TestOpenWSFile_CreatesTemp(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{BodySpoolDir: ""}}
	f, path := d.openWSFile()
	if f == nil || path == "" {
		t.Fatalf("openWSFile")
	}
	_ = f.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing: %v", err)
	}
	_ = os.Remove(path)
}

func TestOpenWSFile_CustomSpoolDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-ws-spool-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	d := &Deps{Cfg: cfgpkg.Config{BodySpoolDir: tempDir}}
	f, path := d.openWSFile()
	if f == nil || path == "" {
		t.Fatalf("openWSFile with custom dir failed")
	}
	_ = f.Close()

	// Verify file is in custom directory
	if !strings.Contains(path, tempDir) {
		t.Errorf("path %q should contain %q", path, tempDir)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file missing: %v", err)
	}

	_ = os.Remove(path)
}

func TestBuildWSPreview_TextWithRSV1_DeflateDisabled(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: false, WSPreviewMaxBytes: 64}}
	// RSV1 is set, but deflate preview is disabled - should use previewBuf
	p, raw := d.buildWSPreview(domain.OpcodeText, true, []byte("plain"), []byte("compressed"))
	if p != "plain" || raw != "plain" {
		t.Errorf("with deflate disabled, should use previewBuf: p=%q raw=%q", p, raw)
	}
}

func TestBuildWSPreview_TextWithRSV1_NoCompressedBuf(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: true, WSPreviewMaxBytes: 64}}
	// RSV1 is set but compressedBuf is empty - should use previewBuf
	p, raw := d.buildWSPreview(domain.OpcodeText, true, []byte("fallback"), nil)
	if p != "fallback" || raw != "fallback" {
		t.Errorf("with no compressedBuf, should use previewBuf: p=%q raw=%q", p, raw)
	}
}

func TestBuildWSPreview_TextWithRSV1_InvalidCompressed(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: true, WSPreviewMaxBytes: 64}}
	// Invalid compressed data - decompression will fail
	invalidCompressed := []byte{0xFF, 0xFF, 0xFF}
	p, raw := d.buildWSPreview(domain.OpcodeText, true, []byte("backup"), invalidCompressed)
	// When decompression fails, function may return empty or partial data
	// The important thing is it doesn't panic
	_ = p
	_ = raw
	// Test passes if no panic occurs
}

func TestBuildWSPreview_TextDeflate_ZeroWSPreviewMaxBytes(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: true, WSPreviewMaxBytes: 0, PreviewMaxBytes: 32}}
	// WSPreviewMaxBytes is 0, should fallback to PreviewMaxBytes
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	_, _ = zw.Write([]byte("test with fallback limit"))
	_ = zw.Close()
	p, _ := d.buildWSPreview(domain.OpcodeText, true, nil, buf.Bytes())
	if len(p) > 32 {
		t.Errorf("should respect PreviewMaxBytes fallback: got len=%d", len(p))
	}
	if p != "test with fallback limit" {
		t.Errorf("decompressed text: %q", p)
	}
}

func TestBuildWSPreview_TextDeflate_LargeData(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{WSDeflatePreview: true, WSPreviewMaxBytes: 10}}
	// Large compressed data should be limited
	var buf bytes.Buffer
	zw, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	_, _ = zw.Write([]byte("this is a very long message that exceeds the preview limit"))
	_ = zw.Close()
	p, _ := d.buildWSPreview(domain.OpcodeText, true, nil, buf.Bytes())
	if len(p) > 10 {
		t.Errorf("preview should be limited to 10 bytes: got len=%d", len(p))
	}
	if p != "this is a " {
		t.Errorf("preview: %q", p)
	}
}

func TestBuildWSPreview_BinaryOpcode(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	// Binary opcode should return hex preview
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	p, raw := d.buildWSPreview(domain.OpcodeBinary, false, data, nil)
	if p != "deadbeef" {
		t.Errorf("binary preview: %q", p)
	}
	if raw != "" {
		t.Errorf("binary raw should be empty: %q", raw)
	}
}

func TestBuildWSPreview_EmptyData(t *testing.T) {
	d := &Deps{Cfg: cfgpkg.Config{}}
	// Empty data
	p, raw := d.buildWSPreview(domain.OpcodeText, false, []byte{}, nil)
	if p != "" || raw != "" {
		t.Errorf("empty text: p=%q raw=%q", p, raw)
	}
}
