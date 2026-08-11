package workerlock

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func newTestLocker(t *testing.T) (*RedisLocker, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	locker, err := NewRedisLocker("redis://" + mr.Addr())
	if err != nil {
		t.Fatalf("NewRedisLocker: %v", err)
	}
	t.Cleanup(func() { _ = locker.Close() })
	return locker, mr
}

// TestRedisLockerRenewExtendsTTL proves the renewal primitive keeps a lease
// alive past its original TTL. It is the deterministic core of the heartbeat
// fix: miniredis does not expire keys in real time, so we fast-forward its
// clock instead.
func TestRedisLockerRenewExtendsTTL(t *testing.T) {
	locker, mr := newTestLocker(t)
	ctx := context.Background()
	const ttl = 50 * time.Millisecond

	token, ok, err := locker.Acquire(ctx, "key", ttl)
	if err != nil || !ok {
		t.Fatalf("acquire: err=%v ok=%v", err, ok)
	}

	// Drain most of the TTL, then renew. A second FastForward past the original
	// TTL must not expire the key while the lease is renewed.
	mr.FastForward(40 * time.Millisecond)
	if err := locker.Renew(ctx, "key", token, ttl); err != nil {
		t.Fatalf("renew: %v", err)
	}
	mr.FastForward(40 * time.Millisecond) // cumulative 80ms > 50ms original TTL

	token2, ok, err := locker.Acquire(ctx, "key", ttl)
	if err != nil {
		t.Fatalf("reacquire after renew: %v", err)
	}
	if ok {
		_ = locker.Release(ctx, "key", token2)
		t.Fatal("lock expired despite Renew: lease was not extended")
	}
	if err := locker.Release(ctx, "key", token); err != nil {
		t.Fatalf("release: %v", err)
	}
}

// TestRedisLockerRenewReportsLostLease verifies that renewing a lease that has
// already expired (or was taken over) surfaces ErrLockNotHeld so the heartbeat
// knows to stop instead of extending someone else's lock.
func TestRedisLockerRenewReportsLostLease(t *testing.T) {
	locker, mr := newTestLocker(t)
	ctx := context.Background()
	const ttl = 50 * time.Millisecond

	token, ok, err := locker.Acquire(ctx, "key", ttl)
	if err != nil || !ok {
		t.Fatalf("acquire: err=%v ok=%v", err, ok)
	}
	mr.FastForward(60 * time.Millisecond) // key expires

	if err := locker.Renew(ctx, "key", token, ttl); !errors.Is(err, ErrLockNotHeld) {
		t.Fatalf("renew after loss: got %v, want ErrLockNotHeld", err)
	}

	// A stolen token must not renew the current holder's lease.
	token2, ok, err := locker.Acquire(ctx, "key", ttl)
	if err != nil || !ok {
		t.Fatalf("reacquire: err=%v ok=%v", err, ok)
	}
	defer func() { _ = locker.Release(ctx, "key", token2) }()
	if err := locker.Renew(ctx, "key", token, ttl); !errors.Is(err, ErrLockNotHeld) {
		t.Fatalf("renew with stale token: got %v, want ErrLockNotHeld", err)
	}
}

// TestWithLockRenewsLeaseForLongCriticalSection is the end-to-end regression
// for the heartbeat. A critical section runs longer than the TTL; a competing
// worker must still not acquire the lock mid-flight. On the pre-fix path the
// lease expires (no renewal) and the competing acquire succeeds.
func TestWithLockRenewsLeaseForLongCriticalSection(t *testing.T) {
	locker, mr := newTestLocker(t)
	ctx := context.Background()
	const ttl = 400 * time.Millisecond // heartbeat ticks at ~200ms

	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	var acquired bool
	var lockErr error
	go func() {
		acquired, lockErr = WithLock(ctx, locker, "long-job", ttl, func(c context.Context) error {
			close(started)
			<-release // hold the critical section until the test is ready
			return nil
		})
		close(done)
	}()
	<-started

	// Advance miniredis's clock well past the original TTL in two steps, giving
	// the real-time heartbeat time to fire (ticks at ~200ms and ~400ms) in
	// between. Without renewal the key is gone after cumulative 500ms > ttl.
	mr.FastForward(300 * time.Millisecond)
	time.Sleep(500 * time.Millisecond)
	mr.FastForward(200 * time.Millisecond) // cumulative 500ms > 400ms TTL

	competeCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	token, ok, err := locker.Acquire(competeCtx, "long-job", ttl)
	if err != nil {
		t.Fatalf("competing acquire: %v", err)
	}
	if ok {
		_ = locker.Release(ctx, "long-job", token)
		t.Fatal("competing Acquire succeeded while the lock holder was still running: lease was not renewed")
	}

	close(release)
	<-done
	if lockErr != nil {
		t.Fatalf("WithLock: %v", lockErr)
	}
	if !acquired {
		t.Fatal("WithLock reported not acquired")
	}

	// Once fn returns, the lease must be released and the heartbeat stopped.
	reacquireCtx, cancel2 := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel2()
	token, ok, err = locker.Acquire(reacquireCtx, "long-job", ttl)
	if err != nil {
		t.Fatalf("reacquire after completion: %v", err)
	}
	if !ok {
		t.Fatal("lock not released after WithLock completed")
	}
	_ = locker.Release(ctx, "long-job", token)
}

// TestWithLockSkipsWhenHeld verifies the non-affected path: a worker that finds
// the lock already held must skip and report not-acquired, and must not disturb
// the holder's lease.
func TestWithLockSkipsWhenHeld(t *testing.T) {
	locker, _ := newTestLocker(t)
	ctx := context.Background()
	const ttl = time.Minute

	if _, ok, err := locker.Acquire(ctx, "job", ttl); err != nil || !ok {
		t.Fatalf("acquire holder: err=%v ok=%v", err, ok)
	}

	called := false
	acquired, err := WithLock(ctx, locker, "job", ttl, func(c context.Context) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithLock: %v", err)
	}
	if acquired {
		t.Fatal("WithLock reported acquired while the lock was held")
	}
	if called {
		t.Fatal("fn must not run when the lock is held by another worker")
	}
}
