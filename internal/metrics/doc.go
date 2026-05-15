// Package metrics provides a simple, allocation-friendly counter registry
// for tracking cronwatch runtime events such as pings received, alerts
// fired, and late-job detections.
//
// Counters are identified by string names and are safe for concurrent
// use without additional locking. The registry can produce a consistent
// snapshot at any time for reporting or HTTP exposure.
package metrics
