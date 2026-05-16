package ratelimit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempStorePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "ratelimit.json")
}

func TestStore_SetAndGet(t *testing.T) {
	s, err := NewStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	now := time.Now().Truncate(time.Second)
	if err := s.Set("job-a", now); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, ok := s.Get("job-a")
	if !ok {
		t.Fatal("expected key to be present")
	}
	if !got.Equal(now) {
		t.Errorf("got %v, want %v", got, now)
	}
}

func TestStore_Get_Missing(t *testing.T) {
	s, err := NewStore(tempStorePath(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected missing key to return ok=false")
	}
}

func TestStore_PersistsAcrossReload(t *testing.T) {
	path := tempStorePath(t)

	s1, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	now := time.Now().Truncate(time.Second)
	if err := s1.Set("job-b", now); err != nil {
		t.Fatalf("Set: %v", err)
	}

	s2, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore reload: %v", err)
	}
	got, ok := s2.Get("job-b")
	if !ok {
		t.Fatal("expected persisted key after reload")
	}
	if !got.Equal(now) {
		t.Errorf("got %v, want %v", got, now)
	}
}

func TestStore_MissingFile_NoError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does_not_exist.json")
	_, err := NewStore(path)
	if err != nil {
		t.Errorf("expected no error for missing file, got: %v", err)
	}
}

func TestStore_BadPath_ReturnsError(t *testing.T) {
	_, err := NewStore("/nonexistent-dir/sub/ratelimit.json")
	if err == nil {
		t.Error("expected error for unwritable path")
	}
}

func TestStore_MultipleKeys_Independent(t *testing.T) {
	s, _ := NewStore(tempStorePath(t))
	t1 := time.Now().Add(-10 * time.Minute).Truncate(time.Second)
	t2 := time.Now().Truncate(time.Second)

	_ = s.Set("alpha", t1)
	_ = s.Set("beta", t2)

	gotA, _ := s.Get("alpha")
	gotB, _ := s.Get("beta")

	if !gotA.Equal(t1) {
		t.Errorf("alpha: got %v, want %v", gotA, t1)
	}
	if !gotB.Equal(t2) {
		t.Errorf("beta: got %v, want %v", gotB, t2)
	}
}

func init() {
	// ensure os is available for os.Remove in flush
	_ = os.Stderr
}
