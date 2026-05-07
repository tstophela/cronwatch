// Package reporter generates human-readable status reports for monitored
// cron jobs by querying the tracker for current job state.
//
// Usage:
//
//	g := reporter.New(trk)
//	report := g.Generate(jobNames)
//	reporter.Write(os.Stdout, report)
//
// Reports include each job's name, status, last-seen timestamp, and the
// number of missed runs detected since the daemon started.
package reporter
