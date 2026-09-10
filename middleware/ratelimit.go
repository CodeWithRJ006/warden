package middleware

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	CheckIdempotency(ctx context.Context, key string, window time.Duration) (bool, error)
}

// MemoryRateLimiter for local tests. In prod, we'd use a Redis-backed implementation.
type MemoryRateLimiter struct {
	mu      sync.Mutex
	counts  map[string]int
	expires map[string]time.Time
}

func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		counts:  make(map[string]int),
		expires: make(map[string]time.Time),
	}
}

func (m *MemoryRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if exp, ok := m.expires[key]; ok && now.After(exp) {
		m.counts[key] = 0
	}

	if m.counts[key] >= limit {
		return false, nil
	}

	m.counts[key]++
	if m.counts[key] == 1 {
		m.expires[key] = now.Add(window)
	}

	return true, nil
}


func (m *MemoryRateLimiter) CheckIdempotency(ctx context.Context, key string, window time.Duration) (bool, error) {
	return true, nil
}

