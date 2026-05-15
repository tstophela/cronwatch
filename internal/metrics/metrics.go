// Package metrics exposes lightweight runtime counters for cronwatch.
// Counters are safe for concurrent use and can be scraped via the
// heartbeat HTTP server or written to a log sink on demand.
package metrics

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Counter is a monotonically increasing uint64.
type Counter struct {
	v uint64
}

func (c *Counter) Inc() { atomic.AddUint64(&c.v, 1) }
func (c *Counter) Add(n uint64) { atomic.AddUint64(&c.v, n) }
func (c *Counter) Value() uint64 { return atomic.LoadUint64(&c.v) }

// Registry holds named counters.
type Registry struct {
	mu       sync.RWMutex
	counters map[string]*Counter
	created  time.Time
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{
		counters: make(map[string]*Counter),
		created:  time.Now(),
	}
}

// Counter returns the named counter, creating it if necessary.
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{}
	r.counters[name] = c
	return c
}

// Snapshot returns a sorted copy of all counter values.
func (r *Registry) Snapshot() map[string]uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]uint64, len(r.counters))
	for k, c := range r.counters {
		out[k] = c.Value()
	}
	return out
}

// Write prints all counters in a stable order to w.
func (r *Registry) Write(w io.Writer) {
	snap := r.Snapshot()
	keys := make([]string, 0, len(snap))
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	uptime := time.Since(r.created).Truncate(time.Second)
	fmt.Fprintf(w, "# cronwatch metrics  uptime=%s\n", uptime)
	for _, k := range keys {
		fmt.Fprintf(w, "%s %d\n", k, snap[k])
	}
}
