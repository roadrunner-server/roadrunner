package serve

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spiral/roadrunner-binary/v2/internal/container"
	"github.com/spiral/roadrunner-binary/v2/internal/meta"

	"github.com/spf13/cobra"
	"github.com/spiral/errors"
	configImpl "github.com/spiral/roadrunner-plugins/v2/config"
)

// NewCommand creates `serve` command.
func NewCommand(cfgPlugin *configImpl.Plugin) *cobra.Command { //nolint:funlen
	return &cobra.Command{
		Use:   "serve",
		Short: "Start RoadRunner server",
		RunE: func(*cobra.Command, []string) error {
			const op = errors.Op("handle_serve_command")

			// create endure container config
			containerCfg, err := container.NewConfig(cfgPlugin)
			if err != nil {
				return errors.E(op, err)
			}

			// set the grace period which would be same for all the plugins
			cfgPlugin.Timeout = containerCfg.GracePeriod
			cfgPlugin.Version = meta.Version()

			// create endure container
			endureContainer, err := container.NewContainer(*containerCfg)
			if err != nil {
				return errors.E(op, err)
			}

			// register config plugin
			if err = endureContainer.Register(cfgPlugin); err != nil {
				return errors.E(op, err)
			}

			// register another container plugins
			for i, plugins := 0, container.Plugins(); i < len(plugins); i++ {
				if err = endureContainer.Register(plugins[i]); err != nil {
					return errors.E(op, err)
				}
			}

			// init container and all services
			if err = endureContainer.Init(); err != nil {
				return errors.E(op, err)
			}

			// start serving the graph
			errCh, err := endureContainer.Serve()
			if err != nil {
				return errors.E(op, err)
			}

			oss, stop := make(chan os.Signal, 2), make(chan struct{}, 1) //nolint:gomnd
			signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

			go func() {
				// first catch - stop the container
				<-oss
				// send signal to stop execution
				stop <- struct{}{}

				// after first hit we are waiting for the second
				// second catch - exit from the process
				<-oss
				fmt.Println("exit forced")
				os.Exit(1)
			}()

			fmt.Printf("[INFO] RoadRunner server started; version: %s, buildtime: %s\n", meta.Version(), meta.BuildTime())

			for {
				select {
				case e := <-errCh:
					fmt.Printf("error occurred: %v, plugin: %s\n", e.Error, e.VertexID)

					// return error, container already stopped internally
					if !containerCfg.RetryOnFail {
						return errors.E(op, e.Error)
					}

				case <-stop: // stop the container after first signal
					fmt.Printf("stop signal received, grace timeout is: %d seconds\n", uint64(containerCfg.GracePeriod.Seconds()))

					if err = endureContainer.Stop(); err != nil {
						fmt.Printf("error occurred during the stopping container: %v\n", err)
					}

					return nil
				}
			}
		},
	}
}
