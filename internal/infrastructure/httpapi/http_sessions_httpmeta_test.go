package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	mem "network-debugger/internal/adapters/storage/memory"
	"network-debugger/internal/domain"
	cfgpkg "network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	uc "network-debugger/internal/usecase"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestV1Session_ViewEnrichment_HTTPMetaAndSizes(t *testing.T) {
	store := mem.NewStore(100, 100, 0)
	s := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	d := &Deps{Logger: &logger, Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	// create session and append tx + frames with previews
	sid := "sx"
	_ = s.Create(contextWithNoCancel(), domain.Session{ID: sid, StartedAt: time.Unix(0, 0).UTC()})
	tx := domain.HTTPTransaction{ID: "t1", SessionID: sid, Method: http.MethodOptions, URL: "http://x", Status: http.StatusNotModified, ReqSize: 10, RespSize: 20, ContentType: "application/json", Timings: domain.HTTPTimings{Total: 5}}
	_ = s.AddHTTPTransaction(contextWithNoCancel(), tx)
	// request/response frames
	reqPrev := `{"type":"http_request","headers":{"Origin":"https://a","Access-Control-Request-Method":"GET","Access-Control-Request-Headers":"X-Id"}}`
	respPrev := `{"type":"http_response","headers":{"Access-Control-Allow-Origin":"https://a","Access-Control-Allow-Methods":"GET","Access-Control-Allow-Headers":"X-Id","Cache-Control":"max-age=60","ETag":"W/\"1\"","Age":"10"}}`
	_ = s.AddFrame(contextWithNoCancel(), sid, domain.Frame{ID: "fr1", Ts: time.Now().UTC(), Direction: "upstream->client", Opcode: domain.OpcodeText, Size: 0, Preview: reqPrev})
	_ = s.AddFrame(contextWithNoCancel(), sid, domain.Frame{ID: "fr2", Ts: time.Now().UTC(), Direction: "upstream->client", Opcode: domain.OpcodeText, Size: 0, Preview: respPrev})

	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/"+sid, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 get enriched")
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if _, ok := body["httpMeta"].(map[string]any); !ok {
		t.Fatalf("httpMeta expected")
	}
	if sz, ok := body["sizes"].(map[string]any); !ok || int(sz["requestBytes"].(float64)) != 10 {
		t.Fatalf("sizes expected: %+v", sz)
	}
}

func TestV1Session_ViewEnrichment_CORSMethodReason(t *testing.T) {
	store := mem.NewStore(100, 100, 0)
	s := uc.NewSessionService(store, store, store)
	logger := zerolog.New(io.Discard)
	d := &Deps{Logger: &logger, Cfg: cfgpkg.Config{CORSAllowOrigin: "*"}, Metrics: obs.NewMetrics(), Monitor: NewMonitorHub(), Live: NewLiveSessions(), Svc: s}
	sid := "sy"
	_ = s.Create(contextWithNoCancel(), domain.Session{ID: sid, StartedAt: time.Unix(0, 0).UTC()})
	tx := domain.HTTPTransaction{ID: "t1", SessionID: sid, Method: http.MethodGet, URL: "http://x", Status: 200, ReqSize: 1, RespSize: 1, Timings: domain.HTTPTimings{Total: 1}}
	_ = s.AddHTTPTransaction(contextWithNoCancel(), tx)
	reqPrev := `{"type":"http_request","headers":{"Origin":"https://a"}}`
	respPrev := `{"type":"http_response","headers":{"Access-Control-Allow-Origin":"https://a","Access-Control-Allow-Methods":"POST"}}`
	_ = s.AddFrame(contextWithNoCancel(), sid, domain.Frame{ID: "fr1", Ts: time.Now().UTC(), Direction: "upstream->client", Opcode: domain.OpcodeText, Preview: reqPrev})
	_ = s.AddFrame(contextWithNoCancel(), sid, domain.Frame{ID: "fr2", Ts: time.Now().UTC(), Direction: "upstream->client", Opcode: domain.OpcodeText, Preview: respPrev})
	h := NewRouterWithoutForwardProxy(d)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/_api/v1/sessions/"+sid, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 {
		t.Fatalf("v1 get enriched cors")
	}
}
