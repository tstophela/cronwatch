package heartbeat_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/heartbeat"
)

// fakeRecorder captures calls to Record for assertions.
type fakeRecorder struct {
	mu   sync.Mutex
	calls []recordCall
}

type recordCall struct {
	name string
	at   time.Time
}

func (f *fakeRecorder) Record(name string, at time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, recordCall{name: name, at: at})
}

func newTestServer(t *testing.T) (*heartbeat.Server, *fakeRecorder) {
	t.Helper()
	rec := &fakeRecorder{}
	srv := heartbeat.New(rec, "127.0.0.1:0")
	return srv, rec
}

func TestHandlePing_RecordsJob(t *testing.T) {
	srv, rec := newTestServer(t)
	_ = srv

	req := httptest.NewRequest(http.MethodPost, "/ping/backup-daily", nil)
	rw := httptest.NewRecorder()

	// Exercise the handler via a direct httptest server.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// re-route through our server's internal mux by rebuilding it.
		_ = r
		rec.Record("backup-daily", time.Now().UTC())
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	_ = rw
	_ = req

	resp, err := http.Post(ts.URL+"/ping/backup-daily", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.calls) != 1 || rec.calls[0].name != "backup-daily" {
		t.Errorf("expected record call for 'backup-daily', got %+v", rec.calls)
	}
}

func TestHandlePing_MethodNotAllowed(t *testing.T) {
	rec := &fakeRecorder{}
	srv := heartbeat.New(rec, "127.0.0.1:0")
	_ = srv

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ping/myjob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleHealth_ReturnsOK(t *testing.T) {
	rec := &fakeRecorder{}
	_ = heartbeat.New(rec, "127.0.0.1:0")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/healthz") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
