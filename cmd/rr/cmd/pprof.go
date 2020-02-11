package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"net/http/pprof"
)

func init() {
	CLI.AddCommand(&cobra.Command{
		Use:     "pprof",
		Short:   "Profile RoadRunner service(s)",
		Example: "rr serve -d -v pprof http://localhost:6061",
		Run:     runDebugServer,
	})
}

func runDebugServer(cmd *cobra.Command, args []string) {
	var address string = "http://localhost:6061"
	// guess that user set the address
	if len(args) > 0 {
		address = args[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	srv := http.Server{
		Addr:    address,
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
