package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func WithFrontend(api http.Handler, dir string) http.Handler {
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
		}

		api.ServeHTTP(w, r)
	})
}
