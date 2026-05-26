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
