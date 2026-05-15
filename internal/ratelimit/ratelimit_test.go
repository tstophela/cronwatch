package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func newLimiterWithClock(cooldown time.Duration, now *time.Time) *Limiter {
	return newWithClock(cooldown, func() time.Time { return *now })
}

func TestAllow_FirstCall_Permitted(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(time.Minute, &now)
	if !l.Allow("job-a") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_WithinCooldown_Suppressed(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(time.Minute, &now)
	l.Allow("job-a") // first call records time
	now = now.Add(30 * time.Second)
	if l.Allow("job-a") {
		t.Fatal("expected second call within cooldown to be suppressed")
	}
}

func TestAllow_AfterCooldown_Permitted(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(time.Minute, &now)
	l.Allow("job-a")
	now = now.Add(61 * time.Second)
	if !l.Allow("job-a") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_DifferentJobs_Independent(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(time.Minute, &now)
	l.Allow("job-a")
	if !l.Allow("job-b") {
		t.Fatal("expected different job to be allowed independently")
	}
}

func TestReset_ClearsSingleJob(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(time.Minute, &now)
	l.Allow("job-a")
	l.Allow("job-b")
	l.Reset("job-a")
	if !l.Allow("job-a") {
		t.Fatal("expected job-a to be allowed after reset")
	}
	// job-b should still be suppressed
	if l.Allow("job-b") {
		t.Fatal("expected job-b to remain suppressed")
	}
}

func TestResetAll_ClearsAllJobs(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(time.Minute, &now)
	l.Allow("job-a")
	l.Allow("job-b")
	l.ResetAll()
	if !l.Allow("job-a") {
		t.Fatal("expected job-a to be allowed after ResetAll")
	}
	if !l.Allow("job-b") {
		t.Fatal("expected job-b to be allowed after ResetAll")
	}
}

func TestAllow_ConcurrentAccess_NoPanic(t *testing.T) {
	now := time.Now()
	l := newLimiterWithClock(10*time.Millisecond, &now)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			job := "job-a"
			if i%2 == 0 {
				job = "job-b"
			}
			l.Allow(job)
		}(i)
	}
	wg.Wait()
}
