package httpapi

import (
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"testing"
)

func TestBuildBaseMux_AppliesPreviewFlags(t *testing.T) {
	oldMax, oldExpose, oldDecomp := previewMaxBytes, exposeSensitiveHeaders, previewDecompress
	defer func() { previewMaxBytes, exposeSensitiveHeaders, previewDecompress = oldMax, oldExpose, oldDecomp }()
	d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 777, ExposeSensitiveHeaders: false, PreviewDecompress: false}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions()}
	_ = buildBaseMux(d)
	if previewMaxBytes != 777 || exposeSensitiveHeaders != false || previewDecompress != false {
		t.Fatalf("flags not applied: %d %v %v", previewMaxBytes, exposeSensitiveHeaders, previewDecompress)
	}
}

func TestNewRouter_ReturnsHandler(t *testing.T) {
	cfg := cfgpkg.Config{
		PreviewMaxBytes:        1024,
		ExposeSensitiveHeaders: false,
		PreviewDecompress:      false,
	}
	metrics := obs.NewMetrics()

	handler := NewRouter(cfg, nil, metrics)

	if handler == nil {
		t.Fatal("NewRouter returned nil handler")
	}
}
