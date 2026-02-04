package presentation

import (
	"io"
	"network-debugger/internal/features/sessions_cli/domain"
	"network-debugger/internal/features/sessions_cli/presentation/sections"
	"network-debugger/internal/features/sessions_cli/presentation/theme"
)

// Renderer prints representation by selected sections.
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
	// For now we only print the line. Other sections will be added in subsequent steps.
	if opts.Fields["line"] {
		if err := r.line.Render(w, v); err != nil {
			return err
		}
	}
	return nil
}
