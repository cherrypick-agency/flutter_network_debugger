package theme

import (
	"github.com/mattn/go-isatty"
	"io"
	"network-debugger/internal/features/sessions_cli/domain"
	"os"
	"runtime"
	"strings"
)

// Colorizer decides whether to print ANSI and selects colors for elements.
type Colorizer struct {
	mode domain.ColorMode
	use  bool
}

func NewColorizer(mode domain.ColorMode, out io.Writer) *Colorizer {
	use := false
	switch mode {
	case domain.ColorAlways:
		use = true
	case domain.ColorNever:
		use = false
	default:
		// auto: if stdout is a TTY
		if f, ok := out.(*os.File); ok {
			use = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
		}
	}
	// windows: modern terminals support ANSI, no fallback needed
	_ = runtime.GOOS
	return &Colorizer{mode: mode, use: use}
}

func (c *Colorizer) wrap(code string, s string) string {
	if !c.use {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// Public helpers for key elements
func (c *Colorizer) Method(method string) string {
	up := strings.ToUpper(method)
	switch up {
	case "GET":
		return c.wrap("32", up) // green
	case "POST":
		return c.wrap("36", up) // cyan
	case "PUT":
		return c.wrap("33", up) // yellow
	case "DELETE":
		return c.wrap("31", up) // red
	case "PATCH":
		return c.wrap("35", up) // magenta
	default:
		return c.wrap("37", up) // gray
	}
}

func (c *Colorizer) Status(code int, s string) string {
	switch {
	case code >= 200 && code < 300:
		return c.wrap("32", s)
	case code >= 300 && code < 400:
		return c.wrap("34", s)
	case code >= 400 && code < 500:
		return c.wrap("33", s)
	default:
		return c.wrap("31", s)
	}
}

func (c *Colorizer) Dim(s string) string { return c.wrap("90", s) }
