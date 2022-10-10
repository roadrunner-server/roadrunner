package serve

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/roadrunner-server/roadrunner/v2/container"
	"github.com/roadrunner-server/roadrunner/v2/internal/meta"

	configImpl "github.com/roadrunner-server/config/v3"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

const (
	rrPrefix string = "rr"
)

// NewCommand creates `serve` command.
func NewCommand(override *[]string, cfgFile *string, silent *bool) *cobra.Command { //nolint:funlen
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
				Path:    *cfgFile,
				Prefix:  rrPrefix,
				Timeout: containerCfg.GracePeriod,
				Flags:   *override,
				Version: meta.Version(),
			}

			// create endure container
			endureContainer, err := container.NewContainer(*containerCfg)
			if err != nil {
				return errors.E(op, err)
			}

			// register plugins
			err = endureContainer.RegisterAll(append(container.Plugins(), cfg)...)
			if err != nil {
				return errors.E(op, err)
			}

			// init container and all services
			err = endureContainer.Init()
			if err != nil {
				return errors.E(op, err)
			}

			// start serving the graph
			errCh, err := endureContainer.Serve()
			if err != nil {
				return errors.E(op, err)
			}

			oss, stop := make(chan os.Signal, 5), make(chan struct{}, 1) //nolint:gomnd
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

			if !*silent {
				fmt.Printf("[INFO] RoadRunner server started; version: %s, buildtime: %s\n", meta.Version(), meta.BuildTime())
			}

			for {
				select {
				case e := <-errCh:
					return fmt.Errorf("error: %w\nplugin: %s", e.Error, e.VertexID)
				case <-stop: // stop the container after first signal
					fmt.Printf("stop signal received, grace timeout is: %0.f seconds\n", containerCfg.GracePeriod.Seconds())

					if err = endureContainer.Stop(); err != nil {
						return fmt.Errorf("error: %w", err)
					}

					return nil
				}
			}
		},
	}
}
