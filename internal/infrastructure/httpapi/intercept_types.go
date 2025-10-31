package httpapi

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// InterceptDirection описывает направление перехвата
type InterceptDirection string

const (
	DirRequest  InterceptDirection = "request"
	DirResponse InterceptDirection = "response"
)

// RuleStringMatch — простая модель сопоставления строковых полей
type RuleStringMatch struct {
	Equals   string   `json:"equals,omitempty"`
	Prefix   string   `json:"prefix,omitempty"`
	Suffix   string   `json:"suffix,omitempty"`
	Contains string   `json:"contains,omitempty"`
	AnyOf    []string `json:"anyOf,omitempty"`
	Regex    string   `json:"regex,omitempty"`
}

// RuleHeaderMatch — сопоставление заголовков
type RuleHeaderMatch struct {
	Name  RuleStringMatch `json:"name"`
	Value RuleStringMatch `json:"value,omitempty"`
}

// RuleStatusMatch — диапазоны статусов (для ответа)
type RuleStatusMatch struct {
	Equals []int `json:"equals,omitempty"`
	From   int   `json:"from,omitempty"`
	To     int   `json:"to,omitempty"`
	Is4xx  bool  `json:"is4xx,omitempty"`
	Is5xx  bool  `json:"is5xx,omitempty"`
}

// InterceptRule — правило перехвата (MVP)
type InterceptRule struct {
	ID             string        `json:"id"`
	Enabled        bool          `json:"enabled"`
	Priority       int           `json:"priority"`
	Action         string        `json:"action"` // request|response|both
	Once           bool          `json:"once"`
	StopProcessing bool          `json:"stopProcessing"`
	When           InterceptWhen `json:"when"`
}

// InterceptWhen — условия (все AND в MVP)
type InterceptWhen struct {
	Method       []string         `json:"method,omitempty"`
	Scheme       []string         `json:"scheme,omitempty"`
	Host         *RuleStringMatch `json:"host,omitempty"`
	Port         *RuleStringMatch `json:"port,omitempty"`
	Path         *RuleStringMatch `json:"path,omitempty"`
	ContentType  *RuleStringMatch `json:"contentType,omitempty"`
	Status       *RuleStatusMatch `json:"responseStatus,omitempty"`
	Header       *RuleHeaderMatch `json:"header,omitempty"`
	BodyContains string           `json:"bodyContains,omitempty"`
}

// InterceptItemState — состояние элемента перехвата
type InterceptItemState string

const (
	StatePending  InterceptItemState = "PENDING"
	StateApplied  InterceptItemState = "APPLIED"
	StateCanceled InterceptItemState = "CANCELED"
	StateTimedOut InterceptItemState = "TIMED_OUT"
)

// InterceptItem — единица ожидания решения пользователя
type InterceptItem struct {
	ID        string             `json:"id"`
	CreatedAt time.Time          `json:"createdAt"`
	Deadline  time.Time          `json:"deadline"`
	Direction InterceptDirection `json:"direction"`
	SessionID string             `json:"sessionId"`
	RuleID    string             `json:"ruleId,omitempty"`

	// Снимки
	Req *HTTPRequestSnapshot  `json:"req,omitempty"`
	Res *HTTPResponseSnapshot `json:"res,omitempty"`

	State InterceptItemState `json:"state"`
}

// HTTPRequestSnapshot — часть, которую можно показать/редактировать
type HTTPRequestSnapshot struct {
	Method        string      `json:"method"`
	URL           string      `json:"url"`
	Headers       http.Header `json:"headers"`
	BodyBase64    string      `json:"bodyBase64,omitempty"`
	BodyTruncated bool        `json:"bodyTruncated,omitempty"`
	ContentType   string      `json:"contentType,omitempty"`
}

// HTTPResponseSnapshot — аналогично для ответа
type HTTPResponseSnapshot struct {
	Status        int         `json:"status"`
	Headers       http.Header `json:"headers"`
	BodyBase64    string      `json:"bodyBase64,omitempty"`
	BodyTruncated bool        `json:"bodyTruncated,omitempty"`
	ContentType   string      `json:"contentType,omitempty"`
}

// Решения пользователя
type HTTPRequestDecision struct {
	Action  string      `json:"action"` // continue|drop
	Method  string      `json:"method,omitempty"`
	URL     string      `json:"url,omitempty"`
	Headers http.Header `json:"headers,omitempty"`
	Body    []byte      `json:"-"`
}

type HTTPResponseDecision struct {
	Action  string      `json:"action"` // continue|drop
	Status  int         `json:"status,omitempty"`
	Headers http.Header `json:"headers,omitempty"`
	Body    []byte      `json:"-"`
}

