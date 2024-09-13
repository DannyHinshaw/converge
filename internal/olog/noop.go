package olog

// NoopLogger implements the LevelLogger interface but discards all log messages.
type NoopLogger struct{}

// NewNoopLogger returns a new NoopLogger.
func NewNoopLogger() NoopLogger {
	return NoopLogger{}
}

// Debugf does nothing.
func (NoopLogger) Debugf(string, ...any) {}

// Debug does nothing.
func (NoopLogger) Debug(...any) {}

// Infof does nothing.
func (NoopLogger) Infof(string, ...any) {}

// Info does nothing.
func (NoopLogger) Info(...any) {}

// Errorf does nothing.
func (NoopLogger) Errorf(string, ...any) {}

// Error does nothing.
func (NoopLogger) Error(...any) {}

// WithName returns the NoopLogger, unaltered.
func (l NoopLogger) WithName(string) LevelLogger {
	return l
}
