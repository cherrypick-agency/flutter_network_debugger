package httpapi

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"
)

type sessionProjectionFilters struct {
	Types             []string
	StatusGroups      []string
	Tags              []string
	CaptureScope      string
	CaptureIDExplicit *int
	IncludeUnassigned bool
	Limit             int
	GroupBy           string
	ScanGraphQL       bool
}

type sessionProjector struct {
	deps *Deps
}

func newSessionProjector(d *Deps) sessionProjector {
	return sessionProjector{deps: d}
}

func (p sessionProjector) listViews(ctx context.Context, filter usecase.SessionFilter, project sessionProjectionFilters) ([]sessionV1, int, error) {
	list, total, err := p.deps.Svc.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].StartedAt.Before(list[j].StartedAt) })

	views := make([]sessionV1, 0, len(list))
	for _, sess := range list {
		view := p.buildView(ctx, sess)
		if !p.matches(ctx, view, project) {
			continue
		}
		views = append(views, view)
		if project.Limit > 0 && len(views) >= project.Limit {
			break
		}
	}
	return views, total, nil
}

func (p sessionProjector) viewByID(ctx context.Context, sessionID string, project sessionProjectionFilters) (*sessionV1, bool, error) {
	sess, ok, err := p.deps.Svc.Get(ctx, sessionID)
	if err != nil || !ok {
		return nil, ok, err
	}
	view := p.buildView(ctx, sess)
	if !p.matches(ctx, view, project) {
		return nil, false, nil
	}
	return &view, true, nil
}

func (p sessionProjector) buildView(ctx context.Context, sess domain.Session) sessionV1 {
	view := sessionV1{Session: sess}
	if meta, sz := p.deps.enrichWithHTTPMeta(ctx, sess); meta != nil || sz != nil {
		view.HttpMeta = meta
		view.Sizes = sz
	}
	return view
}

func (p sessionProjector) matches(ctx context.Context, view sessionV1, project sessionProjectionFilters) bool {
	if !p.passQuickFilters(ctx, view, project) {
		return false
	}
	if len(project.Tags) > 0 && !p.sessionHasAnyTag(ctx, view.ID, project.Tags) {
		return false
	}
	if !p.passCapture(view, project) {
		return false
	}
	return true
}

func (p sessionProjector) passQuickFilters(ctx context.Context, view sessionV1, project sessionProjectionFilters) bool {
	if len(project.Types) > 0 {
		tags := getBaseTags(view)
		needsGraphQL := false
		for _, t := range project.Types {
			if t == "graphql" {
				needsGraphQL = true
				break
			}
		}
		if needsGraphQL && project.ScanGraphQL {
			if _, ok := tags["graphql"]; !ok && detectGraphQLByBody(ctx, p.deps, view.ID) {
				tags["graphql"] = struct{}{}
			}
		}
		if !hasAnyTag(project.Types, tags) {
			return false
		}
	}
	if len(project.StatusGroups) > 0 && !matchesAnyStatusGroup(project.StatusGroups, view.HttpMeta) {
		return false
	}
	return true
}

func (p sessionProjector) sessionHasAnyTag(ctx context.Context, sessionID string, want []string) bool {
	if p.deps.TagsSvc == nil || len(want) == 0 {
		return true
	}
	tags, err := p.deps.TagsSvc.GetSessionTags(ctx, sessionID)
	if err != nil || len(tags) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		set[strings.ToLower(strings.TrimSpace(tag.TagName))] = struct{}{}
	}
	for _, wantTag := range want {
		if _, ok := set[strings.ToLower(strings.TrimSpace(wantTag))]; ok {
			return true
		}
	}
	return false
}

func (p sessionProjector) passCapture(view sessionV1, project sessionProjectionFilters) bool {
	if project.CaptureIDExplicit != nil {
		return view.CaptureID != nil && *view.CaptureID == *project.CaptureIDExplicit
	}
	switch strings.ToLower(strings.TrimSpace(project.CaptureScope)) {
	case "all":
		if !project.IncludeUnassigned && view.CaptureID == nil {
			return false
		}
		return true
	case "current", "":
		if project.IncludeUnassigned {
			return true
		}
		cur := -1
		if repo := sessionsRepoOf(p.deps.Svc); repo != nil {
			if rs, ok := repo.(interface{ RecordingState() (bool, int) }); ok {
				_, cur = rs.RecordingState()
			}
		}
		return view.CaptureID != nil && *view.CaptureID == cur
	default:
		return true
	}
}

