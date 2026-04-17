package httpapi

import (
	"encoding/json"
	"net/http"
)

type responseDelayDTO struct {
	Enabled bool   `json:"enabled"`
	Value   string `json:"value"` // "1500" or range "1000-3000"
}

type settingsDTO struct {
	ResponseDelay       responseDelayDTO `json:"responseDelay"`
	FontScale           float64          `json:"fontScale,omitempty"`
	HighlightTheme      string           `json:"highlightTheme,omitempty"`
	HighlightThemeLight string           `json:"highlightThemeLight,omitempty"`
	HighlightThemeDark  string           `json:"highlightThemeDark,omitempty"`
}

type settingsUpdateDTO struct {
	ResponseDelay       *responseDelayDTO `json:"responseDelay,omitempty"`
	FontScale           float64           `json:"fontScale,omitempty"`
	HighlightTheme      string            `json:"highlightTheme,omitempty"`
	HighlightThemeLight string            `json:"highlightThemeLight,omitempty"`
	HighlightThemeDark  string            `json:"highlightThemeDark,omitempty"`
}

func (d *Deps) handleV1Settings(w http.ResponseWriter, r *http.Request) {
	service := newSettingsAdminService(d)
	switch r.Method {
	case http.MethodGet:
		out := service.load(contextWithNoCancel())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
		return
	case http.MethodPost:
		var in settingsUpdateDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid json", nil)
			return
		}
		out, apiErr := service.save(contextWithNoCancel(), in)
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
