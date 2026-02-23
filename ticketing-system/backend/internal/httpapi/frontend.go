package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// WithBasePath mounts a handler at a specific base path
func WithBasePath(handler http.Handler, basePath string) http.Handler {
	if basePath == "" || basePath == "/" {
		return handler
	}

	r := chi.NewRouter()

	// Health check at root
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Mount at base path with prefix stripping
	r.Mount(basePath, http.StripPrefix(basePath, handler))

	return r
}

// WithFrontend serves static frontend files with SPA fallback support
func WithFrontend(api http.Handler, dir string, basePath string) http.Handler {
	if dir == "" {
		return api
	}

	indexPath := filepath.Join(dir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return api
	}

	fileServer := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			if r.URL.Path == "" || r.URL.Path == "/" {
				http.ServeFile(w, r, indexPath)
				return
			}

			cleanPath := filepath.Clean(r.URL.Path)
			cleanPath = strings.TrimPrefix(cleanPath, "/")
			if cleanPath != "." {
				filePath := filepath.Join(dir, cleanPath)
				if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
					fileServer.ServeHTTP(w, r)
					return
				}
			}

			// Browser navigation requests should always receive the SPA shell,
			// even when the path matches API-like prefixes such as /projects/*.
			if wantsHTML(r) {
				http.ServeFile(w, r, indexPath)
				return
			}

			// SPA fallback - serve index.html for non-API routes
			if !strings.HasPrefix(r.URL.Path, "/rest/v1") {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		api.ServeHTTP(w, r)
	})
}

func wantsHTML(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	if !strings.Contains(accept, "text/html") {
		return false
	}
	// Only treat true browser page navigations as SPA document requests.
	// Fetch/XHR API calls can still carry broad Accept headers in some clients.
	if strings.EqualFold(r.Header.Get("X-Requested-With"), "XMLHttpRequest") {
		return false
	}
	if dest := strings.ToLower(strings.TrimSpace(r.Header.Get("Sec-Fetch-Dest"))); dest != "" && dest != "document" {
		return false
	}
	if mode := strings.ToLower(strings.TrimSpace(r.Header.Get("Sec-Fetch-Mode"))); mode != "" && mode != "navigate" && mode != "nested-navigate" {
		return false
	}
	return true
}
