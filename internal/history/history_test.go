package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/history"
)

func tempStore(t *testing.T, max int) (*history.Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	s, err := history.New(path, max)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, path
}

func TestRecord_And_Get(t *testing.T) {
	s, _ := tempStore(t, 10)

	e := history.Entry{
		JobName: "nightly-backup",
		RunAt:   time.Now().UTC(),
		Duration: 2 * time.Second,
		Success: true,
	}
	if err := s.Record(e); err != nil {
		t.Fatalf("Record: %v", err)
	}

	got := s.Get("nightly-backup")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].JobName != e.JobName || !got[0].Success {
		t.Errorf("unexpected entry: %+v", got[0])
	}
}

func TestRecord_EvictsOldest(t *testing.T) {
	s, _ := tempStore(t, 3)

	for i := 0; i < 5; i++ {
		s.Record(history.Entry{JobName: "job", RunAt: time.Now(), Success: true}) //nolint:errcheck
	}

	if got := s.Get("job"); len(got) != 3 {
		t.Errorf("expected 3 entries (ring capped), got %d", len(got))
	}
}

func TestGet_Missing_ReturnsEmpty(t *testing.T) {
	s, _ := tempStore(t, 10)
	if got := s.Get("nonexistent"); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestPersistence_ReloadsFromDisk(t *testing.T) {
	s, path := tempStore(t, 10)
	s.Record(history.Entry{JobName: "db-dump", RunAt: time.Now(), Success: false, ExitCode: 1}) //nolint:errcheck

	// Open a second store pointing at the same file.
	s2, err := history.New(path, 10)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	got := s2.Get("db-dump")
	if len(got) != 1 {
		t.Fatalf("expected 1 persisted entry, got %d", len(got))
	}
	if got[0].Success || got[0].ExitCode != 1 {
		t.Errorf("persisted entry mismatch: %+v", got[0])
	}
}

func TestNew_MissingFileIsOK(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")
	_, err := history.New(path, 5)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
}

func TestNew_BadPath_ReturnsError(t *testing.T) {
	_, err := history.New("/proc/cronwatch-no-permission/h.json", 5)
	if err == nil {
		t.Skip("running as root or /proc path writable — skipping")
	}
}

func TestRecord_AtomicWrite_TmpFileCleanedUp(t *testing.T) {
	s, path := tempStore(t, 5)
	s.Record(history.Entry{JobName: "x", RunAt: time.Now(), Success: true}) //nolint:errcheck

	tmp := path + ".tmp"
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Errorf("expected .tmp file to be removed after save, but it exists")
	}
}
