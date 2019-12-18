package log

import "github.com/sirupsen/logrus"

type rpcServer struct {
	l logrus.FieldLogger
}

// Entry that represent a log entry to send to the logrus logger
type Entry struct {
	Level   string
	Message string
	Fields  interface{}
}

// Log writes the entry on the registered logger.
// Using "fatal" & "panic" levels exit the process.
func (r *rpcServer) Log(entry Entry, result *bool) error {
	l, err := logrus.ParseLevel(entry.Level)

	if err != nil {
		*result = false

		return err
	}

	fields, ok := entry.Fields.(map[string]interface{})

	if !ok && entry.Fields != nil {
		fields = logrus.Fields{"data": entry.Fields}
	}

	r.l.WithFields(fields).Log(l, entry.Message)

	*result = true

	return nil
}
