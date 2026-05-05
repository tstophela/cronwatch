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
// Tracker is safe for concurrent use.
package tracker
