package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"path/filepath"
	"regexp"
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

	if len(r.HostPattern) > 4096 {
		return "BAD_RULE", "hostPattern too long (max 4096)", map[string]any{"field": "hostPattern"}, false
	}
	if len(r.PathPattern) > 4096 {
		return "BAD_RULE", "pathPattern too long (max 4096)", map[string]any{"field": "pathPattern"}, false
	}
	if len(r.ContentTypeOverride) > 512 {
		return "BAD_RULE", "contentTypeOverride too long (max 512)", map[string]any{"field": "contentTypeOverride"}, false
	}
	if r.FilePath != nil && len(*r.FilePath) > 4096 {
		return "BAD_RULE", "filePath too long (max 4096)", map[string]any{"field": "filePath"}, false
	}
	if r.TargetURLTemplate != "" && len(r.TargetURLTemplate) > 4096 {
		return "BAD_RULE", "targetURLTemplate too long (max 4096)", map[string]any{"field": "targetURLTemplate"}, false
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
	service := newMappingAdminService(d)
	switch r.Method {
	case http.MethodGet:
		cfg := service.loadConfig(contextWithNoCancel())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var in mappingConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}
		out, apiErr := service.saveConfig(contextWithNoCancel(), in)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
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
	service := newMappingAdminService(d)
	switch r.Method {
	case http.MethodGet:
		out, apiErr := service.listRules(r.Context())
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	case http.MethodPost:
		// either upsert single record, or reorder (if path ends with /reorder)
		if strings.HasSuffix(r.URL.Path, "/reorder") {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			var ids []string
			if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
				writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
				return
			}
			if apiErr := service.reorderRules(r.Context(), ids); apiErr != nil {
				writeSessionAPIError(w, apiErr)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var in mapRuleInputDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}
		out, apiErr := service.upsertRule(r.Context(), in)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
	}
}

func (d *Deps) handleMappingRuleByID(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/_api/v1/mapping/rules/")
	if apiErr := newMappingAdminService(d).deleteRule(r.Context(), id); apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
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
	out, apiErr := newMappingAdminService(d).uploadBlob(r)
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
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
	_ = f.Sync()
	_ = f.Close()
	abs, _ := filepath.Abs(f.Name())
	return abs, false, nil
}
