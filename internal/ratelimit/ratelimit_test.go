package ratelimit

import (
	"testing"
	"time"
)

func newLimiterWithClock(cooldown time.Duration, now func() time.Time) *Limiter {
	l := New(cooldown)
	l.now = now
	return l
}

func TestAllow_FirstCall_Permitted(t *testing.T) {
	l := New(5 * time.Minute)
	if !l.Allow("backup") {
		t.Fatal("expected first alert to be allowed")
	}
}

func TestAllow_WithinCooldown_Suppressed(t *testing.T) {
	base := time.Now()
	l := newLimiterWithClock(10*time.Minute, func() time.Time { return base })

	l.Allow("backup") // record first send

	// advance by less than cooldown
	l.now = func() time.Time { return base.Add(5 * time.Minute) }
	if l.Allow("backup") {
		t.Fatal("expected alert to be suppressed within cooldown")
	}
}

func TestAllow_AfterCooldown_Permitted(t *testing.T) {
	base := time.Now()
	l := newLimiterWithClock(10*time.Minute, func() time.Time { return base })

	l.Allow("backup")

	l.now = func() time.Time { return base.Add(11 * time.Minute) }
	if !l.Allow("backup") {
		t.Fatal("expected alert to be allowed after cooldown")
	}
}

func TestAllow_DifferentJobs_Independent(t *testing.T) {
	base := time.Now()
	l := newLimiterWithClock(10*time.Minute, func() time.Time { return base })

	l.Allow("jobA")

	if !l.Allow("jobB") {
		t.Fatal("expected independent job to be allowed")
	}
}

func TestReset_ClearsState(t *testing.T) {
	base := time.Now()
	l := newLimiterWithClock(10*time.Minute, func() time.Time { return base })

	l.Allow("backup")
	l.Reset("backup")

	if !l.Allow("backup") {
		t.Fatal("expected alert to be allowed after reset")
	}
}

func TestNextAllowed_NoHistory_ReturnsZero(t *testing.T) {
	l := New(5 * time.Minute)
	if !l.NextAllowed("missing").IsZero() {
		t.Fatal("expected zero time for unseen job")
	}
}

func TestNextAllowed_ReturnsCorrectTime(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	cooldown := 15 * time.Minute
	l := newLimiterWithClock(cooldown, func() time.Time { return base })

	l.Allow("sync")

	want := base.Add(cooldown)
	got := l.NextAllowed("sync")
	if !got.Equal(want) {
		t.Fatalf("NextAllowed: got %v, want %v", got, want)
	}
}
