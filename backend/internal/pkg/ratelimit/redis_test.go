package ratelimit

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisStore_Allow(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	store, err := NewRedisStore("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRedisStore: %v", err)
	}
	for i := 0; i < 3; i++ {
		if !store.Allow("127.0.0.1", 3, time.Minute) {
			t.Fatalf("attempt %d should pass", i+1)
		}
	}
	if store.Allow("127.0.0.1", 3, time.Minute) {
		t.Fatal("4th attempt should be blocked")
	}
}

func TestRedisStore_Allow_FailsBoundedOnRedisError(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	store, err := NewRedisStore("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRedisStore: %v", err)
	}

	// Happy path: the limit is enforced while Redis is up.
	for i := 0; i < 3; i++ {
		if !store.Allow("127.0.0.1", 3, time.Minute) {
			t.Fatalf("attempt %d should pass while Redis is up", i+1)
		}
	}
	if store.Allow("127.0.0.1", 3, time.Minute) {
		t.Fatal("4th attempt should be blocked while Redis is up")
	}

	// Kill Redis. The next INCR will fail; the store must NOT fail open
	// (return true = unlimited). It must fall back to the bounded in-memory
	// store and keep enforcing the configured limit.
	mr.Close()
	for i := 0; i < 3; i++ {
		if !store.Allow("down-ip", 3, time.Minute) {
			t.Fatalf("fallback attempt %d should pass", i+1)
		}
	}
	if store.Allow("down-ip", 3, time.Minute) {
		t.Fatal("fallback 4th attempt should be blocked — fail-bounded, not fail-open")
	}
}
