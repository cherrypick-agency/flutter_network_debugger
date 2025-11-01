package httpapi

import (
	"reflect"
	"sync"
	"testing"

	"network-debugger/internal/domain"
)

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "single value",
			input: "value",
			want:  []string{"value"},
		},
		{
			name:  "multiple values",
			input: "one,two,three",
			want:  []string{"one", "two", "three"},
		},
		{
			name:  "values with spaces",
			input: " one , two , three ",
			want:  []string{"one", "two", "three"},
		},
		{
			name:  "mixed case",
			input: "One,TWO,Three",
			want:  []string{"one", "two", "three"},
		},
		{
			name:  "empty values",
			input: "one,,two,  ,three",
			want:  []string{"one", "two", "three"},
		},
		{
			name:  "only commas",
			input: ",,,",
			want:  []string{},
		},
		{
			name:  "only spaces",
			input: "  ,  ,  ",
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCSV(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasAnyTag(t *testing.T) {
	tests := []struct {
		name string
		want []string
		tags map[string]struct{}
		exp  bool
	}{
		{
			name: "empty want - always true",
			want: []string{},
			tags: map[string]struct{}{"foo": {}},
			exp:  true,
		},
		{
			name: "nil want - always true",
			want: nil,
			tags: map[string]struct{}{"foo": {}},
			exp:  true,
		},
		{
			name: "has matching tag",
			want: []string{"foo", "bar"},
			tags: map[string]struct{}{"foo": {}, "baz": {}},
			exp:  true,
		},
		{
			name: "no matching tag",
			want: []string{"foo", "bar"},
			tags: map[string]struct{}{"baz": {}, "qux": {}},
			exp:  false,
		},
		{
			name: "empty tags",
			want: []string{"foo"},
			tags: map[string]struct{}{},
			exp:  false,
		},
		{
			name: "single match",
			want: []string{"test"},
			tags: map[string]struct{}{"test": {}},
			exp:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAnyTag(tt.want, tt.tags)
			if got != tt.exp {
				t.Errorf("hasAnyTag() = %v, want %v", got, tt.exp)
			}
		})
	}
}

func TestMatchesAnyStatusGroup(t *testing.T) {
	tests := []struct {
		name   string
		groups []string
		meta   *httpMetaV1
		want   bool
	}{
		{
			name:   "empty groups - always true",
			groups: []string{},
			meta:   &httpMetaV1{Status: 404},
			want:   true,
		},
		{
			name:   "nil groups - always true",
			groups: nil,
			meta:   &httpMetaV1{Status: 404},
			want:   true,
		},
		{
			name:   "nil meta",
			groups: []string{"4xx"},
			meta:   nil,
			want:   false,
		},
		{
			name:   "matches 1xx",
			groups: []string{"1xx"},
			meta:   &httpMetaV1{Status: 101},
			want:   true,
		},
		{
			name:   "matches 2xx",
			groups: []string{"2xx"},
			meta:   &httpMetaV1{Status: 200},
			want:   true,
		},
		{
			name:   "matches 3xx",
			groups: []string{"3xx"},
			meta:   &httpMetaV1{Status: 301},
			want:   true,
		},
		{
			name:   "matches 4xx",
			groups: []string{"4xx"},
			meta:   &httpMetaV1{Status: 404},
			want:   true,
		},
		{
			name:   "matches 5xx",
			groups: []string{"5xx"},
			meta:   &httpMetaV1{Status: 500},
			want:   true,
		},
		{
			name:   "no match",
			groups: []string{"2xx"},
			meta:   &httpMetaV1{Status: 404},
			want:   false,
		},
		{
			name:   "multiple groups one matches",
			groups: []string{"2xx", "4xx"},
			meta:   &httpMetaV1{Status: 404},
			want:   true,
		},
		{
			name:   "edge case 100",
			groups: []string{"1xx"},
			meta:   &httpMetaV1{Status: 100},
			want:   true,
		},
		{
			name:   "edge case 199",
			groups: []string{"1xx"},
			meta:   &httpMetaV1{Status: 199},
			want:   true,
		},
		{
			name:   "edge case 200",
			groups: []string{"2xx"},
			meta:   &httpMetaV1{Status: 200},
			want:   true,
		},
		{
			name:   "edge case 299",
			groups: []string{"2xx"},
			meta:   &httpMetaV1{Status: 299},
			want:   true,
		},
		{
			name:   "status 99 no match",
			groups: []string{"1xx"},
			meta:   &httpMetaV1{Status: 99},
			want:   false,
		},
		{
			name:   "status 600 no match",
			groups: []string{"5xx"},
			meta:   &httpMetaV1{Status: 600},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesAnyStatusGroup(tt.groups, tt.meta)
			if got != tt.want {
				t.Errorf("matchesAnyStatusGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeBaseTags(t *testing.T) {
	tests := []struct {
		name     string
		session  sessionV1
		wantTags []string
	}{
		{
			name: "https url",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test1",
					Target: "https://example.com/path",
					Kind:   "http",
				},
			},
			wantTags: []string{"https"},
		},
		{
			name: "http url",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test2",
					Target: "http://example.com/path",
					Kind:   "http",
				},
			},
			wantTags: []string{"http"},
		},
		{
			name: "ws url",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test3",
					Target: "ws://example.com/socket",
					Kind:   "ws",
				},
			},
			wantTags: []string{"ws"},
		},
		{
			name: "wss url",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test4",
					Target: "wss://example.com/socket",
					Kind:   "ws",
				},
			},
			wantTags: []string{"ws"},
		},
		{
			name: "graphql path",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test5",
					Target: "https://example.com/graphql",
					Kind:   "http",
				},
			},
			wantTags: []string{"https", "graphql"},
		},
		{
			name: "json mime",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test6",
					Target: "https://example.com/api",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "application/json",
				},
			},
			wantTags: []string{"https", "json"},
		},
		{
			name: "form mime",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test7",
					Target: "https://example.com/submit",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "application/x-www-form-urlencoded",
				},
			},
			wantTags: []string{"https", "form"},
		},
		{
			name: "xml mime",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test8",
					Target: "https://example.com/soap",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "text/xml",
				},
			},
			wantTags: []string{"https", "xml"},
		},
		{
			name: "javascript mime",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test9",
					Target: "https://example.com/script.js",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "application/javascript",
				},
			},
			wantTags: []string{"https", "js"},
		},
		{
			name: "css mime",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test10",
					Target: "https://example.com/style.css",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "text/css",
				},
			},
			wantTags: []string{"https", "css"},
		},
		{
			name: "media mime - image",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test11",
					Target: "https://example.com/photo.jpg",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "image/jpeg",
				},
			},
			wantTags: []string{"https", "media"},
		},
		{
			name: "media mime - audio",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test12",
					Target: "https://example.com/audio.mp3",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "audio/mpeg",
				},
			},
			wantTags: []string{"https", "media"},
		},
		{
			name: "document mime - html",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test13",
					Target: "https://example.com/page.html",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "text/html",
				},
			},
			wantTags: []string{"https", "document"},
		},
		{
			name: "document mime - pdf",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test14",
					Target: "https://example.com/doc.pdf",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "application/pdf",
				},
			},
			wantTags: []string{"https", "document"},
		},
		{
			name: "other - no mime but has https",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test15",
					Target: "https://example.com/unknown",
					Kind:   "http",
				},
			},
			wantTags: []string{"https"},
		},
		{
			name: "graphql mime",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test16",
					Target: "https://example.com/api",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "application/graphql",
				},
			},
			wantTags: []string{"https", "graphql"},
		},
		{
			name: "multipart form",
			session: sessionV1{
				Session: domain.Session{
					ID:     "test17",
					Target: "https://example.com/upload",
					Kind:   "http",
				},
				HttpMeta: &httpMetaV1{
					Mime: "multipart/form-data",
				},
			},
			wantTags: []string{"https", "form"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeBaseTags(tt.session)
			// Convert result to slice for easier comparison
			var gotTags []string
			for tag := range got {
				gotTags = append(gotTags, tag)
			}

			// Check that all expected tags are present
			for _, wantTag := range tt.wantTags {
				if _, ok := got[wantTag]; !ok {
					t.Errorf("computeBaseTags() missing tag %q, got tags: %v", wantTag, gotTags)
				}
			}
		})
	}
}

