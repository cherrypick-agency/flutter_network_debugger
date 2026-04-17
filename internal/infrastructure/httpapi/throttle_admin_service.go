package httpapi

import (
	"context"
	"net/http"

	sdomain "network-debugger/internal/features/settings/domain"
)

type throttleAdminService struct {
	d *Deps
}

func newThrottleAdminService(d *Deps) throttleAdminService {
	return throttleAdminService{d: d}
}

func (s throttleAdminService) load(ctx context.Context) throttleDTO {
	out := s.runtimeDTO()
	if s.d.Settings != nil {
		if rs, err := s.d.Settings.Load(ctx); err == nil {
			out.Enabled = rs.ThrottleEnabled
			out.DownKbps = rs.ThrottleDownKbps
			out.UpKbps = rs.ThrottleUpKbps
			out.PacketLossPct = rs.ThrottlePacketLoss
			out.LatencyMs = rs.ThrottleLatencyMs
			out.LatencyJitter = rs.ThrottleLatencyJitter
			out.Offline = rs.ThrottleOffline
		}
	}
	out.Presets = defaultThrottlePresets()
	return out
}

func (s throttleAdminService) apply(ctx context.Context, in throttleDTO) throttleDTO {
	s.d.Cfg.ThrottleEnabled = in.Enabled
	s.d.Cfg.ThrottleDownKbps = in.DownKbps
	s.d.Cfg.ThrottleUpKbps = in.UpKbps
	s.d.Cfg.ThrottlePacketLoss = in.PacketLossPct
	s.d.Cfg.ThrottleLatencyMs = in.LatencyMs
	s.d.Cfg.ThrottleLatencyJitter = in.LatencyJitter
	s.d.Cfg.ThrottleOffline = in.Offline

	if s.d.Settings != nil {
		cur, _ := s.d.Settings.Load(ctx)
		cur.ThrottleEnabled = s.d.Cfg.ThrottleEnabled
		cur.ThrottleDownKbps = s.d.Cfg.ThrottleDownKbps
		cur.ThrottleUpKbps = s.d.Cfg.ThrottleUpKbps
		cur.ThrottlePacketLoss = s.d.Cfg.ThrottlePacketLoss
		cur.ThrottleLatencyMs = s.d.Cfg.ThrottleLatencyMs
		cur.ThrottleLatencyJitter = s.d.Cfg.ThrottleLatencyJitter
		cur.ThrottleOffline = s.d.Cfg.ThrottleOffline
		cur.ResponseDelayMs = s.d.Cfg.ResponseDelayMs
		cur.ResponseDelayMinMs = s.d.Cfg.ResponseDelayMinMs
		cur.ResponseDelayMaxMs = s.d.Cfg.ResponseDelayMaxMs
		_, _ = s.d.Settings.SaveRuntime(ctx, cur)
	}

	return s.runtimeDTO()
}

func (s throttleAdminService) listProfiles(ctx context.Context) ([]sdomain.ThrottleProfile, *sessionAPIError) {
	if s.d.Settings == nil {
		return []sdomain.ThrottleProfile{}, nil
	}
	items, err := s.d.Settings.ListProfiles(ctx)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "LIST_FAILED",
			Message: err.Error(),
		}
	}
	return items, nil
}

func (s throttleAdminService) upsertProfile(ctx context.Context, in throttleProfileDTO) (sdomain.ThrottleProfile, *sessionAPIError) {
	if s.d.Settings == nil {
		return sdomain.ThrottleProfile{}, &sessionAPIError{
			Status:  http.StatusServiceUnavailable,
			Code:    "FEATURE_DISABLED",
			Message: "throttle profiles not available",
		}
	}
	p, err := s.d.Settings.UpsertProfile(ctx, sdomain.ThrottleProfile{
		ID:            in.ID,
		Name:          in.Name,
		DownKbps:      in.DownKbps,
		UpKbps:        in.UpKbps,
		PacketLossPct: in.PacketLossPct,
		LatencyMs:     in.LatencyMs,
		LatencyJitter: in.LatencyJitter,
		Offline:       in.Offline,
	})
	if err != nil {
		return sdomain.ThrottleProfile{}, &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "UPSERT_FAILED",
			Message: err.Error(),
		}
	}
	return p, nil
}

func (s throttleAdminService) deleteProfile(ctx context.Context, id string) *sessionAPIError {
	if s.d.Settings == nil {
		return &sessionAPIError{
			Status:  http.StatusServiceUnavailable,
			Code:    "FEATURE_DISABLED",
			Message: "throttle profiles not available",
		}
	}
	if err := s.d.Settings.DeleteProfile(ctx, id); err != nil {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "DELETE_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s throttleAdminService) runtimeDTO() throttleDTO {
	return throttleDTO{
		Enabled:       s.d.Cfg.ThrottleEnabled,
		DownKbps:      s.d.Cfg.ThrottleDownKbps,
		UpKbps:        s.d.Cfg.ThrottleUpKbps,
		PacketLossPct: s.d.Cfg.ThrottlePacketLoss,
		LatencyMs:     s.d.Cfg.ThrottleLatencyMs,
		LatencyJitter: s.d.Cfg.ThrottleLatencyJitter,
		Offline:       s.d.Cfg.ThrottleOffline,
	}
}

func defaultThrottlePresets() []any {
	return []any{
		map[string]any{"name": "No throttling", "downKbps": 0, "upKbps": 0, "packetLossPct": 0, "latencyMs": 0, "latencyJitter": 0},
		map[string]any{"name": "3G", "downKbps": 400, "upKbps": 400, "packetLossPct": 0, "latencyMs": 100, "latencyJitter": 20},
		map[string]any{"name": "Slow 4G", "downKbps": 1500, "upKbps": 1500, "packetLossPct": 0, "latencyMs": 50, "latencyJitter": 10},
		map[string]any{"name": "Fast 4G", "downKbps": 3000, "upKbps": 3000, "packetLossPct": 0, "latencyMs": 30, "latencyJitter": 5},
		map[string]any{"name": "Satellite", "downKbps": 2000, "upKbps": 1000, "packetLossPct": 2, "latencyMs": 600, "latencyJitter": 50},
		map[string]any{"name": "Offline", "offline": true},
	}
}