// match helpers
func (m *RuleStringMatch) matches(s string) bool {
	if m == nil {
		return true
	}
	if m.Equals != "" && s == m.Equals {
		return true
	}
	if m.Prefix != "" && strings.HasPrefix(s, m.Prefix) {
		return true
	}
	if m.Suffix != "" && strings.HasSuffix(s, m.Suffix) {
		return true
	}
	if m.Contains != "" && strings.Contains(s, m.Contains) {
		return true
	}
	if len(m.AnyOf) > 0 {
		for _, v := range m.AnyOf {
			if s == v {
				return true
			}
		}
	}
	if m.Regex != "" {
		if rx, err := regexp.Compile(m.Regex); err == nil {
			return rx.MatchString(s)
		}
	}
	return m.Equals == "" && m.Prefix == "" && m.Suffix == "" && m.Contains == "" && len(m.AnyOf) == 0 && m.Regex == ""
}

func isEmptyRuleStringMatch(m RuleStringMatch) bool {
	return m.Equals == "" && m.Prefix == "" && m.Suffix == "" && m.Contains == "" && len(m.AnyOf) == 0 && m.Regex == ""
}

func (h *RuleHeaderMatch) matches(headers http.Header) bool {
	if h == nil {
		return true
	}
	for name, values := range headers {
		if !h.Name.matches(name) {
			continue
		}
		if isEmptyRuleStringMatch(h.Value) {
			return true
		}
		for _, v := range values {
			if h.Value.matches(v) {
				return true
			}
		}
	}
	return false
}

func (s *RuleStatusMatch) matches(code int) bool {
	if s == nil {
		return true
	}
	if len(s.Equals) > 0 {
		for _, c := range s.Equals {
			if c == code {
				return true
			}
		}
	}
	if s.From > 0 || s.To > 0 {
		lo := s.From
		hi := s.To
		if lo == 0 {
			lo = 100
		}
		if hi == 0 {
			hi = 599
		}
		if code >= lo && code <= hi {
			return true
		}
	}
	if s.Is4xx && code >= 400 && code <= 499 {
		return true
	}
	if s.Is5xx && code >= 500 && code <= 599 {
		return true
	}
	return len(s.Equals) == 0 && !s.Is4xx && !s.Is5xx && s.From == 0 && s.To == 0
}

// requestMatch проверяет условие для запроса
func (w *InterceptWhen) requestMatch(r *http.Request, bodyPreview string) bool {
	if w == nil {
		return true
	}
	if len(w.Method) > 0 {
		ok := false
		m := strings.ToUpper(r.Method)
		for _, mm := range w.Method {
			if strings.ToUpper(mm) == m {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if len(w.Scheme) > 0 {
		s := ""
		if r.URL != nil {
			s = strings.ToLower(r.URL.Scheme)
		}
		ok := false
		for _, sc := range w.Scheme {
			if strings.ToLower(sc) == s {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if w.Host != nil {
		host := ""
		if r.URL != nil {
			host = r.URL.Hostname()
		}
		if !w.Host.matches(host) {
			return false
		}
	}
	if w.Port != nil {
		port := ""
		if r.URL != nil {
			port = r.URL.Port()
		}
		if !w.Port.matches(port) {
			return false
		}
	}
	if w.Path != nil {
		p := ""
		if r.URL != nil {
			p = r.URL.EscapedPath()
		}
		if !w.Path.matches(p) {
			return false
		}
	}
	if w.ContentType != nil {
		ct := strings.ToLower(r.Header.Get("Content-Type"))
		if !w.ContentType.matches(ct) {
			return false
		}
	}
	if w.Header != nil && !w.Header.matches(r.Header) {
		return false
	}
	if w.BodyContains != "" && !strings.Contains(bodyPreview, w.BodyContains) {
		return false
	}
	return true
}

// responseMatch проверяет условие для ответа
func (w *InterceptWhen) responseMatch(resp *http.Response, bodyPreview string) bool {
	if w == nil {
		return true
	}
	if w.Status != nil && !w.Status.matches(resp.StatusCode) {
		return false
	}
	if w.ContentType != nil {
		ct := strings.ToLower(resp.Header.Get("Content-Type"))
		if !w.ContentType.matches(ct) {
			return false
		}
	}
	if w.Header != nil && !w.Header.matches(resp.Header) {
		return false
	}
	if w.BodyContains != "" && !strings.Contains(bodyPreview, w.BodyContains) {
		return false
	}
	return true
}

// urlString безопасно возвращает строку URL для снапшота
func urlString(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
