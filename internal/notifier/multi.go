package notifier

import (
	"context"
	"errors"
	"fmt"
)

// Multi fans a single Send call out to several Notifiers.
// All backends are attempted; errors are joined and returned together.
type Multi struct {
	backends []Notifier
}

// NewMulti wraps one or more Notifiers into a fan-out notifier.
func NewMulti(backends ...Notifier) *Multi {
	return &Multi{backends: backends}
}

// Send calls every backend and collects errors.
func (m *Multi) Send(ctx context.Context, subject, body string) error {
	var errs []error
	for i, n := range m.backends {
		if err := n.Send(ctx, subject, body); err != nil {
			errs = append(errs, fmt.Errorf("backend[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}
