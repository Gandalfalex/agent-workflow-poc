package httpapi

import (
	"context"
	"net/http"
	"strings"
	"ticketing-system/backend/internal/auth"
)

type ctxKey string

const (
	userKey     ctxKey = "currentUser"
	authUserKey ctxKey = "authUser"
)

func requireAuth(h *API) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Context().Value(SessionAuthScopes) == nil {
				next.ServeHTTP(w, r)
				return
			}

			token := readSessionToken(r, h.cookieName)
			if token == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
				return
			}

			user, err := h.auth.Verify(r.Context(), token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid session")
				return
			}
			h.syncUser(r, user)

			ctx := context.WithValue(r.Context(), userKey, mapUser(user))
			ctx = context.WithValue(ctx, authUserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func readSessionToken(r *http.Request, cookieName string) string {
	cookie, err := r.Cookie(cookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	return ""
}

func currentUser(ctx context.Context) (userResponse, bool) {
	value := ctx.Value(userKey)
	if value == nil {
		return userResponse{}, false
	}
	user, ok := value.(userResponse)
	return user, ok
}

func authUser(ctx context.Context) (auth.User, bool) {
	value := ctx.Value(authUserKey)
	if value == nil {
		return auth.User{}, false
	}
	user, ok := value.(auth.User)
	return user, ok
}
