// Package alerter provides a simple multi-sink alerting system for cronwatch.
//
// An Alerter dispatches Alert values to one or more Sink implementations.
// Built-in sinks include LogSink (writes to a logger) and MemorySink
// (stores alerts in memory, primarily for testing).
//
// Example usage:
//
//	logSink := alerter.NewLogSink(nil)
//	a := alerter.New(logSink)
//	a.Error("daily-backup", "non-zero exit code: 2")
package alerter
