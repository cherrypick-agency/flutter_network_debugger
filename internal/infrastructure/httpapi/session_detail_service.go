package httpapi

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/domain"
)

const maxSessionFrameBodyBytes = 100 * 1024 * 1024

type sessionAPIError struct {
	Status  int
	Code    string
	Message string
	Details any
}

type sessionPage[T any] struct {
	Items []T
	Next  string
}

type legacyEventView struct {
	ID          string    `json:"id"`
	Ts          time.Time `json:"ts"`
	Namespace   string    `json:"namespace"`
	Event       string    `json:"event"`
	Name        string    `json:"name"`
	AckID       *int64    `json:"ackId,omitempty"`
	ArgsPreview string    `json:"argsPreview"`
}

type sessionFrameBody struct {
	Data        []byte
	ContentType string
	Source      string
	FrameID     string
}

type sessionDetailService struct {
	d *Deps
}

func newSessionDetailService(d *Deps) sessionDetailService {
	return sessionDetailService{d: d}
}

func (s sessionDetailService) getLegacySession(ctx context.Context, sessionID string) (domain.Session, bool, *sessionAPIError) {
	session, ok, err := s.d.Svc.Get(ctx, sessionID)
	if err != nil {
		return domain.Session{}, false, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "SESSION_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{"id": sessionID},
		}
	}
	if !ok {
		return domain.Session{}, false, &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "session not found",
			Details: map[string]any{"id": sessionID},
		}
	}
	return session, true, nil
}

func (s sessionDetailService) getProjectedSession(ctx context.Context, sessionID string) (*sessionV1, bool, *sessionAPIError) {
	session, ok, err := s.d.Svc.Get(ctx, sessionID)
	if err != nil {
		return nil, false, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "SESSION_GET_FAILED",
			Message: err.Error(),
		}
	}
	if !ok {
		return nil, false, &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "session not found",
			Details: map[string]any{"id": sessionID},
		}
	}
	view := newSessionProjector(s.d).buildView(ctx, session)
	return &view, true, nil
}

func (s sessionDetailService) deleteProjectedSession(ctx context.Context, sessionID string) {
	if s.d.TagsSvc != nil {
		_ = s.d.TagsSvc.DeleteAllSessionTags(ctx, sessionID)
		_ = s.d.TagsSvc.DeleteAllAnnotations(ctx, sessionID)
	}
	_ = s.d.Svc.Delete(ctx, sessionID)
}

func (s sessionDetailService) listFrames(ctx context.Context, sessionID, from string, limit int) (sessionPage[domain.Frame], *sessionAPIError) {
	items, next, err := s.d.Svc.ListFrames(ctx, sessionID, from, limit)
	if err != nil {
		return sessionPage[domain.Frame]{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "FRAMES_LIST_FAILED",
			Message: err.Error(),
			Details: map[string]any{"id": sessionID},
		}
	}
	return sessionPage[domain.Frame]{Items: items, Next: next}, nil
}

func (s sessionDetailService) listLegacyEvents(ctx context.Context, sessionID, from string, limit int) (sessionPage[legacyEventView], *sessionAPIError) {
	events, apiErr := s.listEvents(ctx, sessionID, from, limit)
	if apiErr != nil {
		return sessionPage[legacyEventView]{}, apiErr
	}
	items := make([]legacyEventView, 0, len(events.Items))
	for _, event := range events.Items {
		items = append(items, legacyEventView{
			ID:          event.ID,
			Ts:          event.Ts,
			Namespace:   event.Namespace,
			Event:       event.Name,
			Name:        event.Name,
			AckID:       event.AckID,
			ArgsPreview: event.ArgsPreview,
		})
	}
	return sessionPage[legacyEventView]{Items: items, Next: events.Next}, nil
}

func (s sessionDetailService) listEvents(ctx context.Context, sessionID, from string, limit int) (sessionPage[domain.Event], *sessionAPIError) {
	items, next, err := s.d.Svc.ListEvents(ctx, sessionID, from, limit)
	if err != nil {
		return sessionPage[domain.Event]{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "EVENTS_LIST_FAILED",
			Message: err.Error(),
			Details: map[string]any{"id": sessionID},
		}
	}
	return sessionPage[domain.Event]{Items: items, Next: next}, nil
}

func (s sessionDetailService) listHTTPTransactions(ctx context.Context, sessionID, from string, limit int) (sessionPage[domain.HTTPTransaction], *sessionAPIError) {
	items, next, err := s.d.Svc.ListHTTPTransactions(ctx, sessionID, from, limit)
	if err != nil {
		return sessionPage[domain.HTTPTransaction]{}, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "HTTP_LIST_FAILED",
			Message: err.Error(),
			Details: map[string]any{"id": sessionID},
		}
	}
	return sessionPage[domain.HTTPTransaction]{Items: items, Next: next}, nil
}

