package ratelimit

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// persistedEntry is the on-disk representation of a rate-limit record.
type persistedEntry struct {
	LastFired time.Time `json:"last_fired"`
}

// Store persists rate-limit state to disk so that cooldowns survive restarts.
type Store struct {
	mu   sync.Mutex
	path string
	data map[string]persistedEntry
}

// NewStore loads (or creates) a persistent store at the given path.
func NewStore(path string) (*Store, error) {
	s := &Store{
		path: path,
		data: make(map[string]persistedEntry),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

// Get returns the last-fired time for a job key, and whether it was found.
func (s *Store) Get(key string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.data[key]
	return e.LastFired, ok
}

// Set records the last-fired time for a job key and flushes to disk.
func (s *Store) Set(key string, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = persistedEntry{LastFired: t}
	return s.flush()
}

// load reads the JSON file from disk into memory.
func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.data)
}

// flush writes the current in-memory state to disk atomically.
func (s *Store) flush() error {
	tmp := s.path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.data); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, s.path)
}
