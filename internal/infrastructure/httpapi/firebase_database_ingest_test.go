package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"

	"github.com/rs/zerolog"
)

func newFirebaseIngestRouter(t *testing.T) (http.Handler, *uc.SessionService) {
	t.Helper()
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", IngestAllowRemote: true, PreviewMaxBytes: 1000, WSBodyMaxBytes: 4096},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	return NewRouterWithoutForwardProxy(deps), svc
}

func TestFirebaseIngest_HappyPath_WithBodyAndClose(t *testing.T) {
	h, svc := newFirebaseIngestRouter(t)

	body := base64.StdEncoding.EncodeToString([]byte(`{"full":"payload"}`))
	reqBody := map[string]any{
		"session": map[string]any{
			"id":        "fb-test-1",
			"target":    "https://demo-default-rtdb.firebaseio.com/users/123",
			"captureId": "current",
		},
		"frames": []map[string]any{
			{
				"id":           "fr-1",
				"direction":    "client->upstream",
				"opcode":       "text",
				"preview":      `{"type":"firebase_database","op":"set","path":"/users/123"}`,
				"body":         body,
				"bodyEncoding": "base64",
			},
		},
		"close": true,
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	sess, ok, err := svc.Get(context.Background(), "fb-test-1")
	if err != nil || !ok {
		t.Fatalf("session not found: ok=%v err=%v", ok, err)
	}
	if sess.Kind != firebaseSessionKind {
		t.Fatalf("kind=%s want=%s", sess.Kind, firebaseSessionKind)
	}
	if sess.ClosedAt == nil {
		t.Fatal("expected closed session")
	}

	frames, _, err := svc.ListFrames(context.Background(), "fb-test-1", "", 10)
	if err != nil {
		t.Fatalf("ListFrames err=%v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("len(frames)=%d want=1", len(frames))
	}
	if frames[0].BodyFile == "" {
		t.Fatal("expected body file to be stored")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/fb-test-1/frames/fr-1/body", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("body status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"full":"payload"`) {
		t.Fatalf("unexpected body content: %s", rr.Body.String())
	}
}

func TestFirebaseIngest_BadJSON(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", strings.NewReader("{not-json"))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"BAD_JSON"`) {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}

func TestFirebaseIngest_ConflictWithOtherKind(t *testing.T) {
	h, svc := newFirebaseIngestRouter(t)

	err := svc.Create(context.Background(), domain.Session{
		ID:         "same-id",
		Target:     "wss://example.com/socket",
		ClientAddr: "127.0.0.1:1234",
		StartedAt:  time.Now().UTC(),
		Kind:       "ws",
	})
	if err != nil {
		t.Fatalf("seed create err=%v", err)
	}

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "same-id",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"type":"firebase_database","op":"set"}`,
			},
		},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_FilterTypesFirebase(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-test-types",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"type":"firebase_database","op":"onValue"}`,
			},
		},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("ingest status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?types=firebase&captures=all&includeUnassigned=true", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", rr.Code, rr.Body.String())
	}

	var out struct {
		Items []struct {
			ID   string `json:"id"`
			Kind string `json:"kind"`
		} `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode err=%v body=%s", err, rr.Body.String())
	}
	if len(out.Items) == 0 {
		t.Fatalf("expected at least one item")
	}
	found := false
	for _, it := range out.Items {
		if it.ID == "fb-test-types" && it.Kind == firebaseSessionKind {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("firebase session not found in filtered list: %+v", out.Items)
	}
}

func TestFirebaseIngest_AuthRequiredWhenTokenConfigured(t *testing.T) {
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", AdminToken: "secret-token"},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(deps)

	reqBody := map[string]any{
		"session": map[string]any{"id": "fb-auth-1", "target": "https://demo-default-rtdb.firebaseio.com/users/1"},
		"frames":  []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	req.Header.Set("X-Admin-Token", "secret-token")
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_CaptureVisibilityWhenRecordingStopped(t *testing.T) {
	h, svc := newFirebaseIngestRouter(t)

	repo, ok := svc.SessionsRepoUnsafe().(*memory.Store)
	if !ok {
		t.Fatalf("unexpected repo type")
	}
	repo.StopCapture()

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-capture-1",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
			// captureId omitted -> should default to current and stay visible
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"type":"firebase_database","op":"set"}`,
			},
		},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("ingest status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/_api/v1/sessions?limit=100", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", rr.Code, rr.Body.String())
	}
	var out struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode err=%v body=%s", err, rr.Body.String())
	}
	found := false
	for _, it := range out.Items {
		if it.ID == "fb-capture-1" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("session not visible in default list: %+v", out.Items)
	}
}

