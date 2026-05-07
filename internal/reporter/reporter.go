// Package reporter provides a summary report of all monitored cron job statuses.
package reporter

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/cronwatch/internal/tracker"
)

// Report holds a snapshot of job statuses at a point in time.
type Report struct {
	GeneratedAt time.Time
	Entries     []Entry
}

// Entry represents the status of a single cron job.
type Entry struct {
	JobName    string
	Status     string
	LastSeen   time.Time
	MissedRuns int
}

// Generator builds reports from tracker state.
type Generator struct {
	trk *tracker.Tracker
}

// New creates a new Generator backed by the given Tracker.
func New(trk *tracker.Tracker) *Generator {
	return &Generator{trk: trk}
}

// Generate produces a Report from current tracker state for the given job names.
func (g *Generator) Generate(jobNames []string) Report {
	r := Report{
		GeneratedAt: time.Now().UTC(),
	}

	for _, name := range jobNames {
		state, ok := g.trk.Get(name)
		if !ok {
			r.Entries = append(r.Entries, Entry{
				JobName: name,
				Status:  "unknown",
			})
			continue
		}
		r.Entries = append(r.Entries, Entry{
			JobName:    name,
			Status:     string(state.Status),
			LastSeen:   state.LastSeen,
			MissedRuns: state.MissedRuns,
		})
	}

	sort.Slice(r.Entries, func(i, j int) bool {
		return r.Entries[i].JobName < r.Entries[j].JobName
	})

	return r
}

// Write renders the report as plain text to w.
func Write(w io.Writer, r Report) {
	fmt.Fprintf(w, "cronwatch report — %s\n", r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "%-30s %-10s %-25s %s\n", "JOB", "STATUS", "LAST SEEN", "MISSED")
	fmt.Fprintln(w, "--------------------------------------------------------------------------------")
	for _, e := range r.Entries {
		lastSeen := "never"
		if !e.LastSeen.IsZero() {
			lastSeen = e.LastSeen.Format(time.RFC3339)
		}
		fmt.Fprintf(w, "%-30s %-10s %-25s %d\n", e.JobName, e.Status, lastSeen, e.MissedRuns)
	}
}
