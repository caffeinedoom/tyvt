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
	mu           sync.Mutex
	lastRequest  time.Time
	minInterval  time.Duration
	requestCount int
	maxRequests  int
	keyQuotas    map[string]*KeyQuota
}

const (
	DailyLimit   = 500
	MonthlyLimit = 15500
)

func New(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
		maxRequests: 4,
		keyQuotas:   make(map[string]*KeyQuota),
	}
}

func (rl *RateLimiter) CanMakeRequest(apiKey string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	quota, exists := rl.keyQuotas[apiKey]
	if !exists {
		quota = &KeyQuota{
			LastReset:  time.Now().Truncate(24 * time.Hour),
			MonthReset: time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC),
		}
		rl.keyQuotas[apiKey] = quota
	}

	now := time.Now()

	if now.Sub(quota.LastReset) >= 24*time.Hour {
		quota.DailyCount = 0
		quota.LastReset = now.Truncate(24 * time.Hour)
	}

	if now.Month() != quota.MonthReset.Month() || now.Year() != quota.MonthReset.Year() {
		quota.MonthlyCount = 0
		quota.MonthReset = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	if quota.DailyCount >= DailyLimit {
		return fmt.Errorf("daily quota exceeded for key (500/day)")
	}

	if quota.MonthlyCount >= MonthlyLimit {
		return fmt.Errorf("monthly quota exceeded for key (15,500/month)")
	}

	return nil
}

func (rl *RateLimiter) Wait(ctx context.Context, apiKey string) error {
	rl.mu.Lock()

	if err := rl.canMakeRequestUnsafe(apiKey); err != nil {
		rl.mu.Unlock()
		return err
	}

	now := time.Now()
	var waitTime time.Duration

	if !rl.lastRequest.IsZero() {
		elapsed := now.Sub(rl.lastRequest)
		if elapsed < rl.minInterval {
			waitTime = rl.minInterval - elapsed
		}
	}

	rl.mu.Unlock()

	if waitTime > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	rl.mu.Lock()
	rl.lastRequest = time.Now()
	rl.requestCount++

	quota := rl.keyQuotas[apiKey]
	quota.DailyCount++
	quota.MonthlyCount++
	rl.mu.Unlock()

	return nil
}

func (rl *RateLimiter) canMakeRequestUnsafe(apiKey string) error {
	quota, exists := rl.keyQuotas[apiKey]
	if !exists {
		quota = &KeyQuota{
			LastReset:  time.Now().Truncate(24 * time.Hour),
			MonthReset: time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC),
		}
		rl.keyQuotas[apiKey] = quota
	}

	now := time.Now()

	if now.Sub(quota.LastReset) >= 24*time.Hour {
		quota.DailyCount = 0
		quota.LastReset = now.Truncate(24 * time.Hour)
	}

	if now.Month() != quota.MonthReset.Month() || now.Year() != quota.MonthReset.Year() {
		quota.MonthlyCount = 0
		quota.MonthReset = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	if quota.DailyCount >= DailyLimit {
		return fmt.Errorf("daily quota exceeded for key (500/day)")
	}

	if quota.MonthlyCount >= MonthlyLimit {
		return fmt.Errorf("monthly quota exceeded for key (15,500/month)")
	}

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