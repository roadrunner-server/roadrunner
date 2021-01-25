package logger

import (
	"hash/fnv"
	"strings"
	"time"

	"github.com/fatih/color"
	"go.uber.org/zap/zapcore"
)

var colorMap = []func(string, ...interface{}) string{
	color.HiYellowString,
	color.HiGreenString,
	color.HiBlueString,
	color.HiRedString,
	color.HiCyanString,
	color.HiMagentaString,
}

// ColoredLevelEncoder colorizes log levels.
func ColoredLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		enc.AppendString(color.HiWhiteString(level.CapitalString()))
	case zapcore.InfoLevel:
		enc.AppendString(color.HiCyanString(level.CapitalString()))
	case zapcore.WarnLevel:
		enc.AppendString(color.HiYellowString(level.CapitalString()))
	case zapcore.ErrorLevel, zapcore.DPanicLevel:
		enc.AppendString(color.HiRedString(level.CapitalString()))
	case zapcore.PanicLevel, zapcore.FatalLevel:
		enc.AppendString(color.HiMagentaString(level.CapitalString()))
	}
}

// ColoredNameEncoder colorizes service names.
func ColoredNameEncoder(s string, enc zapcore.PrimitiveArrayEncoder) {
	if len(s) < 12 {
		s += strings.Repeat(" ", 12-len(s))
	}

	enc.AppendString(color.HiGreenString(s))
}

// ColoredHashedNameEncoder colorizes service names and assigns different colors to different names.
func ColoredHashedNameEncoder(s string, enc zapcore.PrimitiveArrayEncoder) {
	if len(s) < 12 {
		s += strings.Repeat(" ", 12-len(s))
	}

	colorID := stringHash(s, len(colorMap))
	enc.AppendString(colorMap[colorID](s))
}

// UTCTimeEncoder encodes time into short UTC specific timestamp.
func UTCTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006/01/02 15:04:05"))
}

// returns string hash
func stringHash(name string, base int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	return int(h.Sum32()) % base
}
