package runtime

import (
	"net"
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"regexp"
	"sort"
	"strings"
	"sync"

	mdomain "network-debugger/internal/features/mapping/domain"
)

// Decision описывает результат сопоставления
type Decision struct {
	Kind   mdomain.Kind
	RuleID string
	// Local
	LocalFilePath       *string
	LocalBlobPath       *string
	StatusOverride      int
	ContentTypeOverride string
	// Remote
	RemoteURL    string
	PreserveHost bool
}

// Manager хранит предкомпилированные правила и осуществляет быстрый matching
type Manager struct {
	mu    sync.RWMutex
	rules []compiledRule
	// отслеживание изменений локальных файлов (без внешних зависимостей)
	fileMTime    map[string]int64
	onFileChange func(ruleID string, path string)
}

type compiledRule struct {
	src        mdomain.MapRule
	methodsSet map[string]struct{}
	isRegex    bool
	invalid    bool
	hostRe     *regexp.Regexp
	pathRe     *regexp.Regexp
}

func New() *Manager { return &Manager{rules: []compiledRule{}, fileMTime: make(map[string]int64)} }

// Update заменяет набор правил (ожидается отсортированный по priority ASC)
func (m *Manager) Update(rules []mdomain.MapRule) {
	// На всякий случай сортируем сами, чтобы рантайм не зависел от порядка списка.
	sorted := append([]mdomain.MapRule(nil), rules...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	cr := make([]compiledRule, 0, len(rules))
	for _, r := range sorted {
		c := compiledRule{src: r, methodsSet: make(map[string]struct{})}
		for _, me := range r.Methods {
			c.methodsSet[strings.ToUpper(strings.TrimSpace(me))] = struct{}{}
		}
		if r.PatternType == mdomain.PatternRegex {
			c.isRegex = true
			if strings.TrimSpace(r.HostPattern) != "" {
				re, err := regexp.Compile(r.HostPattern)
				if err != nil {
					c.invalid = true
				} else {
					c.hostRe = re
				}
			}
			if strings.TrimSpace(r.PathPattern) != "" {
				re, err := regexp.Compile(r.PathPattern)
				if err != nil {
					c.invalid = true
				} else {
					c.pathRe = re
				}
			}
		}
		cr = append(cr, c)
	}
	m.mu.Lock()
	m.rules = cr
	// переинициализируем кеш mtime для локальных путей
	m.fileMTime = make(map[string]int64)
	for _, r := range sorted {
		if r.FilePath != nil && *r.FilePath != "" {
			if fi, err := os.Stat(*r.FilePath); err == nil {
				m.fileMTime[*r.FilePath] = fi.ModTime().Unix()
			}
		}
	}
	m.mu.Unlock()
}

// EvalRequest возвращает решение по маппингу.
//
// Важно: если remote-правило имеет stopProcessing=false, мы применяем его
// (как переписывание URL) и продолжаем матчить следующие правила на уже
// переписанном URL.
func (m *Manager) EvalRequest(r *http.Request) (Decision, bool) {
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()

	method := strings.ToUpper(r.Method)

	var curURL *url.URL
	if r.URL != nil {
		u := *r.URL
		curURL = &u
	}
	curHostHeader := r.Host

	host := ""
	path := ""
	if curURL != nil {
		host = curURL.Hostname()
		path = curURL.Path
	}
	if host == "" && strings.TrimSpace(curHostHeader) != "" {
		host = hostNoPort(strings.TrimSpace(curHostHeader))
	}

	var lastRemote Decision
	haveRemote := false

	for _, cr := range rules {
		if cr.invalid {
			continue
		}
		if !cr.src.Enabled {
			continue
		}
		if len(cr.methodsSet) > 0 {
			if _, ok := cr.methodsSet[method]; !ok {
				continue
			}
		}
		if !matchHostPath(cr, host, path) {
			continue
		}
		// Совпало — формируем решение
		if cr.src.Kind == mdomain.KindLocal {
			// проверим изменение файла (best-effort)
			if cr.src.FilePath != nil && *cr.src.FilePath != "" {
				p := *cr.src.FilePath
				if fi, err := os.Stat(p); err == nil {
					mt := fi.ModTime().Unix()
					cb, changed := m.maybeUpdateMTime(p, mt)
					if changed && cb != nil {
						cb(cr.src.ID, p)
					}
				}
			}
			return Decision{
				Kind:                mdomain.KindLocal,
				RuleID:              cr.src.ID,
				LocalFilePath:       cr.src.FilePath,
				LocalBlobPath:       cr.src.BlobPath,
				StatusOverride:      nonZero(cr.src.StatusOverride, 200),
				ContentTypeOverride: cr.src.ContentTypeOverride,
			}, true
		}
		// Remote
		to := cr.src.TargetURLTemplate
		if cr.isRegex && cr.pathRe != nil && curURL != nil {
			// применим подстановки $1.. по path
			to = cr.pathRe.ReplaceAllString(curURL.Path, cr.src.TargetURLTemplate)
		}
		to = applyTemplateTokens(to, &http.Request{Method: r.Method, URL: curURL, Host: curHostHeader})
		// нормализуем целевой URL
		if u, err := url.Parse(to); err == nil {
			lastRemote = Decision{
				Kind:         mdomain.KindRemote,
				RuleID:       cr.src.ID,
				RemoteURL:    u.String(),
				PreserveHost: cr.src.PreserveHost,
			}
			haveRemote = true

			curURL = u
			curHostHeader = u.Host
			host = u.Hostname()
			path = u.Path

			if cr.src.StopProcessing {
				return lastRemote, true
			}
			continue
		}
		// если шаблон не распарсился — игнорируем правило
	}
	if haveRemote {
		return lastRemote, true
	}
	return Decision{}, false
}

func (m *Manager) maybeUpdateMTime(path string, mt int64) (func(ruleID string, path string), bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	last, ok := m.fileMTime[path]
	if ok && last == mt {
		return m.onFileChange, false
	}
	m.fileMTime[path] = mt
	return m.onFileChange, true
}

func matchHostPath(cr compiledRule, host, path string) bool {
	if cr.invalid {
		return false
	}
	// Host
	if cr.isRegex {
		if cr.hostRe != nil && !cr.hostRe.MatchString(host) {
			return false
		}
	} else {
		if strings.TrimSpace(cr.src.HostPattern) != "" {
			ok, err := stdpath.Match(cr.src.HostPattern, host)
			if err != nil || !ok {
				return false
			}
		}
	}
	// Path
	if cr.isRegex {
		if cr.pathRe != nil && !cr.pathRe.MatchString(path) {
			return false
		}
	} else {
		if strings.TrimSpace(cr.src.PathPattern) != "" {
			ok, err := stdpath.Match(cr.src.PathPattern, path)
			if err != nil || !ok {
				return false
			}
		}
	}
	return true
}

func nonZero(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

func hostNoPort(hostport string) string {
	h, _, err := net.SplitHostPort(hostport)
	if err == nil && h != "" {
		return h
	}
	// net.SplitHostPort падает на "example.com" без порта — это нормальный кейс.
	return hostport
}

func applyTemplateTokens(tpl string, r *http.Request) string {
	if r == nil || r.URL == nil {
		return tpl
	}
	u := r.URL
	path := u.EscapedPath()
	if path == "" {
		path = u.Path
	}
	full := u.String()
	port := u.Port()
	portWithColon := ""
	if port != "" {
		portWithColon = ":" + port
	}
	return strings.NewReplacer(
		"{scheme}", u.Scheme,
		"{host}", u.Host,
		"{hostname}", u.Hostname(),
		"{port}", port,
		"{portWithColon}", portWithColon,
		"{path}", path,
		"{query}", u.RawQuery,
		"{rawQuery}", u.RawQuery,
		"{url}", full,
		"{full}", full,
	).Replace(tpl)
}

// SetOnFileChange регистрирует коллбек на изменения локальных файлов правил
func (m *Manager) SetOnFileChange(cb func(ruleID string, path string)) {
	m.mu.Lock()
	m.onFileChange = cb
	m.mu.Unlock()
}
