package mocklogger

import (
	"github.com/roadrunner-server/endure/v2/dep"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the interface that the mock logger provides via Endure DI.
type Logger interface {
	NamedLogger(string) *zap.Logger
}

// ZapLoggerMock is a mock logger plugin for integration tests.
// It captures all log entries for later assertion via ObservedLogs.
type ZapLoggerMock struct {
	l *zap.Logger
}

// ZapTestLogger creates a new mock logger plugin and returns the plugin
// instance along with an ObservedLogs for asserting on log messages.
func ZapTestLogger(enab zapcore.LevelEnabler) (*ZapLoggerMock, *ObservedLogs) {
	core, logs := New(enab)
	obsLog := zap.New(core, zap.Development())

	return &ZapLoggerMock{
		l: obsLog,
	}, logs
}

func (z *ZapLoggerMock) Init() error {
	return nil
}

func (z *ZapLoggerMock) Serve() chan error {
	return make(chan error, 1)
}

func (z *ZapLoggerMock) Stop() error {
	return z.l.Sync()
}

func (z *ZapLoggerMock) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Logger)(nil), z.ProvideLogger),
	}
}

func (z *ZapLoggerMock) Weight() uint {
	return 100
}

// ProvideLogger returns the Log instance for Endure dependency injection.
func (z *ZapLoggerMock) ProvideLogger() *Log {
	return NewLog(z.l)
}

// Log wraps a zap.Logger to satisfy the Logger interface.
type Log struct {
	base *zap.Logger
}

// NewLog creates a new Log from a zap.Logger.
func NewLog(log *zap.Logger) *Log {
	return &Log{
		base: log,
	}
}

// NamedLogger returns the underlying zap.Logger regardless of name.
func (l *Log) NamedLogger(string) *zap.Logger {
	return l.base
}
