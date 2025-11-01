package runtime

import (
	"net/http"
	"net/url"
	"os"
	stdpath "path"
	"regexp"
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
	hostRe     *regexp.Regexp
	pathRe     *regexp.Regexp
}

func New() *Manager { return &Manager{rules: []compiledRule{}, fileMTime: make(map[string]int64)} }

// Update заменяет набор правил (ожидается отсортированный по priority ASC)
func (m *Manager) Update(rules []mdomain.MapRule) {
	cr := make([]compiledRule, 0, len(rules))
	for _, r := range rules {
		c := compiledRule{src: r, methodsSet: make(map[string]struct{})}
		for _, me := range r.Methods {
			c.methodsSet[strings.ToUpper(strings.TrimSpace(me))] = struct{}{}
		}
		if r.PatternType == mdomain.PatternRegex {
			c.isRegex = true
			if strings.TrimSpace(r.HostPattern) != "" {
				if re, err := regexp.Compile(r.HostPattern); err == nil {
					c.hostRe = re
				}
			}
			if strings.TrimSpace(r.PathPattern) != "" {
				if re, err := regexp.Compile(r.PathPattern); err == nil {
					c.pathRe = re
				}
			}
		}
		cr = append(cr, c)
	}
	m.mu.Lock()
	m.rules = cr
	// переинициализируем кеш mtime для локальных путей
	for _, r := range rules {
		if r.FilePath != nil && *r.FilePath != "" {
			if fi, err := os.Stat(*r.FilePath); err == nil {
				m.fileMTime[*r.FilePath] = fi.ModTime().Unix()
			}
		}
	}
	m.mu.Unlock()
}

// EvalRequest возвращает решение по первому совпавшему правилу
func (m *Manager) EvalRequest(r *http.Request) (Decision, bool) {
	m.mu.RLock()
	rules := m.rules
	m.mu.RUnlock()
	method := strings.ToUpper(r.Method)
	host := ""
	path := ""
	if r.URL != nil {
		host = r.URL.Hostname()
		path = r.URL.Path
	}
	for _, cr := range rules {
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
				if fi, err := os.Stat(*cr.src.FilePath); err == nil {
					mt := fi.ModTime().Unix()
					if last, ok := m.fileMTime[*cr.src.FilePath]; !ok || last != mt {
						m.fileMTime[*cr.src.FilePath] = mt
						if m.onFileChange != nil {
							m.onFileChange(cr.src.ID, *cr.src.FilePath)
						}
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
		if cr.isRegex && cr.pathRe != nil && r.URL != nil {
			// применим подстановки $1.. по path
			to = cr.pathRe.ReplaceAllString(r.URL.String(), cr.src.TargetURLTemplate)
		}
		// нормализуем целевой URL
		if u, err := url.Parse(to); err == nil {
			return Decision{
				Kind:         mdomain.KindRemote,
				RuleID:       cr.src.ID,
				RemoteURL:    u.String(),
				PreserveHost: cr.src.PreserveHost,
			}, true
		}
		// если шаблон не распарсился — игнорируем правило
	}
	return Decision{}, false
}

func matchHostPath(cr compiledRule, host, path string) bool {
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

// SetOnFileChange регистрирует коллбек на изменения локальных файлов правил
func (m *Manager) SetOnFileChange(cb func(ruleID string, path string)) {
	m.mu.Lock()
	m.onFileChange = cb
	m.mu.Unlock()
}
