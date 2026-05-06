package alerter

import (
	"fmt"
	"log"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds the details of a single alert event.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	OccurredAt time.Time
}

// Sink is the interface that alert destinations must implement.
type Sink interface {
	Send(a Alert) error
}

// Alerter dispatches alerts to one or more sinks.
type Alerter struct {
	sinks []Sink
}

// New creates an Alerter with the provided sinks.
func New(sinks ...Sink) *Alerter {
	return &Alerter{sinks: sinks}
}

// Fire sends an alert at the given level to all registered sinks.
// Errors from individual sinks are logged but do not abort delivery to others.
func (a *Alerter) Fire(level Level, jobName, message string) {
	alert := Alert{
		JobName:    jobName,
		Level:      level,
		Message:    message,
		OccurredAt: time.Now().UTC(),
	}
	for _, s := range a.sinks {
		if err := s.Send(alert); err != nil {
			log.Printf("alerter: sink error for job %q: %v", jobName, err)
		}
	}
}

// Warn is a convenience wrapper for Level=WARN.
func (a *Alerter) Warn(jobName, message string) {
	a.Fire(LevelWarn, jobName, message)
}

// Error is a convenience wrapper for Level=ERROR.
func (a *Alerter) Error(jobName, message string) {
	a.Fire(LevelError, jobName, message)
}

// FormatMessage returns a human-readable alert string.
func FormatMessage(a Alert) string {
	return fmt.Sprintf("[%s] %s — %s (at %s)",
		a.Level, a.JobName, a.Message, a.OccurredAt.Format(time.RFC3339))
}
