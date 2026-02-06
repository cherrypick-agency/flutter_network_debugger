package usecase

import (
	"context"
	"encoding/base64"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"network-debugger/internal/domain"
	idomain "network-debugger/internal/features/intercept/domain"
	"network-debugger/pkg/shared/id"
)

// EventBroadcaster abstracts monitor hub broadcasting
type EventBroadcaster interface {
	Broadcast(ev domain.MonitorEvent)
}

// MetricsCollector abstracts metrics collection
type MetricsCollector interface {
	IncInterceptsTotal(outcome, direction string)
	SetInterceptsQueue(size float64)
}

// pendingEntry stores internal state for pending decision
type pendingEntry struct {
	item      idomain.InterceptItem
	decisionR chan any
	timer     *time.Timer
}

// InterceptorManager manages intercept queue and rule matching
type InterceptorManager struct {
	mu      sync.RWMutex
	cfg     idomain.InterceptConfig
	monitor EventBroadcaster
	metrics MetricsCollector
	closed  bool

	rules []idomain.InterceptRule
	items map[string]*pendingEntry
	queue []string
}

// NewInterceptorManager creates a new InterceptorManager
func NewInterceptorManager(cfg idomain.InterceptConfig, monitor EventBroadcaster, metrics MetricsCollector) *InterceptorManager {
	return &InterceptorManager{
		cfg:     cfg,
		monitor: monitor,
		metrics: metrics,
		rules:   []idomain.InterceptRule{},
		items:   make(map[string]*pendingEntry),
		queue:   make([]string, 0, 64),
	}
}

func (m *InterceptorManager) interceptTimeout() time.Duration {
	m.mu.RLock()
	timeoutMs := m.cfg.TimeoutMs
	m.mu.RUnlock()
	timeout := time.Duration(timeoutMs) * time.Millisecond
	if timeout <= 0 {
		return 60 * time.Second
	}
	return timeout
}

// UpdateRules completely replaces the rule set and sorts by priority
func (m *InterceptorManager) UpdateRules(rules []idomain.InterceptRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]idomain.InterceptRule, len(rules))
	copy(out, rules)
	sort.Slice(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	m.rules = out
}

// UpdateConfig updates the intercept configuration
func (m *InterceptorManager) UpdateConfig(cfg idomain.InterceptConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = cfg
}

// Config returns the current intercept configuration
func (m *InterceptorManager) Config() idomain.InterceptConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg
}

// ListRules returns a copy of the current rules
func (m *InterceptorManager) ListRules() []idomain.InterceptRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]idomain.InterceptRule, len(m.rules))
	copy(cp, m.rules)
	return cp
}

// ListPending returns pending intercept items up to limit
func (m *InterceptorManager) ListPending(limit int) []idomain.InterceptItem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > len(m.queue) {
		limit = len(m.queue)
	}
	out := make([]idomain.InterceptItem, 0, limit)
	for i := 0; i < limit; i++ {
		if pe := m.items[m.queue[i]]; pe != nil {
			out = append(out, pe.item)
		}
	}
	return out
}

// PeekItem returns a snapshot of the queue element if it is still PENDING
func (m *InterceptorManager) PeekItem(itemID string) (idomain.InterceptItem, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if pe, ok := m.items[itemID]; ok && pe.item.State == idomain.StatePending {
		return pe.item, true
	}
	return idomain.InterceptItem{}, false
}

