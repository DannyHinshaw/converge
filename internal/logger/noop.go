package logger

// NoopLogger is a logger that does nothing.
type NoopLogger struct{}

// Debug handles logging the given message at the debug level.
func (l *NoopLogger) Debug(msg string, args ...any) {} //nolint:revive //unused

// Info handles logging the given message at the info level.
func (l *NoopLogger) Info(msg string, args ...any) {} //nolint:revive //unused

// Warn handles logging the given message at the warn level.
func (l *NoopLogger) Warn(msg string, args ...any) {} //nolint:revive //unused

// Error handles logging the given message at the error level.
func (l *NoopLogger) Error(msg string, args ...any) {} //nolint:revive //unused

// WithGroup creates a new child/group Logger from the current logger.
func (l *NoopLogger) WithGroup(group string) LevelLogger { //nolint:revive //unused
	return l
}

// Verbose returns true if the logger is in verbose mode.
func (l *NoopLogger) Verbose() bool {
	return false
}
