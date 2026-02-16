package httpapi

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
	"network-debugger/pkg/shared/redact"
)

const (
	firebaseIngestMaxRequestBytes = 16 << 20 // 16MB
	firebaseIngestMaxFrames       = 200
	firebaseSessionKind           = "firebase_database"
)

type firebaseIngestRequest struct {
	Session firebaseIngestSession `json:"session"`
	Frames  []firebaseIngestFrame `json:"frames"`
	Close   bool                  `json:"close"`
	Error   string                `json:"error"`
}

type firebaseIngestSession struct {
	ID        string    `json:"id"`
	Target    string    `json:"target"`
	StartedAt time.Time `json:"startedAt"`
	CaptureID string    `json:"captureId"`
}

type firebaseIngestFrame struct {
	ID           string    `json:"id"`
	Ts           time.Time `json:"ts"`
	Direction    string    `json:"direction"`
	Opcode       string    `json:"opcode"`
	ContentType  string    `json:"contentType"`
	Preview      string    `json:"preview"`
	Body         *string   `json:"body"`
	BodyEncoding *string   `json:"bodyEncoding"`
}

func (d *Deps) handleV1IngestFirebaseDatabase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !d.firebaseIngestAuthOK(r) {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "admin token required", nil)
		return
	}
	if d.Svc == nil {
		writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "session service unavailable", nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, firebaseIngestMaxRequestBytes)
	defer r.Body.Close()

	var req firebaseIngestRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", err.Error(), nil)
		return
	}
	var trailing any
	if err := dec.Decode(&trailing); err == nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "request must contain exactly one JSON object", nil)
		return
	} else if !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "request must contain exactly one JSON object", nil)
		return
	}

	if errMsg := validateFirebaseIngestRequest(req); errMsg != "" {
		writeError(w, http.StatusBadRequest, "VALIDATION", errMsg, nil)
		return
	}

	target := strings.TrimSpace(req.Session.Target)
	u, err := url.Parse(target)
	if err != nil || u.Scheme == "" || u.Host == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION", "session.target must be a valid URL", nil)
		return
	}

	startedAt := req.Session.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	} else {
		startedAt = startedAt.UTC()
	}

	sid := strings.TrimSpace(req.Session.ID)
	existing, ok, err := d.Svc.Get(r.Context(), sid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "SESSION_GET_FAILED", err.Error(), map[string]any{"id": sid})
		return
	}
	if ok {
		if existing.Kind != firebaseSessionKind {
			writeError(w, http.StatusConflict, "CONFLICT", "session id is already used by another kind", map[string]any{"id": sid, "kind": existing.Kind})
			return
		}
		if strings.TrimSpace(existing.Target) != target {
			writeError(w, http.StatusConflict, "CONFLICT", "session id is already bound to another target", map[string]any{"id": sid})
			return
		}
		if existing.ClosedAt != nil {
			writeError(w, http.StatusConflict, "CONFLICT", "session already closed", map[string]any{"id": sid})
			return
		}
	}

	if !ok {
		sess := domain.Session{
			ID:         sid,
			Target:     target,
			ClientAddr: r.RemoteAddr,
			StartedAt:  startedAt,
			Kind:       firebaseSessionKind,
		}
		applyCaptureIDIfRequested(d, &sess, req.Session.CaptureID)
		if err := d.Svc.Create(r.Context(), sess); err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_CREATE_FAILED", err.Error(), map[string]any{"id": sid})
			return
		}
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_started", ID: sid})
	}

	for i := range req.Frames {
		freq := req.Frames[i]
		ts := freq.Ts
		if ts.IsZero() {
			ts = time.Now().UTC()
		} else {
			ts = ts.UTC()
		}

		opcode := parseFirebaseOpcode(freq.Opcode)
		if opcode == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid frame opcode", map[string]any{"frameId": freq.ID})
			return
		}
		direction := parseFirebaseDirection(freq.Direction, freq.Preview)
		if direction == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION", "invalid frame direction", map[string]any{"frameId": freq.ID})
			return
		}

		preview := sanitizeFirebasePreview(freq.Preview, d.Cfg.PreviewMaxBytes)
		if preview == "" {
			writeError(w, http.StatusBadRequest, "VALIDATION", "frame.preview must be non-empty", map[string]any{"frameId": freq.ID})
			return
		}

		bodyBytes, err := decodeFirebaseBody(freq.Body, freq.BodyEncoding, d.Cfg.WSBodyMaxBytes)
		if err != nil {
			if err == http.ErrContentLength {
				writeError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", err.Error(), map[string]any{"frameId": freq.ID})
			} else {
				writeError(w, http.StatusBadRequest, "VALIDATION", err.Error(), map[string]any{"frameId": freq.ID})
			}
			return
		}

		size := len(preview)
		if len(bodyBytes) > 0 {
			size = len(bodyBytes)
		}

		frame := domain.Frame{
			ID:        strings.TrimSpace(freq.ID),
			Ts:        ts,
			Direction: direction,
			Opcode:    opcode,
			Size:      size,
			Preview:   preview,
		}
		if _, exists, getErr := d.Svc.GetFrameByID(r.Context(), sid, frame.ID); getErr == nil && exists {
			// Повторный фрейм с тем же id считаем ретраем клиента и не дублируем.
			continue
		}
		if err := d.Svc.AddFrame(r.Context(), sid, frame); err != nil {
			writeError(w, http.StatusInternalServerError, "FRAME_ADD_FAILED", err.Error(), map[string]any{"sessionId": sid, "frameId": frame.ID})
			return
		}

		if len(bodyBytes) > 0 {
			fpath, err := d.spoolBodyBytes(bodyBytes, "firebase")
			if err != nil {
				writeError(w, http.StatusInternalServerError, "BODY_SPOOL_FAILED", err.Error(), map[string]any{"sessionId": sid, "frameId": frame.ID})
				return
			}
			if fpath != "" {
				if err := d.Svc.UpdateFrameBodyFile(r.Context(), sid, frame.ID, fpath); err != nil {
					writeError(w, http.StatusInternalServerError, "FRAME_UPDATE_FAILED", err.Error(), map[string]any{"sessionId": sid, "frameId": frame.ID})
					return
				}
				d.Svc.AddSpoolFile(r.Context(), sid, fpath)
			}
		}

		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "frame_added", ID: sid, Ref: frame.ID})
	}

	if req.Close {
		var errPtr *string
		if strings.TrimSpace(req.Error) != "" {
			msg := strings.TrimSpace(req.Error)
			errPtr = &msg
		}
		if err := d.Svc.SetClosed(r.Context(), sid, time.Now().UTC(), errPtr); err != nil {
			writeError(w, http.StatusInternalServerError, "SESSION_CLOSE_FAILED", err.Error(), map[string]any{"id": sid})
			return
		}
		d.broadcastMonitorEvent(domain.MonitorEvent{Type: "session_ended", ID: sid})
	}

	w.WriteHeader(http.StatusNoContent)
}

