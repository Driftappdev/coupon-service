package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	wrapperratelimit "coupon-service/pkg/wrapper/ratelimit"
)

var (
	rateLimitOnce sync.Once
	rateLimitMW   func(http.Handler) http.Handler
)

func RateLimit() func(http.Handler) http.Handler {
	rateLimitOnce.Do(func() {
		limiter := wrapperratelimit.NewPerIP(120, time.Minute)
		rateLimitMW = func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				allowed, err := limiter.Allow(context.Background(), r.RemoteAddr)
				if err != nil {
					http.Error(w, "rate limiter unavailable", http.StatusServiceUnavailable)
					return
				}
				if !allowed {
					http.Error(w, "too many requests", http.StatusTooManyRequests)
					return
				}
				next.ServeHTTP(w, r)
			})
		}
	})
	return rateLimitMW
}
