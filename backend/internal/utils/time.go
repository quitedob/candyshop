package utils

import (
	"sync/atomic"
	"time"
)

// currentTime is stored as atomic.Value for thread-safe access
var currentTime atomic.Value

func init() {
	currentTime.Store(time.Now)
}

// Now returns the current time, using the configured time function
func Now() time.Time {
	if fn, ok := currentTime.Load().(func() time.Time); ok {
		return fn()
	}
	return time.Now()
}

// SetTimeFunc sets a custom time function for testing
func SetTimeFunc(fn func() time.Time) {
	currentTime.Store(fn)
}

// ResetTimeFunc resets the time function to default
func ResetTimeFunc() {
	currentTime.Store(time.Now)
}
