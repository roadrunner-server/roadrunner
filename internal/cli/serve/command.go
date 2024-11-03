package serve

import (
	"fmt"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/roadrunner/v2024/container"
	"github.com/roadrunner-server/roadrunner/v2024/internal/meta"
	"github.com/roadrunner-server/roadrunner/v2024/internal/sdnotify"
	"os"
	"os/signal"
	"syscall"

	configImpl "github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

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

			// Process config and get temporary file path
			tempFile, err := container.ProcessConfig(*cfgFile)
			if err != nil {
				return errors.E(op, err)
			}
			if len(tempFile) > 0 {
				*cfgFile = tempFile
				defer func() {
					_ = os.Remove(tempFile)
				}()
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
				if !*silent {
					fmt.Println(fmt.Errorf("[WARN] Failed to parse log level, using default (error): %w", err))
				}
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

			oss, stop := make(chan os.Signal, 5), make(chan struct{}, 1) //nolint:gomnd
			signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

			go func() {
				// first catch - stop the container
				<-oss
				// send signal to stop execution
				stop <- struct{}{}

				// notify about stopping
				_, _ = sdnotify.SdNotify(sdnotify.Stopping)

				// after first hit we are waiting for the second catch - exit from the process
				<-oss
				fmt.Println("exit forced")
				os.Exit(1)
			}()

			if !*silent {
				fmt.Printf("[INFO] RoadRunner server started; version: %s, buildtime: %s\n", meta.Version(), meta.BuildTime())
			}

			// at this moment, we're almost sure that the container is running (almost- because we don't know if the plugins won't report an error on the next step)
			notified, err := sdnotify.SdNotify(sdnotify.Ready)
			if err != nil {
				if !*silent {
					fmt.Printf("[WARN] sdnotify: %s\n", err)
				}
			}

			if !*silent {
				if notified {
					fmt.Println("[INFO] sdnotify: notified")
					stopCh := make(chan struct{}, 1)
					if containerCfg.WatchdogSec > 0 {
						fmt.Printf("[INFO] sdnotify: watchdog enabled, timeout: %d seconds\n", containerCfg.WatchdogSec)
						sdnotify.StartWatchdog(containerCfg.WatchdogSec, stopCh)
					}

					// if notified -> notify about stop
					defer func() {
						stopCh <- struct{}{}
					}()
				} else {
					fmt.Println("[INFO] sdnotify: not notified")
				}
			}

			for {
				select {
				case e := <-errCh:
					return fmt.Errorf("error: %w\nplugin: %s", e.Error, e.VertexID)
				case <-stop: // stop the container after first signal
					fmt.Printf("stop signal received, grace timeout is: %0.f seconds\n", containerCfg.GracePeriod.Seconds())

					if err = cont.Stop(); err != nil {
						return fmt.Errorf("error: %w", err)
					}

					return nil
				}
			}
		},
	}
}
