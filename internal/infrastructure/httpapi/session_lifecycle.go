package httpapi

import (
	"context"

	"network-debugger/internal/domain"
)

func (d *Deps) clearSessionsAndNotify(ctx context.Context) error {
	if err := d.Svc.ClearAll(ctx); err != nil {
		return err
	}
	if d.Live != nil {
		d.Live.CloseAll()
	}
	if d.Monitor != nil {
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "sessions_cleared", ID: "*"})
	}
	return nil
}
