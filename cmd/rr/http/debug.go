package http

import (
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"

	"github.com/spf13/cobra"
	"github.com/spiral/roadrunner/cmd/rr/debug"
	"github.com/spiral/roadrunner/service/http"
)

func init() {
	cobra.OnInitialize(func() {
		if rr.Debug {
			svc, _ := rr.Container.Get(http.ID)
			if svc, ok := svc.(*http.Service); ok {
				svc.AddListener(debug.Listener(rr.Logger))
			}
		}
	})
}
