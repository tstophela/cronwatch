package ratelimit

import (
	"sync"
	"time"
)

// AlertMiddleware wraps an alert-firing function and suppresses repeated
// alerts for the same job within a configurable cooldown window.
type AlertMiddleware struct {
	limiter *Limiter
}

// NewAlertMiddleware returns an AlertMiddleware backed by an in-memory
// rate limiter with the given cooldown duration.
func NewAlertMiddleware(cooldown time.Duration) *AlertMiddleware {
	return &AlertMiddleware{
		limiter: New(cooldown),
	}
}

// Wrap returns a new function that calls fn only when the rate limiter
// permits an alert for the given job name. Suppressed calls are silently
// dropped; the boolean return indicates whether fn was invoked.
func (m *AlertMiddleware) Wrap(fn func(job string)) func(job string) bool {
	return func(job string) bool {
		if !m.limiter.Allow(job) {
			return false
		}
		fn(job)
		return true
	}
}

// MultiMiddleware applies independent cooldowns per job across multiple
// concurrent goroutines safely.
type MultiMiddleware struct {
	mu       sync.Mutex
	cooldown time.Duration
	limiters map[string]*Limiter
}

// NewMultiMiddleware returns a MultiMiddleware where each job gets its
// own isolated Limiter created lazily on first use.
func NewMultiMiddleware(cooldown time.Duration) *MultiMiddleware {
	return &MultiMiddleware{
		cooldown: cooldown,
		limiters: make(map[string]*Limiter),
	}
}

// Allow returns true if the alert for job should be forwarded, false if
// it is within the cooldown window and should be suppressed.
func (m *MultiMiddleware) Allow(job string) bool {
	m.mu.Lock()
	l, ok := m.limiters[job]
	if !ok {
		l = New(m.cooldown)
		m.limiters[job] = l
	}
	m.mu.Unlock()
	return l.Allow(job)
}

// Reset clears the cooldown state for a specific job, allowing the next
// alert to pass through immediately.
func (m *MultiMiddleware) Reset(job string) {
	m.mu.Lock()
	delete(m.limiters, job)
	m.mu.Unlock()
}
