package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
)

type mappingConfigDTO struct {
	Enabled     bool `json:"enabled"`
	UploadMaxMB int  `json:"uploadMaxMB"`
}

// mapRuleDTO — DTO for external response (including time metadata).
type mapRuleDTO struct {
	ID                  string    `json:"id"`
	Enabled             bool      `json:"enabled"`
	Priority            int       `json:"priority"`
	Kind                string    `json:"kind"`
	StopProcessing      bool      `json:"stopProcessing"`
	Methods             []string  `json:"methods"`
	HostPattern         string    `json:"hostPattern"`
	PathPattern         string    `json:"pathPattern"`
	PatternType         string    `json:"patternType"`
	FilePath            *string   `json:"filePath"`
	BlobPath            *string   `json:"blobPath"`
	StatusOverride      int       `json:"statusOverride"`
	ContentTypeOverride string    `json:"contentTypeOverride"`
	TargetURLTemplate   string    `json:"targetURLTemplate"`
	PreserveHost        bool      `json:"preserveHost"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// mapRuleInputDTO — DTO for receiving from frontend.
//
// Important: some fields are pointers to accept JSON null without crashing
// and to distinguish "not sent" from "sent default value".
type mapRuleInputDTO struct {
	ID             string   `json:"id"`
	Enabled        bool     `json:"enabled"`
	Priority       int      `json:"priority"`
	Kind           string   `json:"kind"`
	StopProcessing bool     `json:"stopProcessing"`
	Methods        []string `json:"methods"`

	HostPattern string `json:"hostPattern"`
	PathPattern string `json:"pathPattern"`
	PatternType string `json:"patternType"`

	FilePath *string `json:"filePath"`
	BlobPath *string `json:"blobPath"`

	StatusOverride      *int    `json:"statusOverride"`
	ContentTypeOverride *string `json:"contentTypeOverride"`
	TargetURLTemplate   *string `json:"targetURLTemplate"`

	PreserveHost bool       `json:"preserveHost"`
	UpdatedAt    *time.Time `json:"updatedAt,omitempty"`
}

func toDTO(d mdomain.MapRule) mapRuleDTO {
	return mapRuleDTO{
		ID:                  d.ID,
		Enabled:             d.Enabled,
		Priority:            d.Priority,
		Kind:                string(d.Kind),
		StopProcessing:      d.StopProcessing,
		Methods:             append([]string{}, d.Methods...),
		HostPattern:         d.HostPattern,
		PathPattern:         d.PathPattern,
		PatternType:         string(d.PatternType),
		FilePath:            d.FilePath,
		BlobPath:            d.BlobPath,
		StatusOverride:      d.StatusOverride,
		ContentTypeOverride: d.ContentTypeOverride,
		TargetURLTemplate:   d.TargetURLTemplate,
		PreserveHost:        d.PreserveHost,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
}

func fromDTO(r mapRuleInputDTO) mdomain.MapRule {
	// normalize methods: discard empty/whitespace
	methods := make([]string, 0, len(r.Methods))
	for _, m := range r.Methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		if m != "" {
			methods = append(methods, m)
		}
	}

	status := 200
	if r.StatusOverride != nil && *r.StatusOverride != 0 {
		status = *r.StatusOverride
	}
	ct := ""
	if r.ContentTypeOverride != nil {
		ct = *r.ContentTypeOverride
	}
	target := ""
	if r.TargetURLTemplate != nil {
		target = *r.TargetURLTemplate
	}

	return mdomain.MapRule{
		ID:                  r.ID,
		Enabled:             r.Enabled,
		Priority:            normalizePriority(r.Priority),
		Kind:                normalizeKind(r.Kind),
		StopProcessing:      r.StopProcessing,
		Methods:             methods,
		HostPattern:         r.HostPattern,
		PathPattern:         r.PathPattern,
		PatternType:         normalizePatternType(r.PatternType),
		FilePath:            r.FilePath,
		BlobPath:            r.BlobPath,
		StatusOverride:      status,
		ContentTypeOverride: ct,
		TargetURLTemplate:   target,
		PreserveHost:        r.PreserveHost,
	}
}

func normalizePriority(v int) int {
	if v <= 0 {
		return 100
	}
	return v
}

func normalizeKind(v string) mdomain.Kind {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return mdomain.KindRemote
	}
	return mdomain.Kind(v)
}

func normalizePatternType(v string) mdomain.PatternType {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return mdomain.PatternGlob
	}
	return mdomain.PatternType(v)
}

func validateRule(r mdomain.MapRule) (string, string, map[string]any, bool) {
	if r.Kind != mdomain.KindLocal && r.Kind != mdomain.KindRemote {
		return "BAD_RULE", "kind must be local or remote", map[string]any{"field": "kind", "kind": r.Kind}, false
	}
	if r.PatternType != mdomain.PatternGlob && r.PatternType != mdomain.PatternRegex {
		return "BAD_RULE", "patternType must be glob or regex", map[string]any{"field": "patternType", "patternType": r.PatternType}, false
	}
	if r.Priority <= 0 {
		return "BAD_RULE", "priority must be > 0", map[string]any{"field": "priority", "priority": r.Priority}, false
	}
	if r.StatusOverride < 100 || r.StatusOverride > 599 {
		return "BAD_RULE", "statusOverride must be in 100..599", map[string]any{"field": "statusOverride", "statusOverride": r.StatusOverride}, false
	}
	if len(r.Methods) > 20 {
		return "BAD_RULE", "too many HTTP methods", map[string]any{"field": "methods", "methods": len(r.Methods)}, false
	}
	for _, m := range r.Methods {
		if m == "" {
			continue
		}
		for _, ch := range m {
			if ch < 'A' || ch > 'Z' {
				return "BAD_RULE", "method must be uppercase token like GET", map[string]any{"field": "methods", "method": m}, false
			}
		}
	}

	if strings.ContainsAny(r.HostPattern, "\n\r\t") {
		return "BAD_RULE", "hostPattern must not contain control whitespace", map[string]any{"field": "hostPattern"}, false
	}
	if strings.ContainsAny(r.PathPattern, "\n\r\t") {
		return "BAD_RULE", "pathPattern must not contain control whitespace", map[string]any{"field": "pathPattern"}, false
	}

	switch r.PatternType {
	case mdomain.PatternRegex:
		if strings.TrimSpace(r.HostPattern) != "" {
			if _, err := regexp.Compile(r.HostPattern); err != nil {
				return "BAD_RULE", "invalid host regex", map[string]any{"field": "hostPattern", "hostPattern": r.HostPattern}, false
			}
		}
		if strings.TrimSpace(r.PathPattern) != "" {
			if _, err := regexp.Compile(r.PathPattern); err != nil {
				return "BAD_RULE", "invalid path regex", map[string]any{"field": "pathPattern", "pathPattern": r.PathPattern}, false
			}
		}
	case mdomain.PatternGlob:
		if strings.TrimSpace(r.HostPattern) != "" {
			if _, err := stdpath.Match(r.HostPattern, "example.com"); err != nil {
				return "BAD_RULE", "invalid host glob", map[string]any{"field": "hostPattern", "hostPattern": r.HostPattern}, false
			}
		}
		if strings.TrimSpace(r.PathPattern) != "" {
			if _, err := stdpath.Match(r.PathPattern, "/"); err != nil {
				return "BAD_RULE", "invalid path glob", map[string]any{"field": "pathPattern", "pathPattern": r.PathPattern}, false
			}
		}
	}

	if strings.ContainsAny(r.ContentTypeOverride, "\n\r") {
		return "BAD_RULE", "contentTypeOverride must not contain CR/LF", map[string]any{"field": "contentTypeOverride"}, false
	}
	if r.FilePath != nil && strings.ContainsAny(*r.FilePath, "\n\r") {
		return "BAD_RULE", "filePath must not contain CR/LF", map[string]any{"field": "filePath"}, false
	}
	if r.BlobPath != nil && strings.ContainsAny(*r.BlobPath, "\n\r") {
		return "BAD_RULE", "blobPath must not contain CR/LF", map[string]any{"field": "blobPath"}, false
	}
	if r.FilePath != nil && strings.TrimSpace(*r.FilePath) != "" &&
		r.BlobPath != nil && strings.TrimSpace(*r.BlobPath) != "" {
		return "BAD_RULE", "only one of filePath/blobPath must be set", map[string]any{"field": "filePath/blobPath"}, false
	}

	if r.Kind == mdomain.KindRemote {
		tpl := strings.TrimSpace(r.TargetURLTemplate)
		if tpl == "" {
			return "BAD_RULE", "targetURLTemplate is required for remote rules", map[string]any{"field": "targetURLTemplate"}, false
		}
		if strings.ContainsAny(tpl, "\n\r\t ") {
			return "BAD_RULE", "targetURLTemplate must not contain whitespace", map[string]any{"field": "targetURLTemplate"}, false
		}
		// best-effort parse (with token substitution)
		mat := materializeRemoteTemplateForValidation(tpl)
		u, err := url.Parse(mat)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "BAD_RULE", "targetURLTemplate must be a valid absolute URL template", map[string]any{"field": "targetURLTemplate", "targetURLTemplate": tpl}, false
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return "BAD_RULE", "targetURLTemplate scheme must be http or https", map[string]any{"field": "targetURLTemplate", "scheme": u.Scheme}, false
		}
	}
	// Local: file/blob are optional — can return empty body with desired status.
	return "", "", nil, true
}

func materializeRemoteTemplateForValidation(tpl string) string {
	// Replace tokens with valid values to verify that resulting URL will parse.
	// This allows accepting templates like "https://example.com{path}".
	return strings.NewReplacer(
		"{scheme}", "https",
		"{host}", "example.com",
		"{hostname}", "example.com",
		"{port}", "443",
		"{portWithColon}", ":443",
		"{path}", "/test",
		"{query}", "a=1",
		"{rawQuery}", "a=1",
		"{url}", "https://example.com/test?a=1",
		"{full}", "https://example.com/test?a=1",
	).Replace(tpl)
}

// handleMappingConfig returns/accepts feature config (MVP without persistence)
func (d *Deps) handleMappingConfig(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		cfg := mappingConfigDTO{
			Enabled:     d.Cfg.MappingEnabled,
			UploadMaxMB: d.Cfg.MappingUploadMaxMB,
		}
		// If DB is available — get actual values from runtime_settings
		if d.Settings != nil {
			if rs, err := d.Settings.Load(contextWithNoCancel()); err == nil {
				cfg.Enabled = rs.MappingEnabled
				if rs.MappingUploadMaxMB > 0 {
					cfg.UploadMaxMB = rs.MappingUploadMaxMB
				}
			}
		}
		if cfg.UploadMaxMB <= 0 {
			cfg.UploadMaxMB = 20
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cfg); err != nil {
			log.Printf("[MappingHandler] encode error: %v", err)
		}
	case http.MethodPost:
		// Accept and save to runtime_settings (if available)
		var in mappingConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}
		if in.UploadMaxMB <= 0 {
			in.UploadMaxMB = 20
		}
		// upper limit — to avoid accidentally setting gigabytes and killing disk
		if in.UploadMaxMB > 512 {
			writeError(w, http.StatusBadRequest, "BAD_VALUE", "uploadMaxMB must be <= 512", nil)
			return
		}

		d.CfgMu.Lock()
		d.Cfg.MappingEnabled = in.Enabled
		d.Cfg.MappingUploadMaxMB = in.UploadMaxMB
		d.CfgMu.Unlock()

		if d.Settings != nil {
			cur, _ := d.Settings.Load(contextWithNoCancel())
			cur.MappingEnabled = in.Enabled
			cur.MappingUploadMaxMB = in.UploadMaxMB
			_, _ = d.Settings.SaveRuntime(contextWithNoCancel(), cur)
		}

		log.Printf("[MappingHandler] config updated: enabled=%v uploadMaxMB=%d", in.Enabled, in.UploadMaxMB)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(in); err != nil {
			log.Printf("[MappingHandler] encode error: %v", err)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
	}
}

// handleMappingRules manages rules CRUD
func (d *Deps) handleMappingRules(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if d.Mapping == nil {
		writeError(w, http.StatusServiceUnavailable, "NO_SERVICE", "mapping service unavailable", nil)
		return
	}
	switch r.Method {
	case http.MethodGet:
		list, err := d.Mapping.List(r.Context())
		if err != nil {
			log.Printf("[MappingHandler] list rules: %v", err)
			writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), nil)
			return
		}
		out := make([]mapRuleDTO, 0, len(list))
		for _, it := range list {
			out = append(out, toDTO(it))
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(out); err != nil {
			log.Printf("[MappingHandler] encode error: %v", err)
		}
	case http.MethodPost:
		// either upsert single record, or reorder (if path ends with /reorder)
		if strings.HasSuffix(r.URL.Path, "/reorder") {
			var ids []string
			if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
				return
			}
			seen := make(map[string]struct{}, len(ids))
			for _, idv := range ids {
				idv = strings.TrimSpace(idv)
				if idv == "" {
					writeError(w, http.StatusBadRequest, "BAD_IDS", "ids must not contain empty values", map[string]any{"field": "ids"})
					return
				}
				if _, ok := seen[idv]; ok {
					writeError(w, http.StatusBadRequest, "BAD_IDS", "ids must not contain duplicates", map[string]any{"field": "ids", "id": idv})
					return
				}
				seen[idv] = struct{}{}
			}
			// validate that submitted IDs contain ALL existing rule IDs
			allRules, err := d.Mapping.List(r.Context())
			if err != nil {
				log.Printf("[MappingHandler] reorder: failed to list rules: %v", err)
				writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), nil)
				return
			}
			if len(ids) != len(allRules) {
				writeError(w, http.StatusBadRequest, "BAD_IDS", "submitted IDs must contain all existing rule IDs", map[string]any{
					"field":    "ids",
					"expected": len(allRules),
					"got":      len(ids),
				})
				return
			}
			for _, rule := range allRules {
				if _, ok := seen[rule.ID]; !ok {
					writeError(w, http.StatusBadRequest, "BAD_IDS", "submitted IDs must contain all existing rule IDs", map[string]any{
						"field":     "ids",
						"missingId": rule.ID,
					})
					return
				}
			}
			if err := d.Mapping.Reorder(r.Context(), ids); err != nil {
				var nf mdomain.RuleNotFoundError
				if errors.As(err, &nf) {
					writeError(
						w,
						http.StatusConflict,
						"BAD_IDS",
						"rule was deleted by another user during reorder, please reload",
						map[string]any{"field": "ids", "id": nf.ID},
					)
					return
				}
				log.Printf("[MappingHandler] reorder rules: %v", err)
				writeError(w, http.StatusInternalServerError, "REORDER_FAILED", err.Error(), nil)
				return
			}
			// update runtime
			if d.MapRt != nil {
				if rules, err := d.Mapping.List(r.Context()); err == nil {
					d.MapRt.Update(rules)
				}
			}
			log.Printf("[MappingHandler] rules reordered: count=%d", len(ids))
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var in mapRuleInputDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}

		// for orphan blob cleanup and optimistic locking: find existing rule if editing
		oldBlob := ""
		if strings.TrimSpace(in.ID) != "" {
			if old, err := d.Mapping.GetByID(r.Context(), strings.TrimSpace(in.ID)); err == nil && old != nil {
				if old.BlobPath != nil && strings.TrimSpace(*old.BlobPath) != "" {
					oldBlob = *old.BlobPath
				}
				// optimistic locking: compare updatedAt timestamps
				if in.UpdatedAt != nil {
					if !old.UpdatedAt.Truncate(time.Second).Equal(in.UpdatedAt.Truncate(time.Second)) {
						writeError(w, http.StatusConflict, "CONFLICT", "rule was modified by another user, please reload", map[string]any{
							"field":           "updatedAt",
							"serverUpdatedAt": old.UpdatedAt,
							"clientUpdatedAt": *in.UpdatedAt,
						})
						return
					}
				}
			}
		}

		rule := fromDTO(in)
		if code, msg, det, ok := validateRule(rule); !ok {
			writeError(w, http.StatusBadRequest, code, msg, det)
			return
		}
		saved, err := d.Mapping.Upsert(r.Context(), rule)
		if err != nil {
			log.Printf("[MappingHandler] upsert rule: %v", err)
			writeError(w, http.StatusInternalServerError, "UPSERT_FAILED", err.Error(), nil)
			return
		}
		if d.MapRt != nil {
			if rules, err := d.Mapping.List(r.Context()); err == nil {
				d.MapRt.Update(rules)
			}
		}
		// if blobPath changed/removed — try to delete old file (safely)
		if oldBlob != "" && (saved.BlobPath == nil || strings.TrimSpace(*saved.BlobPath) == "" || strings.TrimSpace(*saved.BlobPath) != strings.TrimSpace(oldBlob)) {
			d.tryRemoveOrphanMappingBlob(r.Context(), oldBlob)
		}
		log.Printf("[MappingHandler] rule upserted: id=%s kind=%s", saved.ID, saved.Kind)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(toDTO(saved)); err != nil {
			log.Printf("[MappingHandler] encode error: %v", err)
		}
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
	}
}

func (d *Deps) handleMappingRuleByID(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if d.Mapping == nil {
		writeError(w, http.StatusServiceUnavailable, "NO_SERVICE", "mapping service unavailable", nil)
		return
	}
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/_api/v1/mapping/rules/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "MISSING_ID", "", nil)
		return
	}
	// remember blobPath before deletion (to remove orphan file)
	oldBlob := ""
	if old, err := d.Mapping.GetByID(r.Context(), id); err == nil && old != nil {
		if old.BlobPath != nil && strings.TrimSpace(*old.BlobPath) != "" {
			oldBlob = *old.BlobPath
		}
	}
	if err := d.Mapping.Delete(r.Context(), id); err != nil {
		log.Printf("[MappingHandler] delete rule: %v", err)
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
		return
	}
	if d.MapRt != nil {
		if rules, err := d.Mapping.List(r.Context()); err == nil {
			d.MapRt.Update(rules)
		}
	}
	if oldBlob != "" {
		d.tryRemoveOrphanMappingBlob(r.Context(), oldBlob)
	}
	log.Printf("[MappingHandler] rule deleted: id=%s", id)
	w.WriteHeader(http.StatusNoContent)
}

func (d *Deps) tryRemoveOrphanMappingBlob(ctx context.Context, blobPath string) {
	blobPath = strings.TrimSpace(blobPath)
	if blobPath == "" || d.Mapping == nil {
		return
	}
	// only delete files that we created ourselves
	base := filepath.Base(blobPath)
	if !strings.HasPrefix(base, "gpx-map-") {
		return
	}

	spoolDir := d.Cfg.BodySpoolDir
	if spoolDir == "" {
		spoolDir = os.TempDir()
	}
	spoolAbs, _ := filepath.Abs(spoolDir)
	blobAbs, _ := filepath.Abs(blobPath)
	if !isWithinDir(spoolAbs, blobAbs) {
		return
	}

	// if someone still references this blob — don't touch it
	list, err := d.Mapping.List(ctx)
	if err != nil {
		return
	}
	for _, r := range list {
		if r.BlobPath != nil && strings.TrimSpace(*r.BlobPath) != "" {
			abs, _ := filepath.Abs(*r.BlobPath)
			if abs == blobAbs {
				return
			}
		}
	}
	_ = os.Remove(blobAbs)
}

func isWithinDir(dirAbs, pathAbs string) bool {
	rel, err := filepath.Rel(dirAbs, pathAbs)
	if err != nil {
		return false
	}
	rel = filepath.Clean(rel)
	return rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && rel != ".."
}

// handleMappingUpload — file upload (Web)
func (d *Deps) handleMappingUpload(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
		return
	}
	// Determine limit (MB → bytes). Source of truth — runtime settings / config.
	uploadMaxMB := d.Cfg.MappingUploadMaxMB
	if uploadMaxMB <= 0 {
		uploadMaxMB = 20
	}
	// For dev-mode you can decrease limit via query param, but not increase
	if v := r.URL.Query().Get("maxMB"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 && n < uploadMaxMB {
			uploadMaxMB = n
		}
	}
	maxBytes := int64(uploadMaxMB) * 1024 * 1024

	maxMem := int64(32 << 20)
	if maxBytes > 0 && maxBytes < maxMem {
		maxMem = maxBytes
	}
	if err := r.ParseMultipartForm(maxMem); err != nil {
		log.Printf("[MappingHandler] bad multipart: %v", err)
		writeError(w, http.StatusBadRequest, "BAD_MULTIPART", err.Error(), nil)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "NO_FILE", "file part is required", nil)
		return
	}
	defer file.Close()

	path, tooLarge, err := d.spoolMultipartFile(file, maxBytes, "map")
	if err != nil {
		if errors.Is(err, errSpoolTooLarge) || tooLarge {
			writeError(
				w,
				http.StatusRequestEntityTooLarge,
				"FILE_TOO_LARGE",
				"file exceeds upload limit",
				map[string]any{"maxBytes": maxBytes, "maxMB": uploadMaxMB, "fileName": hdr.Filename},
			)
			return
		}
		log.Printf("[MappingHandler] spool file: %v", err)
		writeError(w, http.StatusInternalServerError, "SPOOL_FAILED", "failed to store file", nil)
		return
	}
	if path == "" {
		log.Printf("[MappingHandler] spool file: empty path")
		writeError(w, http.StatusInternalServerError, "SPOOL_FAILED", "failed to store file", nil)
		return
	}

	// Very rough content-type estimate based on filename extension
	ct := hdr.Header.Get("Content-Type")
	if ct == "" {
		ct = guessContentTypeByName(hdr.Filename)
	}

	log.Printf("[MappingHandler] file uploaded: path=%s name=%s", path, hdr.Filename)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"blobPath":    path,
		"fileName":    hdr.Filename,
		"contentType": ct,
	}); err != nil {
		log.Printf("[MappingHandler] encode error: %v", err)
	}
}

func guessContentTypeByName(name string) string {
	low := strings.ToLower(name)
	switch {
	case strings.HasSuffix(low, ".js"):
		return "application/javascript"
	case strings.HasSuffix(low, ".css"):
		return "text/css"
	case strings.HasSuffix(low, ".json"):
		return "application/json"
	case strings.HasSuffix(low, ".png"):
		return "image/png"
	case strings.HasSuffix(low, ".jpg") || strings.HasSuffix(low, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(low, ".svg"):
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

var errSpoolTooLarge = errors.New("spool too large")

func (d *Deps) spoolMultipartFile(file multipart.File, maxBytes int64, kind string) (string, bool, error) {
	dir := d.Cfg.BodySpoolDir
	if dir == "" {
		dir = os.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, err
	}
	f, err := os.CreateTemp(dir, "gpx-"+kind+"-*.bin")
	if err != nil {
		return "", false, err
	}
	defer func() { _ = f.Sync() }()

	// maxBytes + 1: to distinguish "fits exactly" from "over limit".
	n, copyErr := io.CopyN(f, file, maxBytes+1)
	if copyErr != nil && copyErr != io.EOF && copyErr != io.ErrUnexpectedEOF {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", false, copyErr
	}
	if n > maxBytes {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", true, errSpoolTooLarge
	}
	_ = f.Close()
	abs, _ := filepath.Abs(f.Name())
	return abs, false, nil
}
