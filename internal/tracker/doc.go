// Package tracker maintains in-memory execution state for monitored cron jobs.
//
// A Tracker records the last successful execution time of each job and tracks
// how many consecutive intervals have been missed. Based on a configurable
// threshold, a job transitions through the following states:
//
//	  unknown  →  ok  →  late  →  failed
//	                ↑________________|
//	                (Record resets to ok)
//
// # States
//
// The unknown state applies to jobs that have never been recorded. Once a
// successful execution is recorded, the job moves to ok. If the job misses
// one or more expected executions, it transitions to late, and then to failed
// once the configured miss threshold is exceeded. A successful Record call
// always resets the job back to ok regardless of its current state.
//
// # Concurrency
//
// Tracker is safe for concurrent use by multiple goroutines.
package tracker
