package middleware

import "net/http"

func Validation() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) && r.Body == nil {
				http.Error(w, "request payload is required", http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
