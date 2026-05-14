// Package digest provides a periodic summary (digest) of cron job health.
//
// A Digest is created with a Tracker, a Sender, and an interval. On each
// tick it calls reporter.GenerateSnapshot to collect current job states,
// formats a human-readable table, and delivers it via the Sender interface
// (typically a notifier.Multi wrapping one or more webhook targets).
//
// Usage:
//
//	d := digest.New(tracker, sender, 24*time.Hour, logger)
//	go d.Start(ctx)
package digest
