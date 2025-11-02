package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
)

// --- DTOs ---

type composeHeaderDTO struct{ Key, Value string }
type composeQueryDTO struct{ Key, Value string }
type composeFormFieldDTO struct{ Key, Value string }
type composePartDTO struct {
	Name        string `json:"name"`
	IsFile      bool   `json:"isFile"`
	FileName    string `json:"fileName,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	Base64      string `json:"base64,omitempty"`
	Value       string `json:"value,omitempty"`
}
type composeAuthDTO struct {
	Type         string `json:"type"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	BearerToken  string `json:"bearerToken,omitempty"`
	APIKey       string `json:"apiKey,omitempty"`
	APIKeyHeader string `json:"apiKeyHeader,omitempty"`
}
type composeBodyDTO struct {
	Mode      string                `json:"mode"`
	Raw       string                `json:"raw,omitempty"`
	JSON      string                `json:"json,omitempty"`
	Form      []composeFormFieldDTO `json:"form,omitempty"`
	Multipart []composePartDTO      `json:"multipart,omitempty"`
}
type composeTemplateDTO struct {
	ID      string             `json:"id"`
	Name    string             `json:"name"`
	Method  string             `json:"method"`
	URL     string             `json:"url"`
	Headers []composeHeaderDTO `json:"headers"`
	Query   []composeQueryDTO  `json:"query"`
	Body    composeBodyDTO     `json:"body"`
	Auth    *composeAuthDTO    `json:"auth,omitempty"`
}

type composeSendRequest struct {
	Template composeTemplateDTO `json:"template"`
}

type composeSendResponse struct {
	SessionID   string                 `json:"sessionId"`
	Transaction domain.HTTPTransaction `json:"transaction"`
	RespPreview string                 `json:"respPreview"`
	RespBody    []byte                 `json:"respBody,omitempty"`
}

type composeConfigDTO struct {
	MaxUploadMB int `json:"maxUploadMB"`
}

type composeHistoryItemDTO struct {
	ID        string    `json:"id"`
	Template  string    `json:"templateId,omitempty"`
	SessionID string    `json:"sessionId,omitempty"`
	Method    string    `json:"method"`
	URL       string    `json:"url"`
	Status    int       `json:"status"`
	When      time.Time `json:"when"`
}

type composeHistoryResponse struct {
	Items []composeHistoryItemDTO `json:"items"`
	Next  string                  `json:"next,omitempty"`
}

// --- Handlers ---

func (d *Deps) handleComposeSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req composeSendRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json", nil)
		return
	}
	tpl := dtoToDomainTemplate(req.Template)
	// set timestamps for new templates
	if tpl.CreatedAt.IsZero() {
		tpl.CreatedAt = time.Now().UTC()
	}
	tpl.UpdatedAt = time.Now().UTC()

	res, err := d.Compose.Send(r.Context(), usecase.ComposeSendCmd{Template: tpl})
	if err != nil {
		writeError(w, http.StatusBadGateway, "SEND_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(composeSendResponse{SessionID: res.SessionID, Transaction: res.Transaction, RespPreview: res.RespPreview, RespBody: res.RespBody})
}

func (d *Deps) handleComposeGetLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	lib, err := d.Compose.LoadLibrary(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "LIB_LOAD_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lib)
}

// GET /_api/v1/compose/config -> returns runtime config for compose UI
func (d *Deps) handleComposeConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(composeConfigDTO{MaxUploadMB: d.Cfg.ComposeMaxUploadMB})
}

// GET /_api/v1/compose/history?limit=20
func (d *Deps) handleComposeHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	items, next, err := d.Compose.ListHistory(r.Context(), "", limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "HISTORY_FAILED", err.Error(), nil)
		return
	}
	out := make([]composeHistoryItemDTO, 0, len(items))
	for _, it := range items {
		out = append(out, composeHistoryItemDTO{ID: it.ID, Template: it.Template, SessionID: it.Session, Method: it.Method, URL: it.URL, Status: it.Status, When: it.When})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(composeHistoryResponse{Items: out, Next: next})
}

