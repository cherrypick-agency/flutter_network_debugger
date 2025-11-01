package httpapi

import (
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"network-debugger/internal/infrastructure/config"
	obs "network-debugger/internal/infrastructure/observability"
	"network-debugger/pkg/shared/id"
)

// pendingEntry хранит внутреннее состояние ожидания решения
type pendingEntry struct {
	item      InterceptItem
	decisionR chan any // *HTTPRequestDecision | *HTTPResponseDecision | string("cancel")
	timer     *time.Timer
}

type InterceptorManager struct {
	mu      sync.RWMutex
	cfg     *config.Config
	monitor *MonitorHub
	metrics *obs.Metrics

	// Правила, отсортированные по Priority ASC
	rules []InterceptRule

	// Очередь и элементы
	items map[string]*pendingEntry
	queue []string
}

func NewInterceptorManager(cfg *config.Config, monitor *MonitorHub, metrics *obs.Metrics) *InterceptorManager {
	return &InterceptorManager{
		cfg:     cfg,
		monitor: monitor,
		metrics: metrics,
		rules:   []InterceptRule{},
		items:   make(map[string]*pendingEntry),
		queue:   make([]string, 0, 64),
	}
}

// UpdateRules полностью заменяет набор правил и сортирует по приоритету
func (m *InterceptorManager) UpdateRules(rules []InterceptRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// нормализуем и сортируем
	out := make([]InterceptRule, 0, len(rules))
	for _, r := range rules {
		if r.ID == "" {
			r.ID = id.New()
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Priority < out[j].Priority })
	m.rules = out
}

func (m *InterceptorManager) ListRules() []InterceptRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make([]InterceptRule, len(m.rules))
	copy(cp, m.rules)
	return cp
}

func (m *InterceptorManager) ListPending(limit int) []InterceptItem {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > len(m.queue) {
		limit = len(m.queue)
	}
	out := make([]InterceptItem, 0, limit)
	for i := 0; i < limit; i++ {
		if pe := m.items[m.queue[i]]; pe != nil {
			out = append(out, pe.item)
		}
	}
	return out
}

// PeekItem возвращает снапшот элемента в очереди, если он ещё PENDING
func (m *InterceptorManager) PeekItem(id string) (InterceptItem, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if pe, ok := m.items[id]; ok && pe.item.State == StatePending {
		return pe.item, true
	}
	return InterceptItem{}, false
}

// helper: loopback detection для простого решения auth
func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	return ip != nil && (ip.IsLoopback() || ip.IsUnspecified())
}

// enqueue создаёт pending элемент и запускает таймер
func (m *InterceptorManager) enqueue(it InterceptItem) *pendingEntry {
	pe := &pendingEntry{item: it, decisionR: make(chan any, 1)}
	timeout := time.Duration(m.cfg.InterceptTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	pe.timer = time.AfterFunc(timeout, func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if pe.item.State != StatePending {
			return
		}
		pe.item.State = StateTimedOut
		// По умолчанию — auto-continue
		select {
		case pe.decisionR <- "timeout":
		default:
		}
		if m.monitor != nil {
			m.monitor.Broadcast(MonitorEvent{Type: "intercept_timeout", ID: pe.item.SessionID, Ref: pe.item.ID})
		}
		if m.metrics != nil {
			m.metrics.InterceptsTotal.WithLabelValues("timeout", string(pe.item.Direction)).Inc()
		}
	})
	m.mu.Lock()
	// overflow strategy
	if m.cfg.InterceptQueueMax > 0 && len(m.queue) >= m.cfg.InterceptQueueMax {
		if strings.ToLower(m.cfg.InterceptOverflow) == "drop-new" {
			if m.metrics != nil {
				m.metrics.InterceptsTotal.WithLabelValues("overflow_drop", string(it.Direction)).Inc()
				m.metrics.InterceptsQueue.Set(float64(len(m.queue)))
			}
			m.mu.Unlock()
			return nil
		}
		// auto-continue-oldest: снимаем самый старый без вмешательства
		oldest := m.queue[0]
		m.queue = m.queue[1:]
		if old := m.items[oldest]; old != nil {
			old.item.State = StateTimedOut
			select {
			case old.decisionR <- "overflow":
			default:
			}
			delete(m.items, oldest)
			if m.metrics != nil {
				m.metrics.InterceptsTotal.WithLabelValues("overflow_autocontinue", string(old.item.Direction)).Inc()
			}
		}
	}
	m.items[it.ID] = pe
	m.queue = append(m.queue, it.ID)
	queueLen := len(m.queue)
	m.mu.Unlock()
	if m.monitor != nil {
		m.monitor.Broadcast(MonitorEvent{Type: "intercept_created", ID: it.SessionID, Ref: it.ID})
	}
	if m.metrics != nil {
		m.metrics.InterceptsTotal.WithLabelValues("created", string(it.Direction)).Inc()
		m.metrics.InterceptsQueue.Set(float64(queueLen))
	}
	return pe
}

// finalize снимает элемент из очереди
func (m *InterceptorManager) finalize(id string, newState InterceptItemState) (*pendingEntry, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	pe, ok := m.items[id]
	if !ok {
		return nil, false
	}
	if pe.item.State != StatePending {
		return nil, false
	}
	pe.item.State = newState
	// удаляем из очереди
	for i, qid := range m.queue {
		if qid == id {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			break
		}
	}
	delete(m.items, id)
	if pe.timer != nil {
		pe.timer.Stop()
	}
	if m.metrics != nil {
		m.metrics.InterceptsQueue.Set(float64(len(m.queue)))
	}
	return pe, true
}

