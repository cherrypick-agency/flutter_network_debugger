package httpapi

import (
	"encoding/json"
	"fmt"
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
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		tags, err := d.TagsSvc.ListPredefinedTags(ctx)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "TAGS_LIST_FAILED", err.Error(), nil)
			return
		}

		dtos := make([]predefinedTagDTO, 0, len(tags))
		for _, tag := range tags {
			dtos = append(dtos, predefinedTagToDTO(tag))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": dtos})

	case http.MethodPost:
		var req struct {
			Name         string `json:"name"`
			Color        string `json:"color"`
			Category     string `json:"category"`
			DisplayOrder int    `json:"displayOrder"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}

		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "name is required", nil)
			return
		}
		if len(req.Name) > 100 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tag name exceeds maximum length of 100 characters", nil)
			return
		}
		if req.Color == "" {
			req.Color = "#808080"
		}
		if req.Category == "" {
			req.Category = "general"
		}

		err := d.TagsSvc.CreatePredefinedTag(ctx, req.Name, req.Color, req.Category, req.DisplayOrder)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "TAG_CREATE_FAILED", err.Error(), nil)
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

	if id == "" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "tag id required", nil)
		return
	}

	if r.Method == http.MethodDelete {
		err := d.TagsSvc.DeletePredefinedTag(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "TAG_DELETE_FAILED", err.Error(), nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleSessionTags handles GET/POST/DELETE for session tags
func (d *Deps) handleSessionTags(w http.ResponseWriter, r *http.Request) {
	// path: /_api/v1/sessions/{id}/tags[/{tagName}]
	path := strings.TrimPrefix(r.URL.Path, "/_api/v1/sessions/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[1] != "tags" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "invalid path", nil)
		return
	}

	sessionID := parts[0]
	ctx := r.Context()

	// DELETE /_api/v1/sessions/{id}/tags/{tagName}
	if len(parts) == 3 && r.Method == http.MethodDelete {
		tagName := parts[2]
		err := d.TagsSvc.RemoveTagFromSession(ctx, sessionID, tagName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "TAG_REMOVE_FAILED", err.Error(), nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET /_api/v1/sessions/{id}/tags
	if len(parts) == 2 && r.Method == http.MethodGet {
		tags, err := d.TagsSvc.GetSessionTags(ctx, sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "TAGS_GET_FAILED", err.Error(), nil)
			return
		}

		dtos := make([]sessionTagDTO, 0, len(tags))
		for _, tag := range tags {
			dtos = append(dtos, sessionTagToDTO(tag))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": dtos})
		return
	}

	// POST /_api/v1/sessions/{id}/tags
	if len(parts) == 2 && r.Method == http.MethodPost {
		var req struct {
			TagName string `json:"tagName"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}

		if req.TagName == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tagName is required", nil)
			return
		}
		if len(req.TagName) > 100 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tag name exceeds maximum length of 100 characters", nil)
			return
		}

		err := d.TagsSvc.AddTagToSession(ctx, sessionID, req.TagName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "TAG_ADD_FAILED", err.Error(), nil)
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

	if len(req.SessionIDs) == 0 || len(req.TagNames) == 0 {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "sessionIds and tagNames required", nil)
		return
	}

	// Rate limiting: check maximum allowed items
	if len(req.SessionIDs) > maxBulkSessions {
		writeError(w, http.StatusBadRequest, "TOO_MANY_SESSIONS",
			fmt.Sprintf("maximum %d sessions allowed in a single bulk operation", maxBulkSessions), nil)
		return
	}
	if len(req.TagNames) > maxBulkTags {
		writeError(w, http.StatusBadRequest, "TOO_MANY_TAGS",
			fmt.Sprintf("maximum %d tags allowed in a single bulk operation", maxBulkTags), nil)
		return
	}

	// Validate tag name lengths
	for _, tagName := range req.TagNames {
		if len(tagName) > 100 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "tag name exceeds maximum length of 100 characters", nil)
			return
		}
	}

	ctx := r.Context()
	var err error

	switch req.Operation {
	case "add":
		err = d.TagsSvc.BulkAddTags(ctx, req.SessionIDs, req.TagNames)
	case "remove":
		err = d.TagsSvc.BulkRemoveTags(ctx, req.SessionIDs, req.TagNames)
	default:
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "operation must be 'add' or 'remove'", nil)
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "BULK_OPERATION_FAILED", err.Error(), nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleSessionAnnotations handles GET/POST/DELETE for session annotations
func (d *Deps) handleSessionAnnotations(w http.ResponseWriter, r *http.Request) {
	// path: /_api/v1/sessions/{id}/annotations[/{key}]
	path := strings.TrimPrefix(r.URL.Path, "/_api/v1/sessions/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[1] != "annotations" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "invalid path", nil)
		return
	}

	sessionID := parts[0]
	ctx := r.Context()

	// DELETE /_api/v1/sessions/{id}/annotations/{key}
	if len(parts) == 3 && r.Method == http.MethodDelete {
		key := parts[2]
		err := d.TagsSvc.DeleteAnnotation(ctx, sessionID, key)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ANNOTATION_DELETE_FAILED", err.Error(), nil)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// GET /_api/v1/sessions/{id}/annotations
	if len(parts) == 2 && r.Method == http.MethodGet {
		annotations, err := d.TagsSvc.GetSessionAnnotations(ctx, sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ANNOTATIONS_GET_FAILED", err.Error(), nil)
			return
		}

		dtos := make([]sessionAnnotationDTO, 0, len(annotations))
		for _, ann := range annotations {
			dtos = append(dtos, sessionAnnotationToDTO(ann))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": dtos})
		return
	}

	// POST /_api/v1/sessions/{id}/annotations
	if len(parts) == 2 && r.Method == http.MethodPost {
		var req struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error(), nil)
			return
		}

		if req.Key == "" {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "key is required", nil)
			return
		}
		if len(req.Key) > 255 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "key exceeds maximum length of 255 characters", nil)
			return
		}
		if len(req.Value) > 10000 {
			writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "value exceeds maximum length of 10000 characters", nil)
			return
		}

		err := d.TagsSvc.UpsertAnnotation(ctx, sessionID, req.Key, req.Value)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ANNOTATION_UPSERT_FAILED", err.Error(), nil)
			return
		}

		w.WriteHeader(http.StatusCreated)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
