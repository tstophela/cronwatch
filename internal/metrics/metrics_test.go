package metrics

import (
	"bytes"
	"strings"
	"sync"
	"testing"
)

func TestCounter_IncAndValue(t *testing.T) {
	r := New()
	c := r.Counter("pings")
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %d", c.Value())
	}
	c.Inc()
	c.Inc()
	if c.Value() != 2 {
		t.Fatalf("expected 2, got %d", c.Value())
	}
}

func TestCounter_Add(t *testing.T) {
	r := New()
	c := r.Counter("alerts")
	c.Add(10)
	if c.Value() != 10 {
		t.Fatalf("expected 10, got %d", c.Value())
	}
}

func TestRegistry_SameNameReturnsSameCounter(t *testing.T) {
	r := New()
	a := r.Counter("x")
	b := r.Counter("x")
	if a != b {
		t.Fatal("expected same pointer for same name")
	}
}

func TestRegistry_Snapshot_Isolated(t *testing.T) {
	r := New()
	r.Counter("a").Inc()
	r.Counter("b").Add(5)
	snap := r.Snapshot()
	if snap["a"] != 1 {
		t.Fatalf("a: expected 1, got %d", snap["a"])
	}
	if snap["b"] != 5 {
		t.Fatalf("b: expected 5, got %d", snap["b"])
	}
	// Mutating after snapshot should not affect snap.
	r.Counter("a").Inc()
	if snap["a"] != 1 {
		t.Fatal("snapshot was not isolated")
	}
}

func TestRegistry_Write_ContainsCounters(t *testing.T) {
	r := New()
	r.Counter("late_jobs").Add(3)
	r.Counter("pings").Add(7)
	var buf bytes.Buffer
	r.Write(&buf)
	out := buf.String()
	if !strings.Contains(out, "late_jobs 3") {
		t.Errorf("missing late_jobs line in output:\n%s", out)
	}
	if !strings.Contains(out, "pings 7") {
		t.Errorf("missing pings line in output:\n%s", out)
	}
	if !strings.Contains(out, "# cronwatch metrics") {
		t.Errorf("missing header in output:\n%s", out)
	}
}

func TestCounter_ConcurrentInc(t *testing.T) {
	r := New()
	c := r.Counter("concurrent")
	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}
	wg.Wait()
	if c.Value() != goroutines {
		t.Fatalf("expected %d, got %d", goroutines, c.Value())
	}
}
