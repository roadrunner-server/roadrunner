package logger

import (
	"fmt"

	"go.uber.org/zap"
	core "go.uber.org/zap/zapcore"
)

type ZapAdapter struct {
	zl *zap.Logger
}

// NewZapAdapter ... which uses general log interface
func NewZapAdapter(zapLogger *zap.Logger) *ZapAdapter {
	return &ZapAdapter{
		zl: zapLogger.WithOptions(zap.AddCallerSkip(1)),
	}
}

func separateFields(keyVals []interface{}) ([]zap.Field, []interface{}) {
	var fields []zap.Field
	var pairedKeyVals []interface{}

	for key := range keyVals {
		switch value := keyVals[key].(type) {
		case zap.Field:
			fields = append(fields, value)
		case core.ObjectMarshaler:
			fields = append(fields, zap.Inline(value))
		default:
			pairedKeyVals = append(pairedKeyVals, value)
		}
	}
	return fields, pairedKeyVals
}

func (log *ZapAdapter) fields(keyvals []interface{}) []zap.Field {
	// separate any zap fields from other structs
	zapFields, keyvals := separateFields(keyvals)

	// we should have even number of keys and values
	if len(keyvals)%2 != 0 {
		return []zap.Field{zap.Error(fmt.Errorf("odd number of keyvals pairs: %v", keyvals))}
	}

	fields := make([]zap.Field, 0, len(keyvals)/2+len(zapFields))
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", keyvals[i])
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	// add all the fields
	fields = append(fields, zapFields...)

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