func TestFirebaseIngest_DefaultDirectionFromPreviewOp(t *testing.T) {
	h, svc := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-dir-1",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"type":"firebase_database","op":"onValue"}`,
			},
		},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	frames, _, err := svc.ListFrames(context.Background(), "fb-dir-1", "", 10)
	if err != nil {
		t.Fatalf("ListFrames err=%v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("len(frames)=%d want=1", len(frames))
	}
	if frames[0].Direction != domain.DirectionUpstreamToClient {
		t.Fatalf("direction=%s want=%s", frames[0].Direction, domain.DirectionUpstreamToClient)
	}
}

func TestFirebaseIngest_AuthRemoteDeniedWithoutToken(t *testing.T) {
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", IngestAllowRemote: false},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(deps)

	reqBody := map[string]any{
		"session": map[string]any{"id": "fb-auth-remote", "target": "https://demo-default-rtdb.firebaseio.com/users/1"},
		"frames":  []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	req.RemoteAddr = "8.8.8.8:5050"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_AuthPrivateLanAllowedWithoutToken(t *testing.T) {
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", IngestAllowRemote: false},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(deps)

	reqBody := map[string]any{
		"session": map[string]any{"id": "fb-auth-lan", "target": "https://demo-default-rtdb.firebaseio.com/users/1"},
		"frames":  []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	req.RemoteAddr = "192.168.1.55:5050"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_AuthRemoteAllowedWhenFlagEnabled(t *testing.T) {
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", IngestAllowRemote: true},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(deps)

	reqBody := map[string]any{
		"session": map[string]any{"id": "fb-auth-remote-ok", "target": "https://demo-default-rtdb.firebaseio.com/users/1"},
		"frames":  []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
	}
	raw, _ := json.Marshal(reqBody)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	req.RemoteAddr = "8.8.8.8:5050"
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_InvalidBodyEncoding(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-bad-encoding",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":           "fr-1",
				"preview":      `{"op":"set"}`,
				"body":         "abc",
				"bodyEncoding": "hex",
			},
		},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_OversizeBodyReturns413(t *testing.T) {
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", IngestAllowRemote: true, WSBodyMaxBytes: 4},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(deps)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-big-body",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"op":"set"}`,
				"body":    "12345",
			},
		},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_ConflictWhenSameKindButDifferentTarget(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	first := map[string]any{
		"session": map[string]any{
			"id":     "fb-same-id-different-target",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
	}
	raw1, _ := json.Marshal(first)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw1))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("first status=%d body=%s", rr.Code, rr.Body.String())
	}

	second := map[string]any{
		"session": map[string]any{
			"id":     "fb-same-id-different-target",
			"target": "https://another-default-rtdb.firebaseio.com/users/999",
		},
		"frames": []map[string]any{{"id": "fr-2", "preview": `{"op":"set"}`}},
	}
	raw2, _ := json.Marshal(second)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw2))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_ConflictWhenSessionClosed(t *testing.T) {
	h, svc := newFirebaseIngestRouter(t)

	seed := map[string]any{
		"session": map[string]any{
			"id":     "fb-closed-conflict",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
		"close":  true,
	}
	rawSeed, _ := json.Marshal(seed)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(rawSeed))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("seed status=%d body=%s", rr.Code, rr.Body.String())
	}

	sess, ok, err := svc.Get(context.Background(), "fb-closed-conflict")
	if err != nil || !ok || sess.ClosedAt == nil {
		t.Fatalf("seed session invalid: ok=%v err=%v closed=%v", ok, err, sess.ClosedAt != nil)
	}

	second := map[string]any{
		"session": map[string]any{
			"id":     "fb-closed-conflict",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{{"id": "fr-2", "preview": `{"op":"set"}`}},
	}
	rawSecond, _ := json.Marshal(second)
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(rawSecond))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_InvalidTarget(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-invalid-target",
			"target": "not-a-url",
		},
		"frames": []map[string]any{{"id": "fr-1", "preview": `{"op":"set"}`}},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestFirebaseIngest_PreviewRedactedAndTrimmed(t *testing.T) {
	store := memory.NewStore(100, 1000, 0)
	svc := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	deps := &Deps{
		Logger:  &logger,
		Cfg:     cfgpkg.Config{CORSAllowOrigin: "*", IngestAllowRemote: true, PreviewMaxBytes: 48},
		Metrics: obs.NewMetrics(),
		Monitor: NewMonitorHub(),
		Live:    NewLiveSessions(),
		Svc:     svc,
	}
	h := NewRouterWithoutForwardProxy(deps)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-preview-sanitize",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"authorization":"Bearer super-secret-token","op":"set","path":"/users/123","value":{"k":"v"}}`,
			},
		},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	frames, _, err := svc.ListFrames(context.Background(), "fb-preview-sanitize", "", 10)
	if err != nil || len(frames) != 1 {
		t.Fatalf("frames err=%v len=%d", err, len(frames))
	}
	if len(frames[0].Preview) > 48 {
		t.Fatalf("preview len=%d want<=48", len(frames[0].Preview))
	}
}

func TestFirebaseIngest_IdempotentDuplicateFrameID(t *testing.T) {
	h, svc := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-idempotent-frames",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{
				"id":      "fr-1",
				"preview": `{"op":"set","path":"/users/123"}`,
			},
		},
	}
	raw, _ := json.Marshal(reqBody)
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNoContent {
			t.Fatalf("iter=%d status=%d body=%s", i, rr.Code, rr.Body.String())
		}
	}

	frames, _, err := svc.ListFrames(context.Background(), "fb-idempotent-frames", "", 10)
	if err != nil {
		t.Fatalf("ListFrames err=%v", err)
	}
	if len(frames) != 1 {
		t.Fatalf("len(frames)=%d want=1", len(frames))
	}
	sess, ok, err := svc.Get(context.Background(), "fb-idempotent-frames")
	if err != nil || !ok {
		t.Fatalf("Get session ok=%v err=%v", ok, err)
	}
	if sess.Frames.Total != 1 {
		t.Fatalf("frames.total=%d want=1", sess.Frames.Total)
	}
}

