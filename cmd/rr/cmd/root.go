// Copyright (c) 2018 SpiralScout
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spiral/roadrunner/cmd/util"
	"github.com/spiral/roadrunner/service"
	"github.com/spiral/roadrunner/service/limit"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
)

// Services bus for all the commands.
var (
	cfgFile, workDir, logFormat string
	override                    []string
	mergeJson                   string

	// Verbose enables verbosity mode (container specific).
	Verbose bool

	// Debug enables debug mode (service specific).
	Debug bool

	// Logger - shared logger.
	Logger = logrus.New()

	// Container - shared service bus.
	Container = service.NewContainer(Logger)

	// CLI is application endpoint.
	CLI = &cobra.Command{
		Use:           "rr",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       Version, // allows to use `--version` flag
		Short: util.Sprintf(
			"<green>RoadRunner</reset>, PHP Application Server\nVersion: <yellow+hb>%s</reset>, %s",
			Version,
			BuildTime,
		),
	}
)

// Execute adds all child commands to the CLI command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the CLI.
func Execute() {
	if err := CLI.Execute(); err != nil {
		util.ExitWithError(err)
	}
}

func init() {
	CLI.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	CLI.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
	CLI.PersistentFlags().StringVarP(&logFormat, "logFormat", "l", "color", "select log formatter (color, json, plain)")
	CLI.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is .rr.yaml)")
	CLI.PersistentFlags().StringVarP(&workDir, "workDir", "w", "", "work directory")
	CLI.PersistentFlags().StringVarP(&mergeJson, "jsonConfig", "j", "", "merge json configuration")

	CLI.PersistentFlags().StringArrayVarP(
		&override,
		"override",
		"o",
		nil,
		"override config value (dot.notation=value)",
	)

	cobra.OnInitialize(func() {
		if Verbose {
			Logger.SetLevel(logrus.DebugLevel)
		}

		configureLogger(logFormat)

		cfg, err := util.LoadConfig(cfgFile, []string{"."}, ".rr", override, mergeJson)
		if err != nil {
			Logger.Warnf("config: %s", err)
			return
		}

		if workDir != "" {
			if err := os.Chdir(workDir); err != nil {
				util.ExitWithError(err)
			}
		}

		if err := Container.Init(cfg); err != nil {
			util.ExitWithError(err)
		}

		// global watcher config
		if Verbose {
			wcv, _ := Container.Get(limit.ID)
			if wcv, ok := wcv.(*limit.Service); ok {
				wcv.AddListener(func(event int, ctx interface{}) {
					util.LogEvent(Logger, event, ctx)
				})
			}
		}

		// if debug --> also run pprof service
		if Debug {
			go runDebugServer()
		}
	})
}
func runDebugServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	srv := http.Server{
		Addr:    ":6061",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func configureLogger(format string) {
	util.Colorize = false
	switch format {
	case "color", "default":
		util.Colorize = true
		Logger.Formatter = &logrus.TextFormatter{ForceColors: true}
	case "plain":
		Logger.Formatter = &logrus.TextFormatter{DisableColors: true}
	case "json":
		Logger.Formatter = &logrus.JSONFormatter{}
	}
}
