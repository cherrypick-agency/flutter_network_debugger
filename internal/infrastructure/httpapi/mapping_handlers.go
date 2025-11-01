package httpapi

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
)

type mappingConfigDTO struct {
	Enabled     bool `json:"enabled"`
	UploadMaxMB int  `json:"uploadMaxMB"`
}

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

func fromDTO(r mapRuleDTO) mdomain.MapRule {
	// нормализуем методы: пустые/пробелы выбрасываем
	methods := make([]string, 0, len(r.Methods))
	for _, m := range r.Methods {
		m = strings.TrimSpace(strings.ToUpper(m))
		if m != "" {
			methods = append(methods, m)
		}
	}
	return mdomain.MapRule{
		ID:                  r.ID,
		Enabled:             r.Enabled,
		Priority:            r.Priority,
		Kind:                mdomain.Kind(r.Kind),
		StopProcessing:      r.StopProcessing,
		Methods:             methods,
		HostPattern:         r.HostPattern,
		PathPattern:         r.PathPattern,
		PatternType:         mdomain.PatternType(r.PatternType),
		FilePath:            r.FilePath,
		BlobPath:            r.BlobPath,
		StatusOverride:      r.StatusOverride,
		ContentTypeOverride: r.ContentTypeOverride,
		TargetURLTemplate:   r.TargetURLTemplate,
		PreserveHost:        r.PreserveHost,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}

// handleMappingConfig возвращает/принимает конфиг фичи (MVP без персистентности)
func (d *Deps) handleMappingConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := mappingConfigDTO{Enabled: true, UploadMaxMB: 20}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cfg)
	case http.MethodPost:
		// MVP: принимаем тело, но не сохраняем — вернём как есть
		var in mappingConfigDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(in)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
	}
}

// handleMappingRules управляет CRUD правил
func (d *Deps) handleMappingRules(w http.ResponseWriter, r *http.Request) {
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
			if err := d.Mapping.Reorder(r.Context(), ids); err != nil {
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
		var in mapRuleDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "BAD_JSON", "invalid JSON", nil)
			return
		}
		saved, err := d.Mapping.Upsert(r.Context(), fromDTO(in))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "UPSERT_FAILED", err.Error(), nil)
			return
		}
		if d.MapRt != nil {
			if rules, err := d.Mapping.List(r.Context()); err == nil {
				d.MapRt.Update(rules)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(toDTO(saved))
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
	}
}

func (d *Deps) handleMappingRuleByID(w http.ResponseWriter, r *http.Request) {
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
	if err := d.Mapping.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
		return
	}
	if d.MapRt != nil {
		if rules, err := d.Mapping.List(r.Context()); err == nil {
			d.MapRt.Update(rules)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleMappingUpload — загрузка файла (Web)
func (d *Deps) handleMappingUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "", nil)
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB form buffer
		writeError(w, http.StatusBadRequest, "BAD_MULTIPART", err.Error(), nil)
		return
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "NO_FILE", "file part is required", nil)
		return
	}
	defer file.Close()

	// Определим лимит (MB → bytes)
	uploadMaxMB := 20
	if v := r.URL.Query().Get("maxMB"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			uploadMaxMB = n
		}
	}
	maxBytes := int64(uploadMaxMB) * 1024 * 1024

	// Скопируем в спул
	pr, pw := io.Pipe()
	go func(src multipart.File) {
		defer pw.Close()
		_, _ = io.Copy(pw, src)
	}(file)
	path, err := d.spoolBody(pr, maxBytes, "map")
	if err != nil || path == "" {
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
