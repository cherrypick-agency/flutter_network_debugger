package httpapi

import "network-debugger/internal/domain"

type sessionMonitorAdapter struct {
	monitor  *MonitorHub
	realtime *sessionRealtimeService
}

func newSessionMonitorAdapter(monitor *MonitorHub, realtime *sessionRealtimeService) *sessionMonitorAdapter {
	return &sessionMonitorAdapter{
		monitor:  monitor,
		realtime: realtime,
	}
}

func (a *sessionMonitorAdapter) bind() {
	if a == nil || a.monitor == nil || a.realtime == nil {
		return
	}
	go a.run()
}

func (a *sessionMonitorAdapter) run() {
	sub := a.monitor.Subscribe()
	defer a.monitor.Unsubscribe(sub)

	for ev := range sub {
		a.handleMonitorEvent(ev)
	}
}

func (a *sessionMonitorAdapter) handleMonitorEvent(ev domain.MonitorEvent) {
	a.realtime.handleMonitorEvent(ev)
}
