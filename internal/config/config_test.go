package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Valid(t *testing.T) {
	content := `
check_interval: 2m
jobs:
  - name: backup
    schedule: "0 2 * * *"
    timeout: 10m
    command: /usr/local/bin/backup.sh
alerts:
  email: ops@example.com
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 2*time.Minute {
		t.Errorf("expected check_interval 2m, got %v", cfg.CheckInterval)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("expected job name 'backup', got %q", cfg.Jobs[0].Name)
	}
	if cfg.Jobs[0].Timeout != 10*time.Minute {
		t.Errorf("expected timeout 10m, got %v", cfg.Jobs[0].Timeout)
	}
}

func TestLoad_DefaultsApplied(t *testing.T) {
	content := `
jobs:
  - name: cleanup
    schedule: "@daily"
`
	path := writeTempConfig(t, content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != time.Minute {
		t.Errorf("expected default check_interval 1m, got %v", cfg.CheckInterval)
	}
	if cfg.Jobs[0].Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Jobs[0].Timeout)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/cronwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_NoJobs(t *testing.T) {
	content := `check_interval: 1m\njobs: []\n`
	path := writeTempConfig(t, content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for empty jobs, got nil")
	}
}
