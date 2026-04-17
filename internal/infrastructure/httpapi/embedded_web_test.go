package httpapi

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestNewEmbeddedWebMux_RoutesAPIAndSPA(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-API-Path", r.URL.Path)
		_, _ = w.Write([]byte("api"))
	})

	webRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>app</html>")},
		"main.js":    &fstest.MapFile{Data: []byte("console.log('ok');")},
	}

	handler := NewEmbeddedWebMux(api, fs.FS(webRoot))

	t.Run("api routes take precedence", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/_api/v1/version", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if rec.Header().Get("X-API-Path") != "/_api/v1/version" {
			t.Fatalf("X-API-Path = %q", rec.Header().Get("X-API-Path"))
		}
		if strings.TrimSpace(rec.Body.String()) != "api" {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})

	t.Run("static asset is served directly", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/main.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		body, _ := io.ReadAll(rec.Body)
		if strings.TrimSpace(string(body)) != "console.log('ok');" {
			t.Fatalf("body = %q", string(body))
		}
	})

	t.Run("spa fallback serves index", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/deep/link", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
			t.Fatalf("content-type = %q", ct)
		}
		if strings.TrimSpace(rec.Body.String()) != "<html>app</html>" {
			t.Fatalf("body = %q", rec.Body.String())
		}
	})
}
