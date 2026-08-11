package ratelimit

import (
	"testing"
	"time"
)

func TestMemoryStore_Allow(t *testing.T) {
	s := NewMemoryStore()
	for i := 0; i < 3; i++ {
		if !s.Allow("127.0.0.1", 3, time.Minute) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if s.Allow("127.0.0.1", 3, time.Minute) {
		t.Fatal("fourth attempt should be blocked")
	}
}

func TestMemoryStore_WindowElapseResetsCount(t *testing.T) {
	m := &memoryStore{attempts: make(map[string][]time.Time), maxKeys: defaultMaxKeys}
	// The key is at its limit, but every attempt is far outside the window.
	m.attempts["ip"] = []time.Time{
		time.Now().Add(-2 * time.Hour),
		time.Now().Add(-2*time.Hour + time.Second),
		time.Now().Add(-2*time.Hour + 2*time.Second),
	}
	// A request after the window elapsed must be allowed: the stale history is
	// dropped rather than carried forward to block the caller forever.
	if !m.Allow("ip", 3, time.Minute) {
		t.Fatal("request after window elapse should be allowed")
	}
	// Only the fresh attempt is recorded — stale timestamps are not leaked.
	if got := len(m.attempts["ip"]); got != 1 {
		t.Fatalf("expected 1 recorded attempt after window elapse, got %d", got)
	}
}

func TestMemoryStore_CapsSize(t *testing.T) {
	m := &memoryStore{attempts: make(map[string][]time.Time), maxKeys: 2}
	// More distinct keys than the cap, all active within the window.
	for _, k := range []string{"a", "b", "c", "d"} {
		if !m.Allow(k, 10, time.Minute) {
			t.Fatalf("key %q should be allowed", k)
		}
	}
	if len(m.attempts) > 2 {
		t.Fatalf("map should be capped at 2 entries, got %d", len(m.attempts))
	}
}

func TestMemoryStore_CapsSizeOldestEvictedFirst(t *testing.T) {
	m := &memoryStore{attempts: make(map[string][]time.Time), maxKeys: 2}
	// Simulate history where "old" has the least-recent last activity.
	now := time.Now()
	m.attempts["old"] = []time.Time{now.Add(-30 * time.Second)}
	m.attempts["new"] = []time.Time{now.Add(-time.Second)}
	// Inserting a third active key must evict the least-recently-active one.
	m.Allow("third", 10, time.Minute)
	if len(m.attempts) > 2 {
		t.Fatalf("map should be capped at 2 entries, got %d", len(m.attempts))
	}
	if _, ok := m.attempts["old"]; ok {
		t.Fatal("least-recently-active key should have been evicted first")
	}
}
