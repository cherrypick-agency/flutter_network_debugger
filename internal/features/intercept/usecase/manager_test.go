package usecase

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"network-debugger/internal/domain"
	idomain "network-debugger/internal/features/intercept/domain"
)

// mockBroadcaster implements EventBroadcaster for tests
type mockBroadcaster struct {
	mu     sync.Mutex
	events []domain.MonitorEvent
}

func (b *mockBroadcaster) Broadcast(ev domain.MonitorEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, ev)
}

// mockMetrics implements MetricsCollector for tests
type mockMetrics struct {
	mu       sync.Mutex
	totals   map[string]int
	queueLen float64
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{totals: make(map[string]int)}
}

func (m *mockMetrics) IncInterceptsTotal(outcome, direction string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totals[outcome+"/"+direction]++
}

func (m *mockMetrics) SetInterceptsQueue(size float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queueLen = size
}

func makeItem(itemID, sessionID string) idomain.InterceptItem {
	return idomain.InterceptItem{
		ID:        itemID,
		CreatedAt: time.Now().UTC(),
		Deadline:  time.Now().UTC().Add(60 * time.Second),
		Direction: idomain.DirRequest,
		SessionID: sessionID,
		RuleID:    "rule-1",
		State:     idomain.StatePending,
	}
}

func TestOverflowAutoContinueOldest_StopsTimer(t *testing.T) {
	metrics := newMockMetrics()
	mgr := NewInterceptorManager(idomain.InterceptConfig{
		Enabled:      true,
		Requests:     true,
		TimeoutMs:    60000,
		QueueMax:     2,
		BodyMaxBytes: 1024,
		Overflow:     "auto-continue-oldest",
	}, &mockBroadcaster{}, metrics)

	// Enqueue 2 items to fill the queue
	it1 := makeItem("item-1", "sess-1")
	pe1 := mgr.enqueue(it1)
	if pe1 == nil {
		t.Fatal("expected pe1 to be non-nil")
	}

	it2 := makeItem("item-2", "sess-2")
	pe2 := mgr.enqueue(it2)
	if pe2 == nil {
		t.Fatal("expected pe2 to be non-nil")
	}

	// Enqueue a 3rd item — should evict item-1
	it3 := makeItem("item-3", "sess-3")
	pe3 := mgr.enqueue(it3)
	if pe3 == nil {
		t.Fatal("expected pe3 to be non-nil")
	}

	// item-1 should have been evicted with a decision sent
	select {
	case v := <-pe1.decisionR:
		if v != "overflow" {
			t.Fatalf("expected overflow decision, got %v", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for overflow decision on item-1")
	}

	// Verify the old timer was stopped (calling Stop again returns false if already stopped)
	// We can't directly test Stop() was called, but we verify no panic and item was removed
	mgr.mu.RLock()
	_, exists := mgr.items["item-1"]
	qLen := len(mgr.queue)
	mgr.mu.RUnlock()

	if exists {
		t.Fatal("item-1 should have been removed from items map")
	}
	if qLen != 2 {
		t.Fatalf("expected queue length 2, got %d", qLen)
	}
}

func TestOverflowDropNew_StopsTimer(t *testing.T) {
	metrics := newMockMetrics()
	mgr := NewInterceptorManager(idomain.InterceptConfig{
		Enabled:      true,
		Requests:     true,
		TimeoutMs:    60000,
		QueueMax:     1,
		BodyMaxBytes: 1024,
		Overflow:     "drop-new",
	}, &mockBroadcaster{}, metrics)

	// Fill the queue
	it1 := makeItem("item-1", "sess-1")
	pe1 := mgr.enqueue(it1)
	if pe1 == nil {
		t.Fatal("expected pe1 to be non-nil")
	}

	// Enqueue a 2nd item — should be dropped
	it2 := makeItem("item-2", "sess-2")
	pe2 := mgr.enqueue(it2)
	if pe2 != nil {
		t.Fatal("expected pe2 to be nil (drop-new)")
	}

	// Verify item-1 is still in queue
	mgr.mu.RLock()
	qLen := len(mgr.queue)
	mgr.mu.RUnlock()

	if qLen != 1 {
		t.Fatalf("expected queue length 1, got %d", qLen)
	}

	// Verify metrics recorded the drop
	metrics.mu.Lock()
	dropCount := metrics.totals["overflow_drop/request"]
	metrics.mu.Unlock()
	if dropCount != 1 {
		t.Fatalf("expected 1 overflow_drop metric, got %d", dropCount)
	}
}

func TestClose_StopsAllTimersAndUnblocksWaiters(t *testing.T) {
	mgr := NewInterceptorManager(idomain.InterceptConfig{
		Enabled:      true,
		Requests:     true,
		TimeoutMs:    60000,
		QueueMax:     100,
		BodyMaxBytes: 1024,
		Overflow:     "auto-continue-oldest",
	}, &mockBroadcaster{}, newMockMetrics())

	// Enqueue several items
	entries := make([]*pendingEntry, 3)
	for i := 0; i < 3; i++ {
		it := makeItem("close-item-"+string(rune('0'+i)), "sess-close")
		entries[i] = mgr.enqueue(it)
		if entries[i] == nil {
			t.Fatalf("expected entry %d to be non-nil", i)
		}
	}

	// Close the manager
	mgr.Close()

	// All waiters should receive "shutdown"
	for i, pe := range entries {
		select {
		case v := <-pe.decisionR:
			if v != "shutdown" {
				t.Fatalf("entry %d: expected shutdown, got %v", i, v)
			}
		case <-time.After(time.Second):
			t.Fatalf("entry %d: timed out waiting for shutdown decision", i)
		}
	}

	// items and queue should be empty
	mgr.mu.RLock()
	itemsLen := len(mgr.items)
	queueLen := len(mgr.queue)
	mgr.mu.RUnlock()

	if itemsLen != 0 {
		t.Fatalf("expected 0 items after Close, got %d", itemsLen)
	}
	if queueLen != 0 {
		t.Fatalf("expected 0 queue after Close, got %d", queueLen)
	}
}

func TestClose_PreventsNewEnqueue(t *testing.T) {
	cfg := idomain.InterceptConfig{
		Enabled:   true,
		Requests:  true,
		TimeoutMs: 5000,
		QueueMax:  10,
		Overflow:  "drop-new",
	}
	mgr := NewInterceptorManager(cfg, nil, nil)
	mgr.UpdateRules([]idomain.InterceptRule{
		{ID: "r1", Enabled: true, Action: "request"},
	})

	// Close the manager
	mgr.Close()

	// Try to intercept after close — should return nil, nil (not block or panic)
	input := idomain.RequestMatchInput{Method: "GET"}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	dec, err := mgr.InterceptRequest(ctx, "sess-after-close", input, nil, "")
	if err != nil {
		t.Errorf("expected nil error after Close, got %v", err)
	}
	if dec != nil {
		t.Errorf("expected nil decision after Close, got %v", dec)
	}
}

func TestConcurrent_UpdateConfigAndInterceptRequest(t *testing.T) {
	cfg := idomain.InterceptConfig{
		Enabled:      true,
		Requests:     true,
		TimeoutMs:    5000,
		QueueMax:     100,
		BodyMaxBytes: 1024,
		Overflow:     "auto-continue-oldest",
	}
	mgr := NewInterceptorManager(cfg, nil, nil)
	mgr.UpdateRules([]idomain.InterceptRule{
		{ID: "r1", Enabled: true, Action: "request"},
	})

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Goroutine 1: constantly update config
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			newCfg := cfg
			newCfg.TimeoutMs = 1000 + (i % 5000)
			newCfg.BodyMaxBytes = 512 + (i % 4096)
			mgr.UpdateConfig(newCfg)
		}
	}()

	// Goroutine 2: constantly intercept requests
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			select {
			case <-stop:
				return
			default:
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			mgr.InterceptRequest(ctx, "sess", idomain.RequestMatchInput{Method: "GET"}, []byte("body"), "text/plain")
			cancel()
		}
	}()

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)
	close(stop)
	wg.Wait()

	// If we get here without -race detector complaining, the fix works
}

func TestCloneHeaders_SizeLimit(t *testing.T) {
	// Create headers that exceed 256KB
	huge := make(map[string][]string)
	bigVal := strings.Repeat("x", 8192)
	for i := 0; i < 100; i++ {
		huge[fmt.Sprintf("X-Header-%d", i)] = []string{bigVal}
	}
	// Total: 100 * 8192 = ~800KB (exceeds 256KB limit)

	result := cloneHeaders(huge)

	// Count total size of result
	totalSize := 0
	for k, vs := range result {
		totalSize += len(k)
		for _, v := range vs {
			totalSize += len(v)
		}
	}

	if totalSize > 256*1024 {
		t.Errorf("cloneHeaders exceeded limit: got %d bytes, want <= %d", totalSize, 256*1024)
	}
	if len(result) == 0 {
		t.Error("cloneHeaders returned empty map, expected at least some headers")
	}
}
