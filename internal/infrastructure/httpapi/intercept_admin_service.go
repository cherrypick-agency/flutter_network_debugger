package httpapi

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	interceptdomain "network-debugger/internal/features/intercept/domain"
)

type interceptContinueRequestPayload struct {
	Action  string              `json:"action"`
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	BodyB64 *string             `json:"bodyBase64"`
}

type interceptContinueResponsePayload struct {
	Action  string              `json:"action"`
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	BodyB64 *string             `json:"bodyBase64"`
}

type interceptItemPath struct {
	ID     string
	Action string
}

type interceptAdminService struct {
	d *Deps
}

func newInterceptAdminService(d *Deps) interceptAdminService {
	return interceptAdminService{d: d}
}

func (s interceptAdminService) ensureAvailable() *sessionAPIError {
	if s.d.InterceptSvc == nil {
		return &sessionAPIError{
			Status:  http.StatusServiceUnavailable,
			Code:    "FEATURE_DISABLED",
			Message: "intercept feature not available",
		}
	}
	return nil
}

func (s interceptAdminService) listRules() ([]interceptdomain.InterceptRule, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return nil, apiErr
	}
	return s.d.InterceptSvc.ListRules(), nil
}

func (s interceptAdminService) loadConfig() (interceptdomain.InterceptConfig, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return interceptdomain.InterceptConfig{}, apiErr
	}
	return s.d.InterceptSvc.Manager().Config(), nil
}

func (s interceptAdminService) saveRules(ctx context.Context, rules []interceptdomain.InterceptRule) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if err := s.d.InterceptSvc.UpdateRules(ctx, rules); err != nil {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "VALIDATION",
			Message: err.Error(),
		}
	}
	return nil
}

func (s interceptAdminService) saveConfig(ctx context.Context, cfg interceptdomain.InterceptConfig) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	cfg.SetDefaults()
	if err := s.d.InterceptSvc.UpdateConfig(ctx, cfg); err != nil {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "VALIDATION",
			Message: err.Error(),
		}
	}
	return nil
}

func (s interceptAdminService) parsePendingLimit(raw string) (int, *sessionAPIError) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_PARAM",
			Message: "limit must be an integer",
		}
	}
	if limit < 0 {
		return 0, &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_PARAM",
			Message: "limit must be non-negative",
		}
	}
	return limit, nil
}

func (s interceptAdminService) listPending(limit int) ([]interceptdomain.InterceptItem, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return nil, apiErr
	}
	return s.d.InterceptSvc.Manager().ListPending(limit), nil
}

func (s interceptAdminService) parseItemPath(path string) (interceptItemPath, *sessionAPIError) {
	const base = "/_api/v1/intercept/items/"
	if !strings.HasPrefix(path, base) {
		return interceptItemPath{}, &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "",
		}
	}
	rest := strings.TrimPrefix(path, base)
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return interceptItemPath{}, &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_ID",
			Message: "missing id",
		}
	}
	itemPath := interceptItemPath{ID: parts[0]}
	if len(parts) > 1 {
		itemPath.Action = parts[1]
	}
	return itemPath, nil
}

func (s interceptAdminService) pendingItem(id string) (interceptdomain.InterceptItem, *sessionAPIError) {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return interceptdomain.InterceptItem{}, apiErr
	}
	item, ok := s.d.InterceptSvc.Manager().PeekItem(id)
	if !ok {
		return interceptdomain.InterceptItem{}, &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "no such pending item",
		}
	}
	return item, nil
}

func (s interceptAdminService) continueRequest(id string, payload interceptContinueRequestPayload) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if payload.URL != "" {
		u, err := url.Parse(payload.URL)
		if err != nil || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") {
			return &sessionAPIError{
				Status:  http.StatusBadRequest,
				Code:    "INVALID_URL",
				Message: "URL must be valid http or https",
			}
		}
	}
	body, apiErr := decodeInterceptBody(payload.BodyB64)
	if apiErr != nil {
		return apiErr
	}
	ok := s.d.InterceptSvc.Manager().ContinueRequest(id, &interceptdomain.HTTPRequestDecision{
		Action:  payload.Action,
		Method:  payload.Method,
		URL:     payload.URL,
		Headers: payload.Headers,
		Body:    body,
	})
	if !ok {
		return &sessionAPIError{
			Status:  http.StatusConflict,
			Code:    "CONFLICT",
			Message: "already finalized",
		}
	}
	return nil
}

func (s interceptAdminService) continueResponse(id string, payload interceptContinueResponsePayload) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if payload.Status != 0 && (payload.Status < 100 || payload.Status > 599) {
		return &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_STATUS",
			Message: "status code must be 100-599",
		}
	}
	body, apiErr := decodeInterceptBody(payload.BodyB64)
	if apiErr != nil {
		return apiErr
	}
	ok := s.d.InterceptSvc.Manager().ContinueResponse(id, &interceptdomain.HTTPResponseDecision{
		Action:  payload.Action,
		Status:  payload.Status,
		Headers: payload.Headers,
		Body:    body,
	})
	if !ok {
		return &sessionAPIError{
			Status:  http.StatusConflict,
			Code:    "CONFLICT",
			Message: "already finalized",
		}
	}
	return nil
}

func (s interceptAdminService) cancel(id string) *sessionAPIError {
	if apiErr := s.ensureAvailable(); apiErr != nil {
		return apiErr
	}
	if ok := s.d.InterceptSvc.Manager().Cancel(id); !ok {
		return &sessionAPIError{
			Status:  http.StatusConflict,
			Code:    "CONFLICT",
			Message: "already finalized",
		}
	}
	return nil
}

func decodeInterceptBody(bodyB64 *string) ([]byte, *sessionAPIError) {
	if bodyB64 == nil {
		return nil, nil
	}
	body, err := base64.StdEncoding.DecodeString(*bodyB64)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusBadRequest,
			Code:    "BAD_BASE64",
			Message: "invalid base64 body",
		}
	}
	return body, nil
}
