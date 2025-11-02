package presentation

import (
	"io"
	"network-debugger/internal/features/sessions_cli/domain"
	"network-debugger/internal/features/sessions_cli/presentation/sections"
	"network-debugger/internal/features/sessions_cli/presentation/theme"
)

// Renderer печатает представление по выбранным секциям.
type Renderer interface {
	RenderHTTP(w io.Writer, v domain.HTTPView, opts domain.Options) error
}

type defaultRenderer struct {
	line *sections.LineSection
}

func NewDefaultRenderer(opts domain.Options, out io.Writer) Renderer {
	colors := theme.NewColorizer(opts.Color, out)
	return &defaultRenderer{line: sections.NewLineSection(colors)}
}

func (r *defaultRenderer) RenderHTTP(w io.Writer, v domain.HTTPView, opts domain.Options) error {
	// Пока печатаем только строку. Остальные секции добавим в следующих шагах.
	if opts.Fields["line"] {
		if err := r.line.Render(w, v); err != nil {
			return err
		}
	}
	return nil
}