func (s sessionDetailService) frameBody(ctx context.Context, sessionID, frameID string) (*sessionFrameBody, *sessionAPIError) {
	frame, found, err := s.d.Svc.GetFrameByID(ctx, sessionID, frameID)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "FRAME_GET_FAILED",
			Message: err.Error(),
			Details: map[string]any{"sessionId": sessionID, "frameId": frameID},
		}
	}
	if !found {
		return nil, &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "FRAME_NOT_FOUND",
			Message: "frame not found",
			Details: map[string]any{"frameId": frameID},
		}
	}
	if frame.BodyFile == "" {
		return &sessionFrameBody{
			Data:        []byte(frame.Preview),
			ContentType: "text/plain; charset=utf-8",
			Source:      "preview",
		}, nil
	}

	fileInfo, err := os.Stat(frame.BodyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &sessionAPIError{
				Status:  http.StatusGone,
				Code:    "BODY_FILE_EXPIRED",
				Message: "Body file no longer available (cleaned up)",
				Details: map[string]any{"frameId": frameID},
			}
		}
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "BODY_STAT_FAILED",
			Message: err.Error(),
			Details: map[string]any{"frameId": frameID, "bodyFile": frame.BodyFile},
		}
	}
	if fileInfo.Size() >= maxSessionFrameBodyBytes {
		return nil, &sessionAPIError{
			Status:  http.StatusRequestEntityTooLarge,
			Code:    "BODY_TOO_LARGE",
			Message: fmt.Sprintf("body file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileInfo.Size(), maxSessionFrameBodyBytes),
			Details: map[string]any{"frameId": frameID, "fileSize": fileInfo.Size(), "maxSize": maxSessionFrameBodyBytes},
		}
	}

	bodyData, err := os.ReadFile(frame.BodyFile)
	if err != nil {
		return nil, &sessionAPIError{
			Status:  http.StatusInternalServerError,
			Code:    "BODY_READ_FAILED",
			Message: err.Error(),
			Details: map[string]any{"frameId": frameID, "bodyFile": frame.BodyFile},
		}
	}

	return &sessionFrameBody{
		Data:        bodyData,
		ContentType: "application/octet-stream",
		Source:      "file",
		FrameID:     frameID,
	}, nil
}

func (s sessionDetailService) writeLegacyExport(w http.ResponseWriter, r *http.Request, sessionID string) *sessionAPIError {
	session, ok, apiErr := s.getLegacySession(r.Context(), sessionID)
	if apiErr != nil {
		return apiErr
	}
	if !ok {
		return &sessionAPIError{
			Status:  http.StatusNotFound,
			Code:    "NOT_FOUND",
			Message: "session not found",
			Details: map[string]any{"id": sessionID},
		}
	}

	useGzip := false
	if raw := r.URL.Query().Get("gzip"); raw == "1" || raw == "true" {
		useGzip = true
	} else if strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip") {
		useGzip = true
	}

	var outWriter interface{ Write([]byte) (int, error) }
	var gz *gzip.Writer

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=network-debugger_session_"+sessionID+".json")
	if useGzip {
		w.Header().Set("Content-Encoding", "gzip")
		gz = gzip.NewWriter(w)
		outWriter = gz
	} else {
		outWriter = w
	}

	flush := func() {
		if gz != nil {
			_ = gz.Flush()
		}
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	write := func(data []byte) { _, _ = outWriter.Write(data) }

	payload, _ := json.Marshal(session)
	write([]byte("{\"session\":"))
	write(payload)
	write([]byte(",\"frames\":["))
	s.streamLegacyFrames(write, flush, r.Context(), sessionID)
	write([]byte("],\"events\":["))
	s.streamLegacyEvents(write, flush, r.Context(), sessionID)
	write([]byte("]}"))

	if gz != nil {
		_ = gz.Close()
	}
	return nil
}

func (s sessionDetailService) streamLegacyFrames(write func([]byte), flush func(), ctx context.Context, sessionID string) {
	from := ""
	first := true
	for {
		page, apiErr := s.listFrames(ctx, sessionID, from, 1000)
		if apiErr != nil {
			break
		}
		for _, frame := range page.Items {
			payload, _ := json.Marshal(frame)
			if !first {
				write([]byte(","))
			} else {
				first = false
			}
			write(payload)
		}
		flush()
		if page.Next == "" {
			break
		}
		from = page.Next
	}
}

func (s sessionDetailService) streamLegacyEvents(write func([]byte), flush func(), ctx context.Context, sessionID string) {
	from := ""
	first := true
	for {
		page, apiErr := s.listEvents(ctx, sessionID, from, 1000)
		if apiErr != nil {
			break
		}
		for _, event := range page.Items {
			payload, _ := json.Marshal(event)
			if !first {
				write([]byte(","))
			} else {
				first = false
			}
			write(payload)
		}
		flush()
		if page.Next == "" {
			break
		}
		from = page.Next
	}
}

func writeSessionAPIError(w http.ResponseWriter, apiErr *sessionAPIError) {
	if apiErr == nil {
		return
	}
	writeError(w, apiErr.Status, apiErr.Code, apiErr.Message, apiErr.Details)
}

func writeSessionFrameBody(w http.ResponseWriter, body *sessionFrameBody) {
	if body == nil {
		return
	}
	w.Header().Set("Content-Type", body.ContentType)
	w.Header().Set("X-Body-Source", body.Source)
	if body.Source == "file" {
		w.Header().Set("X-Frame-Id", body.FrameID)
		w.Header().Set("Content-Length", strconv.Itoa(len(body.Data)))
	}
	if _, err := w.Write(body.Data); err != nil {
		log.Printf("Error writing session frame body response: %v", err)
	}
}
