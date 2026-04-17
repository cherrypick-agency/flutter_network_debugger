package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	mdomain "network-debugger/internal/features/mapping/domain"
)

type mappingAdminService struct {
	d *Deps
}

func newMappingAdminService(d *Deps) mappingAdminService {
	return mappingAdminService{d: d}
}

func (s mappingAdminService) loadConfig(ctx context.Context) mappingConfigDTO {
	s.d.CfgMu.RLock()
	cfg := mappingConfigDTO{
		Enabled:     s.d.Cfg.MappingEnabled,
		UploadMaxMB: s.d.Cfg.MappingUploadMaxMB,
	}
	s.d.CfgMu.RUnlock()

	if s.d.Settings != nil {
		if rs, err := s.d.Settings.Load(ctx); err == nil {
			cfg.Enabled = rs.MappingEnabled
			if rs.MappingUploadMaxMB > 0 {
				cfg.UploadMaxMB = rs.MappingUploadMaxMB
			}
		}
	}
	if cfg.UploadMaxMB <= 0 {
		cfg.UploadMaxMB = 20
	}
	return cfg
}

func (s mappingAdminService) saveConfig(ctx context.Context, in mappingConfigDTO) (mappingConfigDTO, *sessionAPIError) {
	if in.UploadMaxMB <= 0 {
		in.UploadMaxMB = 20
	}
	if in.UploadMaxMB > 512 {
		return mappingConfigDTO{}, &sessionAPIError{
			Status:  400,
			Code:    "BAD_VALUE",
			Message: "uploadMaxMB must be <= 512",
		}
	}

	s.d.CfgMu.Lock()
	s.d.Cfg.MappingEnabled = in.Enabled
	s.d.Cfg.MappingUploadMaxMB = in.UploadMaxMB
	s.d.CfgMu.Unlock()

	if s.d.Settings != nil {
		cur, _ := s.d.Settings.Load(ctx)
		cur.MappingEnabled = in.Enabled
		cur.MappingUploadMaxMB = in.UploadMaxMB
		_, _ = s.d.Settings.SaveRuntime(ctx, cur)
	}

	return in, nil
}

func (s mappingAdminService) ensureRulesAvailable() *sessionAPIError {
	if s.d.Mapping == nil {
		return &sessionAPIError{
			Status:  503,
			Code:    "NO_SERVICE",
			Message: "mapping service unavailable",
		}
	}
	return nil
}

func (s mappingAdminService) listRules(ctx context.Context) ([]mapRuleDTO, *sessionAPIError) {
	if apiErr := s.ensureRulesAvailable(); apiErr != nil {
		return nil, apiErr
	}
	list, err := s.d.Mapping.List(ctx)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  500,
			Code:    "LIST_FAILED",
			Message: "internal error",
		}
	}
	out := make([]mapRuleDTO, 0, len(list))
	for _, it := range list {
		out = append(out, toDTO(it))
	}
	return out, nil
}

func (s mappingAdminService) reorderRules(ctx context.Context, ids []string) *sessionAPIError {
	if apiErr := s.ensureRulesAvailable(); apiErr != nil {
		return apiErr
	}
	for i := range ids {
		ids[i] = strings.TrimSpace(ids[i])
	}
	seen := make(map[string]struct{}, len(ids))
	for _, idv := range ids {
		if idv == "" {
			return &sessionAPIError{
				Status:  400,
				Code:    "BAD_IDS",
				Message: "ids must not contain empty values",
				Details: map[string]any{"field": "ids"},
			}
		}
		if _, ok := seen[idv]; ok {
			return &sessionAPIError{
				Status:  400,
				Code:    "BAD_IDS",
				Message: "ids must not contain duplicates",
				Details: map[string]any{"field": "ids", "id": idv},
			}
		}
		seen[idv] = struct{}{}
	}

	allRules, err := s.d.Mapping.List(ctx)
	if err != nil {
		return &sessionAPIError{
			Status:  500,
			Code:    "LIST_FAILED",
			Message: "internal error",
		}
	}
	if len(ids) != len(allRules) {
		return &sessionAPIError{
			Status:  400,
			Code:    "BAD_IDS",
			Message: "submitted IDs must contain all existing rule IDs",
			Details: map[string]any{"field": "ids", "expected": len(allRules), "got": len(ids)},
		}
	}
	for _, rule := range allRules {
		if _, ok := seen[rule.ID]; !ok {
			return &sessionAPIError{
				Status:  400,
				Code:    "BAD_IDS",
				Message: "submitted IDs must contain all existing rule IDs",
				Details: map[string]any{"field": "ids", "missingId": rule.ID},
			}
		}
	}

	if err := s.d.Mapping.Reorder(ctx, ids); err != nil {
		var nf mdomain.RuleNotFoundError
		if errors.As(err, &nf) {
			return &sessionAPIError{
				Status:  409,
				Code:    "BAD_IDS",
				Message: "rule was deleted by another user during reorder, please reload",
				Details: map[string]any{"field": "ids", "id": nf.ID},
			}
		}
		return &sessionAPIError{
			Status:  500,
			Code:    "REORDER_FAILED",
			Message: "internal error",
		}
	}
	s.refreshRuntime(ctx)
	return nil
}

