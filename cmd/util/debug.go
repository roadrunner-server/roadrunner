package util

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner"
	"strings"
)

// LogEvent outputs rr event into given logger and return false if event was not handled.
func LogEvent(logger *logrus.Logger, event int, ctx interface{}) bool {
	switch event {
	case roadrunner.EventWorkerKill:
		w := ctx.(*roadrunner.Worker)
		logger.Warning(Sprintf(
			"<white+hb>worker.%v</reset> <yellow>killed</red>",
			*w.Pid,
		))
		return true
	case roadrunner.EventWorkerError:
		err := ctx.(roadrunner.WorkerError)
		logger.Error(Sprintf(
			"<white+hb>worker.%v</reset> <red>%s</reset>",
			*err.Worker.Pid,
			err.Caused,
		))
		return true
	}

	// outputs
	switch event {
	case roadrunner.EventStderrOutput:
		logger.Warning(Sprintf(
			"<yellow>%s</reset>",
			strings.Trim(string(ctx.([]byte)), "\r\n"),
		))
		return true
	}

	// rr server events
	switch event {
	case roadrunner.EventServerFailure:
		logger.Error(Sprintf("<red>server is dead</reset>"))
		return true
	}

	// pool events
	switch event {
	case roadrunner.EventPoolConstruct:
		logger.Debug(Sprintf("<cyan>new worker pool</reset>"))
		return true
	case roadrunner.EventPoolError:
		logger.Error(Sprintf("<red>%s</reset>", ctx))
		return true
	}

	return false
}
