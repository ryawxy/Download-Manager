package internal

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int
	rate       time.Duration
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity int, rate time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		rate:       rate,
		tokens:     capacity,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(float64(elapsed) / float64(tb.rate))

	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.capacity)
		tb.lastRefill = now
	}
}

func (tb *TokenBucket) WaitAndTake(tokens int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	for tb.tokens < tokens {
		needed := tokens - tb.tokens
		waitTime := time.Duration(needed) * tb.rate

		tb.mu.Unlock()
		time.Sleep(waitTime)
		tb.mu.Lock()
		tb.refill()
	}

	tb.tokens -= tokens
}
