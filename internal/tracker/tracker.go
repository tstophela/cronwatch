package tracker

import (
	"sync"
	"time"
)

// Status represents the last known execution state of a cron job.
type Status int

const (
	StatusUnknown Status = iota
	StatusOK
	StatusLate
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "ok"
	case StatusLate:
		return "late"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Entry holds tracking data for a single monitored job.
type Entry struct {
	JobName     string
	LastSeen    time.Time
	Status      Status
	MissedCount int
}

// Tracker maintains in-memory state for all monitored cron jobs.
type Tracker struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{
		entries: make(map[string]*Entry),
	}
}

// Record marks a job as having executed successfully at the given time.
func (t *Tracker) Record(jobName string, at time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.entries[jobName]
	if !ok {
		e = &Entry{JobName: jobName}
		t.entries[jobName] = e
	}
	e.LastSeen = at
	e.Status = StatusOK
	e.MissedCount = 0
}

// MarkLate increments the missed counter and sets status to Late or Failed.
func (t *Tracker) MarkLate(jobName string, threshold int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	e, ok := t.entries[jobName]
	if !ok {
		e = &Entry{JobName: jobName}
		t.entries[jobName] = e
	}
	e.MissedCount++
	if e.MissedCount >= threshold {
		e.Status = StatusFailed
	} else {
		e.Status = StatusLate
	}
}

// Get returns a copy of the entry for jobName, and whether it existed.
func (t *Tracker) Get(jobName string) (Entry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.entries[jobName]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all tracked entries.
func (t *Tracker) All() []Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Entry, 0, len(t.entries))
	for _, e := range t.entries {
		out = append(out, *e)
	}
	return out
}

// Delete removes the entry for jobName from the tracker.
// It is a no-op if the job is not currently tracked.
func (t *Tracker) Delete(jobName string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, jobName)
}
