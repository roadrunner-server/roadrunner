package http

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spiral/roadrunner"
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/cmd/util"
	rrhttp "github.com/spiral/roadrunner/service/http"
	"net"
	"net/http"
	"strings"
	"time"
)

func init() {
	cobra.OnInitialize(func() {
		if rr.Debug {
			svc, _ := rr.Container.Get(rrhttp.ID)
			if svc, ok := svc.(*rrhttp.Service); ok {
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

	// http events
	switch event {
	case rrhttp.EventResponse:
		e := ctx.(*rrhttp.ResponseEvent)
		s.logger.Info(util.Sprintf(
			"<cyan+h>%s</reset> %s %s <white+hb>%s</reset> %s",
			e.Request.RemoteAddr,
			elapsed(e.Elapsed()),
			statusColor(e.Response.Status),
			e.Request.Method,
			e.Request.URI,
		))

	case rrhttp.EventError:
		e := ctx.(*rrhttp.ErrorEvent)

		if _, ok := e.Error.(roadrunner.JobError); ok {
			s.logger.Info(util.Sprintf(
				"<cyan+h>%s</reset> %s %s <white+hb>%s</reset> %s",
				addr(e.Request.RemoteAddr),
				elapsed(e.Elapsed()),
				statusColor(500),
				e.Request.Method,
				uri(e.Request),
			))
		} else {
			s.logger.Info(util.Sprintf(
				"<cyan+h>%s</reset> %s %s <white+hb>%s</reset> %s <red>%s</reset>",
				addr(e.Request.RemoteAddr),
				elapsed(e.Elapsed()),
				statusColor(500),
				e.Request.Method,
				uri(e.Request),
				e.Error,
			))
		}
	}
}

func statusColor(status int) string {
	if status < 300 {
		return util.Sprintf("<green>%v</reset>", status)
	}

	if status < 400 {
		return util.Sprintf("<cyan>%v</reset>", status)
	}

	if status < 500 {
		return util.Sprintf("<yellow>%v</reset>", status)
	}

	return util.Sprintf("<red>%v</reset>", status)
}

func uri(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s%s", r.Host, r.URL.String())
	}

	return fmt.Sprintf("http://%s%s", r.Host, r.URL.String())
}

// fits duration into 5 characters
func elapsed(d time.Duration) string {
	var v string
	switch {
	case d > 100*time.Second:
		v = fmt.Sprintf("%.1fs", d.Seconds())
	case d > 10*time.Second:
		v = fmt.Sprintf("%.2fs", d.Seconds())
	case d > time.Second:
		v = fmt.Sprintf("%.3fs", d.Seconds())
	case d > 100*time.Millisecond:
		v = fmt.Sprintf("%.0fms", d.Seconds()*1000)
	case d > 10*time.Millisecond:
		v = fmt.Sprintf("%.1fms", d.Seconds()*1000)
	default:
		v = fmt.Sprintf("%.2fms", d.Seconds()*1000)
	}

	if d > time.Second {
		return util.Sprintf("<red>{%v}</reset>", v)
	}

	if d > time.Millisecond*500 {
		return util.Sprintf("<yellow>{%v}</reset>", v)
	}

	return util.Sprintf("<gray+hb>{%v}</reset>", v)
}

func addr(addr string) string {
	// otherwise, return remote address as is
	if !strings.ContainsRune(addr, ':') {
		return addr
	}

	addr, _, _ = net.SplitHostPort(addr)
	return addr
}
