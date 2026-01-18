package platform

import (
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)
	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}
}

func TestRateLimiterAllowsInitialBurst(t *testing.T) {
	// Rate limiter with 5 requests per second
	rl := NewRateLimiter(5, time.Second)

	// Should allow up to 5 requests immediately
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("request %d should have been allowed", i+1)
		}
	}
}

func TestRateLimiterBlocksExcessRequests(t *testing.T) {
	// Rate limiter with 3 requests per second
	rl := NewRateLimiter(3, time.Second)

	// Exhaust the bucket
	for i := 0; i < 3; i++ {
		rl.Allow()
	}

	// 4th request should be blocked
	if rl.Allow() {
		t.Error("4th request should have been blocked")
	}
}

func TestRateLimiterRefillsOverTime(t *testing.T) {
	// Rate limiter with 2 requests per 100ms
	rl := NewRateLimiter(2, 100*time.Millisecond)

	// Exhaust the bucket
	rl.Allow()
	rl.Allow()

	// Should be blocked
	if rl.Allow() {
		t.Error("should be blocked after exhausting bucket")
	}

	// Wait for refill (100ms should add 2 tokens)
	time.Sleep(110 * time.Millisecond)

	// Should allow again
	if !rl.Allow() {
		t.Error("should be allowed after waiting for refill")
	}
}

func TestRateLimiterWait(t *testing.T) {
	// Rate limiter with 1 request per 50ms
	rl := NewRateLimiter(1, 50*time.Millisecond)

	// Exhaust the bucket
	rl.Allow()

	// Wait should block until token is available
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	// Should have waited approximately 50ms
	if elapsed < 40*time.Millisecond {
		t.Errorf("Wait returned too quickly: %v", elapsed)
	}
}

func TestRateLimiterConcurrentAccess(t *testing.T) {
	// Rate limiter with 100 requests per second
	rl := NewRateLimiter(100, time.Second)

	var wg sync.WaitGroup
	allowed := 0
	blocked := 0
	var mu sync.Mutex

	// Spawn 200 goroutines trying to acquire tokens
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if rl.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			} else {
				mu.Lock()
				blocked++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should have allowed exactly 100 and blocked 100
	if allowed != 100 {
		t.Errorf("expected 100 allowed, got %d", allowed)
	}
	if blocked != 100 {
		t.Errorf("expected 100 blocked, got %d", blocked)
	}
}

func TestPolymarketRateLimiter(t *testing.T) {
	// Polymarket rate limit is 100/min = 100 per 60 seconds
	rl := NewRateLimiter(100, time.Minute)

	// Should allow 100 requests
	for i := 0; i < 100; i++ {
		if !rl.Allow() {
			t.Errorf("request %d should have been allowed", i+1)
		}
	}

	// 101st should be blocked
	if rl.Allow() {
		t.Error("101st request should have been blocked")
	}
}

func TestKalshiRateLimiter(t *testing.T) {
	// Kalshi rate limit is 30/min = 30 per 60 seconds
	rl := NewRateLimiter(30, time.Minute)

	// Should allow 30 requests
	for i := 0; i < 30; i++ {
		if !rl.Allow() {
			t.Errorf("request %d should have been allowed", i+1)
		}
	}

	// 31st should be blocked
	if rl.Allow() {
		t.Error("31st request should have been blocked")
	}
}
