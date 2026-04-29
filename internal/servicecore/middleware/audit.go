package middleware

import (
	"log"
	"net/http"
	"time"
)

func Audit() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("http_audit method=%s path=%s latency_ms=%d", r.Method, r.URL.Path, time.Since(start).Milliseconds())
		})
	}
}
