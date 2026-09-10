package middleware

import (
	"context"
	"time"
)

type IdempotencyStore interface {
	CheckOrSet(ctx context.Context, key string, expiration time.Duration) (bool, error)
}

