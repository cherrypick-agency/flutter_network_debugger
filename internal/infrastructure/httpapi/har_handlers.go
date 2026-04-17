package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleExportHAR handles POST /_api/v1/export/har
// Accepts JSON body with export options
func (d *Deps) handleExportHAR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	var opts HARExportOptions

	if r.Method == http.MethodPost {
		// Parse JSON body
		var body struct {
			SessionIDs       []string `json:"sessionIds"`
			IncludeBodies    *bool    `json:"includeBodies"`
			IncludeSensitive *bool    `json:"includeSensitive"`
			Minify           *bool    `json:"minify"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
			return
		}

		opts.SessionIDs = body.SessionIDs
		opts.IncludeBodies = body.IncludeBodies != nil && *body.IncludeBodies
		opts.IncludeSensitive = body.IncludeSensitive != nil && *body.IncludeSensitive
		opts.Minify = body.Minify != nil && *body.Minify
	} else {
		// Parse query parameters (GET fallback)
		q := r.URL.Query()

		if ids := q.Get("sessionIds"); ids != "" {
			opts.SessionIDs = strings.Split(ids, ",")
		}

		opts.IncludeBodies = parseBoolParam(q.Get("includeBodies"), true)
		opts.IncludeSensitive = parseBoolParam(q.Get("includeSensitive"), false)
		opts.Minify = parseBoolParam(q.Get("minify"), false)
	}

	exportHARWithOptions(w, r, d, opts)
}

// handleImportHAR handles POST /_api/v1/import/har
// Accepts HAR file upload and creates sessions/transactions
func (d *Deps) handleImportHAR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", nil)
		return
	}

	// Parse import mode from query parameter (default: merge)
	mode, apiErr := parseHARImportMode(r.URL.Query().Get("mode"))
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}

	// Limit request body to 500MB
	r.Body = http.MaxBytesReader(w, r.Body, 500*1024*1024)

	// Parse HAR from body
	var harData struct {
		Log harLog `json:"log"`
	}

	if err := json.NewDecoder(r.Body).Decode(&harData); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HAR", "Failed to parse HAR format: "+err.Error(), nil)
		return
	}

	summary, apiErr := newHARService(d).importLog(r.Context(), harData.Log, mode)
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}

	// Return summary
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"imported":   summary.Imported,
		"failed":     summary.Failed,
		"total":      summary.Total,
		"sessionIds": summary.SessionIDs,
	})
}
