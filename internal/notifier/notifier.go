// Package notifier provides pluggable notification channels (e.g. email, webhook)
// that can be wired into the alerter pipeline.
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Notifier is the interface every notification backend must satisfy.
type Notifier interface {
	Send(ctx context.Context, subject, body string) error
}

// WebhookNotifier posts a JSON payload to a remote URL.
type WebhookNotifier struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewWebhook constructs a WebhookNotifier. If timeout is zero, 10 s is used.
func NewWebhook(url string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		URL:     url,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

type webhookPayload struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
	SentAt  string `json:"sent_at"`
}

// Send marshals the alert as JSON and POSTs it to the configured URL.
func (w *WebhookNotifier) Send(ctx context.Context, subject, body string) error {
	payload := webhookPayload{
		Subject: subject,
		Body:    body,
		SentAt:  time.Now().UTC().Format(time.RFC3339),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notifier: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("notifier: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("notifier: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notifier: unexpected status %d from webhook", resp.StatusCode)
	}
	return nil
}
