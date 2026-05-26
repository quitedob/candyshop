// Package ratelimit provides a pluggable rate-limit store (memory default; Redis when REDIS_URL is set).
package ratelimit

import (
	"sync"
	"time"
)

// Store tracks attempts within a sliding window.
type Store interface {
	Allow(key string, limit int, window time.Duration) bool
}

type memoryStore struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

// NewMemoryStore creates an in-process rate limit store.
func NewMemoryStore() Store {
	return &memoryStore{attempts: make(map[string][]time.Time)}
}

func (m *memoryStore) Allow(key string, limit int, window time.Duration) bool {
	now := time.Now()
	cutoff := now.Add(-window)
	m.mu.Lock()
	defer m.mu.Unlock()
	history := m.attempts[key]
	filtered := history[:0]
	for _, t := range history {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= limit {
		m.attempts[key] = filtered
		return false
	}
	filtered = append(filtered, now)
	m.attempts[key] = filtered
	return true
}
