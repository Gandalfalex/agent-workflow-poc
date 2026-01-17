package httpapi

import (
	"net/http"
	"strings"
)

func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	allowAll := false
	for _, origin := range allowedOrigins {
		value := strings.TrimSpace(origin)
		if value == "" {
			continue
		}
		if value == "*" {
			allowAll = true
			continue
		}
		allowed[value] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowAll || allowed[origin]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				methods := "GET,POST,PATCH,PUT,DELETE,OPTIONS"
				w.Header().Set("Access-Control-Allow-Methods", methods)
				headers := r.Header.Get("Access-Control-Request-Headers")
				if headers == "" {
					headers = "Content-Type, Authorization"
				}
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
