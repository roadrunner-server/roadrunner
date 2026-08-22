package mocklogger

import (
	"log/slog"

	"github.com/roadrunner-server/endure/v2/dep"
)

// Logger is the interface that the mock logger provides via Endure DI.
type Logger interface {
	NamedLogger(string) *slog.Logger
}

// LoggerMock is a mock logger plugin for integration tests.
// It captures all log records for later assertion via ObservedLogs.
type LoggerMock struct {
	l *slog.Logger
}

// TestLogger creates a new mock logger plugin and returns the plugin
// instance along with an ObservedLogs for asserting on log messages.
func TestLogger(level slog.Leveler) (*LoggerMock, *ObservedLogs) {
	handler, logs := NewObserver(level)

	return &LoggerMock{
		l: slog.New(handler),
	}, logs
}

func (z *LoggerMock) Init() error {
	return nil
}

func (z *LoggerMock) Serve() chan error {
	return make(chan error, 1)
}

func (z *LoggerMock) Stop() error {
	return nil
}

func (z *LoggerMock) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Logger)(nil), z.ProvideLogger),
	}
}

func (z *LoggerMock) Weight() uint {
	return 100
}

// ProvideLogger returns the Log instance for Endure dependency injection.
func (z *LoggerMock) ProvideLogger() *Log {
	return NewLog(z.l)
}

// Log wraps a slog.Logger to satisfy the Logger interface.
type Log struct {
	base *slog.Logger
}

// NewLog creates a new Log from a slog.Logger.
func NewLog(log *slog.Logger) *Log {
	return &Log{
		base: log,
	}
}

// NamedLogger returns the underlying slog.Logger scoped with the given name.
func (l *Log) NamedLogger(name string) *slog.Logger {
	return l.base.With(slog.String("logger", name))
}
