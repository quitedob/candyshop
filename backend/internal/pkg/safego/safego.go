// Package safego provides goroutine launchers with panic recovery.
//
// Background workers and async hooks should use Go (or Run) instead of bare
// `go fn()` so a panic in one goroutine cannot crash the entire process.
package safego

import (
	"log"
	"runtime/debug"
)

// Go launches fn in a new goroutine, recovering from any panic and logging it
// with the supplied label and a captured stack trace.
//
// Use this for fire-and-forget background work where a panic must not crash
// the process (webhook delivery, event hooks, periodic workers, async email
// dispatch, etc.). The label should identify the call site for log triage.
func Go(label string, fn func()) {
	if fn == nil {
		return
	}
	go Run(label, fn)
}

// Run executes fn synchronously with panic recovery. Useful inside an existing
// goroutine (e.g. a worker loop) to scope recovery to a single iteration
// without spawning a new goroutine.
func Run(label string, fn func()) {
	if fn == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			if label == "" {
				label = "safego"
			}
			log.Printf("panic recovered in %s: %v\n%s", label, r, debug.Stack())
		}
	}()
	fn()
}
