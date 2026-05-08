package notifier_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notifier"
)

func TestWebhook_Send_Success(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected Content-Type: %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	wh := notifier.NewWebhook(srv.URL, 5*time.Second)
	err := wh.Send(context.Background(), "job failed", "backup-db missed its window")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["subject"] != "job failed" {
		t.Errorf("subject mismatch: %q", received["subject"])
	}
	if received["body"] != "backup-db missed its window" {
		t.Errorf("body mismatch: %q", received["body"])
	}
	if received["sent_at"] == "" {
		t.Error("sent_at should not be empty")
	}
}

func TestWebhook_Send_Non2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	wh := notifier.NewWebhook(srv.URL, 5*time.Second)
	err := wh.Send(context.Background(), "s", "b")
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestWebhook_Send_BadURL(t *testing.T) {
	wh := notifier.NewWebhook("http://127.0.0.1:0/nope", 500*time.Millisecond)
	err := wh.Send(context.Background(), "s", "b")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestNewWebhook_DefaultTimeout(t *testing.T) {
	wh := notifier.NewWebhook("http://example.com", 0)
	if wh.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", wh.Timeout)
	}
}
