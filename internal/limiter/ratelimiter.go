package limiter

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	mu           sync.Mutex
	lastRequest  time.Time
	minInterval  time.Duration
	requestCount int
	maxRequests  int
}

func New(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
		maxRequests: 4,
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	if !rl.lastRequest.IsZero() {
		elapsed := now.Sub(rl.lastRequest)

		if elapsed < rl.minInterval {
			waitTime := rl.minInterval - elapsed

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	rl.lastRequest = time.Now()
	rl.requestCount++

	return nil
}

func (rl *RateLimiter) GetRequestCount() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.requestCount
}

func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.requestCount = 0
	rl.lastRequest = time.Time{}
}