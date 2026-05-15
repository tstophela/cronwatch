// Package ratelimit provides per-job alert rate limiting to prevent
// alert storms when a job repeatedly fails in a short window.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks when alerts were last sent for each job and suppresses
// duplicate alerts within a configurable cooldown period.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[string]time.Time
	now      func() time.Time
}

// New creates a Limiter with the given cooldown duration. Alerts for the
// same job fired within the cooldown window will be suppressed.
func New(cooldown time.Duration) *Limiter {
	return &Limiter{
		cooldown: cooldown,
		lastSent: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if an alert for jobName should be sent, and records
// the send time. Returns false if the job is still within the cooldown
// window from the last alert.
func (l *Limiter) Allow(jobName string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if last, ok := l.lastSent[jobName]; ok {
		if now.Sub(last) < l.cooldown {
			return false
		}
	}
	l.lastSent[jobName] = now
	return true
}

// Reset clears the rate-limit state for a specific job, allowing the next
// alert to pass through immediately. Useful when a job recovers.
func (l *Limiter) Reset(jobName string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.lastSent, jobName)
}

// NextAllowed returns the time at which the next alert for jobName will be
// permitted. Returns the zero time if the job has no recorded send.
func (l *Limiter) NextAllowed(jobName string) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	if last, ok := l.lastSent[jobName]; ok {
		return last.Add(l.cooldown)
	}
	return time.Time{}
}