func (p sessionProjector) aggregateViews(views []sessionV1, groupBy string) []map[string]any {
	agg := make(map[string]int)
	for _, view := range views {
		key := sessionGroupKey(view, groupBy)
		agg[key]++
	}
	groups := make([]map[string]any, 0, len(agg))
	for key, count := range agg {
		groups = append(groups, map[string]any{"key": key, "count": count})
	}
	sort.Slice(groups, func(i, j int) bool {
		ci := groups[i]["count"].(int)
		cj := groups[j]["count"].(int)
		if ci != cj {
			return ci > cj
		}
		ki := groups[i]["key"].(string)
		kj := groups[j]["key"].(string)
		return ki < kj
	})
	return groups
}

func sessionGroupKey(view sessionV1, groupBy string) string {
	gb := strings.ToLower(strings.TrimSpace(groupBy))
	switch gb {
	case "method":
		if view.HttpMeta != nil && strings.TrimSpace(view.HttpMeta.Method) != "" {
			return strings.ToUpper(view.HttpMeta.Method)
		}
		return "unknown"
	case "mime", "contenttype":
		if view.HttpMeta != nil && strings.TrimSpace(view.HttpMeta.Mime) != "" {
			mime := strings.ToLower(view.HttpMeta.Mime)
			if i := strings.IndexByte(mime, ';'); i >= 0 {
				mime = strings.TrimSpace(mime[:i])
			}
			return mime
		}
		return "unknown"
	case "status", "statusgroup":
		if view.HttpMeta != nil && view.HttpMeta.Status > 0 {
			statusGroup := view.HttpMeta.Status / 100
			if statusGroup >= 1 && statusGroup <= 5 {
				return strconv.Itoa(statusGroup) + "xx"
			}
			return strconv.Itoa(view.HttpMeta.Status)
		}
		return "unknown"
	case "host", "domain", "target_host":
		fallthrough
	default:
		key := normalizeProjectionHost(view.Target)
		if key == "" {
			return "unknown"
		}
		return key
	}
}

func projectionFilterFromSIO(f sioFilters) sessionProjectionFilters {
	return sessionProjectionFilters{
		Types:             f.Types,
		StatusGroups:      f.StatusGroups,
		Tags:              f.Tags,
		CaptureScope:      f.CaptureScope,
		CaptureIDExplicit: f.CaptureIDExplicit,
		IncludeUnassigned: f.IncludeUnassigned,
		Limit:             f.Limit,
		GroupBy:           f.GroupBy,
	}
}

func projectionListFilterFromSIO(f sioFilters) usecase.SessionFilter {
	out := usecase.SessionFilter{Q: f.Q, Target: f.Target, Limit: 1000, Offset: 0}
	if f.CaptureIDExplicit != nil {
		out.CaptureID = f.CaptureIDExplicit
	} else {
		switch strings.ToLower(strings.TrimSpace(f.CaptureScope)) {
		case "all":
			out.CaptureID = nil
			if f.IncludeUnassigned {
				out.IncludeUnassigned = true
			}
		case "current", "":
			v := -1
			out.CaptureID = &v
		}
	}
	if f.IncludeUnassigned {
		out.IncludeUnassigned = true
	}
	return out
}

func normalizeProjectionHost(targetURL string) string {
	key := targetURL
	scheme := ""
	if i := strings.Index(key, "://"); i >= 0 {
		scheme = strings.ToLower(key[:i])
		key = key[i+3:]
	}
	if at := strings.IndexByte(key, '@'); at >= 0 {
		key = key[at+1:]
	}
	if j := strings.IndexByte(key, '/'); j >= 0 {
		key = key[:j]
	}
	if scheme == "http" && strings.HasSuffix(key, ":80") {
		key = key[:len(key)-3]
	} else if scheme == "https" && strings.HasSuffix(key, ":443") {
		key = key[:len(key)-4]
	}
	return key
}
