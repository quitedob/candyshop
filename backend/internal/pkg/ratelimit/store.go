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
	maxKeys  int
}

// defaultMaxKeys bounds the in-memory map so a high volume of distinct keys
// (e.g. distinct client IPs) cannot grow the store without bound.
const defaultMaxKeys = 100_000

// NewMemoryStore creates an in-process rate limit store. The store is bounded:
// a key whose window has fully elapsed is dropped on its next access (lazy TTL),
// and the map is capped at defaultMaxKeys entries with the least-recently-active
// keys evicted first when the cap is exceeded.
func NewMemoryStore() Store {
	return &memoryStore{
		attempts: make(map[string][]time.Time),
		maxKeys:  defaultMaxKeys,
	}
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
	if len(filtered) == 0 {
		// Lazy TTL: once a key's history has fully fallen out of the window it
		// no longer counts toward the limit, so drop the map entry entirely.
		delete(m.attempts, key)
	}
	if len(filtered) >= limit {
		if len(filtered) > 0 {
			m.attempts[key] = filtered
		}
		return false
	}
	filtered = append(filtered, now)
	m.attempts[key] = filtered
	m.sweepIfNeeded(cutoff)
	return true
}

// sweepIfNeeded keeps the map bounded to maxKeys entries. It runs only after an
// insert, so the steady-state map size cannot exceed the cap. With fewer
// distinct keys than the cap (the common case) it is a no-op.
func (m *memoryStore) sweepIfNeeded(cutoff time.Time) {
	if m.maxKeys <= 0 || len(m.attempts) <= m.maxKeys {
		return
	}
	// Pass 1: drop keys whose history has entirely fallen out of the window.
	for k, h := range m.attempts {
		filtered := h[:0]
		for _, t := range h {
			if t.After(cutoff) {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) == 0 {
			delete(m.attempts, k)
		} else {
			m.attempts[k] = filtered
		}
	}
	// Pass 2: genuinely concurrent distinct keys still over the cap — evict the
	// least-recently-active keys oldest-first.
	for len(m.attempts) > m.maxKeys {
		var oldestKey string
		var oldest time.Time
		for k, h := range m.attempts {
			last := h[len(h)-1]
			if oldestKey == "" || last.Before(oldest) {
				oldestKey, oldest = k, last
			}
		}
		delete(m.attempts, oldestKey)
	}
}
