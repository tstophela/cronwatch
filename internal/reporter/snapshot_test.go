package reporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cronwatch/internal/tracker"
)

func TestGenerateSnapshot_Empty(t *testing.T) {
	tr := newTracker()
	snap := GenerateSnapshot(tr, []string{})

	if len(snap.Jobs) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(snap.Jobs))
	}
	if snap.GeneratedAt.IsZero() {
		t.Fatal("expected GeneratedAt to be set")
	}
}

func TestGenerateSnapshot_KnownJobs_SortedByName(t *testing.T) {
	tr := newTracker()
	tr.Record("zebra", time.Now())
	tr.Record("alpha", time.Now())
	tr.Record("mango", time.Now())

	snap := GenerateSnapshot(tr, []string{"zebra", "alpha", "mango"})

	if len(snap.Jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(snap.Jobs))
	}
	names := []string{snap.Jobs[0].Name, snap.Jobs[1].Name, snap.Jobs[2].Name}
	expected := []string{"alpha", "mango", "zebra"}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("position %d: got %q, want %q", i, n, expected[i])
		}
	}
}

func TestGenerateSnapshot_UnknownJobSkipped(t *testing.T) {
	tr := newTracker()
	tr.Record("known", time.Now())

	snap := GenerateSnapshot(tr, []string{"known", "ghost"})

	if len(snap.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(snap.Jobs))
	}
	if snap.Jobs[0].Name != "known" {
		t.Errorf("unexpected job name %q", snap.Jobs[0].Name)
	}
}

func TestWriteSnapshot_RoundTrip(t *testing.T) {
	tr := newTracker()
	tr.Record("backup", time.Now())

	snap := GenerateSnapshot(tr, []string{"backup"})

	tmp := filepath.Join(t.TempDir(), "snapshot.json")
	if err := WriteSnapshot(tmp, snap); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got Snapshot
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(got.Jobs) != 1 {
		t.Fatalf("expected 1 job in snapshot, got %d", len(got.Jobs))
	}
	if got.Jobs[0].Name != "backup" {
		t.Errorf("unexpected job name %q", got.Jobs[0].Name)
	}
}

func TestWriteSnapshot_BadPath(t *testing.T) {
	snap := Snapshot{GeneratedAt: time.Now()}
	err := WriteSnapshot("/nonexistent/dir/snapshot.json", snap)
	if err == nil {
		t.Fatal("expected error for bad path, got nil")
	}
}
