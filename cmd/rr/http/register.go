package http

import (
	rr "github.com/spiral/roadrunner/cmd/rr/cmd"
	"github.com/spiral/roadrunner/http"
	"github.com/spf13/cobra"
)

func init() {
	rr.Services.Register(&http.Service{})

	rr.CLI.AddCommand(&cobra.Command{
		Use:   "http:reload",
		Short: "Reload RoadRunner worker pools for the HTTP service",
		Run:   reloadHandler,
	})

	rr.CLI.AddCommand(&cobra.Command{
		Use:   "http:workers",
		Short: "List workers associated with RoadRunner HTTP service",
		Run:   workersHandler,
	})
}
