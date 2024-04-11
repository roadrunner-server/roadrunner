package cli

import (
	"context"
	stderr "errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/roadrunner-server/errors"
	"github.com/roadrunner-server/roadrunner/v2024/internal/cli/jobs"
	"github.com/roadrunner-server/roadrunner/v2024/internal/cli/reset"
	"github.com/roadrunner-server/roadrunner/v2024/internal/cli/serve"
	"github.com/roadrunner-server/roadrunner/v2024/internal/cli/stop"
	"github.com/roadrunner-server/roadrunner/v2024/internal/cli/workers"
	dbg "github.com/roadrunner-server/roadrunner/v2024/internal/debug"
	"github.com/roadrunner-server/roadrunner/v2024/internal/meta"
	"github.com/spf13/cobra"
)

const (
	// env var name: path to the .env file
	envDotenv   string = "DOTENV_PATH"
	pidFileName string = ".pid"
)

// NewCommand creates root command.
func NewCommand(cmdName string) *cobra.Command { //nolint:funlen,gocognit
	// path to the .rr.yaml
	cfgFile := toPtr("")
	// pidfile path
	pidFile := toPtr(false)
	// force stop RR
	forceStop := toPtr(false)
	// override config values
	override := &[]string{}
	// do not print startup message
	silent := toPtr(false)
	// enable experimental features
	experimental := toPtr(false)

	// working directory
	var workDir string
	// path to the .env file
	var dotenv string
	// debug mode
	var debug bool

	cmd := &cobra.Command{
		Use:           cmdName,
		Short:         "High-performance PHP application server, process manager written in Golang and powered with ❤️ (by SpiralScout)",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       fmt.Sprintf("%s (build time: %s, %s), OS: %s, arch: %s", meta.Version(), meta.BuildTime(), runtime.Version(), runtime.GOOS, runtime.GOARCH),
		PersistentPreRunE: func(*cobra.Command, []string) error {
			// cfgFile could be defined by user or default `.rr.yaml`
			// this check added just to be safe
			if cfgFile == nil || *cfgFile == "" {
				return errors.Str("no configuration file provided")
			}

			// if user set the wd, change the current wd
			if workDir != "" {
				if err := os.Chdir(workDir); err != nil {
					return err
				}
			}

			// try to get the absolute path to the configuration
			if absPath, err := filepath.Abs(*cfgFile); err == nil {
				*cfgFile = absPath // switch a config path to the absolute

				// if workDir is empty - force working absPath related to config file
				if workDir == "" {
					if err = os.Chdir(filepath.Dir(absPath)); err != nil {
						return err
					}
				}
			}

			if v, ok := os.LookupEnv(envDotenv); ok { // read a path to the dotenv file from environment variable
				dotenv = v
			}

			if dotenv != "" {
				err := godotenv.Load(dotenv)
				if err != nil {
					return err
				}
			}

			if debug {
				srv := dbg.NewServer()
				exit := make(chan os.Signal, 1)
				stpErr := make(chan error, 1)
				signal.Notify(exit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

				go func() {
					errS := srv.Start(":6061")
					// errS is always non-nil, this is just double check
					if errS != nil && stderr.Is(errS, http.ErrServerClosed) {
						return
					}
					// if we have another type of error - record it
					stpErr <- errS
				}()

				go func() {
					for {
						select {
						case e := <-stpErr:
							// no need to stop the server
							fmt.Println(fmt.Errorf("[ERROR] debug server stopped with error: %w", e))

							return
						case <-exit:
							_ = srv.Stop(context.Background())
						}
					}
				}()
			}

			// user wanted to write a .pid file
			if *pidFile {
				f, err := os.Create(pidFileName)
				if err != nil {
					return err
				}
				defer func() {
					_ = f.Close()
				}()

				_, err = f.WriteString(strconv.Itoa(os.Getpid()))
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	f := cmd.PersistentFlags()

	f.BoolVarP(experimental, "enable-experimental", "e", false, "enable experimental features")
	f.BoolVarP(forceStop, "force", "f", false, "force stop")
	f.BoolVarP(pidFile, "pid", "p", false, "create a .pid file")
	f.StringVarP(cfgFile, "config", "c", ".rr.yaml", "config file")
	f.StringVarP(&workDir, "WorkDir", "w", "", "working directory")
	f.StringVarP(&dotenv, "dotenv", "", "", fmt.Sprintf("dotenv file [$%s]", envDotenv))
	f.BoolVarP(&debug, "debug", "d", false, "debug mode")
	f.BoolVarP(silent, "silent", "s", false, "do not print startup message")
	f.StringArrayVarP(override, "override", "o", nil, "override config value (dot.notation=value)")

	cmd.AddCommand(
		workers.NewCommand(cfgFile, override),
		reset.NewCommand(cfgFile, override, silent),
		serve.NewCommand(override, cfgFile, silent, experimental),
		stop.NewCommand(silent, forceStop),
		jobs.NewCommand(cfgFile, override, silent),
	)

	return cmd
}

func toPtr[T any](val T) *T {
	return &val
}