func TestGetBaseTags(t *testing.T) {
	// Clear cache before test
	baseTagsCache = sync.Map{}

	t.Run("caching works", func(t *testing.T) {
		session := sessionV1{
			Session: domain.Session{
				ID:     "cache-test",
				Target: "https://example.com/api",
				Kind:   "http",
			},
			HttpMeta: &httpMetaV1{
				Mime: "application/json",
			},
		}

		// First call - should compute
		tags1 := getBaseTags(session)
		if len(tags1) == 0 {
			t.Fatal("getBaseTags() returned empty tags")
		}

		// Second call - should use cache
		tags2 := getBaseTags(session)
		if len(tags2) == 0 {
			t.Fatal("getBaseTags() returned empty tags on second call")
		}

		// Verify tags are the same
		if len(tags1) != len(tags2) {
			t.Errorf("cached tags differ: first=%d, second=%d", len(tags1), len(tags2))
		}
	})

	t.Run("cache invalidation on change", func(t *testing.T) {
		session := sessionV1{
			Session: domain.Session{
				ID:     "invalidation-test",
				Target: "https://example.com/api",
				Kind:   "http",
			},
			HttpMeta: &httpMetaV1{
				Mime: "application/json",
			},
		}

		// First call
		tags1 := getBaseTags(session)

		// Change mime type
		session.HttpMeta.Mime = "text/xml"

		// Second call should recompute
		tags2 := getBaseTags(session)

		// Should have different tags (json vs xml)
		_, hasJSON1 := tags1["json"]
		_, hasXML1 := tags1["xml"]
		_, hasJSON2 := tags2["json"]
		_, hasXML2 := tags2["xml"]

		if !hasJSON1 || hasXML1 {
			t.Error("first call should have json tag, not xml")
		}
		if hasJSON2 || !hasXML2 {
			t.Error("second call should have xml tag, not json")
		}
	})
}

func TestUrlParser_Parse(t *testing.T) {
	parser := &urlParser{}

	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "valid http url",
			raw:     "http://example.com/path",
			wantErr: false,
		},
		{
			name:    "valid https url",
			raw:     "https://example.com/path",
			wantErr: false,
		},
		{
			name:    "valid ws url",
			raw:     "ws://example.com/socket",
			wantErr: false,
		},
		{
			name:    "valid wss url",
			raw:     "wss://example.com/socket",
			wantErr: false,
		},
		{
			name:    "url with query",
			raw:     "https://example.com/path?key=value",
			wantErr: false,
		},
		{
			name:    "url with fragment",
			raw:     "https://example.com/path#section",
			wantErr: false,
		},
		{
			name:    "invalid url - spaces",
			raw:     "http://example .com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.parse(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("parse() returned nil without error")
			}
		})
	}
}
