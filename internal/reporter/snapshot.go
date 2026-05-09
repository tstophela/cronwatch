package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/cronwatch/internal/tracker"
)

// Snapshot represents a point-in-time capture of all tracked job states.
type Snapshot struct {
	GeneratedAt time.Time            `json:"generated_at"`
	Jobs        []tracker.JobStatus  `json:"jobs"`
}

// GenerateSnapshot collects the current status for all known jobs and
// returns a Snapshot sorted alphabetically by job name.
func GenerateSnapshot(tr *tracker.Tracker, jobNames []string) Snapshot {
	snap := Snapshot{
		GeneratedAt: time.Now().UTC(),
	}

	for _, name := range jobNames {
		if st, ok := tr.Get(name); ok {
			snap.Jobs = append(snap.Jobs, st)
		}
	}

	sort.Slice(snap.Jobs, func(i, j int) bool {
		return snap.Jobs[i].Name < snap.Jobs[j].Name
	})

	return snap
}

// WriteSnapshot serialises the Snapshot as indented JSON to the given file path.
// The file is created or truncated on each call.
func WriteSnapshot(path string, snap Snapshot) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("reporter: create snapshot file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		return fmt.Errorf("reporter: encode snapshot: %w", err)
	}
	return nil
}
