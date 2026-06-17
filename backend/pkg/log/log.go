package log

import (
	"log/slog"
	"os"

	"github.com/Marlliton/slogpretty"
)

type Meta map[string]any

type Logger interface {
	Info(msg string, meta ...Meta)
	Error(msg string, meta ...Meta)
	Warn(msg string, meta ...Meta)
	Debug(msg string, meta ...Meta)
}

type LogContext struct {
	File string
	Line int
}

type ConsoleLogger struct {
	logger *slog.Logger
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		logger: slog.New(
			slogpretty.New(
				os.Stdout,
				&slogpretty.Options{
					Level:      slog.LevelDebug,
					AddSource:  false,                        // Show file location
					Colorful:   true,                         // Enable colors. Default is true
					Multiline:  true,                         // Pretty print for complex data
					TimeFormat: slogpretty.DefaultTimeFormat, // Custom format (e.g., time.Kitchen)
				},
			),
		),
	}
}

func (l *ConsoleLogger) contextFields() []any {
	return []any{}
}

func (l *ConsoleLogger) Info(msg string, meta ...Meta) {
	logsData := l.contextFields()
	for _, m := range meta {
		for k, v := range m {
			logsData = append(logsData, k, v)
		}
	}
	l.logger.Info(msg, logsData...)
}

func (l *ConsoleLogger) Error(msg string, meta ...Meta) {
	logsData := l.contextFields()
	for _, m := range meta {
		for k, v := range m {
			logsData = append(logsData, k, v)
		}
	}
	l.logger.Error(msg, logsData...)
}

func (l *ConsoleLogger) Warn(msg string, meta ...Meta) {
	logsData := l.contextFields()
	for _, m := range meta {
		for k, v := range m {
			logsData = append(logsData, k, v)
		}
	}
	l.logger.Warn(msg, logsData...)
}

func (l *ConsoleLogger) Debug(msg string, meta ...Meta) {
	logsData := l.contextFields()
	for _, m := range meta {
		for k, v := range m {
			logsData = append(logsData, k, v)
		}
	}
	l.logger.Debug(msg, logsData...)
}
