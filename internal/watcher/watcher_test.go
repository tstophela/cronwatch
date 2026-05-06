package watcher_test

import (
	"testing"
	"time"

	"github.com/example/cronwatch/internal/alerter"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/tracker"
	"github.com/example/cronwatch/internal/watcher"
)

func newStack(jobs []config.Job) (*config.Config, *tracker.Tracker, *alerter.MemorySink, *alerter.Alerter) {
	cfg := &config.Config{Jobs: jobs}
	tr := tracker.New()
	sink := alerter.NewMemorySink()
	al := alerter.New(sink)
	return cfg, tr, sink, al
}

func TestCheckAll_NeverSeen_OverGrace_Fires(t *testing.T) {
	jobs := []config.Job{
		{Name: "backup", Schedule: "* * * * *", GracePeriod: 1 * time.Minute},
	}
	cfg, tr, sink, al := newStack(jobs)
	w := watcher.New(cfg, tr, al, time.Hour)

	// Advance time so the job is well past its grace period.
	now := time.Now().Add(10 * time.Minute)
	w.CheckAll(now) // exported for testing via internal test helper

	if len(sink.Alerts()) == 0 {
		t.Fatal("expected an alert for never-seen overdue job")
	}
}

func TestCheckAll_RecentlySeen_NoAlert(t *testing.T) {
	jobs := []config.Job{
		{Name: "cleanup", Schedule: "* * * * *", GracePeriod: 2 * time.Minute},
	}
	cfg, tr, sink, al := newStack(jobs)
	w := watcher.New(cfg, tr, al, time.Hour)

	now := time.Now()
	tr.Record("cleanup", true)
	w.CheckAll(now)

	if len(sink.Alerts()) != 0 {
		t.Fatalf("expected no alerts, got %d", len(sink.Alerts()))
	}
}

func TestCheckAll_LateJob_Warns(t *testing.T) {
	jobs := []config.Job{
		{Name: "report", Schedule: "* * * * *", GracePeriod: 30 * time.Second},
	}
	cfg, tr, sink, al := newStack(jobs)
	w := watcher.New(cfg, tr, al, time.Hour)

	// Record a run 5 minutes ago so the job is late.
	tr.RecordAt("report", true, time.Now().Add(-5*time.Minute))
	w.CheckAll(time.Now())

	alerts := sink.Alerts()
	if len(alerts) == 0 {
		t.Fatal("expected a warn alert for late job")
	}
}

func TestStart_Stop_DoesNotPanic(t *testing.T) {
	cfg := &config.Config{}
	tr := tracker.New()
	al := alerter.New(alerter.NewMemorySink())
	w := watcher.New(cfg, tr, al, 50*time.Millisecond)
	w.Start()
	time.Sleep(120 * time.Millisecond)
	w.Stop()
}
