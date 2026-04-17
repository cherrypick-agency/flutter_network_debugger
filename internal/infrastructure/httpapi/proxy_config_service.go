package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	proxyruntime "github.com/777genius/proxykit/proxyruntime"
	pdomain "network-debugger/internal/features/proxy/domain"
)

type proxyConfigAdminService struct {
	d *Deps
}

func newProxyConfigAdminService(d *Deps) proxyConfigAdminService {
	return proxyConfigAdminService{d: d}
}

func (s proxyConfigAdminService) ensureAvailable() *sessionAPIError {
	if s.d.ProxySvc == nil || s.d.ProxyRt == nil {
		return &sessionAPIError{
			Status:  http.StatusServiceUnavailable,
			Code:    "FEATURE_DISABLED",
			Message: "proxy feature not available",
		}
	}
	return nil
}

func (s proxyConfigAdminService) load(ctx context.Context) (proxyCfgDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return proxyCfgDTO{}, apiErr
	}
	pc, err := s.d.ProxySvc.Load(ctx)
	if err != nil {
		return proxyCfgDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "LOAD_FAILED",
			Message: err.Error(),
		}
	}
	return proxyDTOFromConfig(pc), nil
}

func (s proxyConfigAdminService) saveAndApply(ctx context.Context, in proxyCfgDTO) (proxyCfgDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return proxyCfgDTO{}, apiErr
	}
	if apiErr := validateProxyConfigDTO(in); apiErr != nil {
		return proxyCfgDTO{}, apiErr
	}

	pc, err := s.d.ProxySvc.Load(ctx)
	if err != nil {
		return proxyCfgDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "LOAD_FAILED",
			Message: err.Error(),
		}
	}

	mergeProxyConfigDTO(&pc, in)
	saved, err := s.d.ProxySvc.Save(ctx, pc)
	if err != nil {
		return proxyCfgDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "SAVE_FAILED",
			Message: err.Error(),
		}
	}

	if err := s.d.ProxyRt.Apply(ctx, proxyruntime.ApplyConfig{
		ForwardEnabled: saved.ForwardEnabled,
		ForwardAddr:    saved.ForwardAddr,
		SocksEnabled:   saved.SocksEnabled,
		SocksAddr:      saved.SocksAddr,
		SocksAuthMode:  saved.SocksAuthMode,
		SocksUser:      saved.SocksUser,
		SocksPass:      saved.SocksPass,
	}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.d.handleForwardOrNotFound(w, r)
	})); err != nil {
		return proxyCfgDTO{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "APPLY_FAILED",
			Message: err.Error(),
		}
	}

	return proxyDTOFromConfig(saved), nil
}

func validateProxyConfigDTO(in proxyCfgDTO) *sessionAPIError {
	if in.Forward.Port < 0 || in.Forward.Port > 65535 || in.Socks.Port < 0 || in.Socks.Port > 65535 {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_PORT",
			Message: "port must be 0..65535",
		}
	}
	if in.Socks.AuthMode != "" && in.Socks.AuthMode != "none" && in.Socks.AuthMode != "userpass" {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_AUTH_MODE",
			Message: "authMode must be none or userpass",
		}
	}
	if in.Forward.Enabled && in.Socks.Enabled && in.Forward.Port > 0 && in.Forward.Port == in.Socks.Port {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "PORT_CONFLICT",
			Message: "forward and socks ports must be different",
		}
	}
	return nil
}

func mergeProxyConfigDTO(pc *pdomain.ProxyConfig, in proxyCfgDTO) {
	pc.ForwardEnabled = in.Forward.Enabled
	if in.Forward.Port > 0 {
		pc.ForwardAddr = ":" + strconv.Itoa(in.Forward.Port)
	} else if strings.TrimSpace(in.Forward.Addr) != "" {
		pc.ForwardAddr = strings.TrimSpace(in.Forward.Addr)
	}

	pc.SocksEnabled = in.Socks.Enabled
	if in.Socks.Port > 0 {
		pc.SocksAddr = ":" + strconv.Itoa(in.Socks.Port)
	} else if strings.TrimSpace(in.Socks.Addr) != "" {
		pc.SocksAddr = strings.TrimSpace(in.Socks.Addr)
	}
	if in.Socks.AuthMode != "" {
		pc.SocksAuthMode = in.Socks.AuthMode
	}
	if in.Socks.User != "" {
		pc.SocksUser = in.Socks.User
	}
	if in.Socks.Pass != "" {
		pc.SocksPass = in.Socks.Pass
	}
}

func proxyDTOFromConfig(pc pdomain.ProxyConfig) proxyCfgDTO {
	out := proxyCfgDTO{}
	out.Forward.Enabled = pc.ForwardEnabled
	out.Forward.Addr = pc.ForwardAddr
	out.Forward.Port = addrPort(pc.ForwardAddr)
	out.Socks.Enabled = pc.SocksEnabled
	out.Socks.Addr = pc.SocksAddr
	out.Socks.Port = addrPort(pc.SocksAddr)
	out.Socks.AuthMode = pc.SocksAuthMode
	return out
}
