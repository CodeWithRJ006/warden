package pii

import (
	"context"
	"fmt"
	"sync"
)

// Vault stores raw sensitive data and issues deterministic tokens.
type Vault interface {
	Store(ctx context.Context, entityType string, raw string) (string, error)
	Retrieve(ctx context.Context, token string) (string, error)
}

// MemoryVault is an in-memory implementation for Day 3.
type MemoryVault struct {
	mu      sync.RWMutex
	store   map[string]string
	counter map[string]int
}

func NewMemoryVault() *MemoryVault {
	return &MemoryVault{
		store:   make(map[string]string),
		counter: make(map[string]int),
	}
}

func (v *MemoryVault) Store(ctx context.Context, entityType string, raw string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.counter[entityType]++
	token := fmt.Sprintf("[%s_%03d]", entityType, v.counter[entityType])
	v.store[token] = raw
	return token, nil
}

func (v *MemoryVault) Retrieve(ctx context.Context, token string) (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	raw, ok := v.store[token]
	if !ok {
		return "", fmt.Errorf("token not found: %s", token)
	}
	return raw, nil
}