// POST/PUT /_api/v1/compose/library/collections -> upsert collection
func (d *Deps) handleComposeUpsertCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var col domain.ComposeCollection
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&col); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json", nil)
		return
	}
	lib, err := d.Compose.UpsertCollection(r.Context(), col)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPSERT_COLLECTION_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lib)
}

// DELETE /_api/v1/compose/library/collections/{id}
func (d *Deps) handleComposeDeleteCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/_api/v1/compose/library/collections/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing id", nil)
		return
	}
	lib, err := d.Compose.DeleteCollection(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_COLLECTION_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lib)
}

// POST /_api/v1/compose/library/requests_move/{id}
// Body: { "collectionId": "...", "folderId": "...", "index": 0 }
func (d *Deps) handleComposeMoveRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/_api/v1/compose/library/requests_move/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing id", nil)
		return
	}
	var body struct {
		CollectionID string `json:"collectionId"`
		FolderID     string `json:"folderId"`
		Index        int    `json:"index"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json", nil)
		return
	}
	lib, err := d.Compose.MoveRequest(r.Context(), id, body.CollectionID, body.FolderID, body.Index)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "MOVE_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lib)
}

// POST /_api/v1/compose/library/requests -> upsert (create/update by id)
func (d *Deps) handleComposeUpsertRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var tpl composeTemplateDTO
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&tpl); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid json", nil)
		return
	}
	dom := dtoToDomainTemplate(tpl)
	dom.UpdatedAt = time.Now().UTC()
	if dom.CreatedAt.IsZero() {
		dom.CreatedAt = time.Now().UTC()
	}
	lib, err := d.Compose.UpsertRequest(r.Context(), dom)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "UPSERT_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lib)
}

// DELETE /_api/v1/compose/library/requests/{id}
func (d *Deps) handleComposeDeleteRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/_api/v1/compose/library/requests/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing id", nil)
		return
	}
	lib, err := d.Compose.DeleteRequest(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(lib)
}

// --- mapping ---

func dtoToDomainTemplate(x composeTemplateDTO) domain.ComposeRequestTemplate {
	hdrs := make([]domain.ComposeHeader, 0, len(x.Headers))
	for _, h := range x.Headers {
		hdrs = append(hdrs, domain.ComposeHeader{Key: h.Key, Value: h.Value})
	}
	qs := make([]domain.ComposeQueryParam, 0, len(x.Query))
	for _, q := range x.Query {
		qs = append(qs, domain.ComposeQueryParam{Key: q.Key, Value: q.Value})
	}
	var auth *domain.ComposeAuth
	if x.Auth != nil {
		auth = &domain.ComposeAuth{Type: x.Auth.Type, Username: x.Auth.Username, Password: x.Auth.Password, BearerToken: x.Auth.BearerToken, APIKey: x.Auth.APIKey, APIKeyHeader: x.Auth.APIKeyHeader}
	}
	body := domain.ComposeBody{Mode: domain.ComposeBodyMode(x.Body.Mode)}
	body.Raw = x.Body.Raw
	body.JSON = x.Body.JSON
	if len(x.Body.Form) > 0 {
		body.Form = make([]domain.ComposeFormField, 0, len(x.Body.Form))
		for _, f := range x.Body.Form {
			body.Form = append(body.Form, domain.ComposeFormField{Key: f.Key, Value: f.Value})
		}
	}
	if len(x.Body.Multipart) > 0 {
		body.Multipart = make([]domain.ComposeMultipartPart, 0, len(x.Body.Multipart))
		for _, p := range x.Body.Multipart {
			body.Multipart = append(body.Multipart, domain.ComposeMultipartPart{Name: p.Name, IsFile: p.IsFile, FileName: p.FileName, ContentType: p.ContentType, Base64: p.Base64, Value: p.Value})
		}
	}
	return domain.ComposeRequestTemplate{
		ID: x.ID, Name: x.Name, Method: strings.ToUpper(x.Method), URL: x.URL,
		Headers: hdrs, Query: qs, Body: body, Auth: auth,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
}
