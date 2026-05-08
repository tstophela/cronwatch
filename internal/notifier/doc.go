// Package notifier defines the Notifier interface and concrete backends
// (webhook, …) that deliver alert messages to external systems.
//
// Each backend implements:
//
//	Send(ctx context.Context, subject, body string) error
//
// Backends are intentionally decoupled from the alerter package so they can
// be tested and swapped independently.
package notifier
