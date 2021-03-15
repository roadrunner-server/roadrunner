package logger

// Logger is an general RR log interface
type Logger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
}

// With creates a child logger and adds structured context to it
type WithLogger interface {
	With(keyvals ...interface{}) Logger
}
