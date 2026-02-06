package domain

import (
	"regexp"
	"strings"
	"sync"
)

const regexCacheMaxSize = 256

var (
	regexCacheMu sync.Mutex
	regexCacheM  = make(map[string]*regexp.Regexp, regexCacheMaxSize)
)

func cachedRegexp(pattern string) (*regexp.Regexp, error) {
	regexCacheMu.Lock()
	if rx, ok := regexCacheM[pattern]; ok {
		regexCacheMu.Unlock()
		return rx, nil
	}
	regexCacheMu.Unlock()

	rx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	regexCacheMu.Lock()
	if len(regexCacheM) >= regexCacheMaxSize {
		// evict one random entry (map iteration order is random in Go)
		for k := range regexCacheM {
			delete(regexCacheM, k)
			break
		}
	}
	regexCacheM[pattern] = rx
	regexCacheMu.Unlock()
	return rx, nil
}

// RequestMatchInput contains pure data for request matching (no net/http dependency)
type RequestMatchInput struct {
	Method      string
	Scheme      string
	Host        string
	Port        string
	Path        string
	ContentType string
	Headers     map[string][]string
	BodyPreview string
}

// ResponseMatchInput contains pure data for response matching (no net/http dependency)
type ResponseMatchInput struct {
	StatusCode  int
	ContentType string
	Headers     map[string][]string
	BodyPreview string
}

// MatchesRequest checks condition for request
func (w *InterceptWhen) MatchesRequest(in RequestMatchInput) bool {
	if w == nil {
		return true
	}
	if len(w.Method) > 0 {
		ok := false
		m := strings.ToUpper(in.Method)
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
		s := strings.ToLower(in.Scheme)
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
	if w.Host != nil && !w.Host.Matches(in.Host) {
		return false
	}
	if w.Port != nil && !w.Port.Matches(in.Port) {
		return false
	}
	if w.Path != nil && !w.Path.Matches(in.Path) {
		return false
	}
	if w.ContentType != nil && !w.ContentType.Matches(strings.ToLower(in.ContentType)) {
		return false
	}
	if w.Header != nil && !w.Header.MatchesHeaders(in.Headers) {
		return false
	}
	if w.BodyContains != "" && !strings.Contains(in.BodyPreview, w.BodyContains) {
		return false
	}
	return true
}

// MatchesResponse checks condition for response
func (w *InterceptWhen) MatchesResponse(in ResponseMatchInput) bool {
	if w == nil {
		return true
	}
	if w.Status != nil && !w.Status.MatchesCode(in.StatusCode) {
		return false
	}
	if w.ContentType != nil && !w.ContentType.Matches(strings.ToLower(in.ContentType)) {
		return false
	}
	if w.Header != nil && !w.Header.MatchesHeaders(in.Headers) {
		return false
	}
	if w.BodyContains != "" && !strings.Contains(in.BodyPreview, w.BodyContains) {
		return false
	}
	return true
}

// Matches checks if a string matches against RuleStringMatch criteria
func (m *RuleStringMatch) Matches(s string) bool {
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
		if rx, err := cachedRegexp(m.Regex); err == nil {
			return rx.MatchString(s)
		}
	}
	return m.Equals == "" && m.Prefix == "" && m.Suffix == "" && m.Contains == "" && len(m.AnyOf) == 0 && m.Regex == ""
}

// IsEmpty returns true if no matching criteria are set
func (m RuleStringMatch) IsEmpty() bool {
	return m.Equals == "" && m.Prefix == "" && m.Suffix == "" && m.Contains == "" && len(m.AnyOf) == 0 && m.Regex == ""
}

// MatchesHeaders checks if headers match the RuleHeaderMatch criteria
func (h *RuleHeaderMatch) MatchesHeaders(headers map[string][]string) bool {
	if h == nil {
		return true
	}
	for name, values := range headers {
		if !h.Name.Matches(name) {
			continue
		}
		if h.Value.IsEmpty() {
			return true
		}
		for _, v := range values {
			if h.Value.Matches(v) {
				return true
			}
		}
	}
	return false
}

// MatchesCode checks if status code matches the RuleStatusMatch criteria
func (s *RuleStatusMatch) MatchesCode(code int) bool {
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
