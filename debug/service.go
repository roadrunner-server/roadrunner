package debug

import (
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/http"
	"github.com/spiral/roadrunner/utils"
	"github.com/spiral/roadrunner"
)

// Default service name.
const Name = "debug"

// Service provide debug callback for system events. With colors!
type Service struct {
	// Logger used to flush all debug events.
	Logger *logrus.Logger

	cfg *Config
}

// Configure must return configure service and return true if service hasStatus enabled. Must return error in case of
// misconfiguration. Services must not be used without proper configuration pushed first.
func (s *Service) Configure(cfg service.Config, c service.Container) (enabled bool, err error) {
	config := &Config{}
	if err := cfg.Unmarshal(config); err != nil {
		return false, err
	}

	if !config.Enable {
		return false, nil
	}

	s.cfg = config

	// registering as middleware
	if h, ok := c.Get(http.Name); ok >= service.StatusConfigured {
		if h, ok := h.(*http.Service); ok {
			h.AddListener(s.listener)
		}
	} else {
		s.Logger.Error("unable to find http server")
	}

	return true, nil
}

// listener listens to http events and generates nice looking output.
func (s *Service) listener(event int, ctx interface{}) {
	// http events
	switch event {
	case http.EventResponse:
		log := ctx.(*http.Log)
		s.Logger.Info(utils.Sprintf("%s <white+hb>%s</reset> %s", statusColor(log.Status), log.Method, log.Uri))
	case http.EventError:
		log := ctx.(*http.Log)

		if _, ok := log.Error.(roadrunner.JobError); ok {
			s.Logger.Info(utils.Sprintf("%s <white+hb>%s</reset> %s", statusColor(log.Status), log.Method, log.Uri))
		} else {
			s.Logger.Info(utils.Sprintf(
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
		s.Logger.Warning(utils.Sprintf(
			"<white+hb>worker.%v</reset> <yellow>killed</red>",
			*w.Pid,
		))

	case roadrunner.EventWorkerError:
		err := ctx.(roadrunner.WorkerError)
		s.Logger.Error(utils.Sprintf(
			"<white+hb>worker.%v</reset> <red>%s</reset>",
			*err.Worker.Pid,
			err.Caused,
		))
	}

	// rr server events
	switch event {
	case roadrunner.EventServerFailure:
		s.Logger.Error(utils.Sprintf("<red>server is dead</reset>"))
	}

	// pool events
	switch event {
	case roadrunner.EventPoolConstruct:
		s.Logger.Debug(utils.Sprintf("<cyan>new worker pool</reset>"))
	case roadrunner.EventPoolError:
		s.Logger.Error(utils.Sprintf("<red>%s</reset>", ctx))
	}
}

// Serve serves.
func (s *Service) Serve() error { return nil }

// Stop stops the service.
func (s *Service) Stop() {}

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
