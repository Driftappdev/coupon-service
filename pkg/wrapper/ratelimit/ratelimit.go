package ratelimit

import (
	"time"

	goratelimit "github.com/driftappdev/libpackage/goratelimit"
)

func NewPerIP(limit int, window time.Duration) *goratelimit.FixedWindowLimiter {
	return goratelimit.NewFixedWindow(limit, window)
}
