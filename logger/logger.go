package logger

import (
	"io"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

// Log is the global logger instance
var Log *Logger

// Logger wraps zerolog.Logger to provide a similar interface to the previous logrus implementation
type Logger struct {
	zl zerolog.Logger
}

// LoggerEntry represents a log entry with context fields
type LoggerEntry struct {
	zl zerolog.Logger
}

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		// Only log error if file exists but couldn't be loaded
		println("Error loading .env file:", err.Error())
	}

	// Configure the global logger
	Log = NewLogger(os.Stdout)
}

// NewLogger creates a new logger instance with the given output
func NewLogger(out io.Writer) *Logger {
	// Set up console writer with colors
	output := zerolog.ConsoleWriter{
		Out:        out,
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    os.Getenv("LOG_NO_COLOR") == "true",
	}

	// Set log level based on environment
	level := zerolog.InfoLevel
	switch os.Getenv("LOG_LEVEL") {
	case "trace":
		level = zerolog.TraceLevel
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	}

	// Backward compatibility with DEBUG env var
	if os.Getenv("DEBUG") == "1" {
		level = zerolog.DebugLevel
	}

	// Create the logger
	zl := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &Logger{zl: zl}
}

// WithField returns a LoggerEntry with the given field added to the context
func (l *Logger) WithField(key string, value interface{}) *LoggerEntry {
	return &LoggerEntry{
		zl: l.zl.With().Interface(key, value).Logger(),
	}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.zl.Info().Msg(msg)
}

// Infof logs an info message with formatting
func (l *Logger) Infof(format string, args ...interface{}) {
	l.zl.Info().Msgf(format, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.zl.Debug().Msg(msg)
}

// Debugf logs a debug message with formatting
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zl.Debug().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.zl.Error().Msg(msg)
}

// Errorf logs an error message with formatting
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zl.Error().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.zl.Warn().Msg(msg)
}

// Warnf logs a warning message with formatting
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zl.Warn().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.zl.Fatal().Msg(msg)
}

// Fatalf logs a fatal message with formatting and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zl.Fatal().Msgf(format, args...)
}

// Trace logs a trace message
func (l *Logger) Trace(msg string) {
	l.zl.Trace().Msg(msg)
}

// Tracef logs a trace message with formatting
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.zl.Trace().Msgf(format, args...)
}

// WithField returns a LoggerEntry with the given field added to the context
func (e *LoggerEntry) WithField(key string, value interface{}) *LoggerEntry {
	return &LoggerEntry{
		zl: e.zl.With().Interface(key, value).Logger(),
	}
}

// Info logs an info message
func (e *LoggerEntry) Info(msg string) {
	e.zl.Info().Msg(msg)
}

// Infof logs an info message with formatting
func (e *LoggerEntry) Infof(format string, args ...interface{}) {
	e.zl.Info().Msgf(format, args...)
}

// Debug logs a debug message
func (e *LoggerEntry) Debug(msg string) {
	e.zl.Debug().Msg(msg)
}

// Debugf logs a debug message with formatting
func (e *LoggerEntry) Debugf(format string, args ...interface{}) {
	e.zl.Debug().Msgf(format, args...)
}

// Error logs an error message
func (e *LoggerEntry) Error(msg string) {
	e.zl.Error().Msg(msg)
}

// Errorf logs an error message with formatting
func (e *LoggerEntry) Errorf(format string, args ...interface{}) {
	e.zl.Error().Msgf(format, args...)
}

// Warn logs a warning message
func (e *LoggerEntry) Warn(msg string) {
	e.zl.Warn().Msg(msg)
}

// Warnf logs a warning message with formatting
func (e *LoggerEntry) Warnf(format string, args ...interface{}) {
	e.zl.Warn().Msgf(format, args...)
}

// Fatal logs a fatal message and exits
func (e *LoggerEntry) Fatal(msg string) {
	e.zl.Fatal().Msg(msg)
}

// Fatalf logs a fatal message with formatting and exits
func (e *LoggerEntry) Fatalf(format string, args ...interface{}) {
	e.zl.Fatal().Msgf(format, args...)
}

// Trace logs a trace message
func (e *LoggerEntry) Trace(msg string) {
	e.zl.Trace().Msg(msg)
}

// Tracef logs a trace message with formatting
func (e *LoggerEntry) Tracef(format string, args ...interface{}) {
	e.zl.Trace().Msgf(format, args...)
}
