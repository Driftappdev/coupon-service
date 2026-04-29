package middleware

import (
	"net/http"
)

func Permission() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Internal-Auth") == "deny" {
				http.Error(w, "permission denied", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
