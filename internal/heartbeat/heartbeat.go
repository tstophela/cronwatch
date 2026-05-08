// Package heartbeat provides an HTTP endpoint for cron jobs to ping
// upon successful completion, updating the tracker with a fresh timestamp.
package heartbeat

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/example/cronwatch/internal/tracker"
)

// Recorder is the subset of tracker.Tracker used by the heartbeat handler.
type Recorder interface {
	Record(name string, at time.Time)
}

// Server listens for incoming pings from cron jobs.
type Server struct {
	tracker Recorder
	addr    string
	server  *http.Server
}

// New creates a heartbeat Server that binds to addr.
func New(t Recorder, addr string) *Server {
	s := &Server{tracker: t, addr: addr}
	mux := http.NewServeMux()
	mux.HandleFunc("/ping/", s.handlePing)
	mux.HandleFunc("/healthz", s.handleHealth)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return s
}

// Start begins serving in a background goroutine.
func (s *Server) Start() {
	go func() {
		log.Printf("heartbeat: listening on %s", s.addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("heartbeat: server error: %v", err)
		}
	}()
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop() error {
	return s.server.Close()
}

// handlePing processes POST /ping/<job-name> requests.
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	name := r.URL.Path[len("/ping/"):]
	if name == "" {
		http.Error(w, "job name required", http.StatusBadRequest)
		return
	}
	s.tracker.Record(name, time.Now().UTC())
	log.Printf("heartbeat: ping received for job %q", name)
	w.WriteHeader(http.StatusNoContent)
	fmt.Fprintf(w, "")
}

// handleHealth is a simple liveness probe.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
}