func TestFirebaseIngest_TooManyFramesValidation(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	frames := make([]map[string]any, 0, firebaseIngestMaxFrames+1)
	for i := 0; i < firebaseIngestMaxFrames+1; i++ {
		frames = append(frames, map[string]any{
			"id":      "fr-" + strconv.Itoa(i),
			"preview": `{"op":"set"}`,
		})
	}
	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-too-many-frames",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": frames,
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"VALIDATION"`) {
		t.Fatalf("unexpected body=%s", rr.Body.String())
	}
}

func TestFirebaseIngest_InvalidCaptureIDValidation(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":        "fb-invalid-capture",
			"target":    "https://demo-default-rtdb.firebaseio.com/users/123",
			"captureId": "abc",
		},
		"frames": []map[string]any{
			{"id": "fr-1", "preview": `{"op":"set"}`},
		},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"VALIDATION"`) {
		t.Fatalf("unexpected body=%s", rr.Body.String())
	}
}

func TestFirebaseIngest_FrameIDTooLongValidation(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	reqBody := map[string]any{
		"session": map[string]any{
			"id":     "fb-long-frame-id",
			"target": "https://demo-default-rtdb.firebaseio.com/users/123",
		},
		"frames": []map[string]any{
			{"id": strings.Repeat("f", 129), "preview": `{"op":"set"}`},
		},
	}
	raw, _ := json.Marshal(reqBody)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", bytes.NewReader(raw))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"VALIDATION"`) {
		t.Fatalf("unexpected body=%s", rr.Body.String())
	}
}

func TestFirebaseIngest_RejectsTrailingJSON(t *testing.T) {
	h, _ := newFirebaseIngestRouter(t)

	payload := `{"session":{"id":"fb-trailing","target":"https://demo-default-rtdb.firebaseio.com/users/123"},"frames":[{"id":"fr-1","preview":"{\"op\":\"set\"}"}]} {"extra":1}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/_api/v1/ingest/firebase_database", strings.NewReader(payload))
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"BAD_JSON"`) {
		t.Fatalf("unexpected body=%s", rr.Body.String())
	}
}
