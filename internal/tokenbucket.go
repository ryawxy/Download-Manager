package internal

import (
	"sync"
	"time"
)

type TokenBucket struct {
	Capacity   int
	Rate       time.Duration
	Tokens     int
	LastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		Capacity:   capacity,
		Rate:       rate,
		Tokens:     capacity,
		LastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.LastRefill)
	tokensToAdd := int(elapsed / tb.Rate)

	if tokensToAdd > 0 {
		tb.Tokens = min(tb.Tokens+tokensToAdd, tb.Capacity)
		tb.LastRefill = now
	}
}

func (tb *TokenBucket) Take(tokens int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.Tokens >= tokens {
		tb.Tokens -= tokens
		return true
	}
	return false
}
