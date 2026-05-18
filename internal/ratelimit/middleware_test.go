package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestAlertMiddleware_Wrap_AllowsFirst(t *testing.T) {
	m := NewAlertMiddleware(5 * time.Minute)

	called := 0
	wrapped := m.Wrap(func(job string) { called++ })

	if ok := wrapped("backup"); !ok {
		t.Fatal("expected first call to be allowed")
	}
	if called != 1 {
		t.Fatalf("expected fn called once, got %d", called)
	}
}

func TestAlertMiddleware_Wrap_SuppressesWithinCooldown(t *testing.T) {
	m := NewAlertMiddleware(5 * time.Minute)

	called := 0
	wrapped := m.Wrap(func(job string) { called++ })

	wrapped("backup")
	ok := wrapped("backup")

	if ok {
		t.Fatal("expected second call to be suppressed")
	}
	if called != 1 {
		t.Fatalf("expected fn called once, got %d", called)
	}
}

func TestMultiMiddleware_Allow_FirstCallPermitted(t *testing.T) {
	m := NewMultiMiddleware(10 * time.Minute)

	if !m.Allow("jobA") {
		t.Fatal("expected first call for jobA to be allowed")
	}
}

func TestMultiMiddleware_Allow_SuppressesRepeat(t *testing.T) {
	m := NewMultiMiddleware(10 * time.Minute)
	m.Allow("jobA")

	if m.Allow("jobA") {
		t.Fatal("expected repeated call for jobA to be suppressed")
	}
}

func TestMultiMiddleware_Allow_IndependentJobs(t *testing.T) {
	m := NewMultiMiddleware(10 * time.Minute)
	m.Allow("jobA")

	if !m.Allow("jobB") {
		t.Fatal("expected jobB to be independent of jobA cooldown")
	}
}

func TestMultiMiddleware_Reset_ClearsState(t *testing.T) {
	m := NewMultiMiddleware(10 * time.Minute)
	m.Allow("jobA")
	m.Reset("jobA")

	if !m.Allow("jobA") {
		t.Fatal("expected jobA to be allowed after reset")
	}
}

func TestMultiMiddleware_ConcurrentAccess(t *testing.T) {
	m := NewMultiMiddleware(1 * time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			job := "job"
			if n%2 == 0 {
				job = "other"
			}
			m.Allow(job)
		}(i)
	}
	wg.Wait() // should not race or panic
}
