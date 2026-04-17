package httpapi

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// NewEmbeddedWebMux mounts the current API surface together with an SPA bundle.
// This keeps binary-specific startup concerns outside httpapi while avoiding
// duplicated route composition across embedded-web binaries.
func NewEmbeddedWebMux(apiHandler http.Handler, webRoot fs.FS) http.Handler {
	mux := http.NewServeMux()
	spa := embeddedSPAHandler{root: webRoot, index: "index.html"}

	// API routes first.
	mux.Handle("/_api/", apiHandler)
	mux.Handle("/api/", apiHandler)

	// Compatibility proxy routes that frontend SDKs already use.
	mux.Handle("/httpproxy", apiHandler)
	mux.Handle("/httpproxy/", apiHandler)
	mux.Handle("/_ws", apiHandler)
	mux.Handle("/_ws/", apiHandler)

	// Static content and SPA fallback last.
	mux.Handle("/", spa)

	return mux
}

type embeddedSPAHandler struct {
	root  fs.FS
	index string
}

func (h embeddedSPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upath := r.URL.Path
	if upath == "" || upath == "/" {
		h.serveFile(w, h.index)
		return
	}

	p := strings.TrimPrefix(path.Clean(upath), "/")
	f, err := h.root.Open(p)
	if err == nil {
		_ = f.Close()
		http.FileServer(http.FS(h.root)).ServeHTTP(w, r)
		return
	}

	h.serveFile(w, h.index)
}

func (h embeddedSPAHandler) serveFile(w http.ResponseWriter, name string) {
	data, err := fs.ReadFile(h.root, name)
	if err != nil {
		http.NotFound(w, &http.Request{})
		return
	}
	if strings.HasSuffix(strings.ToLower(name), ".html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	_, _ = w.Write(data)
}
