package limit

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spiral/roadrunner"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/cmd/util"
	"github.com/spiral/roadrunner/service/limit"
)

func init() {
	cobra.OnInitialize(func() {
		if rr.Debug {
			svc, _ := rr.Container.Get(limit.ID)
			if svc, ok := svc.(*limit.Service); ok {
				svc.AddListener((&debugger{logger: rr.Logger}).listener)
			}
		}
	})
}

// listener provide debug callback for system events. With colors!
type debugger struct{ logger *logrus.Logger }

// listener listens to http events and generates nice looking output.
func (s *debugger) listener(event int, ctx interface{}) {
	if util.LogEvent(s.logger, event, ctx) {
		// handler by default debug package
		return
	}

	// watchers
	switch event {
	case limit.EventTTL:
		w := ctx.(roadrunner.WorkerError)
		s.logger.Debug(util.Sprintf(
			"<white+hb>worker.%v</reset> <yellow>%s</reset>",
			*w.Worker.Pid,
			w.Caused,
		))
		return

	case limit.EventIdleTTL:
		w := ctx.(roadrunner.WorkerError)
		s.logger.Debug(util.Sprintf(
			"<white+hb>worker.%v</reset> <yellow>%s</reset>",
			*w.Worker.Pid,
			w.Caused,
		))
		return

	case limit.EventMaxMemory:
		w := ctx.(roadrunner.WorkerError)
		s.logger.Error(util.Sprintf(
			"<white+hb>worker.%v</reset> <red>%s</reset>",
			*w.Worker.Pid,
			w.Caused,
		))
		return

	case limit.EventExecTTL:
		w := ctx.(roadrunner.WorkerError)
		s.logger.Error(util.Sprintf(
			"<white+hb>worker.%v</reset> <red>%s</reset>",
			*w.Worker.Pid,
			w.Caused,
		))
		return
	}
}
