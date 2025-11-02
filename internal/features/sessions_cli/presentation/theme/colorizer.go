package theme

import (
	"github.com/mattn/go-isatty"
	"io"
	"network-debugger/internal/features/sessions_cli/domain"
	"os"
	"runtime"
	"strings"
)

// Colorizer решает, печатать ли ANSI и подбирает цвета для элементов.
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
		// auto: если stdout — TTY
		if f, ok := out.(*os.File); ok {
			use = isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
		}
	}
	// windows: современные терминалы поддерживают ANSI, fallback не нужен
	_ = runtime.GOOS
	return &Colorizer{mode: mode, use: use}
}

func (c *Colorizer) wrap(code string, s string) string {
	if !c.use {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// Публичные helpers для ключевых элементов
func (c *Colorizer) Method(method string) string {
	up := strings.ToUpper(method)
	switch up {
	case "GET":
		return c.wrap("32", up) // зелёный
	case "POST":
		return c.wrap("36", up) // циан
	case "PUT":
		return c.wrap("33", up) // жёлтый
	case "DELETE":
		return c.wrap("31", up) // красный
	case "PATCH":
		return c.wrap("35", up) // магента
	default:
		return c.wrap("37", up) // серый
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
