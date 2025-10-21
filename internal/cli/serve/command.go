package serve

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/roadrunner/v2025/container"
	"github.com/roadrunner-server/roadrunner/v2025/internal/meta"
	"github.com/roadrunner-server/roadrunner/v2025/internal/sdnotify"

	configImpl "github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

// log outputs a message to stdout if silent mode is not enabled
func log(msg string, silent bool) {
	if !silent {
		fmt.Println(msg)
	}
}

func NewCommand(override *[]string, cfgFile *string, silent *bool, experimental *bool) *cobra.Command { //nolint:funlen
	return &cobra.Command{
		Use:   "serve",
		Short: "Start RoadRunner server",
		RunE: func(*cobra.Command, []string) error {
			const op = errors.Op("handle_serve_command")
			// just to be safe
			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			// create endure container config
			containerCfg, err := container.NewConfig(*cfgFile)
			if err != nil {
				return errors.E(op, err)
			}

			cfg := &configImpl.Plugin{
				Path:                 *cfgFile,
				Timeout:              containerCfg.GracePeriod,
				Flags:                *override,
				Version:              meta.Version(),
				ExperimentalFeatures: *experimental,
			}

			endureOptions := []endure.Options{
				endure.GracefulShutdownTimeout(containerCfg.GracePeriod),
			}

			if containerCfg.PrintGraph {
				endureOptions = append(endureOptions, endure.Visualize())
			}

			// create endure container
			ll, err := container.ParseLogLevel(containerCfg.LogLevel)
			if err != nil {
				log(fmt.Sprintf("[WARN] Failed to parse log level, using default (error): %v", err), *silent)
			}

			cont := endure.New(ll, endureOptions...)

			// register plugins
			err = cont.RegisterAll(append(container.Plugins(), cfg)...)
			if err != nil {
				return errors.E(op, err)
			}

			// init container and all services
			err = cont.Init()
			if err != nil {
				return errors.E(op, err)
			}

			// start serving the graph
			errCh, err := cont.Serve()
			if err != nil {
				return errors.E(op, err)
			}

			oss, stop := make(chan os.Signal, 1), make(chan struct{}, 1)
			signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT, syscall.SIGQUIT)

			restartCh := make(chan os.Signal, 1)
			signal.Notify(restartCh, syscall.SIGUSR2)

			go func() {
				// first catch - stop the container
				<-oss
				// send signal to stop execution
				stop <- struct{}{}

				// notify about stopping
				_, _ = sdnotify.SdNotify(sdnotify.Stopping)

				// after the first hit we are waiting for the second catch - exit from the process
				<-oss
				log("exit forced", *silent)
				os.Exit(1)
			}()

			log(fmt.Sprintf("[INFO] RoadRunner server started; version: %s, buildtime: %s", meta.Version(), meta.BuildTime()), *silent)

			// at this moment, we're almost sure that the container is running (almost- because we don't know if the plugins won't report an error on the next step)
			notified, err := sdnotify.SdNotify(sdnotify.Ready)
			if err != nil {
				log(fmt.Sprintf("[WARN] sdnotify: %s", err), *silent)
			}

			if notified {
				log("[INFO] sdnotify: notified", *silent)
				stopCh := make(chan struct{}, 1)
				if containerCfg.WatchdogSec > 0 {
					log(fmt.Sprintf("[INFO] sdnotify: watchdog enabled, timeout: %d seconds", containerCfg.WatchdogSec), *silent)
					sdnotify.StartWatchdog(containerCfg.WatchdogSec, stopCh)
				}

				// if notified -> notify about stop
				defer func() {
					stopCh <- struct{}{}
				}()
			}

			for {
				select {
				case e := <-errCh:
					return fmt.Errorf("error: %w\nplugin: %s", e.Error, e.VertexID)
				case <-stop: // stop the container after the first signal
					log(fmt.Sprintf("stop signal received, grace timeout is: %0.f seconds", containerCfg.GracePeriod.Seconds()), *silent)

					if err = cont.Stop(); err != nil {
						return fmt.Errorf("error: %w", err)
					}

					return nil

				case <-restartCh:
					log("restart signal [SIGUSR2] received", *silent)
					executable, err := os.Executable()
					if err != nil {
						log(fmt.Sprintf("restart failed: %s", err), *silent)
						return errors.E("failed to restart")
					}
					args := os.Args
					env := os.Environ()

					if err := cont.Stop(); err != nil {
						log(fmt.Sprintf("restart failed: %s", err), *silent)
						return errors.E("failed to restart")
					}

					err = syscall.Exec(executable, args, env)
					if err != nil {
						log(fmt.Sprintf("restart failed: %s", err), *silent)
						return errors.E("failed to restart")
					}

					return nil
				}
			}
		},
	}
}
