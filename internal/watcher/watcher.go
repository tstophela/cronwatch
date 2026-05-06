// Package watcher polls tracked jobs and fires alerts when jobs are late or missing.
package watcher

import (
	"log"
	"time"

	"github.com/example/cronwatch/internal/alerter"
	"github.com/example/cronwatch/internal/config"
	"github.com/example/cronwatch/internal/schedule"
	"github.com/example/cronwatch/internal/tracker"
)

// Watcher periodically checks each configured job for lateness or missed runs.
type Watcher struct {
	cfg      *config.Config
	tracker  *tracker.Tracker
	alerter  *alerter.Alerter
	interval time.Duration
	stopCh   chan struct{}
}

// New creates a Watcher with the given dependencies.
func New(cfg *config.Config, tr *tracker.Tracker, al *alerter.Alerter, interval time.Duration) *Watcher {
	return &Watcher{
		cfg:      cfg,
		tracker:  tr,
		alerter:  al,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start launches the polling loop in a background goroutine.
func (w *Watcher) Start() {
	go w.loop()
}

// Stop signals the polling loop to exit.
func (w *Watcher) Stop() {
	close(w.stopCh)
}

func (w *Watcher) loop() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.checkAll(time.Now())
		case <-w.stopCh:
			return
		}
	}
}

// checkAll evaluates every job in the config against the current time.
func (w *Watcher) checkAll(now time.Time) {
	for _, job := range w.cfg.Jobs {
		prev, err := schedule.PreviousExpected(job.Schedule, now)
		if err != nil {
			log.Printf("watcher: invalid schedule for job %q: %v", job.Name, err)
			continue
		}

		state, ok := w.tracker.Get(job.Name)
		if !ok {
			// Job has never been recorded; check if it should have run by now.
			if now.Sub(prev) > job.GracePeriod {
				w.alerter.Fire(job.Name, "job has never reported and is overdue")
			}
			continue
		}

		if state.LastSeen.Before(prev) {
			w.tracker.MarkLate(job.Name, job.GracePeriod, now.Sub(prev))
			if state.Failed {
				w.alerter.Fire(job.Name, "job missed its scheduled window")
			} else {
				w.alerter.Warn(job.Name, "job is late")
			}
		}
	}
}
