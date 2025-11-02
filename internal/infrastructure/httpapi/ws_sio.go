package httpapi

import (
	"time"

	sio "network-debugger/internal/adapters/decoders/socketio"
	"network-debugger/internal/domain"
	"network-debugger/pkg/shared/id"
)

// recordSIOIfAny пытается распарсить Socket.IO событие из сырого текста raw
// и, если получилось, записывает событие в стор и шлёт уведомление в монитор.
// Возвращает true, если событие распознано и записано.
func (d *Deps) recordSIOIfAny(sessionID string, raw string, frameID string) bool {
	if raw == "" {
		return false
	}
	if nsp, ev, argsJSON, ok := sio.ParseEvent(raw); ok {
		var ack *int64
		if a := tryExtractAckID(raw); a >= 0 {
			ack = &a
		}
		e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: nsp, Name: ev, AckID: ack, ArgsPreview: argsJSON, FrameIDs: []string{frameID}}
		_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
		return true
	}
	// Фолбэки для распространённых форм
	if len(raw) >= 2 && raw[0] == '4' && raw[1] == '3' {
		if a := tryExtractAckID(raw); a >= 0 {
			aa := a
			e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: "", Name: "ack", AckID: &aa, ArgsPreview: "[]", FrameIDs: []string{frameID}}
			_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
			return true
		}
	}
	if len(raw) >= 2 && raw[0] == '4' && raw[1] == '2' {
		if nsp, ev, argsJSON, ok2 := sio.ParseEvent(raw); ok2 {
			var ack *int64
			if a := tryExtractAckID(raw); a >= 0 {
				ack = &a
			}
			e := domain.Event{ID: id.New(), Ts: time.Now().UTC(), Namespace: nsp, Name: ev, AckID: ack, ArgsPreview: argsJSON, FrameIDs: []string{frameID}}
			_ = d.Svc.AddEvent(contextWithNoCancel(), sessionID, e)
			d.broadcastMonitorEvent(domain.MonitorEvent{Type: "event_added", ID: sessionID, Ref: e.ID})
			return true
		}
	}
	return false
}