// enqueue creates a pending element and starts the timer
func (m *InterceptorManager) enqueue(it idomain.InterceptItem) *pendingEntry {
	log.Printf("[InterceptManager] enqueue: direction=%s ruleID=%s sessionID=%s itemID=%s", it.Direction, it.RuleID, it.SessionID, it.ID)
	pe := &pendingEntry{item: it, decisionR: make(chan any, 1)}

	m.mu.Lock()
	// Check closed under lock
	if m.closed {
		m.mu.Unlock()
		return nil
	}

	// Read timeout under lock
	timeout := time.Duration(m.cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	// overflow strategy
	if m.cfg.QueueMax > 0 && len(m.queue) >= m.cfg.QueueMax {
		if strings.ToLower(m.cfg.Overflow) == "drop-new" {
			log.Printf("[InterceptManager] enqueue: overflow drop-new, discarding itemID=%s", it.ID)
			if m.metrics != nil {
				m.metrics.IncInterceptsTotal("overflow_drop", string(it.Direction))
				m.metrics.SetInterceptsQueue(float64(len(m.queue)))
			}
			m.mu.Unlock()
			return nil
		}
		// auto-continue-oldest
		oldest := m.queue[0]
		log.Printf("[InterceptManager] enqueue: overflow auto-continue-oldest, evicting itemID=%s", oldest)
		m.queue = m.queue[1:]
		if old := m.items[oldest]; old != nil {
			old.timer.Stop()
			old.item.State = idomain.StateTimedOut
			select {
			case old.decisionR <- "overflow":
			default:
			}
			delete(m.items, oldest)
			if m.metrics != nil {
				m.metrics.IncInterceptsTotal("overflow_autocontinue", string(old.item.Direction))
			}
		}
	}

	// Create timer UNDER lock (time.AfterFunc is non-blocking, just schedules)
	pe.timer = time.AfterFunc(timeout, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if pe.item.State != idomain.StatePending {
			return
		}
		pe.item.State = idomain.StateTimedOut
		select {
		case pe.decisionR <- "timeout":
		default:
		}
		if m.monitor != nil {
			m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_timeout", ID: pe.item.SessionID, Ref: pe.item.ID})
		}
		if m.metrics != nil {
			m.metrics.IncInterceptsTotal("timeout", string(pe.item.Direction))
		}
	})

	m.items[it.ID] = pe
	m.queue = append(m.queue, it.ID)
	queueLen := len(m.queue)
	m.mu.Unlock()

	if m.monitor != nil {
		m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_created", ID: it.SessionID, Ref: it.ID})
	}
	if m.metrics != nil {
		m.metrics.IncInterceptsTotal("created", string(it.Direction))
		m.metrics.SetInterceptsQueue(float64(queueLen))
	}
	return pe
}

// finalize removes element from queue
func (m *InterceptorManager) finalize(itemID string, newState idomain.InterceptItemState) (*pendingEntry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	pe, ok := m.items[itemID]
	if !ok {
		return nil, false
	}
	if pe.item.State != idomain.StatePending {
		return nil, false
	}
	log.Printf("[InterceptManager] finalize: itemID=%s newState=%s", itemID, newState)
	pe.item.State = newState
	for i, qid := range m.queue {
		if qid == itemID {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			break
		}
	}
	delete(m.items, itemID)
	if pe.timer != nil {
		pe.timer.Stop()
	}
	if m.metrics != nil {
		m.metrics.SetInterceptsQueue(float64(len(m.queue)))
	}
	return pe, true
}

// InterceptRequest -- main blocking request interception
func (m *InterceptorManager) InterceptRequest(ctx context.Context, sessionID string, input idomain.RequestMatchInput, capBody []byte, contentType string) (*idomain.HTTPRequestDecision, error) {
	if m == nil {
		return nil, nil
	}

	var matched *idomain.InterceptRule
	var bodyMaxBytes int
	m.mu.Lock()
	if !m.cfg.Enabled || !m.cfg.Requests || m.closed {
		m.mu.Unlock()
		return nil, nil
	}
	bodyMaxBytes = m.cfg.BodyMaxBytes
	for i := range m.rules {
		ru := &m.rules[i]
		if !ru.Enabled {
			continue
		}
		if ru.Action != "request" && ru.Action != "both" {
			continue
		}
		if ru.When.MatchesRequest(input) {
			matched = ru
			if ru.StopProcessing {
				break
			}
		}
	}
	if matched != nil && matched.Once {
		matched.Enabled = false
	}
	m.mu.Unlock()
	if matched == nil {
		return nil, nil
	}
	log.Printf("[InterceptManager] InterceptRequest: matched ruleID=%s sessionID=%s", matched.ID, sessionID)

	// Build URL from match input parts
	reqURL := input.Scheme + "://" + input.Host
	if input.Port != "" {
		reqURL += ":" + input.Port
	}
	reqURL += input.Path

	it := idomain.InterceptItem{
		ID:        id.New(),
		CreatedAt: time.Now().UTC(),
		Deadline:  time.Now().UTC().Add(m.interceptTimeout()),
		Direction: idomain.DirRequest,
		SessionID: sessionID,
		RuleID:    matched.ID,
		State:     idomain.StatePending,
		Req: &idomain.HTTPRequestSnapshot{
			Method:        input.Method,
			URL:           reqURL,
			Headers:       cloneHeaders(input.Headers),
			BodyBase64:    base64.StdEncoding.EncodeToString(capBody),
			BodyTruncated: len(capBody) >= bodyMaxBytes,
			ContentType:   contentType,
		},
	}
	pe := m.enqueue(it)
	if pe == nil {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v := <-pe.decisionR:
		switch d := v.(type) {
		case *idomain.HTTPRequestDecision:
			if m.monitor != nil {
				m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_applied", ID: it.SessionID, Ref: it.ID})
			}
			return d, nil
		default:
			if m.monitor != nil {
				m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_timeout", ID: it.SessionID, Ref: it.ID})
			}
			return nil, nil
		}
	}
}

