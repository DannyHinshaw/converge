package logger

import "log/slog"

type Option func(*Logger)

// WithHandler sets the handler for the logger.
func WithHandler(h slog.Handler) Option {
	return func(l *Logger) {
		l.handler = h
	}
}

// WithVerbose sets the logger to verbose mode by way of
// setting the log level to debug.
func WithVerbose(v bool) Option {
	return func(l *Logger) {
		if v {
			l.Level = slog.LevelDebug
		}
	}
}
