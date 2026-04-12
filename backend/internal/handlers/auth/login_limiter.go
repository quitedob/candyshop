package auth

import (
	"sync"
	"time"
)

// SEC-7: Per-account brute-force protection for login attempts.
// Tracks consecutive failed login attempts per email and enforces
// a temporary lockout after maxAttempts consecutive failures.

const (
	// maxLoginAttempts is the number of consecutive failures before lockout.
	maxLoginAttempts = 5
	// lockoutDuration is how long the account is locked after exceeding maxLoginAttempts.
	lockoutDuration = 15 * time.Minute
	// maxTrackedAccounts caps tracked emails to prevent memory exhaustion from enumeration.
	maxTrackedAccounts = 50_000
)

// loginAttempt tracks failed login attempts for a single email.
type loginAttempt struct {
	failures  int
	lastFail  time.Time
	lockedAt  time.Time
	isLocked  bool
}

// LoginAttemptTracker is an in-memory per-account brute-force limiter.
type LoginAttemptTracker struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	stopCh   chan struct{}
}

// NewLoginAttemptTracker creates a new tracker and starts a background
// cleanup goroutine that evicts stale entries every 5 minutes.
func NewLoginAttemptTracker() *LoginAttemptTracker {
	t := &LoginAttemptTracker{
		attempts: make(map[string]*loginAttempt),
		stopCh:   make(chan struct{}),
	}
	go t.cleanup()
	return t
}

// Stop terminates the background cleanup goroutine.
func (t *LoginAttemptTracker) Stop() {
	close(t.stopCh)
}

// IsLocked returns true if the email is currently locked out,
// along with how many seconds remain until the lockout expires.
func (t *LoginAttemptTracker) IsLocked(email string) (bool, int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	a, ok := t.attempts[email]
	if !ok {
		return false, 0
	}

	if !a.isLocked {
		return false, 0
	}

	remaining := lockoutDuration - time.Since(a.lockedAt)
	if remaining <= 0 {
		// Lockout expired — reset
		a.isLocked = false
		a.failures = 0
		return false, 0
	}

	return true, int(remaining.Seconds())
}

// RecordFailure records a failed login attempt. If maxLoginAttempts is
// reached, the account is locked for lockoutDuration.
func (t *LoginAttemptTracker) RecordFailure(email string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	a, ok := t.attempts[email]
	if !ok {
		// Evict stale entries if capacity exceeded
		if len(t.attempts) >= maxTrackedAccounts {
			t.evictExpiredLocked()
		}
		a = &loginAttempt{}
		t.attempts[email] = a
	}

	// If previously locked and lockout expired, reset first
	if a.isLocked && time.Since(a.lockedAt) >= lockoutDuration {
		a.isLocked = false
		a.failures = 0
	}

	a.failures++
	a.lastFail = time.Now()

	if a.failures >= maxLoginAttempts {
		a.isLocked = true
		a.lockedAt = time.Now()
	}
}

// RecordSuccess clears all failure tracking for the email (successful login).
func (t *LoginAttemptTracker) RecordSuccess(email string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, email)
}

// cleanup periodically removes expired entries.
func (t *LoginAttemptTracker) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.mu.Lock()
			t.evictExpiredLocked()
			t.mu.Unlock()
		case <-t.stopCh:
			return
		}
	}
}

// evictExpiredLocked removes entries that are no longer locked and have
// no recent activity. Must be called with mu held.
func (t *LoginAttemptTracker) evictExpiredLocked() {
	now := time.Now()
	for email, a := range t.attempts {
		// Remove if lockout expired OR last failure was more than lockoutDuration ago
		if a.isLocked && now.Sub(a.lockedAt) >= lockoutDuration {
			delete(t.attempts, email)
		} else if !a.isLocked && now.Sub(a.lastFail) >= lockoutDuration {
			delete(t.attempts, email)
		}
	}
}
