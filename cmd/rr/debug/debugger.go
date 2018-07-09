package debug

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/cmd/rr/utils"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"net/http"
	"strings"
)

// Listener creates new debug listener.
func Listener(logger *logrus.Logger) func(event int, ctx interface{}) {
	return (&debugger{logger}).listener
}

// listener provide debug callback for system events. With colors!
type debugger struct{ logger *logrus.Logger }

// listener listens to http events and generates nice looking output.
func (s *debugger) listener(event int, ctx interface{}) {
	// http events
	switch event {
	case rrhttp.EventResponse:
		e := ctx.(*rrhttp.ResponseEvent)
		s.logger.Info(utils.Sprintf(
			"<cyan+h>%s</reset> %s <white+hb>%s</reset> %s",
			e.Request.RemoteAddr,
			statusColor(e.Response.Status),
			e.Request.Method,
			e.Request.URI,
		))
	case rrhttp.EventError:
		e := ctx.(*rrhttp.ErrorEvent)

		if _, ok := e.Error.(roadrunner.JobError); ok {
			s.logger.Info(utils.Sprintf(
				"%s <white+hb>%s</reset> %s",
				statusColor(500),
				e.Request.Method,
				uri(e.Request),
			))
		} else {
			s.logger.Info(utils.Sprintf(
				"%s <white+hb>%s</reset> %s <red>%s</reset>",
				statusColor(500),
				e.Request.Method,
				uri(e.Request),
				e.Error,
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

	// outputs
	switch event {
	case roadrunner.EventStderrOutput:
		s.logger.Warning(utils.Sprintf("<yellow+h>%s</reset>", strings.Trim(string(ctx.([]byte)), "\r\n")))
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

	//s.logger.Warning(event, ctx)
}

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

	return utils.Sprintf("<red>%v</reset>", status)
}

// uri fetches full uri from request in a form of string (including https scheme if TLS connection is enabled).
func uri(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s%s", r.Host, r.URL.String())
	}

	return fmt.Sprintf("http://%s%s", r.Host, r.URL.String())
}
