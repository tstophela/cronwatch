// Package ratelimit provides per-job alert suppression to prevent
// alert storms when a job repeatedly fails within a short window.
package ratelimit

import (
	"sync"
	"time"
)

// Clock abstracts time for testability.
type Clock func() time.Time

// Limiter tracks the last alert time per job and suppresses alerts
// that occur within the configured cooldown period.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	clock    Clock
}

// New creates a Limiter with the given cooldown duration.
// Alerts for the same job fired within cooldown are suppressed.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		clock:    time.Now,
	}
}

// newWithClock is used in tests to inject a custom clock.
func newWithClock(cooldown time.Duration, clock Clock) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		clock:    clock,
	}
}

// Allow returns true if an alert for jobName should be sent.
// It returns false if an alert was already sent within the cooldown window.
func (l *Limiter) Allow(jobName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	if t, ok := l.last[jobName]; ok {
		if now.Sub(t) < l.cooldown {
			return false
		}
	}
	l.last[jobName] = now
	return true
}

// Reset clears the rate-limit state for a specific job.
// Useful when a job recovers and the cooldown should restart fresh.
func (l *Limiter) Reset(jobName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, jobName)
}

// ResetAll clears rate-limit state for all jobs.
func (l *Limiter) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.last = make(map[string]time.Time)
}
