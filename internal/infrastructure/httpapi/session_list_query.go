package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"network-debugger/internal/usecase"
)

type sessionListQuery struct {
	Filter  usecase.SessionFilter
	Project sessionProjectionFilters
}

func newLegacySessionListQuery(r *http.Request, d *Deps) usecase.SessionFilter {
	return usecase.SessionFilter{
		Q:          r.URL.Query().Get("q"),
		Target:     r.URL.Query().Get("_target"),
		Limit:      parseListLimit(r, "limit", 50, 0, 0),
		Offset:     parseListOffset(r),
		SessionIDs: findTaggedSessionIDs(r, d, strings.TrimSpace(r.URL.Query().Get("tags"))),
	}
}

func newV1SessionListQuery(r *http.Request, d *Deps) sessionListQuery {
	limit := parseListLimit(r, "limit", 100, 1, 1000)
	filter := usecase.SessionFilter{
		Q:          r.URL.Query().Get("q"),
		Target:     r.URL.Query().Get("_target"),
		Limit:      limit,
		Offset:     parseListOffset(r),
		SessionIDs: findTaggedSessionIDs(r, d, strings.TrimSpace(r.URL.Query().Get("tags"))),
	}
	project := sessionProjectionFilters{
		Types:        splitCSV(r.URL.Query().Get("types")),
		StatusGroups: splitCSV(r.URL.Query().Get("status")),
		Limit:        limit,
		ScanGraphQL:  hasCSVToken(r.URL.Query().Get("scan"), "graphql"),
	}

	captureID := strings.TrimSpace(r.URL.Query().Get("captureId"))
	if captureID != "" {
		if captureID == "current" {
			v := -1
			filter.CaptureID = &v
			project.CaptureScope = "current"
		} else if n, err := strconv.Atoi(captureID); err == nil {
			filter.CaptureID = &n
			project.CaptureIDExplicit = &n
		}
	}
	if include := r.URL.Query().Get("includeUnassigned"); include == "true" || include == "1" {
		filter.IncludeUnassigned = true
		project.IncludeUnassigned = true
	}
	if r.URL.Query().Get("captures") == "all" {
		filter.CaptureID = nil
		filter.IncludeUnassigned = true
		project.CaptureScope = "all"
		project.IncludeUnassigned = true
	}

	return sessionListQuery{Filter: filter, Project: project}
}

func parseListLimit(r *http.Request, key string, fallback, minValue, maxValue int) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get(key))
	if limit <= 0 {
		return fallback
	}
	if minValue > 0 && limit < minValue {
		return minValue
	}
	if maxValue > 0 && limit > maxValue {
		return maxValue
	}
	return limit
}

func parseListOffset(r *http.Request) int {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		return 0
	}
	return offset
}

func findTaggedSessionIDs(r *http.Request, d *Deps, raw string) []string {
	if raw == "" || d.TagsSvc == nil {
		return nil
	}
	names := splitCSVPreserveCase(raw)
	if len(names) == 0 {
		return nil
	}
	ids, err := d.TagsSvc.FindSessionIDsByTags(r.Context(), names)
	if err != nil {
		return nil
	}
	return ids
}

func splitCSVPreserveCase(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func hasCSVToken(raw, want string) bool {
	for _, item := range splitCSV(raw) {
		if item == want {
			return true
		}
	}
	return false
}
