package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// NextRun returns the next scheduled run time for a cron expression.
func NextRun(expr string, from time.Time) (time.Time, error) {
	schedule, err := parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(from), nil
}

// PreviousExpected returns the most recent expected run time before or at `at`.
// It steps back in small increments to find the last scheduled tick.
func PreviousExpected(expr string, at time.Time) (time.Time, error) {
	schedule, err := parse(expr)
	if err != nil {
		return time.Time{}, err
	}

	// Walk backward by checking next run from progressively earlier points.
	// We search within a window of 366 days to handle yearly schedules.
	const maxLookback = 366 * 24 * time.Hour
	const step = time.Minute

	candidate := at.Add(-maxLookback)
	var prev time.Time
	for candidate.Before(at) {
		next := schedule.Next(candidate)
		if next.IsZero() || next.After(at) {
			break
		}
		prev = next
		candidate = next.Add(step)
	}

	if prev.IsZero() {
		return time.Time{}, fmt.Errorf("no scheduled run found before %s for expression %q", at.Format(time.RFC3339), expr)
	}
	return prev, nil
}

// Validate checks whether a cron expression is syntactically valid.
func Validate(expr string) error {
	_, err := parse(expr)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}
	return nil
}

func parse(expr string) (cron.Schedule, error) {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	return parser.Parse(expr)
}
