package ratelimit

import (
	"context"
	"sync"
	"time"
)

type FixedWindowLimiter struct {
	limit   int
	window  time.Duration
	mu      sync.Mutex
	count   int
	started time.Time
}

func NewPerIP(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{limit: limit, window: window, started: time.Now()}
}

func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if time.Since(l.started) > l.window {
		l.started = time.Now()
		l.count = 0
	}
	if l.count >= l.limit {
		return false, nil
	}
	l.count++
	return true, nil
}
