// Package history provides persistent storage for job execution records,
// allowing cronwatch to survive restarts and report on historical trends.
package history

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry represents a single recorded execution of a cron job.
type Entry struct {
	JobName   string        `json:"job_name"`
	RunAt     time.Time     `json:"run_at"`
	Duration  time.Duration `json:"duration_ns"`
	Success   bool          `json:"success"`
	ExitCode  int           `json:"exit_code,omitempty"`
}

// Store holds a bounded ring of execution history entries per job.
type Store struct {
	mu      sync.RWMutex
	path    string
	max     int
	entries map[string][]Entry
}

// New creates a Store that persists to path and keeps at most maxPerJob
// entries per job. Existing data is loaded from path if it exists.
func New(path string, maxPerJob int) (*Store, error) {
	s := &Store{
		path:    path,
		max:     maxPerJob,
		entries: make(map[string][]Entry),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Record appends an entry for the given job, evicting the oldest if needed.
func (s *Store) Record(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := s.entries[e.JobName]
	list = append(list, e)
	if len(list) > s.max {
		list = list[len(list)-s.max:]
	}
	s.entries[e.JobName] = list
	return s.save()
}

// Get returns a copy of all stored entries for jobName.
func (s *Store) Get(jobName string) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, len(s.entries[jobName]))
	copy(out, s.entries[jobName])
	return out
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.entries)
}

func (s *Store) save() error {
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.entries); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(tmp, s.path)
}
