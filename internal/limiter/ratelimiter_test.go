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

func TestRateLimiter_RequestCount(t *testing.T) {
	rl := New(time.Millisecond)
	ctx := context.Background()
	testKey := "test-api-key"

	for i := 0; i < 5; i++ {
		rl.Wait(ctx, testKey)
	}

	if count := rl.GetRequestCount(); count != 5 {
		t.Errorf("Expected 5 requests, got %d", count)
	}

	rl.Reset()

	if count := rl.GetRequestCount(); count != 0 {
		t.Errorf("Expected 0 requests after reset, got %d", count)
	}
}