// InterceptResponse -- blocking response interception
func (m *InterceptorManager) InterceptResponse(ctx context.Context, sessionID string, input idomain.ResponseMatchInput, capBody []byte, contentType string) (*idomain.HTTPResponseDecision, error) {
	if m == nil {
		return nil, nil
	}

	var matched *idomain.InterceptRule
	var bodyMaxBytes int
	m.mu.Lock()
	if !m.cfg.Enabled || !m.cfg.Responses || m.closed {
		m.mu.Unlock()
		return nil, nil
	}
	bodyMaxBytes = m.cfg.BodyMaxBytes
	for i := range m.rules {
		ru := &m.rules[i]
		if !ru.Enabled {
			continue
		}
		if ru.Action != "response" && ru.Action != "both" {
			continue
		}
		if ru.When.MatchesResponse(input) {
			matched = ru
			if ru.StopProcessing {
				break
			}
		}
	}
	if matched != nil && matched.Once {
		matched.Enabled = false
	}
	m.mu.Unlock()
	if matched == nil {
		return nil, nil
	}
	log.Printf("[InterceptManager] InterceptResponse: matched ruleID=%s sessionID=%s", matched.ID, sessionID)

	it := idomain.InterceptItem{
		ID:        id.New(),
		CreatedAt: time.Now().UTC(),
		Deadline:  time.Now().UTC().Add(m.interceptTimeout()),
		Direction: idomain.DirResponse,
		SessionID: sessionID,
		RuleID:    matched.ID,
		State:     idomain.StatePending,
		Res: &idomain.HTTPResponseSnapshot{
			Status:        input.StatusCode,
			Headers:       cloneHeaders(input.Headers),
			BodyBase64:    base64.StdEncoding.EncodeToString(capBody),
			BodyTruncated: len(capBody) >= bodyMaxBytes,
			ContentType:   contentType,
		},
	}
	pe := m.enqueue(it)
	if pe == nil {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v := <-pe.decisionR:
		switch d := v.(type) {
		case *idomain.HTTPResponseDecision:
			if m.monitor != nil {
				m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_applied", ID: it.SessionID, Ref: it.ID})
			}
			return d, nil
		default:
			if m.monitor != nil {
				m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_timeout", ID: it.SessionID, Ref: it.ID})
			}
			return nil, nil
		}
	}
}

// ContinueRequest applies decision to pending request
func (m *InterceptorManager) ContinueRequest(itemID string, d *idomain.HTTPRequestDecision) bool {
	pe, ok := m.finalize(itemID, idomain.StateApplied)
	if !ok {
		return false
	}
	select {
	case pe.decisionR <- d:
	default:
	}
	if m.metrics != nil {
		m.metrics.IncInterceptsTotal("applied", string(pe.item.Direction))
	}
	return true
}

// ContinueResponse applies decision to pending response
func (m *InterceptorManager) ContinueResponse(itemID string, d *idomain.HTTPResponseDecision) bool {
	pe, ok := m.finalize(itemID, idomain.StateApplied)
	if !ok {
		return false
	}
	select {
	case pe.decisionR <- d:
	default:
	}
	if m.metrics != nil {
		m.metrics.IncInterceptsTotal("applied", string(pe.item.Direction))
	}
	return true
}

// Cancel cancels pending item
func (m *InterceptorManager) Cancel(itemID string) bool {
	pe, ok := m.finalize(itemID, idomain.StateCanceled)
	if !ok {
		return false
	}
	select {
	case pe.decisionR <- "cancel":
	default:
	}
	if m.monitor != nil {
		m.monitor.Broadcast(domain.MonitorEvent{Type: "intercept_canceled", ID: pe.item.SessionID, Ref: pe.item.ID})
	}
	if m.metrics != nil {
		m.metrics.IncInterceptsTotal("canceled", string(pe.item.Direction))
	}
	return true
}

// Close gracefully shuts down the manager, stopping all timers and unblocking waiters
func (m *InterceptorManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	for _, pe := range m.items {
		pe.timer.Stop()
		select {
		case pe.decisionR <- "shutdown":
		default:
		}
	}
	m.items = make(map[string]*pendingEntry)
	m.queue = m.queue[:0]
}

const maxCloneHeadersSize = 256 * 1024 // 256KB total headers limit

// cloneHeaders creates a deep copy of headers map
func cloneHeaders(h map[string][]string) map[string][]string {
	if h == nil {
		return nil
	}
	cp := make(map[string][]string, len(h))
	totalSize := 0
	for k, v := range h {
		totalSize += len(k)
		for _, val := range v {
			totalSize += len(val)
		}
		if totalSize > maxCloneHeadersSize {
			break
		}
		cv := make([]string, len(v))
		copy(cv, v)
		cp[k] = cv
	}
	return cp
}
