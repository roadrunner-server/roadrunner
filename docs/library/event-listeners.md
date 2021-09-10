# Event Listeners
RoadRunner server exposes the way to handle internal events using custom event listeners, we can demonstrate how to display console message each time HTTP server responds:

```golang
func main() {
	rr.Logger.Formatter = &logrus.TextFormatter{ForceColors: true}

	rr.Container.Register(env.ID, env.NewService(map[string]string{"rr": rr.Version}))

	rr.Container.Register(rpc.ID, &rpc.Service{})
	rr.Container.Register(http.ID, &http.Service{})
	rr.Container.Register(static.ID, &static.Service{})

    // add event listener to http server
    svc, _ := Container.Get(http.ID)
	svc.(*http.Service).AddListener(myListener)

	// you can register additional commands using cmd.CLI
	rr.Execute()
}
```

Where `myListener` is:

```golang
import (
    "github.com/sirupsen/logrus"
    rrhttp "github.com/spiral/roadrunner/service/http"
)

func myListener(event int, ctx interface{}) {
	switch event {
	case rrhttp.EventResponse:
		e := ctx.(*rrhttp.ResponseEvent)
		logrus.Info(
			"%s %v %s %s",
			e.Request.RemoteAddr,
			e.Response.Status,
			e.Request.Method,
			e.Request.URI,
		)
	case rrhttp.EventError:
		e := ctx.(*rrhttp.ErrorEvent)

		if _, ok := e.Error.(roadrunner.JobError); ok {
			logrus.Info(
				"%v %s %s",
				500,
				e.Request.Method,
				e.Request.URI,
			)
		} else {
			logrus.Info(
				"%v %s %s %s",
				500,
				e.Request.Method,
				e.Request.URI,
				e.Error,
			)
		}
	}
}
```

You can find a list of available events [here](https://godoc.org/github.com/spiral/roadrunner).
