# cronwatch

Lightweight daemon that monitors cron job execution and alerts on failures or drift.

## Installation

```bash
go install github.com/cronwatch/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/cronwatch/cronwatch.git && cd cronwatch && make build
```

## Usage

Define your monitored jobs in `cronwatch.yaml`:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert_on:
      - failure
      - drift
    notify: slack

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 5m
    alert_on:
      - failure
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

Wrap an existing cron job to report its status:

```bash
# In your crontab
0 2 * * * cronwatch exec --job daily-backup -- /usr/local/bin/backup.sh
```

cronwatch will track execution times, detect missed runs, and send alerts when jobs fail or deviate from their expected schedule.

## Configuration

| Field | Description |
|-------|-------------|
| `schedule` | Standard cron expression |
| `timeout` | Maximum allowed run duration |
| `alert_on` | Trigger conditions: `failure`, `drift`, `missed` |
| `notify` | Notification backend (`slack`, `email`, `pagerduty`) |

## License

MIT