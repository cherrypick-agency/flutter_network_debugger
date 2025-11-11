package httpapi

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"network-debugger/internal/infrastructure/config"
)

func TestSpoolBody_ZeroMaxBytes(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	reader := bytes.NewReader([]byte("test data"))
	path, err := d.spoolBody(reader, 0, "req")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path even with zero max bytes")
	}

	if path != "" {
		os.Remove(path)
	}
}

func TestSpoolBody_NegativeMaxBytes(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	reader := bytes.NewReader([]byte("test data"))
	path, err := d.spoolBody(reader, -1, "req")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if path != "" {
		os.Remove(path)
	}
}

func TestSpoolBody_EmptyReader(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	reader := bytes.NewReader([]byte{})
	path, err := d.spoolBody(reader, 100, "resp")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path even for empty reader")
	}

	if path != "" {
		os.Remove(path)
	}
}

func TestSpoolBody_ReadError(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	reader := &errorReader{err: io.ErrUnexpectedEOF}
	path, err := d.spoolBody(reader, 100, "req")

	if err == nil {
		t.Error("expected error for read failure")
	}

	if path != "" {
		os.Remove(path)
	}
}

func TestSpoolBody_InvalidDirectory(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: "/nonexistent/path/that/does/not/exist",
		},
	}

	reader := bytes.NewReader([]byte("test"))
	path, err := d.spoolBody(reader, 100, "req")

	if err == nil {
		t.Error("expected error for invalid directory")
	}

	if path != "" {
		os.Remove(path)
	}
}

func TestSpoolBody_VeryLargeMaxBytes(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	smallData := []byte("small data")
	reader := bytes.NewReader(smallData)
	path, err := d.spoolBody(reader, 1<<62, "req")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if path == "" {
		t.Error("expected non-empty path")
	}

	content, _ := os.ReadFile(path)
	if len(content) != len(smallData) {
		t.Errorf("expected %d bytes, got %d", len(smallData), len(content))
	}

	if path != "" {
		os.Remove(path)
	}
}

func TestSpoolBody_DifferentKinds(t *testing.T) {
	kinds := []string{"req", "resp", "ws", "map", "test"}
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			reader := bytes.NewReader([]byte("test data"))
			path, err := d.spoolBody(reader, 100, kind)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if path == "" {
				t.Error("expected non-empty path")
			}

			baseName := filepath.Base(path)
			if len(baseName) < len("gpx-")+len(kind) || baseName[:len("gpx-")+len(kind)] != "gpx-"+kind {
				t.Errorf("expected filename to contain kind %s, got %s", kind, baseName)
			}

			if path != "" {
				os.Remove(path)
			}
		})
	}
}

func TestSpoolBody_ConcurrentWrites(t *testing.T) {
	d := &Deps{
		Cfg: config.Config{
			BodySpoolDir: t.TempDir(),
		},
	}

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			reader := bytes.NewReader([]byte("test data"))
			path, err := d.spoolBody(reader, 100, "req")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if path != "" {
				os.Remove(path)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}
