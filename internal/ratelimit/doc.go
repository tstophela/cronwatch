// Package ratelimit implements per-job cooldown tracking for the cronwatch
// alerting pipeline.
//
// When a cron job enters a failure or late state it may trigger repeated
// alerts on every watcher tick. The Limiter wraps the alert dispatch path
// and suppresses duplicate notifications for the same job until a
// configurable cooldown period has elapsed.
//
// Typical usage:
//
//	limiter := ratelimit.New(15 * time.Minute)
//	if limiter.Allow(jobName) {
//		alerter.Fire(ctx, jobName, message)
//	}
package ratelimit
