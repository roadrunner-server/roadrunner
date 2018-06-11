package debug

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/http"
	"github.com/spiral/roadrunner/cmd/rr/utils"
	"github.com/spiral/roadrunner"
)

// Listener provide debug callback for system events. With colors!
type listener struct{ logger *logrus.Logger }

// NewListener creates new debug listener.
func NewListener(logger *logrus.Logger) *listener {
	return &listener{logger}
}

// Listener listens to http events and generates nice looking output.
func (s *listener) Listener(event int, ctx interface{}) {
	// http events
	switch event {
	case http.EventResponse:
		log := ctx.(*http.Log)
		s.logger.Info(utils.Sprintf("%s <white+hb>%s</reset> %s", statusColor(log.Status), log.Method, log.Uri))
	case http.EventError:
		log := ctx.(*http.Log)

		if _, ok := log.Error.(roadrunner.JobError); ok {
			s.logger.Info(utils.Sprintf("%s <white+hb>%s</reset> %s", statusColor(log.Status), log.Method, log.Uri))
		} else {
			s.logger.Info(utils.Sprintf(
				"%s <white+hb>%s</reset> %s <red>%s</reset>",
				statusColor(log.Status),
				log.Method,
				log.Uri,
				log.Error,
			))
		}
	}

	switch event {
	case roadrunner.EventWorkerKill:
		w := ctx.(*roadrunner.Worker)
		s.logger.Warning(utils.Sprintf(
			"<white+hb>worker.%v</reset> <yellow>killed</red>",
			*w.Pid,
		))

	case roadrunner.EventWorkerError:
		err := ctx.(roadrunner.WorkerError)
		s.logger.Error(utils.Sprintf(
			"<white+hb>worker.%v</reset> <red>%s</reset>",
			*err.Worker.Pid,
			err.Caused,
		))
	}

	// rr server events
	switch event {
	case roadrunner.EventServerFailure:
		s.logger.Error(utils.Sprintf("<red>server is dead</reset>"))
	}

	// pool events
	switch event {
	case roadrunner.EventPoolConstruct:
		s.logger.Debug(utils.Sprintf("<cyan>new worker pool</reset>"))
	case roadrunner.EventPoolError:
		s.logger.Error(utils.Sprintf("<red>%s</reset>", ctx))
	}
}

// Serve serves.
func (s *listener) Serve() error { return nil }

// Stop stops the Listener.
func (s *listener) Stop() {}

func statusColor(status int) string {
	if status < 300 {
		return utils.Sprintf("<green>%v</reset>", status)
	}

	if status < 400 {
		return utils.Sprintf("<cyan>%v</reset>", status)
	}

	if status < 500 {
		return utils.Sprintf("<yellow>%v</reset>", status)
	}

	return utils.Sprintf("<red+hb>%v</reset>", status)
}
