package alerter

import "log"

// LogSink writes alerts to the standard logger.
type LogSink struct {
	Logger *log.Logger
}

// NewLogSink creates a LogSink using the provided logger.
// If logger is nil, the default log package logger is used.
func NewLogSink(logger *log.Logger) *LogSink {
	if logger == nil {
		logger = log.Default()
	}
	return &LogSink{Logger: logger}
}

// Send formats the alert and writes it via the logger.
func (l *LogSink) Send(a Alert) error {
	l.Logger.Println(FormatMessage(a))
	return nil
}
