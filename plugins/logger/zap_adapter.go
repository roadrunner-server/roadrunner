package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type ZapAdapter struct {
	zl *zap.Logger
}

// Create NewZapAdapter which uses general log interface
func NewZapAdapter(zapLogger *zap.Logger) *ZapAdapter {
	return &ZapAdapter{
		zl: zapLogger.WithOptions(zap.AddCallerSkip(1)),
	}
}

func (log *ZapAdapter) fields(keyvals []interface{}) []zap.Field {
	// we should have even number of keys and values
	if len(keyvals)%2 != 0 {
		return []zap.Field{zap.Error(fmt.Errorf("odd number of keyvals pairs: %v", keyvals))}
	}

	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	return fields
}

func (log *ZapAdapter) Debug(msg string, keyvals ...interface{}) {
	log.zl.Debug(msg, log.fields(keyvals)...)
}

func (log *ZapAdapter) Info(msg string, keyvals ...interface{}) {
	log.zl.Info(msg, log.fields(keyvals)...)
}

func (log *ZapAdapter) Warn(msg string, keyvals ...interface{}) {
	log.zl.Warn(msg, log.fields(keyvals)...)
}

func (log *ZapAdapter) Error(msg string, keyvals ...interface{}) {
	log.zl.Error(msg, log.fields(keyvals)...)
}

func (log *ZapAdapter) With(keyvals ...interface{}) Logger {
	return NewZapAdapter(log.zl.With(log.fields(keyvals)...))
}
