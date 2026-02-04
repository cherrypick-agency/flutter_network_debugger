package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
)

func testDepsForMgr() (*InterceptorManager, *config.Config) {
	cfg := config.Config{
		InterceptEnabled:      true,
		InterceptRequests:     true,
		InterceptResponses:    true,
		InterceptTimeoutMs:    50,
		InterceptQueueMax:     10,
		InterceptBodyMaxBytes: 1024,
		InterceptReencode:     false,
		InterceptOverflow:     "auto-continue-oldest",
	}
	m := NewInterceptorManager(&cfg, NewMonitorHub(), obs.NewMetrics())
	return m, &cfg
}

func TestInterceptorManager_TimeoutAutoContinue(t *testing.T) {
	mgr, _ := testDepsForMgr()
	// rule: any GET
	mgr.UpdateRules([]InterceptRule{{Enabled: true, Priority: 1, Action: "request", When: InterceptWhen{Method: []string{"GET"}}}})
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	done := make(chan struct{})
	go func() {
		dec, err := mgr.InterceptRequest(context.Background(), "sess1", req, "", nil, "")
		if err != nil {
			t.Errorf("err: %v", err)
		}
		if dec != nil {
			t.Errorf("expected auto-continue (nil decision), got %#v", dec)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for auto-continue")
	}
}

func TestInterceptorManager_ContinueRequest(t *testing.T) {
	mgr, _ := testDepsForMgr()
	mgr.UpdateRules([]InterceptRule{{Enabled: true, Priority: 1, Action: "request", When: InterceptWhen{Method: []string{"POST"}}}})
	req, _ := http.NewRequest(http.MethodPost, "http://example.com/", nil)
	done := make(chan *HTTPRequestDecision)
	go func() {
		dec, err := mgr.InterceptRequest(context.Background(), "sess2", req, "", nil, "")
		if err != nil {
			t.Errorf("err: %v", err)
		}
		done <- dec
	}()
	// wait for pending to appear
	var itemID string
	for i := 0; i < 50; i++ {
		items := mgr.ListPending(10)
		if len(items) > 0 {
			itemID = items[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if itemID == "" {
		t.Fatal("no pending item appeared")
	}
	ok := mgr.ContinueRequest(itemID, &HTTPRequestDecision{Action: "continue", Method: "PUT"})
	if !ok {
		t.Fatal("ContinueRequest returned false")
	}
	select {
	case dec := <-done:
		if dec == nil || dec.Method != "PUT" {
			t.Fatalf("unexpected decision: %#v", dec)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting decision")
	}
}

func TestInterceptorManager_Cancel(t *testing.T) {
	mgr, _ := testDepsForMgr()
	mgr.UpdateRules([]InterceptRule{{Enabled: true, Priority: 1, Action: "request", When: InterceptWhen{Method: []string{"GET"}}}})
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	done := make(chan *HTTPRequestDecision)
	go func() {
		dec, _ := mgr.InterceptRequest(context.Background(), "sess3", req, "", nil, "")
		done <- dec
	}()
	var itemID string
	for i := 0; i < 50; i++ {
		items := mgr.ListPending(10)
		if len(items) > 0 {
			itemID = items[0].ID
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if itemID == "" {
		t.Fatal("no pending item")
	}
	if !mgr.Cancel(itemID) {
		t.Fatal("cancel returned false")
	}
	select {
	case dec := <-done:
		if dec != nil {
			t.Fatalf("expected nil decision after cancel, got %#v", dec)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting cancel result")
	}
}

func TestInterceptorManager_OverflowPolicy(t *testing.T) {
	mgr, cfg := testDepsForMgr()
	cfg.InterceptQueueMax = 1
	mgr.UpdateRules([]InterceptRule{{Enabled: true, Priority: 1, Action: "request", When: InterceptWhen{Method: []string{"GET"}}}})
	req1, _ := http.NewRequest(http.MethodGet, "http://a/", nil)
	req2, _ := http.NewRequest(http.MethodGet, "http://b/", nil)
	// first will hang
	go func() { _, _ = mgr.InterceptRequest(context.Background(), "sessA", req1, "", nil, "") }()
	// wait for queue
	for i := 0; i < 50; i++ {
		if len(mgr.ListPending(10)) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// second call will trigger auto-continue-oldest
	done := make(chan struct{})
	go func() { _, _ = mgr.InterceptRequest(context.Background(), "sessB", req2, "", nil, ""); close(done) }()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("second intercept did not return")
	}
	if n := len(mgr.ListPending(10)); n != 1 {
		t.Fatalf("expected pending size 1, got %d", n)
	}
}