func (s mappingAdminService) upsertRule(ctx context.Context, in mapRuleInputDTO) (mapRuleDTO, *sessionAPIError) {
	if apiErr := s.ensureRulesAvailable(); apiErr != nil {
		return mapRuleDTO{}, apiErr
	}

	oldBlob := ""
	if strings.TrimSpace(in.ID) != "" {
		if old, err := s.d.Mapping.GetByID(ctx, strings.TrimSpace(in.ID)); err == nil && old != nil {
			if old.BlobPath != nil && strings.TrimSpace(*old.BlobPath) != "" {
				oldBlob = *old.BlobPath
			}
			if in.UpdatedAt != nil && !old.UpdatedAt.Truncate(time.Second).Equal(in.UpdatedAt.Truncate(time.Second)) {
				return mapRuleDTO{}, &sessionAPIError{
					Status:  409,
					Code:    "CONFLICT",
					Message: "rule was modified by another user, please reload",
					Details: map[string]any{
						"field":           "updatedAt",
						"serverUpdatedAt": old.UpdatedAt,
						"clientUpdatedAt": *in.UpdatedAt,
					},
				}
			}
		}
	}

	rule := fromDTO(in)
	if code, msg, det, ok := validateRule(rule); !ok {
		return mapRuleDTO{}, &sessionAPIError{
			Status:  400,
			Code:    code,
			Message: msg,
			Details: det,
		}
	}
	saved, err := s.d.Mapping.Upsert(ctx, rule)
	if err != nil {
		return mapRuleDTO{}, &sessionAPIError{
			Status:  500,
			Code:    "UPSERT_FAILED",
			Message: "internal error",
		}
	}

	s.refreshRuntime(ctx)

	if oldBlob != "" && (saved.BlobPath == nil || strings.TrimSpace(*saved.BlobPath) == "" || strings.TrimSpace(*saved.BlobPath) != strings.TrimSpace(oldBlob)) {
		s.d.tryRemoveOrphanMappingBlob(ctx, oldBlob)
	}
	return toDTO(saved), nil
}

func (s mappingAdminService) deleteRule(ctx context.Context, id string) *sessionAPIError {
	if apiErr := s.ensureRulesAvailable(); apiErr != nil {
		return apiErr
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return &sessionAPIError{
			Status:  400,
			Code:    "MISSING_ID",
			Message: "",
		}
	}

	oldBlob := ""
	if old, err := s.d.Mapping.GetByID(ctx, id); err == nil && old != nil {
		if old.BlobPath != nil && strings.TrimSpace(*old.BlobPath) != "" {
			oldBlob = *old.BlobPath
		}
	}
	if err := s.d.Mapping.Delete(ctx, id); err != nil {
		return &sessionAPIError{
			Status:  500,
			Code:    "DELETE_FAILED",
			Message: err.Error(),
		}
	}
	s.refreshRuntime(ctx)
	if oldBlob != "" {
		s.d.tryRemoveOrphanMappingBlob(ctx, oldBlob)
	}
	return nil
}

func (s mappingAdminService) refreshRuntime(ctx context.Context) {
	if s.d.MapRt == nil || s.d.Mapping == nil {
		return
	}
	if rules, err := s.d.Mapping.List(ctx); err == nil {
		s.d.MapRt.Update(rules)
	}
}

func (s mappingAdminService) uploadBlob(r *http.Request) (map[string]any, *sessionAPIError) {
	s.d.CfgMu.RLock()
	uploadMaxMB := s.d.Cfg.MappingUploadMaxMB
	s.d.CfgMu.RUnlock()
	if uploadMaxMB <= 0 {
		uploadMaxMB = 20
	}
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
		return nil, &sessionAPIError{
			Status:  400,
			Code:    "BAD_MULTIPART",
			Message: err.Error(),
		}
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		return nil, &sessionAPIError{
			Status:  400,
			Code:    "NO_FILE",
			Message: "file part is required",
		}
	}
	defer file.Close()

	path, tooLarge, err := s.d.spoolMultipartFile(file, maxBytes, "map")
	if err != nil {
		if errors.Is(err, errSpoolTooLarge) || tooLarge {
			return nil, &sessionAPIError{
				Status:  413,
				Code:    "FILE_TOO_LARGE",
				Message: "file exceeds upload limit",
				Details: map[string]any{"maxBytes": maxBytes, "maxMB": uploadMaxMB, "fileName": hdr.Filename},
			}
		}
		return nil, &sessionAPIError{
			Status:  500,
			Code:    "SPOOL_FAILED",
			Message: "failed to store file",
		}
	}
	if path == "" {
		return nil, &sessionAPIError{
			Status:  500,
			Code:    "SPOOL_FAILED",
			Message: "failed to store file",
		}
	}

	ct := hdr.Header.Get("Content-Type")
	if ct == "" {
		ct = guessContentTypeByName(hdr.Filename)
	}
	return map[string]any{
		"blobPath":    path,
		"fileName":    hdr.Filename,
		"contentType": ct,
	}, nil
}
