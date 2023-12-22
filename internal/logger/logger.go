package logger

import (
	"log/slog"
	"os"
)

// LevelLogger is a logger that supports different levels of logging.
type LevelLogger interface {
	// Debug handles logging the given message at the debug level.
	Debug(msg string, args ...any)

	// Info handles logging the given message at the info level.
	Info(msg string, args ...any)

	// Warn handles logging the given message at the warn level.
	Warn(msg string, args ...any)

	// Error handles logging the given message at the error level.
	Error(msg string, args ...any)

	// WithGroup creates a new child/group Logger from the current logger.
	WithGroup(group string) LevelLogger

	// Verbose returns true if the logger is in verbose mode.
	Verbose() bool
}

// Logger is a wrapper around slog.Logger.
type Logger struct {
	// Level is the log Level to use for
	// filtering log verbosity.
	Level slog.Level

	// Slogger is the underlying slog.Logger
	// that is used for logging.
	Slogger *slog.Logger

	// handler is the handler to use for
	// the loggers structured output.
	handler slog.Handler

	// parent is the parent logger to use
	// for creating the Slogger as a child
	// in the New function.
	parent *slog.Logger
}

// New returns a new Logger with the given name and level.
func New(name string, opts ...Option) *Logger {
	// Base logger defaults to error level.
	logger := Logger{
		Level: slog.LevelError,
	}

	// Apply optional configurations.
	for _, opt := range opts {
		opt(&logger)
	}

	// If no handler was provided, default to
	// a text handler that writes to stderr.
	if logger.handler == nil {
		logger.handler = slog.NewTextHandler(
			os.Stderr, &slog.HandlerOptions{
				Level: logger.Level,
			})
	}

	// If a parent logger was provided, use it
	// to create the Slogger as a child.
	var root *slog.Logger
	if logger.parent != nil {
		root = logger.parent
	} else {
		root = slog.New(logger.handler)
	}

	// Create a new logger with the given name.
	logger.Slogger = root.WithGroup(name)

	return &logger
}

// Debug handles proxying the given message to the underlying slog.Logger at the debug level.
func (l *Logger) Debug(msg string, args ...any) {
	l.Slogger.Debug(msg, args...)
}

// Info handles proxying the given message to the underlying slog.Logger at the info level.
func (l *Logger) Info(msg string, args ...any) {
	l.Slogger.Info(msg, args...)
}

// Warn handles proxying the given message to the underlying slog.Logger at the warn level.
func (l *Logger) Warn(msg string, args ...any) {
	l.Slogger.Warn(msg, args...)
}

// Error handles proxying the given message to the underlying slog.Logger at the error level.
func (l *Logger) Error(msg string, args ...any) {
	l.Slogger.Error(msg, args...)
}

// WithGroup creates a new child/group Logger from the current logger.
func (l *Logger) WithGroup(group string) LevelLogger {
	return &Logger{
		Slogger: l.Slogger.WithGroup(group),
	}
}

// Verbose returns true if the logger is in verbose mode.
func (l *Logger) Verbose() bool {
	return l.Level == slog.LevelDebug
}
