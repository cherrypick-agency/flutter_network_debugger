package httpapi

import (
	"io"

	"github.com/rs/zerolog"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"testing"
)

func TestBuildBaseMux_AppliesPreviewFlags(t *testing.T) {
	oldMax, oldExpose, oldDecomp := previewMaxBytes.Load(), exposeSensitiveHeaders.Load(), previewDecompress.Load()
	defer func() {
		previewMaxBytes.Store(oldMax)
		exposeSensitiveHeaders.Store(oldExpose)
		previewDecompress.Store(oldDecomp)
	}()
	logger := zerolog.New(io.Discard)
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 777, ExposeSensitiveHeaders: false, PreviewDecompress: false}, Logger: &logger, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions()}
	_ = buildBaseMux(d)
	if previewMaxBytes.Load() != 777 || exposeSensitiveHeaders.Load() != false || previewDecompress.Load() != false {
		t.Fatalf("flags not applied: %d %v %v", previewMaxBytes.Load(), exposeSensitiveHeaders.Load(), previewDecompress.Load())
	}
}

func TestNewRouter_ReturnsHandler(t *testing.T) {
	cfg := cfgpkg.Config{
		PreviewMaxBytes:        1024,
		ExposeSensitiveHeaders: false,
		PreviewDecompress:      false,
	}
	metrics := obs.NewMetrics()
	logger := zerolog.New(io.Discard)

	handler := NewRouter(cfg, &logger, metrics)

	if handler == nil {
		t.Fatal("NewRouter returned nil handler")
	}
}
