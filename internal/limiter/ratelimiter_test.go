package limiter

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_Wait(t *testing.T) {
	rl := New(100 * time.Millisecond)
	ctx := context.Background()
	testKey := "test-api-key"

	start := time.Now()

	if err := rl.Wait(ctx, testKey); err != nil {
		t.Errorf("First wait should not error: %v", err)
	}

	if err := rl.Wait(ctx, testKey); err != nil {
		t.Errorf("Second wait should not error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected at least 100ms delay, got %v", elapsed)
	}
}

func TestRateLimiter_ContextCancellation(t *testing.T) {
	rl := New(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	testKey := "test-api-key"

	if err := rl.Wait(ctx, testKey); err != nil {
		t.Errorf("First wait should not error: %v", err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := rl.Wait(ctx, testKey)
	elapsed := time.Since(start)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}

	if elapsed > 500*time.Millisecond {
		t.Errorf("Expected cancellation to happen quickly, took %v", elapsed)
	}
}

func TestRateLimiter_QuotaTracking(t *testing.T) {
	rl := New(time.Millisecond)
	ctx := context.Background()
	testKey := "test-api-key"

	// Make 5 requests
	for i := 0; i < 5; i++ {
		if err := rl.Wait(ctx, testKey); err != nil {
			t.Errorf("Request %d failed: %v", i+1, err)
		}
	}

	// Check quota status
	daily, monthly := rl.GetQuotaStatus(testKey)
	if daily != 5 {
		t.Errorf("Expected 5 daily requests, got %d", daily)
	}
	if monthly != 5 {
		t.Errorf("Expected 5 monthly requests, got %d", monthly)
	}

	// Reset and verify
	rl.Reset()
	daily, monthly = rl.GetQuotaStatus(testKey)
	if daily != 0 {
		t.Errorf("Expected 0 daily requests after reset, got %d", daily)
	}
	if monthly != 0 {
		t.Errorf("Expected 0 monthly requests after reset, got %d", monthly)
	}
}

func TestRateLimiter_DailyQuotaLimit(t *testing.T) {
	rl := New(time.Millisecond)
	ctx := context.Background()
	testKey := "test-api-key"

	// Artificially set quota to near limit
	rl.Wait(ctx, testKey) // Initialize quota
	rl.mu.Lock()
	quota := rl.keyQuotas[testKey]
	quota.DailyCount = DailyLimit - 1
	rl.mu.Unlock()

	// This should succeed (at limit)
	if err := rl.Wait(ctx, testKey); err != nil {
		t.Errorf("Request at limit should succeed: %v", err)
	}

	// This should fail (over limit)
	err := rl.Wait(ctx, testKey)
	if err == nil {
		t.Error("Expected error when exceeding daily quota, got nil")
	}
	if err != nil && err.Error() != "daily quota exceeded for key (500/day)" {
		t.Errorf("Expected daily quota error, got: %v", err)
	}
}

func TestRateLimiter_GetQuotaStatus_NonexistentKey(t *testing.T) {
	rl := New(time.Millisecond)

	daily, monthly := rl.GetQuotaStatus("nonexistent-key")
	if daily != 0 || monthly != 0 {
		t.Errorf("Expected 0,0 for nonexistent key, got %d,%d", daily, monthly)
	}
}