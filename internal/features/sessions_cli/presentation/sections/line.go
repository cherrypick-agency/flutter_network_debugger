package sections

import (
	"fmt"
	"io"
	"network-debugger/internal/features/sessions_cli/domain"
	"network-debugger/internal/features/sessions_cli/presentation/theme"
)

// LineSection prints a brief line with main information.
type LineSection struct {
	Colors *theme.Colorizer
}

func NewLineSection(c *theme.Colorizer) *LineSection { return &LineSection{Colors: c} }

func humanBytes(n int) string {
	if n < 1024 {
		return fmt.Sprintf("%dB", n)
	}
	kb := float64(n) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1fKB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.2fMB", mb)
}

func (s *LineSection) Render(w io.Writer, v domain.HTTPView) error {
	ts := v.StartedAt.Local().Format("15:04:05")
	method := s.Colors.Method(v.Method)
	status := s.Colors.Status(v.Status, fmt.Sprintf("%d", v.Status))
	total := s.Colors.Dim(fmt.Sprintf("%dms", v.TotalMs))
	sizes := s.Colors.Dim(fmt.Sprintf("%s ⇄ %s", humanBytes(v.ReqSize), humanBytes(v.RespSize)))
	_, err := fmt.Fprintf(w, "[%s] %s %s  →  %s  (%s, %s)\n", ts, method, v.URL, status, total, sizes)
	return err
}
