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
	r.Mount(basePath, http.StripPrefix(basePath, handler))

	// Health check at root
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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

			// SPA fallback - serve index.html for non-API routes
			if !strings.HasPrefix(r.URL.Path, "/api/") &&
				!strings.HasPrefix(r.URL.Path, "/auth/") &&
				!strings.HasPrefix(r.URL.Path, "/health") {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		api.ServeHTTP(w, r)
	})
}
