package alerter_test

import (
	"strings"
	"testing"
	"time"

	"cronwatch/internal/alerter"
)

func newAlerterWithSink() (*alerter.Alerter, *alerter.MemorySink) {
	sink := &alerter.MemorySink{}
	a := alerter.New(sink)
	return a, sink
}

func TestFire_StoresAlert(t *testing.T) {
	a, sink := newAlerterWithSink()
	a.Fire(alerter.LevelError, "backup", "exit code 1")

	if sink.Len() != 1 {
		t.Fatalf("expected 1 alert, got %d", sink.Len())
	}
	alerts := sink.All()
	if alerts[0].JobName != "backup" {
		t.Errorf("expected job 'backup', got %q", alerts[0].JobName)
	}
	if alerts[0].Level != alerter.LevelError {
		t.Errorf("expected level ERROR, got %q", alerts[0].Level)
	}
}

func TestWarn_SetsLevelWarn(t *testing.T) {
	a, sink := newAlerterWithSink()
	a.Warn("cleanup", "drift detected")

	alerts := sink.All()
	if len(alerts) != 1 || alerts[0].Level != alerter.LevelWarn {
		t.Errorf("expected WARN alert, got %+v", alerts)
	}
}

func TestFire_MultipleSinks(t *testing.T) {
	sink1 := &alerter.MemorySink{}
	sink2 := &alerter.MemorySink{}
	a := alerter.New(sink1, sink2)
	a.Error("sync", "timed out")

	if sink1.Len() != 1 || sink2.Len() != 1 {
		t.Errorf("expected both sinks to receive alert")
	}
}

func TestFormatMessage_ContainsFields(t *testing.T) {
	al := alerter.Alert{
		JobName:    "myjob",
		Level:      alerter.LevelWarn,
		Message:    "late by 5m",
		OccurredAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}
	msg := alerter.FormatMessage(al)
	for _, want := range []string{"WARN", "myjob", "late by 5m", "2024-01-15"} {
		if !strings.Contains(msg, want) {
			t.Errorf("FormatMessage missing %q in: %s", want, msg)
		}
	}
}

func TestMemorySink_Clear(t *testing.T) {
	sink := &alerter.MemorySink{}
	sink.Send(alerter.Alert{JobName: "x", Level: alerter.LevelError})
	sink.Clear()
	if sink.Len() != 0 {
		t.Errorf("expected 0 alerts after Clear, got %d", sink.Len())
	}
}
