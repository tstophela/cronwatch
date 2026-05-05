package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Job defines a single cron job to monitor.
type Job struct {
	Name     string        `yaml:"name"`
	Schedule string        `yaml:"schedule"`
	Timeout  time.Duration `yaml:"timeout"`
	Command  string        `yaml:"command"`
}

// AlertConfig holds alerting destination settings.
type AlertConfig struct {
	Email   string `yaml:"email"`
	SlackURL string `yaml:"slack_url"`
}

// Config is the top-level cronwatch configuration.
type Config struct {
	CheckInterval time.Duration `yaml:"check_interval"`
	Jobs          []Job         `yaml:"jobs"`
	Alerts        AlertConfig   `yaml:"alerts"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate checks required fields and applies defaults.
func (c *Config) validate() error {
	if len(c.Jobs) == 0 {
		return fmt.Errorf("at least one job must be defined")
	}

	seen := make(map[string]bool)
	for i, job := range c.Jobs {
		if job.Name == "" {
			return fmt.Errorf("job[%d]: name is required", i)
		}
		if seen[job.Name] {
			return fmt.Errorf("job[%d]: duplicate job name %q", i, job.Name)
		}
		seen[job.Name] = true
		if job.Schedule == "" {
			return fmt.Errorf("job %q: schedule is required", job.Name)
		}
		if c.Jobs[i].Timeout == 0 {
			c.Jobs[i].Timeout = 30 * time.Second
		}
	}
	if c.CheckInterval == 0 {
		c.CheckInterval = time.Minute
	}
	return nil
}
