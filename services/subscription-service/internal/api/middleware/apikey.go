package middleware

import (
	"crypto/subtle"
	"net/http"
)

// InternalAPIKey returns middleware that checks the X-Internal-API-Key header.
// Comparison uses constant-time equality to prevent timing attacks.
func InternalAPIKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := r.Header.Get("X-Internal-API-Key")
			if subtle.ConstantTimeCompare([]byte(got), []byte(key)) != 1 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
