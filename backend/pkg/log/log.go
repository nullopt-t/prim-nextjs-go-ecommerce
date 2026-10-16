// Package log provides structured, leveled logging abstractions built on Go's standard log/slog.
package log

import (
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/Marlliton/slogpretty"
)

// Meta is an alias for key-value pairs representing contextual structured log attributes.
type Meta map[string]any

// Logger defines the standard interface for leveled structured application logging.
type Logger interface {
	// Info logs informational messages describing normal service events.
	Info(msg string, meta ...Meta)

	// Error logs error events that require investigation or attention.
	Error(msg string, meta ...Meta)

	// Warn logs warning conditions that do not fail requests but indicate potential issues.
	Warn(msg string, meta ...Meta)

	// Debug logs verbose diagnostic details useful for local development and troubleshooting.
	Debug(msg string, meta ...Meta)
}

// LogContext holds source code location information (file path and line number).
// It implements slog.LogValuer to cleanly serialize as a structured slog attribute group.
type LogContext struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// LogValue implements slog.LogValuer so LogContext can be integrated directly into structured logging.
func (lc LogContext) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("file", lc.File),
		slog.Int("line", lc.Line),
	)
}

// ConsoleLogger implements the Logger interface by wrapping an underlying *slog.Logger.
type ConsoleLogger struct {
	mu     sync.RWMutex
	logger *slog.Logger
}

// NewConsoleLogger creates a new ConsoleLogger instance. In production environments
// (determined by APP_ENV or ENV environment variable), it outputs structured JSON via slog.NewJSONHandler
// for log aggregators (e.g. Datadog, Elasticsearch, CloudWatch). In local development, it uses pretty-printed output.
func NewConsoleLogger() *ConsoleLogger {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	return NewEnvironmentLogger(env)
}

// NewEnvironmentLogger creates a ConsoleLogger tailored to the given environment ("production" uses JSONHandler).
func NewEnvironmentLogger(env string) *ConsoleLogger {
	if strings.EqualFold(env, "production") {
		return &ConsoleLogger{
			logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			})),
		}
	}

	return &ConsoleLogger{
		logger: slog.New(
			slogpretty.New(
				os.Stdout,
				&slogpretty.Options{
					Level:      slog.LevelDebug,
					AddSource:  false,
					Colorful:   true,
					Multiline:  true,
					TimeFormat: slogpretty.DefaultTimeFormat,
				},
			),
		),
	}
}

// NewJSONLogger creates a logger with structured JSON output regardless of environment.
func NewJSONLogger(level slog.Level) *ConsoleLogger {
	return &ConsoleLogger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})),
	}
}

// buildAttrs converts metadata maps into a slice of key-value attributes for slog with pre-allocated capacity.
func (l *ConsoleLogger) buildAttrs(meta []Meta) []any {
	if len(meta) == 0 {
		return nil
	}

	totalKeys := 0
	for _, m := range meta {
		totalKeys += len(m)
	}

	logsData := make([]any, 0, totalKeys*2)
	for _, m := range meta {
		for k, v := range m {
			logsData = append(logsData, k, v)
		}
	}

	return logsData
}

// Info logs a message with INFO severity and structured metadata.
func (l *ConsoleLogger) Info(msg string, meta ...Meta) {
	l.logger.Info(msg, l.buildAttrs(meta)...)
}

// Error logs a message with ERROR severity and structured metadata.
func (l *ConsoleLogger) Error(msg string, meta ...Meta) {
	l.logger.Error(msg, l.buildAttrs(meta)...)
}

// Warn logs a message with WARN severity and structured metadata.
func (l *ConsoleLogger) Warn(msg string, meta ...Meta) {
	l.logger.Warn(msg, l.buildAttrs(meta)...)
}

// Debug logs a message with DEBUG severity and structured metadata.
func (l *ConsoleLogger) Debug(msg string, meta ...Meta) {
	l.logger.Debug(msg, l.buildAttrs(meta)...)
}

