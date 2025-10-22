package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type KeyQuota struct {
	DailyCount   int
	MonthlyCount int
	LastReset    time.Time
	MonthReset   time.Time
}

type RateLimiter struct {
	mu          sync.Mutex
	lastRequest time.Time
	minInterval time.Duration
	keyQuotas   map[string]*KeyQuota
}

const (
	DailyLimit   = 500
	MonthlyLimit = 15500
)

func New(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
		keyQuotas:   make(map[string]*KeyQuota),
	}
}

// checkQuota verifies if a request can be made for the given API key
// without exceeding daily or monthly limits. Assumes mutex is already held.
// This is the single source of truth for quota checking logic.
func (rl *RateLimiter) checkQuota(apiKey string) error {
	quota, exists := rl.keyQuotas[apiKey]
	if !exists {
		quota = &KeyQuota{
			LastReset:  time.Now().Truncate(24 * time.Hour),
			MonthReset: time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC),
		}
		rl.keyQuotas[apiKey] = quota
	}

	now := time.Now()

	// Reset daily quota if 24 hours have passed
	if now.Sub(quota.LastReset) >= 24*time.Hour {
		quota.DailyCount = 0
		quota.LastReset = now.Truncate(24 * time.Hour)
	}

	// Reset monthly quota if we're in a new month
	if now.Month() != quota.MonthReset.Month() || now.Year() != quota.MonthReset.Year() {
		quota.MonthlyCount = 0
		quota.MonthReset = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	// Check if limits would be exceeded
	if quota.DailyCount >= DailyLimit {
		return fmt.Errorf("daily quota exceeded for key (500/day)")
	}

	if quota.MonthlyCount >= MonthlyLimit {
		return fmt.Errorf("monthly quota exceeded for key (15,500/month)")
	}

	return nil
}

// Wait blocks until it's safe to make a request, respecting both
// rate limiting intervals and API quota limits.
func (rl *RateLimiter) Wait(ctx context.Context, apiKey string) error {
	rl.mu.Lock()

	// Check quota first while we have the lock
	if err := rl.checkQuota(apiKey); err != nil {
		rl.mu.Unlock()
		return err
	}

	now := time.Now()
	var waitTime time.Duration

	// Calculate how long we need to wait based on last request time
	if !rl.lastRequest.IsZero() {
		elapsed := now.Sub(rl.lastRequest)
		if elapsed < rl.minInterval {
			waitTime = rl.minInterval - elapsed
		}
	}

	rl.mu.Unlock()

	// Wait outside the lock to allow other goroutines to proceed
	if waitTime > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	// Update counters after wait
	rl.mu.Lock()
	rl.lastRequest = time.Now()

	quota := rl.keyQuotas[apiKey]
	quota.DailyCount++
	quota.MonthlyCount++
	rl.mu.Unlock()

	return nil
}

// GetQuotaStatus returns the current quota usage for an API key.
// This is useful for monitoring and logging.
func (rl *RateLimiter) GetQuotaStatus(apiKey string) (dailyUsed, monthlyUsed int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	quota, exists := rl.keyQuotas[apiKey]
	if !exists {
		return 0, 0
	}

	return quota.DailyCount, quota.MonthlyCount
}

// Reset clears the rate limiter state. Primarily used for testing.
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastRequest = time.Time{}
	rl.keyQuotas = make(map[string]*KeyQuota)
}
