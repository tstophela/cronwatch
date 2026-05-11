// Package history provides a lightweight, file-backed ring buffer for
// cron job execution history.
//
// Usage:
//
//	store, err := history.New("/var/lib/cronwatch/history.json", 50)
//	if err != nil { ... }
//
//	// Record a completed execution:
//	store.Record(history.Entry{
//		JobName:  "backup",
//		RunAt:    time.Now(),
//		Duration: 3 * time.Second,
//		Success:  true,
//	})
//
//	// Retrieve recent history:
//	entries := store.Get("backup")
package history
