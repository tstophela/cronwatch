package alerter

import "sync"

// MemorySink stores alerts in memory. Useful for testing.
type MemorySink struct {
	mu     sync.Mutex
	alerts []Alert
}

// Send appends the alert to the in-memory store.
func (m *MemorySink) Send(a Alert) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = append(m.alerts, a)
	return nil
}

// All returns a copy of all received alerts.
func (m *MemorySink) All() []Alert {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := make([]Alert, len(m.alerts))
	for i, a := range m.alerts {
		copy[i] = a
	}
	return copy
}

// Len returns the number of stored alerts.
func (m *MemorySink) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.alerts)
}

// Clear removes all stored alerts.
func (m *MemorySink) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alerts = nil
}
