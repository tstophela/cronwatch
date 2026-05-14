// Package digest generates periodic summary reports of monitored cron job
// health and delivers them via the alerter pipeline.
package digest

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/example/cronwatch/internal/reporter"
	"github.com/example/cronwatch/internal/tracker"
)

// Sender is anything that can deliver a digest message (e.g. a notifier.Multi).
type Sender interface {
	Send(ctx context.Context, subject, body string) error
}

// Digest periodically summarises job health and sends it via Sender.
type Digest struct {
	tracker  *tracker.Tracker
	sender   Sender
	interval time.Duration
	logger   *slog.Logger
}

// New creates a Digest that fires every interval.
func New(t *tracker.Tracker, s Sender, interval time.Duration, logger *slog.Logger) *Digest {
	if logger == nil {
		logger = slog.Default()
	}
	return &Digest{
		tracker:  t,
		sender:   s,
		interval: interval,
		logger:   logger,
	}
}

// Start runs the digest loop until ctx is cancelled.
func (d *Digest) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.send(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Digest) send(ctx context.Context) {
	snap := reporter.GenerateSnapshot(d.tracker)
	body := formatDigest(snap)
	subject := fmt.Sprintf("cronwatch digest — %s", time.Now().UTC().Format(time.RFC1123))
	if err := d.sender.Send(ctx, subject, body); err != nil {
		d.logger.Error("digest send failed", "err", err)
	}
}

func formatDigest(snap *reporter.Snapshot) string {
	if len(snap.Jobs) == 0 {
		return "No jobs tracked yet.\n"
	}
	out := fmt.Sprintf("%-30s %-10s %-8s %s\n", "JOB", "STATUS", "RUNS", "LAST SEEN")
	out += fmt.Sprintf("%s\n", "----------------------------------------------------------------------")
	for _, j := range snap.Jobs {
		last := "never"
		if !j.LastSeen.IsZero() {
			last = j.LastSeen.UTC().Format(time.RFC3339)
		}
		out += fmt.Sprintf("%-30s %-10s %-8d %s\n", j.Name, j.Status, j.RunCount, last)
	}
	return out
}
