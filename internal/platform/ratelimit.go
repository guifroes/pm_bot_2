package platform

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
// It allows a maximum of 'capacity' requests per 'interval'.
type RateLimiter struct {
	capacity   int           // max tokens (requests per interval)
	interval   time.Duration // time window for the rate limit
	tokens     int           // current available tokens
	lastRefill time.Time     // last time tokens were refilled
	mu         sync.Mutex    // protects tokens and lastRefill
}

// NewRateLimiter creates a new rate limiter.
// capacity is the maximum number of requests allowed per interval.
// interval is the time window (e.g., time.Minute for requests per minute).
func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		capacity:   capacity,
		interval:   interval,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

// refill adds tokens based on elapsed time since last refill.
// Must be called with mutex held.
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	// Calculate how many tokens to add based on elapsed time
	// tokens_to_add = (elapsed / interval) * capacity
	tokensToAdd := int(float64(elapsed) / float64(r.interval) * float64(r.capacity))

	if tokensToAdd > 0 {
		r.tokens += tokensToAdd
		if r.tokens > r.capacity {
			r.tokens = r.capacity
		}
		r.lastRefill = now
	}
}

// Allow checks if a request should be allowed.
// Returns true if allowed (and consumes a token), false if rate limited.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// Wait blocks until a token is available, then consumes it.
func (r *RateLimiter) Wait() {
	for {
		r.mu.Lock()
		r.refill()

		if r.tokens > 0 {
			r.tokens--
			r.mu.Unlock()
			return
		}

		// Calculate how long until next token is available
		// time_per_token = interval / capacity
		timePerToken := r.interval / time.Duration(r.capacity)
		r.mu.Unlock()

		// Sleep for a bit less than time_per_token to avoid oversleeping
		time.Sleep(timePerToken / 2)
	}
}

// Remaining returns the number of tokens currently available.
func (r *RateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()
	return r.tokens
}

// Reset restores the rate limiter to full capacity.
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens = r.capacity
	r.lastRefill = time.Now()
}

// Predefined rate limiters for known platforms

// NewPolymarketRateLimiter creates a rate limiter for Polymarket (100/min).
func NewPolymarketRateLimiter() *RateLimiter {
	return NewRateLimiter(100, time.Minute)
}

// NewKalshiRateLimiter creates a rate limiter for Kalshi (30/min).
func NewKalshiRateLimiter() *RateLimiter {
	return NewRateLimiter(30, time.Minute)
}
