package engine

import (
	"sync"
	"time"
)

type TokenBucket struct {
	maxTokens     float64
	tokensPerSec  float64
	currentTokens float64
	lastRefill    time.Time
	mu            sync.Mutex
}

func NewTokenBucket(tokensPerSec int) *TokenBucket {
	return &TokenBucket{
		maxTokens:     float64(tokensPerSec),
		tokensPerSec:  float64(tokensPerSec),
		currentTokens: float64(tokensPerSec),
		lastRefill:    time.Now(),
	}
}

func (tb *TokenBucket) Take() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.currentTokens += elapsed * tb.tokensPerSec
	if tb.currentTokens > tb.maxTokens {
		tb.currentTokens = tb.maxTokens
	}
	tb.lastRefill = now

	for tb.currentTokens < 1 {
		tb.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		tb.mu.Lock()

		now := time.Now()
		elapsed := now.Sub(tb.lastRefill).Seconds()
		tb.currentTokens += elapsed * tb.tokensPerSec
		if tb.currentTokens > tb.maxTokens {
			tb.currentTokens = tb.maxTokens
		}
		tb.lastRefill = now
	}

	tb.currentTokens--
}

func (tb *TokenBucket) TryTake() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.currentTokens += elapsed * tb.tokensPerSec
	if tb.currentTokens > tb.maxTokens {
		tb.currentTokens = tb.maxTokens
	}
	tb.lastRefill = now

	if tb.currentTokens >= 1 {
		tb.currentTokens--
		return true
	}

	return false
}

type RateLimiter interface {
	Take()
	TryTake() bool
}

type NoOpRateLimiter struct{}

func (nr *NoOpRateLimiter) Take()         {}
func (nr *NoOpRateLimiter) TryTake() bool { return true }

func NewRateLimiter(requestsPerSec int) RateLimiter {
	if requestsPerSec <= 0 {
		return &NoOpRateLimiter{}
	}
	return NewTokenBucket(requestsPerSec)
}
