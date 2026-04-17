package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type createPredefinedTagRequest struct {
	Name         string `json:"name"`
	Color        string `json:"color"`
	Category     string `json:"category"`
	DisplayOrder int    `json:"displayOrder"`
}

type sessionTagRequest struct {
	TagName string `json:"tagName"`
}

type sessionAnnotationRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type sessionScopedResourcePath struct {
	SessionID string
	ItemID    string
}

type tagsAdminService struct {
	d *Deps
}

func newTagsAdminService(d *Deps) tagsAdminService {
	return tagsAdminService{d: d}
}

func (s tagsAdminService) ensureAvailable() *sessionAPIError {
	if s.d.TagsSvc == nil {
		return &sessionAPIError{
			Status:  http.StatusServiceUnavailable,
			Code:    "FEATURE_DISABLED",
			Message: "tags feature not available",
		}
	}
	return nil
}

func (s tagsAdminService) listPredefined(ctx context.Context) ([]predefinedTagDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return nil, apiErr
	}
	tags, err := s.d.TagsSvc.ListPredefinedTags(ctx)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "TAGS_LIST_FAILED",
			Message: err.Error(),
		}
	}
	dtos := make([]predefinedTagDTO, 0, len(tags))
	for _, tag := range tags {
		dtos = append(dtos, predefinedTagToDTO(tag))
	}
	return dtos, nil
}

func (s tagsAdminService) createPredefined(ctx context.Context, in createPredefinedTagRequest) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Color = strings.TrimSpace(in.Color)
	in.Category = strings.TrimSpace(in.Category)
	if in.Name == "" {
		return invalidRequestError("name is required")
	}
	if len(in.Name) > 100 {
		return invalidRequestError("tag name exceeds maximum length of 100 characters")
	}
	if in.Color == "" {
		in.Color = "#808080"
	}
	if in.Category == "" {
		in.Category = "general"
	}
	if err := s.d.TagsSvc.CreatePredefinedTag(ctx, in.Name, in.Color, in.Category, in.DisplayOrder); err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "TAG_CREATE_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s tagsAdminService) deletePredefined(ctx context.Context, id string) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if strings.TrimSpace(id) == "" {
		return &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "tag id required",
		}
	}
	if err := s.d.TagsSvc.DeletePredefinedTag(ctx, id); err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "TAG_DELETE_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s tagsAdminService) listSessionTags(ctx context.Context, sessionID string) ([]sessionTagDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return nil, apiErr
	}
	tags, err := s.d.TagsSvc.GetSessionTags(ctx, sessionID)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "TAGS_GET_FAILED",
			Message: err.Error(),
		}
	}
	dtos := make([]sessionTagDTO, 0, len(tags))
	for _, tag := range tags {
		dtos = append(dtos, sessionTagToDTO(tag))
	}
	return dtos, nil
}

func (s tagsAdminService) addSessionTag(ctx context.Context, sessionID string, in sessionTagRequest) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	in.TagName = strings.TrimSpace(in.TagName)
	if in.TagName == "" {
		return invalidRequestError("tagName is required")
	}
	if len(in.TagName) > 100 {
		return invalidRequestError("tag name exceeds maximum length of 100 characters")
	}
	if err := s.d.TagsSvc.AddTagToSession(ctx, sessionID, in.TagName); err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "TAG_ADD_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s tagsAdminService) removeSessionTag(ctx context.Context, sessionID, tagName string) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if err := s.d.TagsSvc.RemoveTagFromSession(ctx, sessionID, tagName); err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "TAG_REMOVE_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s tagsAdminService) bulkOperate(ctx context.Context, req bulkTagOperationDTO) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if len(req.SessionIDs) == 0 || len(req.TagNames) == 0 {
		return invalidRequestError("sessionIds and tagNames required")
	}
	if len(req.SessionIDs) > maxBulkSessions {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "TOO_MANY_SESSIONS",
			Message: fmt.Sprintf("maximum %d sessions allowed in a single bulk operation", maxBulkSessions),
		}
	}
	if len(req.TagNames) > maxBulkTags {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "TOO_MANY_TAGS",
			Message: fmt.Sprintf("maximum %d tags allowed in a single bulk operation", maxBulkTags),
		}
	}
	for _, tagName := range req.TagNames {
		if len(tagName) > 100 {
			return invalidRequestError("tag name exceeds maximum length of 100 characters")
		}
	}

	var err error
	switch req.Operation {
	case "add":
		err = s.d.TagsSvc.BulkAddTags(ctx, req.SessionIDs, req.TagNames)
	case "remove":
		err = s.d.TagsSvc.BulkRemoveTags(ctx, req.SessionIDs, req.TagNames)
	default:
		return invalidRequestError("operation must be 'add' or 'remove'")
	}
	if err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "BULK_OPERATION_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s tagsAdminService) listSessionAnnotations(ctx context.Context, sessionID string) ([]sessionAnnotationDTO, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return nil, apiErr
	}
	annotations, err := s.d.TagsSvc.GetSessionAnnotations(ctx, sessionID)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "ANNOTATIONS_GET_FAILED",
			Message: err.Error(),
		}
	}
	dtos := make([]sessionAnnotationDTO, 0, len(annotations))
	for _, ann := range annotations {
		dtos = append(dtos, sessionAnnotationToDTO(ann))
	}
	return dtos, nil
}

func (s tagsAdminService) upsertSessionAnnotation(ctx context.Context, sessionID string, in sessionAnnotationRequest) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	in.Key = strings.TrimSpace(in.Key)
	if in.Key == "" {
		return invalidRequestError("key is required")
	}
	if len(in.Key) > 255 {
		return invalidRequestError("key exceeds maximum length of 255 characters")
	}
	if len(in.Value) > 10000 {
		return invalidRequestError("value exceeds maximum length of 10000 characters")
	}
	if err := s.d.TagsSvc.UpsertAnnotation(ctx, sessionID, in.Key, in.Value); err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "ANNOTATION_UPSERT_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func (s tagsAdminService) deleteSessionAnnotation(ctx context.Context, sessionID, key string) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if err := s.d.TagsSvc.DeleteAnnotation(ctx, sessionID, key); err != nil {
		return &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "ANNOTATION_DELETE_FAILED",
			Message: err.Error(),
		}
	}
	return nil
}

func parseSessionScopedResourcePath(path, resource string) (sessionScopedResourcePath, *sessionAPIError) {
	trimmed := strings.TrimPrefix(path, "/_api/v1/sessions/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 || parts[1] != resource {
		return sessionScopedResourcePath{}, &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "invalid path",
		}
	}
	out := sessionScopedResourcePath{SessionID: parts[0]}
	if len(parts) > 2 {
		out.ItemID = parts[2]
	}
	return out, nil
}

func invalidRequestError(message string) *sessionAPIError {
	return &sessionAPIError{
		Status:  http.StatusBadRequest,
		Code:    "INVALID_REQUEST",
		Message: message,
	}
}
