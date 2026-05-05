package tracker

import (
	"testing"
	"time"
)

func TestRecord_SetsOK(t *testing.T) {
	tr := New()
	now := time.Now()
	tr.Record("backup", now)

	e, ok := tr.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != StatusOK {
		t.Errorf("expected StatusOK, got %s", e.Status)
	}
	if !e.LastSeen.Equal(now) {
		t.Errorf("expected LastSeen %v, got %v", now, e.LastSeen)
	}
	if e.MissedCount != 0 {
		t.Errorf("expected MissedCount 0, got %d", e.MissedCount)
	}
}

func TestMarkLate_BelowThreshold(t *testing.T) {
	tr := New()
	tr.MarkLate("sync", 3)

	e, ok := tr.Get("sync")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != StatusLate {
		t.Errorf("expected StatusLate, got %s", e.Status)
	}
	if e.MissedCount != 1 {
		t.Errorf("expected MissedCount 1, got %d", e.MissedCount)
	}
}

func TestMarkLate_AtThreshold_SetsFailed(t *testing.T) {
	tr := New()
	for i := 0; i < 3; i++ {
		tr.MarkLate("sync", 3)
	}

	e, _ := tr.Get("sync")
	if e.Status != StatusFailed {
		t.Errorf("expected StatusFailed, got %s", e.Status)
	}
}

func TestRecord_ResetsAfterLate(t *testing.T) {
	tr := New()
	tr.MarkLate("sync", 3)
	tr.MarkLate("sync", 3)
	tr.Record("sync", time.Now())

	e, _ := tr.Get("sync")
	if e.Status != StatusOK {
		t.Errorf("expected StatusOK after record, got %s", e.Status)
	}
	if e.MissedCount != 0 {
		t.Errorf("expected MissedCount reset to 0, got %d", e.MissedCount)
	}
}

func TestGet_Missing(t *testing.T) {
	tr := New()
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	tr := New()
	tr.Record("job1", time.Now())
	tr.Record("job2", time.Now())

	all := tr.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestStatus_String(t *testing.T) {
	cases := map[Status]string{
		StatusOK:      "ok",
		StatusLate:    "late",
		StatusFailed:  "failed",
		StatusUnknown: "unknown",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Errorf("Status(%d).String() = %q, want %q", s, got, want)
		}
	}
}
