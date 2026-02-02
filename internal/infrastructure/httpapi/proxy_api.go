package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	pruntime "network-debugger/internal/infrastructure/proxyruntime"
)

type proxyCfgDTO struct {
	Forward struct {
		Enabled bool   `json:"enabled"`
		Port    int    `json:"port"`
		Addr    string `json:"addr,omitempty"`
	} `json:"forward"`
	Socks struct {
		Enabled  bool   `json:"enabled"`
		Port     int    `json:"port"`
		Addr     string `json:"addr,omitempty"`
		AuthMode string `json:"authMode"`
		User     string `json:"user,omitempty"`
		Pass     string `json:"pass,omitempty"`
	} `json:"socks"`
}

func (d *Deps) handleV1ProxyConfig(w http.ResponseWriter, r *http.Request) {
	if d.ProxySvc == nil || d.ProxyRt == nil {
		writeError(w, http.StatusServiceUnavailable, "FEATURE_DISABLED", "proxy feature not available", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		pc, err := d.ProxySvc.Load(contextWithNoCancel())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "LOAD_FAILED", err.Error(), nil)
			return
		}
		out := proxyCfgDTO{}
		out.Forward.Enabled = pc.ForwardEnabled
		out.Forward.Addr = pc.ForwardAddr
		out.Forward.Port = addrPort(pc.ForwardAddr)
		out.Socks.Enabled = pc.SocksEnabled
		out.Socks.Addr = pc.SocksAddr
		out.Socks.Port = addrPort(pc.SocksAddr)
		out.Socks.AuthMode = pc.SocksAuthMode
		// user/pass не возвращаем, только если установлены и явно нужны – опустим для простоты UX
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	case http.MethodPost:
		var in proxyCfgDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		// Валидация
		if in.Forward.Port < 0 || in.Forward.Port > 65535 || in.Socks.Port < 0 || in.Socks.Port > 65535 {
			writeError(w, http.StatusBadRequest, "BAD_PORT", "port must be 0..65535", nil)
			return
		}
		if in.Socks.AuthMode != "" && in.Socks.AuthMode != "none" && in.Socks.AuthMode != "userpass" {
			writeError(w, http.StatusBadRequest, "BAD_AUTH_MODE", "authMode must be none or userpass", nil)
			return
		}
		// Не даём случайно посадить SOCKS и forward на один и тот же порт:
		// в таком случае клиенты будут успешно подключаться по TCP, но HTTP к proxy-порту работать не будет.
		if in.Forward.Enabled && in.Socks.Enabled && in.Forward.Port > 0 && in.Forward.Port == in.Socks.Port {
			writeError(w, http.StatusBadRequest, "PORT_CONFLICT", "forward and socks ports must be different", nil)
			return
		}

		pc, err := d.ProxySvc.Load(contextWithNoCancel())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "LOAD_FAILED", err.Error(), nil)
			return
		}
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
		// Сохраняем и применяем
		saved, err := d.ProxySvc.Save(contextWithNoCancel(), pc)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "SAVE_FAILED", err.Error(), nil)
			return
		}
		if err := d.ProxyRt.Apply(contextWithNoCancel(), pruntime.ApplyConfig{
			ForwardEnabled: saved.ForwardEnabled,
			ForwardAddr:    saved.ForwardAddr,
			SocksEnabled:   saved.SocksEnabled,
			SocksAddr:      saved.SocksAddr,
			SocksAuthMode:  saved.SocksAuthMode,
			SocksUser:      saved.SocksUser,
			SocksPass:      saved.SocksPass,
		}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { d.handleForwardOrNotFound(w, r) })); err != nil {
			writeError(w, http.StatusInternalServerError, "APPLY_FAILED", err.Error(), nil)
			return
		}

		// Ответим текущей конфигурацией
		out := proxyCfgDTO{}
		out.Forward.Enabled = saved.ForwardEnabled
		out.Forward.Addr = saved.ForwardAddr
		out.Forward.Port = addrPort(saved.ForwardAddr)
		out.Socks.Enabled = saved.SocksEnabled
		out.Socks.Addr = saved.SocksAddr
		out.Socks.Port = addrPort(saved.SocksAddr)
		out.Socks.AuthMode = saved.SocksAuthMode
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func addrPort(addr string) int {
	// Форматы: ":8888" | "127.0.0.1:8888" | "localhost:8888"
	i := strings.LastIndex(addr, ":")
	if i < 0 || i == len(addr)-1 {
		return 0
	}
	p, _ := strconv.Atoi(addr[i+1:])
	return p
}
