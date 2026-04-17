package httpapi

import (
	"context"
	"net/http"
	cfgpkg "network-debugger/internal/infrastructure/config"
	"strconv"
	"strings"
)

type settingsAdminService struct {
	d *Deps
}

func newSettingsAdminService(d *Deps) settingsAdminService {
	return settingsAdminService{d: d}
}

func (s settingsAdminService) load(ctx context.Context) settingsDTO {
	out := settingsDTO{
		ResponseDelay: responseDelayDTOFromConfig(s.d.Cfg),
		FontScale:     1.0,
	}
	if s.d.Settings == nil {
		return out
	}
	if rs, err := s.d.Settings.Load(ctx); err == nil {
		out.ResponseDelay = responseDelayDTO{
			Enabled: false,
			Value:   "",
		}
		if rs.ResponseDelayMinMs > 0 && rs.ResponseDelayMaxMs > 0 {
			out.ResponseDelay.Enabled = true
			out.ResponseDelay.Value = strconv.Itoa(rs.ResponseDelayMinMs) + "-" + strconv.Itoa(rs.ResponseDelayMaxMs)
		} else if rs.ResponseDelayMs > 0 {
			out.ResponseDelay.Enabled = true
			out.ResponseDelay.Value = strconv.Itoa(rs.ResponseDelayMs)
		}
		if rs.FontScale > 0 {
			out.FontScale = rs.FontScale
		}
		out.HighlightTheme = rs.HighlightTheme
		out.HighlightThemeLight = rs.HighlightThemeLight
		out.HighlightThemeDark = rs.HighlightThemeDark
	}
	return out
}

func (s settingsAdminService) save(ctx context.Context, in settingsUpdateDTO) (settingsDTO, *sessionAPIError) {
	if in.ResponseDelay != nil {
		parsed, apiErr := parseResponseDelayDTO(*in.ResponseDelay)
		if apiErr != nil {
			return settingsDTO{}, apiErr
		}
		s.d.Cfg.ResponseDelayMs = parsed.fixed
		s.d.Cfg.ResponseDelayMinMs = parsed.min
		s.d.Cfg.ResponseDelayMaxMs = parsed.max
	}

	if s.d.Settings != nil {
		cur, _ := s.d.Settings.Load(ctx)
		if in.ResponseDelay != nil {
			cur.ResponseDelayMs = s.d.Cfg.ResponseDelayMs
			cur.ResponseDelayMinMs = s.d.Cfg.ResponseDelayMinMs
			cur.ResponseDelayMaxMs = s.d.Cfg.ResponseDelayMaxMs
		}
		if in.FontScale > 0 {
			cur.FontScale = in.FontScale
		}
		if trimmed := strings.TrimSpace(in.HighlightTheme); trimmed != "" {
			cur.HighlightTheme = trimmed
		}
		if trimmed := strings.TrimSpace(in.HighlightThemeLight); trimmed != "" {
			cur.HighlightThemeLight = trimmed
		}
		if trimmed := strings.TrimSpace(in.HighlightThemeDark); trimmed != "" {
			cur.HighlightThemeDark = trimmed
		}
		cur.ThrottleEnabled = s.d.Cfg.ThrottleEnabled
		cur.ThrottleDownKbps = s.d.Cfg.ThrottleDownKbps
		cur.ThrottleUpKbps = s.d.Cfg.ThrottleUpKbps
		cur.ThrottlePacketLoss = s.d.Cfg.ThrottlePacketLoss
		cur.ThrottleLatencyMs = s.d.Cfg.ThrottleLatencyMs
		cur.ThrottleLatencyJitter = s.d.Cfg.ThrottleLatencyJitter
		cur.ThrottleOffline = s.d.Cfg.ThrottleOffline
		_, _ = s.d.Settings.SaveRuntime(ctx, cur)
	}

	return s.load(ctx), nil
}

type parsedResponseDelay struct {
	fixed int
	min   int
	max   int
}

func parseResponseDelayDTO(in responseDelayDTO) (parsedResponseDelay, *sessionAPIError) {
	value := strings.TrimSpace(in.Value)
	if !in.Enabled || value == "" || value == "0" {
		return parsedResponseDelay{}, nil
	}
	if strings.Contains(value, "-") {
		parts := strings.SplitN(value, "-", 2)
		minStr := strings.TrimSpace(parts[0])
		maxStr := strings.TrimSpace(parts[1])
		min, err1 := strconv.Atoi(minStr)
		max, err2 := strconv.Atoi(maxStr)
		if err1 != nil || err2 != nil || min < 0 || max < 0 {
			return parsedResponseDelay{}, &sessionAPIError{
				Status:  http.StatusBadRequest,
				Code:    "BAD_VALUE",
				Message: "value must be number or range like 1000-3000",
			}
		}
		if max < min {
			min, max = max, min
		}
		return parsedResponseDelay{min: min, max: max}, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return parsedResponseDelay{}, &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_VALUE",
			Message: "value must be non-negative integer or range",
		}
	}
	return parsedResponseDelay{fixed: n}, nil
}

func responseDelayDTOFromConfig(cfg cfgpkg.Config) responseDelayDTO {
	if cfg.ResponseDelayMinMs > 0 && cfg.ResponseDelayMaxMs > 0 {
		return responseDelayDTO{
			Enabled: true,
			Value:   strconv.Itoa(cfg.ResponseDelayMinMs) + "-" + strconv.Itoa(cfg.ResponseDelayMaxMs),
		}
	}
	if cfg.ResponseDelayMs > 0 {
		return responseDelayDTO{
			Enabled: true,
			Value:   strconv.Itoa(cfg.ResponseDelayMs),
		}
	}
	return responseDelayDTO{}
}
