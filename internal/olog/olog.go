// Package olog provides a simple leveled logger designed
// for command-line tools and minimalistic applications.
// It avoids JSON or structured logging to maintain a lightweight
// and human-readable format, suitable for CLI UX.
package olog

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Level represents the logging level, which determines
// the severity and importance of log messages.
type Level int

const (
	// LevelDebug is for debug messages.
	LevelDebug Level = iota

	// LevelInfo is for info messages.
	LevelInfo

	// LevelError is for error messages.
	LevelError
)

// levelName is the string representation of a logging level.
//
// Each log level string is padded to be the same length to ensure
// log messages are aligned correctly when printed.
// The longest level name is "error", so shorter names like "info"
// are padded with spaces to match the length of "error".
// This ensures consistent alignment across all logs.
type levelName = string

const (
	// debugLevel is the string representation of the debug level.
	// It is not padded because "debug" is already as long as "error".
	debugLevel levelName = "debug"

	// infoLevel is the string representation of the info level.
	// It is padded with a space to match the length of "error".
	infoLevel levelName = "info "

	// errorLevel is the string representation of the error level.
	errorLevel levelName = "error"
)

// String returns the padded string representation of the logging level.
// This ensures that logs align consistently, even when using different log levels.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return debugLevel
	case LevelInfo:
		return infoLevel
	case LevelError:
		return errorLevel
	default:
		return errorLevel
	}
}

// LevelLogger is an interface for a leveled logger implementation.
type LevelLogger interface {
	// Debugf logs a formatted debug message.
	Debugf(format string, v ...any)

	// Debug logs a debug message.
	Debug(v ...any)

	// Infof logs a formatted info message.
	Infof(format string, v ...any)

	// Info logs an info message.
	Info(v ...any)

	// Errorf logs a formatted error message.
	Errorf(format string, v ...any)

	// Error logs an error message.
	Error(v ...any)

	// WithName returns a new logger instance with a specific
	// name prefix applied to all log messages.
	WithName(name string) LevelLogger
}

// Logger is a simple logger that can be used to log messages at different levels.
type Logger struct {
	// logger is the underlying Go standard library logger.
	logger *log.Logger

	// name is an optional identifier included in all log messages.
	name string

	// level defines the log level threshold.
	level Level

	// callDepth specifies the stack depth for file/line reporting.
	callDepth int
}

// NewLogger creates a new Logger.
func NewLogger(lvl Level, opts ...Option) Logger {
	var flags int
	if lvl == LevelDebug {
		flags = log.Lshortfile
	}

	lg := Logger{
		logger:    log.New(os.Stderr, "", flags),
		level:     lvl,
		callDepth: 3, //nolint:mnd // 3 is the call depth to log from.
	}

	for _, opt := range opts {
		opt(&lg)
	}

	return lg
}

// Option defines a function type that configures the Logger.
type Option func(*Logger)

// WithWriter returns an Option that sets the writer for the logger.
func WithWriter(w io.Writer) Option {
	return func(l *Logger) {
		l.logger.SetOutput(w)
	}
}

// Debugf logs a formatted debug message if the logger is set to LevelDebug.
// It will not output anything if the logger level is higher than LevelDebug.
func (l Logger) Debugf(format string, v ...any) {
	if l.level == LevelDebug {
		l.logf(LevelDebug, format, v...)
	}
}

// Debug logs a debug message if the logger is set to LevelDebug.
// It will not output anything if the logger level is higher than LevelDebug.
func (l Logger) Debug(v ...any) {
	if l.level == LevelDebug {
		l.log(LevelDebug, v...)
	}
}

// Infof logs a formatted info message if the logger is set to LevelInfo or lower.
// It will not output anything if the logger level is higher than LevelInfo.
func (l Logger) Infof(format string, v ...any) {
	if l.level <= LevelInfo {
		l.logf(LevelInfo, format, v...)
	}
}

// Info logs an info message if the logger is set to LevelInfo or lower.
// It will not output anything if the logger level is higher than LevelInfo.
func (l Logger) Info(v ...any) {
	if l.level <= LevelInfo {
		l.log(LevelInfo, v...)
	}
}

// Errorf logs a formatted error message.
func (l Logger) Errorf(format string, v ...any) {
	l.logf(LevelError, format, v...)
}

// Error logs an error message.
func (l Logger) Error(v ...any) {
	l.log(LevelError, v...)
}

// WithName returns a new logger with the given name.
func (l Logger) WithName(name string) LevelLogger {
	c := l.clone()
	c.name = name
	return c
}

// log logs a message at the given level.
func (l Logger) log(lvl Level, v ...any) {
	msg := fmt.Sprintln(v...)

	if l.name != "" {
		msg = "[" + lvl.String() + "] [" + l.name + "]: " + msg
	} else {
		msg = "[" + lvl.String() + "]: " + msg
	}

	if l.level == LevelDebug {
		// Include call depth to show code
		// line reference in verbose mode.
		_ = l.logger.Output(l.callDepth, msg)
	} else {
		// Directly log without call depth,
		// omitting code line reference.
		l.logger.Print(msg)
	}
}

// logf logs a formatted message at the given level.
func (l Logger) logf(lvl Level, format string, v ...any) {
	l.log(lvl, fmt.Sprintf(format, v...))
}

// clone returns a copy of the logger with the same settings.
func (l Logger) clone() Logger {
	return Logger{
		logger:    l.logger,
		level:     l.level,
		callDepth: l.callDepth,
	}
}