func validateFirebaseIngestRequest(req firebaseIngestRequest) string {
	if strings.TrimSpace(req.Session.ID) == "" {
		return "session.id is required"
	}
	if len(strings.TrimSpace(req.Session.ID)) > 128 {
		return "session.id is too long"
	}
	if strings.TrimSpace(req.Session.Target) == "" {
		return "session.target is required"
	}
	if len(strings.TrimSpace(req.Session.Target)) > 2048 {
		return "session.target is too long"
	}
	if rawCapture := strings.TrimSpace(req.Session.CaptureID); rawCapture != "" &&
		!strings.EqualFold(rawCapture, "current") {
		if _, err := strconv.Atoi(rawCapture); err != nil {
			return "session.captureId must be 'current' or integer string"
		}
	}
	if len(req.Frames) == 0 && !req.Close {
		return "frames must contain at least one item"
	}
	if len(req.Frames) > firebaseIngestMaxFrames {
		return "too many frames in one request"
	}
	for i := range req.Frames {
		f := req.Frames[i]
		if strings.TrimSpace(f.ID) == "" {
			return "frame.id is required"
		}
		if len(strings.TrimSpace(f.ID)) > 128 {
			return "frame.id is too long"
		}
		if strings.TrimSpace(f.Preview) == "" {
			return "frame.preview is required"
		}
	}
	return ""
}

