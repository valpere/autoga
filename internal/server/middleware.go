package server

import (
	"net/http"
	"strings"
)

// apiKeyAuth returns a middleware that enforces Bearer token authentication.
// If key is empty the middleware is a no-op (auth disabled).
func apiKeyAuth(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token != key {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
