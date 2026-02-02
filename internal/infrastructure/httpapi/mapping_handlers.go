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
	"strconv"
	"strings"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
)

type mappingConfigDTO struct {
	Enabled     bool `json:"enabled"`
	UploadMaxMB int  `json:"uploadMaxMB"`
}

// mapRuleDTO — DTO для отдачи наружу (включая метадату времени).
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

// mapRuleInputDTO — DTO для приёма с фронта.
//
// Важно: часть полей сделана указателями, чтобы принимать JSON null без падения
// и чтобы отличать "не прислали" от "прислали значение по умолчанию".
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

	PreserveHost bool `json:"preserveHost"`
}

func toDTO(d mdomain.MapRule) mapRuleDTO {
	return mapRuleDTO{
		ID:                  d.ID,
		Enabled:             d.Enabled,
		Priority:            d.Priority,
		Kind:                string(d.Kind),
		StopProcessing:      d.StopProcessing,
		Methods:             append([]string(nil), d.Methods...),
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
	// нормализуем методы: пустые/пробелы выбрасываем
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
		// best-effort parse (с подстановкой токенов)
		mat := materializeRemoteTemplateForValidation(tpl)
		u, err := url.Parse(mat)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return "BAD_RULE", "targetURLTemplate must be a valid absolute URL template", map[string]any{"field": "targetURLTemplate", "targetURLTemplate": tpl}, false
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return "BAD_RULE", "targetURLTemplate scheme must be http or https", map[string]any{"field": "targetURLTemplate", "scheme": u.Scheme}, false
		}
	}
	// Local: file/blob опциональны — можно отдать пустое тело с нужным статусом.
	return "", "", nil, true
}

func materializeRemoteTemplateForValidation(tpl string) string {
	// Подменяем токены на валидные значения, чтобы проверить, что итоговый URL будет парситься.
	// Это позволяет принимать шаблоны вроде "https://example.com{path}".
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

// handleMappingConfig возвращает/принимает конфиг фичи (MVP без персистентности)
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
		// Если доступна БД — берём фактические значения из runtime_settings
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
		_ = json.NewEncoder(w).Encode(cfg)
	case http.MethodPost:
		// Принимаем и сохраняем в runtime_settings (если доступно)
		var in mappingConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}
		if in.UploadMaxMB <= 0 {
			in.UploadMaxMB = 20
		}
		// верхний предел — чтобы случайно не выставить гигабайты и не убить диск
		if in.UploadMaxMB > 512 {
			writeError(w, http.StatusBadRequest, "BAD_VALUE", "uploadMaxMB must be <= 512", nil)
			return
		}

		d.Cfg.MappingEnabled = in.Enabled
		d.Cfg.MappingUploadMaxMB = in.UploadMaxMB

		if d.Settings != nil {
			cur, _ := d.Settings.Load(contextWithNoCancel())
			cur.MappingEnabled = d.Cfg.MappingEnabled
			cur.MappingUploadMaxMB = d.Cfg.MappingUploadMaxMB
			// сохраняем прочие настройки, чтобы не потерять их при апдейте
			cur.ResponseDelayMs = d.Cfg.ResponseDelayMs
			cur.ResponseDelayMinMs = d.Cfg.ResponseDelayMinMs
			cur.ResponseDelayMaxMs = d.Cfg.ResponseDelayMaxMs
			cur.ThrottleEnabled = d.Cfg.ThrottleEnabled
			cur.ThrottleDownKbps = d.Cfg.ThrottleDownKbps
			cur.ThrottleUpKbps = d.Cfg.ThrottleUpKbps
			cur.ThrottlePacketLoss = d.Cfg.ThrottlePacketLoss
			cur.ThrottleOffline = d.Cfg.ThrottleOffline
			_, _ = d.Settings.SaveRuntime(contextWithNoCancel(), cur)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(in)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
	}
}

