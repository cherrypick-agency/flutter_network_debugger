package httpapi

import (
    "testing"
    cfgpkg "network-debugger/internal/infrastructure/config"
    obs "network-debugger/internal/infrastructure/observability"
)

func TestBuildBaseMux_AppliesPreviewFlags(t *testing.T) {
    oldMax, oldExpose, oldDecomp := previewMaxBytes, exposeSensitiveHeaders, previewDecompress
    defer func(){ previewMaxBytes, exposeSensitiveHeaders, previewDecompress = oldMax, oldExpose, oldDecomp }()
    d := &Deps{Cfg: cfgpkg.Config{PreviewMaxBytes: 777, ExposeSensitiveHeaders: false, PreviewDecompress: false}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions()}
    _ = buildBaseMux(d)
    if previewMaxBytes != 777 || exposeSensitiveHeaders != false || previewDecompress != false {
        t.Fatalf("flags not applied: %d %v %v", previewMaxBytes, exposeSensitiveHeaders, previewDecompress)
    }
}


