package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"network-debugger/internal/features/tags/domain"
)

// Rate limiting constants for bulk operations
const (
	maxBulkSessions = 100 // Maximum number of sessions in a single bulk operation
	maxBulkTags     = 50  // Maximum number of tags in a single bulk operation
)

// DTOs for tags and annotations

type predefinedTagDTO struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Color        string    `json:"color"`
	Category     string    `json:"category"`
	DisplayOrder int       `json:"displayOrder"`
	CreatedAt    time.Time `json:"createdAt"`
}

type sessionTagDTO struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	TagName   string    `json:"tagName"`
	CreatedAt time.Time `json:"createdAt"`
}

type sessionAnnotationDTO struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type bulkTagOperationDTO struct {
	Operation  string   `json:"operation"` // "add" or "remove"
	SessionIDs []string `json:"sessionIds"`
	TagNames   []string `json:"tagNames"`
}

// Converters

func predefinedTagToDTO(d domain.PredefinedTag) predefinedTagDTO {
	return predefinedTagDTO{
		ID:           d.ID,
		Name:         d.Name,
		Color:        d.Color,
		Category:     d.Category,
		DisplayOrder: d.DisplayOrder,
		CreatedAt:    d.CreatedAt,
	}
}

func sessionTagToDTO(d domain.SessionTag) sessionTagDTO {
	return sessionTagDTO{
		ID:        d.ID,
		SessionID: d.SessionID,
		TagName:   d.TagName,
		CreatedAt: d.CreatedAt,
	}
}

func sessionAnnotationToDTO(d domain.SessionAnnotation) sessionAnnotationDTO {
	return sessionAnnotationDTO{
		ID:        d.ID,
		SessionID: d.SessionID,
		Key:       d.Key,
		Value:     d.Value,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// Handlers

// handlePredefinedTags handles GET/POST for predefined tags
func (d *Deps) handlePredefinedTags(w http.ResponseWriter, r *http.Request) {
	service := newTagsAdminService(d)
	switch r.Method {
	case http.MethodGet:
		items, apiErr := service.listPredefined(r.Context())
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})

	case http.MethodPost:
		var req createPredefinedTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		if apiErr := service.createPredefined(r.Context(), req); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}

		w.WriteHeader(http.StatusCreated)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handlePredefinedTagByID handles DELETE for a specific predefined tag
func (d *Deps) handlePredefinedTagByID(w http.ResponseWriter, r *http.Request) {
	// path: /_api/v1/tags/predefined/{id}
	path := strings.TrimPrefix(r.URL.Path, "/_api/v1/tags/predefined/")
	id := strings.TrimSuffix(path, "/")

	if r.Method == http.MethodDelete {
		if apiErr := newTagsAdminService(d).deletePredefined(r.Context(), id); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleSessionTags handles GET/POST/DELETE for session tags
func (d *Deps) handleSessionTags(w http.ResponseWriter, r *http.Request) {
	scoped, apiErr := parseSessionScopedResourcePath(r.URL.Path, "tags")
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
	service := newTagsAdminService(d)

	// DELETE /_api/v1/sessions/{id}/tags/{tagName}
	if scoped.ItemID != "" && r.Method == http.MethodDelete {
		if apiErr := service.removeSessionTag(r.Context(), scoped.SessionID, scoped.ItemID); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET /_api/v1/sessions/{id}/tags
	if scoped.ItemID == "" && r.Method == http.MethodGet {
		items, apiErr := service.listSessionTags(r.Context(), scoped.SessionID)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		return
	}

	// POST /_api/v1/sessions/{id}/tags
	if scoped.ItemID == "" && r.Method == http.MethodPost {
		var req sessionTagRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		if apiErr := service.addSessionTag(r.Context(), scoped.SessionID, req); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleBulkTags handles bulk tag operations
func (d *Deps) handleBulkTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req bulkTagOperationDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
		return
	}
	if apiErr := newTagsAdminService(d).bulkOperate(r.Context(), req); apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleSessionAnnotations handles GET/POST/DELETE for session annotations
func (d *Deps) handleSessionAnnotations(w http.ResponseWriter, r *http.Request) {
	scoped, apiErr := parseSessionScopedResourcePath(r.URL.Path, "annotations")
	if apiErr != nil {
		writeSessionAPIError(w, apiErr)
		return
	}
	service := newTagsAdminService(d)

	// DELETE /_api/v1/sessions/{id}/annotations/{key}
	if scoped.ItemID != "" && r.Method == http.MethodDelete {
		if apiErr := service.deleteSessionAnnotation(r.Context(), scoped.SessionID, scoped.ItemID); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET /_api/v1/sessions/{id}/annotations
	if scoped.ItemID == "" && r.Method == http.MethodGet {
		items, apiErr := service.listSessionAnnotations(r.Context(), scoped.SessionID)
		if apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		return
	}

	// POST /_api/v1/sessions/{id}/annotations
	if scoped.ItemID == "" && r.Method == http.MethodPost {
		var req sessionAnnotationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}
		if apiErr := service.upsertSessionAnnotation(r.Context(), scoped.SessionID, req); apiErr != nil {
			writeSessionAPIError(w, apiErr)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
