package schedule_test

import (
	"testing"
	"time"

	"github.com/user/cronwatch/internal/schedule"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestValidate_Valid(t *testing.T) {
	exprs := []string{
		"* * * * *",
		"0 * * * *",
		"0 9 * * 1-5",
		"@hourly",
		"@daily",
		"30 4 1 * *",
	}
	for _, expr := range exprs {
		if err := schedule.Validate(expr); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", expr, err)
		}
	}
}

func TestValidate_Invalid(t *testing.T) {
	exprs := []string{
		"not a cron",
		"60 * * * *",
		"",
	}
	for _, expr := range exprs {
		if err := schedule.Validate(expr); err == nil {
			t.Errorf("expected %q to be invalid, but got no error", expr)
		}
	}
}

func TestNextRun(t *testing.T) {
	from := mustParseTime("2024-01-15T09:00:00Z")
	next, err := schedule.NextRun("0 10 * * *", from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := mustParseTime("2024-01-15T10:00:00Z")
	if !next.Equal(expected) {
		t.Errorf("expected next run %s, got %s", expected, next)
	}
}

func TestNextRun_InvalidExpr(t *testing.T) {
	_, err := schedule.NextRun("bad expr", time.Now())
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}

func TestPreviousExpected(t *testing.T) {
	// Every hour at :00 — previous run before 09:45 should be 09:00
	at := mustParseTime("2024-01-15T09:45:00Z")
	prev, err := schedule.PreviousExpected("0 * * * *", at)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := mustParseTime("2024-01-15T09:00:00Z")
	if !prev.Equal(expected) {
		t.Errorf("expected previous run %s, got %s", expected, prev)
	}
}

func TestPreviousExpected_InvalidExpr(t *testing.T) {
	_, err := schedule.PreviousExpected("bad expr", time.Now())
	if err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}