func parseFirebaseOpcode(v string) domain.Opcode {
	if strings.TrimSpace(v) == "" {
		return domain.OpcodeText
	}
	switch strings.ToLower(strings.TrimSpace(v)) {
	case string(domain.OpcodeText):
		return domain.OpcodeText
	case string(domain.OpcodeBinary):
		return domain.OpcodeBinary
	case string(domain.OpcodePing):
		return domain.OpcodePing
	case string(domain.OpcodePong):
		return domain.OpcodePong
	case string(domain.OpcodeClose):
		return domain.OpcodeClose
	default:
		return ""
	}
}

func parseFirebaseDirection(v string, preview string) domain.Direction {
	if strings.TrimSpace(v) == "" {
		if inferred := inferDirectionFromPreview(preview); inferred != "" {
			return inferred
		}
		return domain.DirectionClientToUpstream
	}
	switch strings.TrimSpace(v) {
	case string(domain.DirectionClientToUpstream):
		return domain.DirectionClientToUpstream
	case string(domain.DirectionUpstreamToClient):
		return domain.DirectionUpstreamToClient
	default:
		return ""
	}
}

func inferDirectionFromPreview(preview string) domain.Direction {
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(preview)), &payload); err != nil {
		return ""
	}
	op, _ := payload["op"].(string)
	op = strings.TrimSpace(strings.ToLower(op))
	if op == "" {
		return ""
	}
	if strings.HasPrefix(op, "on") {
		return domain.DirectionUpstreamToClient
	}
	switch op {
	case "event", "listen_event", "stream_event":
		return domain.DirectionUpstreamToClient
	default:
		return domain.DirectionClientToUpstream
	}
}

func sanitizeFirebasePreview(preview string, max int) string {
	p := strings.TrimSpace(preview)
	if p == "" {
		return ""
	}
	if json.Valid([]byte(p)) {
		p = redact.RedactJSON(p)
	}
	if max > 0 && len(p) > max {
		p = p[:max]
	}
	return p
}

func decodeFirebaseBody(body *string, encoding *string, maxSize int) ([]byte, error) {
	if body == nil {
		return nil, nil
	}
	if maxSize <= 0 {
		maxSize = 4 << 20
	}
	enc := ""
	if encoding != nil {
		enc = strings.ToLower(strings.TrimSpace(*encoding))
	}
	switch enc {
	case "", "utf8", "utf-8":
		data := []byte(*body)
		if len(data) > maxSize {
			return nil, http.ErrContentLength
		}
		return data, nil
	case "base64":
		data, err := base64.StdEncoding.DecodeString(*body)
		if err != nil {
			return nil, err
		}
		if len(data) > maxSize {
			return nil, http.ErrContentLength
		}
		return data, nil
	default:
		return nil, http.ErrNotSupported
	}
}

func applyCaptureIDIfRequested(d *Deps, sess *domain.Session, raw string) {
	req := strings.TrimSpace(raw)
	if req == "" || strings.EqualFold(req, "current") {
		if repo, ok := sessionsRepoOf(d.Svc).(usecase.CaptureControlRepository); ok {
			_, current := repo.RecordingState()
			sess.CaptureID = &current
		}
		return
	}
	if n, err := strconv.Atoi(req); err == nil {
		sess.CaptureID = &n
	}
}

func (d *Deps) firebaseIngestAuthOK(r *http.Request) bool {
	if strings.TrimSpace(d.Cfg.AdminToken) != "" {
		tok := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
		return tok != "" && subtle.ConstantTimeCompare([]byte(tok), []byte(d.Cfg.AdminToken)) == 1
	}
	if d.Cfg.IngestAllowRemote {
		return true
	}
	return isLoopbackOrPrivateAddr(r.RemoteAddr)
}

func isLoopbackOrPrivateAddr(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return strings.EqualFold(host, "localhost")
	}
	if ip.IsLoopback() || ip.IsUnspecified() {
		return true
	}
	if v4 := ip.To4(); v4 != nil {
		if v4[0] == 10 {
			return true
		}
		if v4[0] == 172 && v4[1] >= 16 && v4[1] <= 31 {
			return true
		}
		if v4[0] == 192 && v4[1] == 168 {
			return true
		}
		return false
	}
	// Unique local IPv6 fc00::/7
	return len(ip) == net.IPv6len && (ip[0]&0xfe) == 0xfc
}