// InterceptRequest — основной блокирующий перехват запроса
func (m *InterceptorManager) InterceptRequest(ctx context.Context, sessionID string, r *http.Request, bodyPreview string, capBody []byte, contentType string) (*HTTPRequestDecision, error) {
	if m == nil || !m.cfg.InterceptEnabled || !m.cfg.InterceptRequests {
		return nil, nil
	}
	// Матчим правило с учётом StopProcessing; Once — отключает правило после срабатывания
	var matched *InterceptRule
	m.mu.Lock()
	for i := range m.rules {
		ru := &m.rules[i]
		if !ru.Enabled {
			continue
		}
		if ru.Action != "request" && ru.Action != "both" {
			continue
		}
		if ru.When.requestMatch(r, bodyPreview) {
			matched = ru
			if ru.StopProcessing {
				break
			}
			// если StopProcessing=false — продолжаем искать ниже по приоритету
		}
	}
	if matched != nil && matched.Once {
		matched.Enabled = false
	}
	m.mu.Unlock()
	if matched == nil {
		return nil, nil
	}

	it := InterceptItem{
		ID:        id.New(),
		CreatedAt: time.Now().UTC(),
		Deadline:  time.Now().UTC().Add(time.Duration(m.cfg.InterceptTimeoutMs) * time.Millisecond),
		Direction: DirRequest,
		SessionID: sessionID,
		RuleID:    matched.ID,
		State:     StatePending,
		Req: &HTTPRequestSnapshot{
			Method:        r.Method,
			URL:           urlString(r.URL),
			Headers:       cloneHeader(r.Header),
			BodyBase64:    base64.StdEncoding.EncodeToString(capBody),
			BodyTruncated: len(capBody) >= m.cfg.InterceptBodyMaxBytes,
			ContentType:   contentType,
		},
	}
	pe := m.enqueue(it)
	if pe == nil {
		return nil, nil
	}

	// Ждём решения/таймаута/контекста
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case v := <-pe.decisionR:
		switch d := v.(type) {
		case *HTTPRequestDecision:
			if m.monitor != nil {
				m.monitor.Broadcast(MonitorEvent{Type: "intercept_applied", ID: it.SessionID, Ref: it.ID})
			}
			return d, nil
		default:
			if m.monitor != nil {
				m.monitor.Broadcast(MonitorEvent{Type: "intercept_timeout", ID: it.SessionID, Ref: it.ID})
			}
			return nil, nil // auto-continue
		}
	}
}

// InterceptResponse — блокирующий перехват ответа
func (m *InterceptorManager) InterceptResponse(ctx context.Context, sessionID string, resp *http.Response, bodyPreview string, capBody []byte, contentType string) (*HTTPResponseDecision, error) {
	if m == nil || !m.cfg.InterceptEnabled || !m.cfg.InterceptResponses {
		return nil, nil
	}
	var matched *InterceptRule
	m.mu.Lock()
	for i := range m.rules {
		ru := &m.rules[i]
		if !ru.Enabled {
			continue
		}
		if ru.Action != "response" && ru.Action != "both" {
			continue
		}
		if ru.When.responseMatch(resp, bodyPreview) {
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

	it := InterceptItem{
		ID:        id.New(),
		CreatedAt: time.Now().UTC(),
		Deadline:  time.Now().UTC().Add(time.Duration(m.cfg.InterceptTimeoutMs) * time.Millisecond),
		Direction: DirResponse,
		SessionID: sessionID,
		RuleID:    matched.ID,
		State:     StatePending,
		Res: &HTTPResponseSnapshot{
			Status:        resp.StatusCode,
			Headers:       cloneHeader(resp.Header),
			BodyBase64:    base64.StdEncoding.EncodeToString(capBody),
			BodyTruncated: len(capBody) >= m.cfg.InterceptBodyMaxBytes,
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
		case *HTTPResponseDecision:
			if m.monitor != nil {
				m.monitor.Broadcast(MonitorEvent{Type: "intercept_applied", ID: it.SessionID, Ref: it.ID})
			}
			return d, nil
		default:
			if m.monitor != nil {
				m.monitor.Broadcast(MonitorEvent{Type: "intercept_timeout", ID: it.SessionID, Ref: it.ID})
			}
			return nil, nil
		}
	}
}

// ContinueRequest применяет решение к pending запросу
func (m *InterceptorManager) ContinueRequest(id string, d *HTTPRequestDecision) bool {
	pe, ok := m.finalize(id, StateApplied)
	if !ok {
		return false
	}
	select {
	case pe.decisionR <- d:
	default:
	}
	if m.metrics != nil {
		m.metrics.InterceptsTotal.WithLabelValues("applied", string(pe.item.Direction)).Inc()
	}
	return true
}

// ContinueResponse применяет решение к pending ответу
func (m *InterceptorManager) ContinueResponse(id string, d *HTTPResponseDecision) bool {
	pe, ok := m.finalize(id, StateApplied)
	if !ok {
		return false
	}
	select {
	case pe.decisionR <- d:
	default:
	}
	if m.metrics != nil {
		m.metrics.InterceptsTotal.WithLabelValues("applied", string(pe.item.Direction)).Inc()
	}
	return true
}

// Cancel отменяет pending
func (m *InterceptorManager) Cancel(id string) bool {
	pe, ok := m.finalize(id, StateCanceled)
	if !ok {
		return false
	}
	select {
	case pe.decisionR <- "cancel":
	default:
	}
	if m.monitor != nil {
		m.monitor.Broadcast(MonitorEvent{Type: "intercept_canceled", ID: pe.item.SessionID, Ref: pe.item.ID})
	}
	if m.metrics != nil {
		m.metrics.InterceptsTotal.WithLabelValues("canceled", string(pe.item.Direction)).Inc()
	}
	return true
}
