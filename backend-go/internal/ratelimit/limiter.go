// Package ratelimit provides per-resource rate limiting for the AI Avengers pipeline.
//
// DESIGN PATTERN: Strategy (GoF)
//   RateLimiter interface allows swapping implementations:
//   - TokenBucketLimiter: production (token bucket algorithm)
//   - NoopLimiter: testing / disabled state (Null Object pattern)
//
// SOLID:
//   SRP: each limiter has one job — decide if a request is allowed.
//   OCP: new algorithms (sliding window, leaky bucket) implement the
//        interface without modifying callers.
//   DIP: callers depend on RateLimiter interface, not concrete types.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter is the Strategy interface for rate limiting.
// Callers call Allow() before each operation.
// If Allow() returns false, the caller should return an error or retry.
type RateLimiter interface {
	// Allow checks if the operation identified by key is allowed.
	// key is typically "expert:{expert_id}" or "project:{project_id}".
	// Returns true if allowed, false if rate limit exceeded.
	Allow(ctx context.Context, key string) bool
}

// TokenBucketLimiter implements RateLimiter using the token bucket algorithm.
//
// Each key gets its own bucket. Buckets are created lazily on first access.
// Tokens refill at a constant rate (refillRate tokens per second).
// Burst allows short spikes above the steady-state rate.
//
// WHY token bucket over fixed window:
//   Fixed window allows 2x burst at window boundaries.
//   Token bucket smooths traffic while still allowing short bursts.
//   Matches how LLM providers implement their own rate limits.
type TokenBucketLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*tokenBucket
	capacity   float64 // max tokens (burst size)
	refillRate float64 // tokens per second
}

type tokenBucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewTokenBucketLimiter creates a limiter.
// capacity: max burst (e.g. 5 = allow 5 concurrent requests)
// refillRate: steady-state rate (e.g. 1.0 = 1 request/second)
func NewTokenBucketLimiter(capacity, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		buckets:    make(map[string]*tokenBucket),
		capacity:   capacity,
		refillRate: refillRate,
	}
}

// Allow implements RateLimiter.
// Thread-safe: uses mutex to protect bucket state.
func (l *TokenBucketLimiter) Allow(_ context.Context, key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.buckets[key]
	if !ok {
		// New key: start with full bucket
		bucket = &tokenBucket{
			tokens:     l.capacity,
			lastRefill: time.Now(),
		}
		l.buckets[key] = bucket
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = min(l.capacity, bucket.tokens+elapsed*l.refillRate)
	bucket.lastRefill = now

	if bucket.tokens < 1.0 {
		return false // rate limit exceeded
	}

	bucket.tokens -= 1.0
	return true
}

// ExpertKey returns the rate limit key for a given expert ID.
// Centralizes key format so callers don't construct strings manually.
func ExpertKey(expertID string) string {
	return fmt.Sprintf("expert:%s", expertID)
}

// NoopLimiter is a RateLimiter that always allows.
// Used when rate limiting is disabled or in tests.
// Implements the Null Object pattern (GoF) — callers need no nil checks.
type NoopLimiter struct{}

func (n *NoopLimiter) Allow(_ context.Context, _ string) bool { return true }

// min returns the smaller of two float64 values.
// Avoids importing math just for this.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