// handleMappingRules управляет CRUD правил
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
			writeError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), nil)
			return
		}
		out := make([]mapRuleDTO, 0, len(list))
		for _, it := range list {
			out = append(out, toDTO(it))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	case http.MethodPost:
		// либо upsert одной записи, либо reorder (если path оканчивается на /reorder)
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
			if err := d.Mapping.Reorder(r.Context(), ids); err != nil {
				var nf mdomain.RuleNotFoundError
				if errors.As(err, &nf) {
					writeError(
						w,
						http.StatusBadRequest,
						"BAD_IDS",
						"unknown id in reorder list",
						map[string]any{"field": "ids", "id": nf.ID},
					)
					return
				}
				writeError(w, http.StatusInternalServerError, "REORDER_FAILED", err.Error(), nil)
				return
			}
			// обновим рантайм
			if d.MapRt != nil {
				if rules, err := d.Mapping.List(r.Context()); err == nil {
					d.MapRt.Update(rules)
				}
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		var in mapRuleInputDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}

		// для уборки orphan blobs: найдём старый blobPath, если это апдейт
		oldBlob := ""
		if strings.TrimSpace(in.ID) != "" {
			if list, err := d.Mapping.List(r.Context()); err == nil {
				for _, it := range list {
					if it.ID == strings.TrimSpace(in.ID) && it.BlobPath != nil && strings.TrimSpace(*it.BlobPath) != "" {
						oldBlob = *it.BlobPath
						break
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
			writeError(w, http.StatusInternalServerError, "UPSERT_FAILED", err.Error(), nil)
			return
		}
		if d.MapRt != nil {
			if rules, err := d.Mapping.List(r.Context()); err == nil {
				d.MapRt.Update(rules)
			}
		}
		// если blobPath изменился/убран — попробуем удалить старый файл (безопасно)
		if oldBlob != "" && (saved.BlobPath == nil || strings.TrimSpace(*saved.BlobPath) == "" || strings.TrimSpace(*saved.BlobPath) != strings.TrimSpace(oldBlob)) {
			d.tryRemoveOrphanMappingBlob(r.Context(), oldBlob)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toDTO(saved))
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
	// запомним blobPath до удаления (чтобы убрать orphan файл)
	oldBlob := ""
	if list, err := d.Mapping.List(r.Context()); err == nil {
		for _, it := range list {
			if it.ID == id && it.BlobPath != nil && strings.TrimSpace(*it.BlobPath) != "" {
				oldBlob = *it.BlobPath
				break
			}
		}
	}
	if err := d.Mapping.Delete(r.Context(), id); err != nil {
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
	w.WriteHeader(http.StatusNoContent)
}

func (d *Deps) tryRemoveOrphanMappingBlob(ctx context.Context, blobPath string) {
	blobPath = strings.TrimSpace(blobPath)
	if blobPath == "" || d.Mapping == nil {
		return
	}
	// удаляем только те файлы, которые мы сами создаём
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

	// если на этот blob всё ещё кто-то ссылается — не трогаем
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

// handleMappingUpload — загрузка файла (Web)
func (d *Deps) handleMappingUpload(w http.ResponseWriter, r *http.Request) {
	if !d.interceptAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
		return
	}
	// Определим лимит (MB → bytes). Источник истины — runtime settings / config.
	uploadMaxMB := d.Cfg.MappingUploadMaxMB
	if uploadMaxMB <= 0 {
		uploadMaxMB = 20
	}
	// Для dev-режима можно уменьшить лимит через query param, но не увеличить
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
		writeError(w, http.StatusInternalServerError, "SPOOL_FAILED", "failed to store file", nil)
		return
	}
	if path == "" {
		writeError(w, http.StatusInternalServerError, "SPOOL_FAILED", "failed to store file", nil)
		return
	}

	// Очень приблизительная оценка content-type по расширению имени файла
	ct := hdr.Header.Get("Content-Type")
	if ct == "" {
		ct = guessContentTypeByName(hdr.Filename)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"blobPath":    path,
		"fileName":    hdr.Filename,
		"contentType": ct,
	})
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

	// maxBytes + 1: чтобы отличить "влезло ровно" от "перелимит".
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
