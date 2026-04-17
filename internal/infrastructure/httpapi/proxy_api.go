package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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
	service := newProxyConfigAdminService(d)
	switch r.Method {
	case http.MethodGet:
		out, apiErr := service.load(contextWithNoCancel())
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	case http.MethodPost:
		var in proxyCfgDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		out, apiErr := service.saveAndApply(contextWithNoCancel(), in)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func addrPort(addr string) int {
	// Formats: ":8888" | "127.0.0.1:8888" | "localhost:8888"
	i := strings.LastIndex(addr, ":")
	if i < 0 || i == len(addr)-1 {
		return 0
	}
	p, _ := strconv.Atoi(addr[i+1:])
	return p
}
