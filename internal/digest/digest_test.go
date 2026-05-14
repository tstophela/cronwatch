package digest_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/digest"
	"github.com/example/cronwatch/internal/tracker"
)

// captureSender records the most recent subject/body pair.
type captureSender struct {
	mu      sync.Mutex
	subject string
	body    string
	calls   int
}

func (c *captureSender) Send(_ context.Context, subject, body string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subject = subject
	c.body = body
	c.calls++
	return nil
}

func newTracker(t *testing.T) *tracker.Tracker {
	t.Helper()
	tr, err := tracker.New(3)
	if err != nil {
		t.Fatalf("tracker.New: %v", err)
	}
	return tr
}

func TestDigest_SendsOnTick(t *testing.T) {
	tr := newTracker(t)
	tr.Record("backup", true)

	sender := &captureSender{}
	d := digest.New(tr, sender, 50*time.Millisecond, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	d.Start(ctx)

	sender.mu.Lock()
	defer sender.mu.Unlock()
	if sender.calls == 0 {
		t.Fatal("expected at least one digest send, got 0")
	}
	if !strings.Contains(sender.subject, "cronwatch digest") {
		t.Errorf("subject missing prefix: %q", sender.subject)
	}
	if !strings.Contains(sender.body, "backup") {
		t.Errorf("body missing job name: %q", sender.body)
	}
}

func TestDigest_EmptyTracker(t *testing.T) {
	tr := newTracker(t)
	sender := &captureSender{}
	d := digest.New(tr, sender, 30*time.Millisecond, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	d.Start(ctx)

	sender.mu.Lock()
	defer sender.mu.Unlock()
	if !strings.Contains(sender.body, "No jobs") {
		t.Errorf("expected empty-tracker message, got: %q", sender.body)
	}
}

func TestDigest_StopsOnContextCancel(t *testing.T) {
	tr := newTracker(t)
	sender := &captureSender{}
	d := digest.New(tr, sender, 10*time.Second, nil) // long interval

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		d.Start(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("digest did not stop after context cancel")
	}
}
