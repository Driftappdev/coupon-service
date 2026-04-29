package middleware

import (
	"net/http"
	"sync"
	"time"
)

var idempotencyStore sync.Map

func IdempotencyHTTP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			storeKey := r.Method + ":" + r.URL.Path + ":" + key
			if _, exists := idempotencyStore.LoadOrStore(storeKey, time.Now().Unix()); exists {
				http.Error(w, "duplicate idempotent request", http.StatusConflict)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
