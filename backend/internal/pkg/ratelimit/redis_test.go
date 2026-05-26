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
