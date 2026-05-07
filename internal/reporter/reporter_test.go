package reporter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/cronwatch/internal/reporter"
	"github.com/cronwatch/internal/tracker"
)

func newTracker() *tracker.Tracker {
	return tracker.New()
}

func TestGenerate_UnknownJob(t *testing.T) {
	trk := newTracker()
	g := reporter.New(trk)

	r := g.Generate([]string{"missing-job"})

	if len(r.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(r.Entries))
	}
	if r.Entries[0].Status != "unknown" {
		t.Errorf("expected status 'unknown', got %q", r.Entries[0].Status)
	}
}

func TestGenerate_KnownJob(t *testing.T) {
	trk := newTracker()
	trk.Record("backup", time.Now())
	g := reporter.New(trk)

	r := g.Generate([]string{"backup"})

	if len(r.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(r.Entries))
	}
	if r.Entries[0].Status != "ok" {
		t.Errorf("expected status 'ok', got %q", r.Entries[0].Status)
	}
	if r.Entries[0].LastSeen.IsZero() {
		t.Error("expected LastSeen to be set")
	}
}

func TestGenerate_SortedByName(t *testing.T) {
	trk := newTracker()
	trk.Record("zebra", time.Now())
	trk.Record("alpha", time.Now())
	g := reporter.New(trk)

	r := g.Generate([]string{"zebra", "alpha"})

	if r.Entries[0].JobName != "alpha" || r.Entries[1].JobName != "zebra" {
		t.Errorf("entries not sorted: %v, %v", r.Entries[0].JobName, r.Entries[1].JobName)
	}
}

func TestWrite_ContainsHeaders(t *testing.T) {
	trk := newTracker()
	trk.Record("myjob", time.Now())
	g := reporter.New(trk)
	r := g.Generate([]string{"myjob"})

	var buf bytes.Buffer
	reporter.Write(&buf, r)
	out := buf.String()

	for _, want := range []string{"cronwatch report", "JOB", "STATUS", "LAST SEEN", "myjob", "ok"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}

func TestGenerate_EmptyJobList(t *testing.T) {
	trk := newTracker()
	g := reporter.New(trk)

	r := g.Generate([]string{})

	if len(r.Entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(r.Entries))
	}
	if r.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
}
