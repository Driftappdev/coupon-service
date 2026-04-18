package middleware

import (
	"net/http"
	"sync"
	"time"

	wrapperratelimit "coupon-service/pkg/wrapper/ratelimit"
	goratelimit "github.com/driftappdev/libpackage/goratelimit"
)

var (
	rateLimitOnce sync.Once
	rateLimitMW   func(http.Handler) http.Handler
)

func RateLimit() func(http.Handler) http.Handler {
	rateLimitOnce.Do(func() {
		limiter := wrapperratelimit.NewPerIP(120, time.Minute)
		rateLimitMW = goratelimit.Middleware(goratelimit.MiddlewareConfig{
			Limiter:      limiter,
			KeyExtractor: goratelimit.ByIP(),
		})
	})
	return rateLimitMW
}
