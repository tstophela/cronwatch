// Package heartbeat implements a lightweight HTTP server that cron jobs
// can POST to upon successful completion. Each ping is forwarded to the
// tracker so that the watcher can determine whether a job ran on time.
//
// Endpoints:
//
//	POST /ping/<job-name>  — record a successful execution for the named job.
//	GET  /healthz          — liveness probe; always returns 200 OK.
//
// Usage in a cron job:
//
//	*/5 * * * * /usr/local/bin/myjob && curl -s -X POST http://localhost:8765/ping/myjob
package heartbeat
