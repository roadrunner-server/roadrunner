package mocklogger

import (
	"log/slog"

	"github.com/roadrunner-server/endure/v2/dep"
)

// Logger is the interface that the mock logger provides via Endure DI.
type Logger interface {
	NamedLogger(string) *slog.Logger
}

// SlogLoggerMock is a mock logger plugin for integration tests.
// It captures all log entries for later assertion via ObservedLogs.
type SlogLoggerMock struct {
	l *slog.Logger
}

// SlogTestLogger creates a new mock logger plugin and returns the plugin
// instance along with an ObservedLogs for asserting on log messages.
func SlogTestLogger(level slog.Level) (*SlogLoggerMock, *ObservedLogs) {
	handler, logs := NewObserverHandler(level)
	obsLog := slog.New(handler)

	return &SlogLoggerMock{
		l: obsLog,
	}, logs
}

func (z *SlogLoggerMock) Init() error {
	return nil
}

func (z *SlogLoggerMock) Serve() chan error {
	return make(chan error, 1)
}

func (z *SlogLoggerMock) Stop() error {
	return nil
}

func (z *SlogLoggerMock) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Logger)(nil), z.ProvideLogger),
	}
}

func (z *SlogLoggerMock) Weight() uint {
	return 100
}

// ProvideLogger returns the Log instance for Endure dependency injection.
func (z *SlogLoggerMock) ProvideLogger() *Log {
	return NewLog(z.l)
}

// Log wraps an slog.Logger to satisfy the Logger interface.
type Log struct {
	base *slog.Logger
}

// NewLog creates a new Log from an slog.Logger.
func NewLog(log *slog.Logger) *Log {
	return &Log{
		base: log,
	}
}

// NamedLogger returns the underlying slog.Logger scoped with the given name.
func (l *Log) NamedLogger(name string) *slog.Logger {
	return l.base.With(slog.String("plugin", name))
}
