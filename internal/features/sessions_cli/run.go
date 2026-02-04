package sessions_cli

import (
	"context"
	"io"
	"network-debugger/internal/features/sessions_cli/adapters"
	"network-debugger/internal/features/sessions_cli/application"
	"network-debugger/internal/features/sessions_cli/domain"
	"network-debugger/internal/features/sessions_cli/presentation"
	httpapi "network-debugger/internal/infrastructure/httpapi"
)

// Run — feature entry point from main.
func Run(ctx context.Context, deps *httpapi.Deps, opts domain.Options, out io.Writer) error {
	opts.EnsureFieldsFromPreset()
	ev := adapters.NewMonitorEventStream(deps.Monitor)
	q := adapters.NewServiceQueries(deps.Svc)
	rend := presentation.NewDefaultRenderer(opts, out)
	app := &application.PrinterApp{
		Events: ev,
		Sess:   q,
		Tx:     q,
		Frames: q,
		Rend:   rend,
	}
	return app.Run(ctx, out, opts)
}
