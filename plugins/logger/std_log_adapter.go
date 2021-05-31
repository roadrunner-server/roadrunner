package logger

import (
	"github.com/spiral/roadrunner/v2/utils"
)

// StdLogAdapter can be passed to the http.Server or any place which required standard logger to redirect output
// to the logger plugin
type StdLogAdapter struct {
	log Logger
}

// Write io.Writer interface implementation
func (s *StdLogAdapter) Write(p []byte) (n int, err error) {
	s.log.Error("server internal error", "message", utils.AsString(p))
	return len(p), nil
}

// NewStdAdapter constructs StdLogAdapter
func NewStdAdapter(log Logger) *StdLogAdapter {
	logAdapter := &StdLogAdapter{
		log: log,
	}

	return logAdapter
}
