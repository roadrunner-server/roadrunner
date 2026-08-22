package mocklogger

import (
	"log/slog"

	"github.com/roadrunner-server/endure/v2/dep"
)

type SlogLoggerMock struct {
	l *slog.Logger
}

type Logger interface {
	NamedLogger(string) *slog.Logger
}

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

func (z *SlogLoggerMock) ProvideLogger() *Log {
	return NewLogger(z.l)
}

type Log struct {
	base *slog.Logger
}

func NewLogger(log *slog.Logger) *Log {
	return &Log{
		base: log,
	}
}

func (l *Log) NamedLogger(string) *slog.Logger {
	return l.base
}
